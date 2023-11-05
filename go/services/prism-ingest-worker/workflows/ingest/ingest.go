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
	ingestv1 "code.prism.io/proto/workflow/gen/prism/ingest/v1"
)

type (
	IngestObjectWorkflow struct {
		input *ingestv1.IngestObjectInput
	}
)

var (
	_ ingestv1.IngestObjectWorkflow = (*IngestObjectWorkflow)(nil)
	_ ingestv1.IngestActivities     = (*Activities)(nil)
)

const (
	heartbeatInterval = 5 * time.Second
)

func (wf *IngestObjectWorkflow) Execute(ctx workflow.Context) error {
	transformResp, err := ingestv1.TransformToParquet(ctx, &ingestv1.TransformToParquetRequest{
		TenantId:    wf.input.Req.TenantId,
		Table:       wf.input.Req.Table,
		Source:      wf.input.Req.Source,
		Destination: wf.input.Req.Destination,
		Location:    wf.input.Req.Location,
	})
	if err != nil {
		return err
	}

	return ingestv1.RecordNewPartition(ctx, &ingestv1.RecordNewPartitionRequest{
		TenantId:  wf.input.Req.TenantId,
		Table:     wf.input.Req.Table,
		Partition: transformResp.Partition,
	})
}

func (a *Activities) RecordNewPartition(ctx context.Context, input *ingestv1.RecordNewPartitionRequest) error {
	client, err := a.metaClientProvider()
	if err != nil {
		return err
	}

	var columns []*commonv1.Column
	for _, column := range input.Partition.Columns {
		columns = append(columns, &commonv1.Column{
			Name: column.Name,
			Type: commonv1.ColumnType(column.Type),
		})
	}

	_, err = client.RecordNewPartition(ctx, &metav1.RecordNewPartitionRequest{
		TenantId:  input.TenantId,
		TableName: input.Table,
		Partition: &commonv1.Partition{
			Name: input.Partition.Name,
			Size: int64(input.Partition.Size),
			TimeRange: &commonv1.TimeRange{
				StartTime: input.Partition.MinTimestamp,
				EndTime:   input.Partition.MaxTimestamp,
			},
		},
		Columns: columns,
	})
	return err

}

func (a *Activities) TransformToParquet(ctx context.Context, input *ingestv1.TransformToParquetRequest) (*ingestv1.TransformToParquetResponse, error) {
	args := []string{
		"--source", input.Source,
		"--location", input.Location,
		"--destination", input.Destination,
		"--tenant-id", input.TenantId,
		"--table", input.Table,
	}
	if a.ingestConfig.S3Endpoint != "" {
		args = append(args, "--s3-endpoint", a.ingestConfig.S3Endpoint)
	}

	logger := activity.GetLogger(ctx)
	logger.Info("starting ingest binary", "args", args)
	cmd := exec.CommandContext(ctx, a.ingestConfig.IngestBinaryPath, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var result ingestv1.Partition
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
				return nil, err
			}
		case err := <-doneChan:
			return &ingestv1.TransformToParquetResponse{
				Partition: &result,
			}, err
		case <-ticker.C:
			activity.RecordHeartbeat(ctx, nil)
		}
	}
}
