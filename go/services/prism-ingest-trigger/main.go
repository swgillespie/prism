package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"code.prism.io/go/services/prism-infra-worker/workflows/ingest"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	lambda.Start(func(ctx context.Context, event events.S3Event) {
		c, err := client.Dial(client.Options{})
		if err != nil {
			logger.Error("failed to dial temporal", zap.Error(err))
			return
		}

		for _, record := range event.Records {
			opts := client.StartWorkflowOptions{
				ID:                    fmt.Sprintf("ingest-%s-%s", record.S3.Bucket.Name, record.S3.Object.Key),
				TaskQueue:             "infra-worker",
				WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
			}
			_, err := c.ExecuteWorkflow(ctx, opts, "Ingest", ingest.IngestInput{
				Event: record,
			})
			if err != nil {
				logger.Error("failed to start workflow", zap.Error(err))
			}
		}
	})
}
