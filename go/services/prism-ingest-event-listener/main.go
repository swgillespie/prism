package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.uber.org/zap"

	"code.prism.io/go/services/prism-ingest-event-listener/config"
)

func newConfigProvider() (config.Provider, error) {
	path := os.Getenv("PRISM_INGEST_EVENT_LISTENER_CONFIG")
	if path == "" {
		return nil, errors.New("PRISM_INGEST_EVENT_LISTENER_CONFIG environment variable is required")
	}

	return config.NewYAMLProvider(path)
}

func newSQSClient(listenerConfig *config.IngestEventListener) (*sqs.SQS, error) {
	session, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, err
	}

	opts := aws.NewConfig()
	if listenerConfig.SQSEndpoint != "" {
		opts = opts.WithEndpoint(listenerConfig.SQSEndpoint).WithDisableSSL(true)
	}

	return sqs.New(session, opts), nil
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalln("failed to initialize logger: ", err)
	}

	config, err := newConfigProvider()
	if err != nil {
		logger.Fatal("failed to get config", zap.Error(err))
	}

	sqsClient, err := newSQSClient(config.GetIngestEventListener())
	if err != nil {
		logger.Fatal("failed to get SQS client: ", zap.Error(err))
	}

	ingestConfig := config.GetIngestEventListener()
	handler := NewEventHandler(logger, config)
	for {
		result, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl: &ingestConfig.SQSQueueEndpoint,
			AttributeNames: aws.StringSlice([]string{
				"SentTimestamp",
			}),
			MaxNumberOfMessages: aws.Int64(int64(ingestConfig.MaxMessages)),
			MessageAttributeNames: aws.StringSlice([]string{
				"All",
			}),
			WaitTimeSeconds: aws.Int64(int64(ingestConfig.PollTimeoutSeconds)),
		})
		if err != nil {
			logger.Error("failed to receive message from SQS", zap.Error(err))
		}

		for _, msg := range result.Messages {
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(ingestConfig.MessageHandleTimeoutSeconds)*time.Second)
				defer cancel()
				err := handler.HandleMessage(ctx, msg)
				if err != nil {
					logger.Error("failed to handle message", zap.Error(err))
				}

				_, err = sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
					QueueUrl:      aws.String(ingestConfig.SQSQueueEndpoint),
					ReceiptHandle: msg.ReceiptHandle,
				})
				if err != nil {
					logger.Error("failed to delete message", zap.Error(err))
				}
			}()
		}
	}
}
