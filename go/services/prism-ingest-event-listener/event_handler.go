package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"code.prism.io/go/services/prism-ingest-event-listener/config"
	ingestv1 "code.prism.io/proto/workflow/gen/prism/ingest/v1"
)

type (
	EventHandler struct {
		logger *zap.Logger
		config config.Provider
	}
)

var (
	keyRegex = regexp.MustCompile(`tenant_id=(?P<tenant>[^/]*)/table=(?P<table>[^/]*)/(?P<file>.*)`)
)

func (e *EventHandler) HandleMessage(ctx context.Context, message *sqs.Message) error {
	if message.MessageId == nil {
		e.logger.Warn("skipping message due to missing message id")
		return nil
	}

	if message.Body == nil {
		e.logger.Warn("skipping message due to missing body")
		return nil
	}

	messageID := *message.MessageId
	body := *message.Body
	var event events.S3Event
	if err := json.Unmarshal([]byte(body), &event); err != nil {
		return fmt.Errorf("failed to unmarshal S3 event: %w", err)
	}

	for _, record := range event.Records {
		if record.EventName != "ObjectCreated:Put" {
			return nil
		}

		key := record.S3.Object.URLDecodedKey
		matches := keyRegex.FindStringSubmatch(key)
		if matches == nil {
			return fmt.Errorf("failed to match key: %s", key)
		}

		tenantID, table := matches[1], matches[2]
		if err := e.dispatchWorkflow(ctx, messageID, tenantID, table, record.S3.Object.URLDecodedKey, record.S3.Bucket.Name); err != nil {
			return fmt.Errorf("failed to dispatch ingest workflow: %w", err)
		}
	}

	return nil
}

func (e *EventHandler) dispatchWorkflow(ctx context.Context, messageID, tenantID, table, key, sourceBucket string) error {
	temporalConfig := e.config.GetTemporal()
	ingestConfig := e.config.GetIngestEventListener()
	temporalClient, err := client.Dial(client.Options{
		HostPort: temporalConfig.Endpoint,
	})
	if err != nil {
		return fmt.Errorf("failed to dial temporal: %w", err)
	}

	ingestClient := ingestv1.NewIngestClient(temporalClient)
	run, err := ingestClient.IngestObjectAsync(ctx, &ingestv1.IngestObjectRequest{
		TenantId:         tenantID,
		Table:            table,
		Source:           sourceBucket,
		Destination:      ingestConfig.DestinationBucket,
		Location:         key,
		IdempotencyToken: messageID,
	})
	if err != nil {
		return fmt.Errorf("failed to dispatch ingest workflow: %w", err)
	}

	e.logger.Info("dispatched ingest workflow", zap.String("workflow_id", run.ID()), zap.String("run_id", run.RunID()))
	return nil
}

func NewEventHandler(logger *zap.Logger, config config.Provider) *EventHandler {
	return &EventHandler{
		logger: logger,
		config: config,
	}
}
