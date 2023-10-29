package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"

	commonv1 "code.prism.io/proto/common/gen/go/prism/common/v1"
	metav1 "code.prism.io/proto/rpc/gen/go/prism/meta/v1"
)

type (
	IngestInput struct {
		TenantID    string
		Table       string
		Source      string
		Destination string
		Location    string
	}

	IngestTransformToParquetInput struct {
		TenantID    string
		Table       string
		Source      string
		Destination string
		Location    string
	}

	IngestRecordNewPartitionInput struct {
		TenantID  string
		Table     string
		Partition Partition
	}

	// Keep synchronized with prism-ingest/src/ingest.rs!
	Partition struct {
		Name    string `json:"name"`
		Size    uint64 `json:"size"`
		MaxTS   int64  `json:"max_ts"`
		MinTS   int64  `json:"min_ts"`
		Columns []Column
	}

	Column struct {
		Name     string `json:"name"`
		DataType string `json:"data_type"`
	}
)

const (
	heartbeatInterval = 5 * time.Second
)

func (w *workflows) Ingest(ctx workflow.Context, input IngestInput) error {
	ctx = workflow.WithActivityOptions(ctx, getActivityOptions())
	var partition Partition
	err := workflow.ExecuteActivity(ctx, (*activities).IngestTransformToParquet, IngestTransformToParquetInput{
		TenantID:    input.TenantID,
		Table:       input.Table,
		Source:      input.Source,
		Destination: input.Destination,
		Location:    input.Location,
	}).Get(ctx, &partition)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, (*activities).IngestRecordNewPartition, IngestRecordNewPartitionInput{
		TenantID:  "test",
		Table:     "web_requests",
		Partition: partition,
	}).Get(ctx, nil)
	return err
}

func (a *activities) IngestTransformToParquet(ctx context.Context, input IngestTransformToParquetInput) (Partition, error) {
	cmd := exec.CommandContext(ctx, a.ingestConfig.IngestBinaryPath,
		"--source", input.Source,
		"--location", input.Location,
		"--destination", input.Destination,
		"--tenant-id", input.TenantID,
		"--table", input.Table,
	)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return Partition{}, err
	}

	var result Partition
	doneChan := make(chan error)
	go func() {
		err := cmd.Wait()
		if err != nil {
			doneChan <- err
			return
		}

		err = json.Unmarshal(buf.Bytes(), &result)
		doneChan <- err
		close(doneChan)
	}()

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			if err := cmd.Process.Kill(); err != nil {
				return Partition{}, err
			}
		case err := <-doneChan:
			return result, err
		case <-ticker.C:
			activity.RecordHeartbeat(ctx, nil)
		}
	}
}

func (a *activities) IngestRecordNewPartition(ctx context.Context, input IngestRecordNewPartitionInput) error {
	client, err := a.metaClientProvider()
	if err != nil {
		return err
	}

	var columns []*commonv1.Column
	for _, column := range input.Partition.Columns {
		columns = append(columns, &commonv1.Column{
			Name: column.Name,
			Type: ingestorTypeToProto(column.DataType),
		})
	}

	_, err = client.RecordNewPartition(ctx, &metav1.RecordNewPartitionRequest{
		TenantId:  input.TenantID,
		TableName: input.Table,
		Partition: &commonv1.Partition{
			Name: input.Partition.Name,
			Size: int64(input.Partition.Size),
			TimeRange: &commonv1.TimeRange{
				StartTime: input.Partition.MinTS,
				EndTime:   input.Partition.MaxTS,
			},
		},
		Columns: columns,
	})
	return err
}

func ingestorTypeToProto(ingestorType string) commonv1.ColumnType {
	switch ingestorType {
	case "Int64":
		return commonv1.ColumnType_COLUMN_TYPE_INT64
	case "String":
		return commonv1.ColumnType_COLUMN_TYPE_UTF8
	case "Timestamp":
		return commonv1.ColumnType_COLUMN_TYPE_TIMESTAMP
	default:
		return commonv1.ColumnType_COLUMN_TYPE_UNSPECIFIED
	}
}

func getActivityOptions() workflow.ActivityOptions {
	return workflow.ActivityOptions{
		ScheduleToStartTimeout: 1 * time.Minute,
		StartToCloseTimeout:    1 * time.Minute,
		HeartbeatTimeout:       30 * time.Second,
	}
}
