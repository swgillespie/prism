package ingest

import (
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

type (
	IngestInput struct {
		Event events.S3EventRecord
	}

	IngestTransformToParquetInput struct {
		TenantID          string
		Table             string
		DestinationBucket string
		Event             events.S3EventRecord
	}
)

const (
	heartbeatInterval = 5 * time.Second
)

func (w *workflows) Ingest(ctx workflow.Context, input IngestInput) error {
	err := workflow.ExecuteActivity(ctx, (*activities).IngestTransformToParquet, IngestTransformToParquetInput{
		TenantID:          "test",
		Table:             "web_requests",
		DestinationBucket: "test",
		Event:             input.Event,
	}).Get(ctx, nil)
	return err
}

func (a *activities) IngestTransformToParquet(ctx context.Context, input IngestTransformToParquetInput) error {
	cmd := exec.CommandContext(ctx, "prism-ingest",
		"--source-bucket", input.Event.S3.Bucket.Name,
		"--location", input.Event.S3.Object.Key,
		"--destination-bucket", input.DestinationBucket,
		"--region", input.Event.AWSRegion,
		"--tenant-id", input.TenantID,
		"--table", input.Table,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}

	doneChan := make(chan error)
	go func() {
		doneChan <- cmd.Wait()
		close(doneChan)
	}()

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			if err := cmd.Process.Kill(); err != nil {
				return err
			}
		case err := <-doneChan:
			return err
		case <-ticker.C:
			activity.RecordHeartbeat(ctx, nil)
		}
	}
}
