package next

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"

	"code.prism.io/go/services/prism-ingest-worker/config"
	metav1 "code.prism.io/proto/rpc/gen/go/prism/meta/v1"
	ingestv1 "code.prism.io/proto/workflow/gen/prism/ingest/v1"
)

type (
	Workflows struct{}

	Activities struct {
		ingestConfig       *config.Ingest
		metaClientProvider MetaClientProvider
	}

	IngestObjectWorkflow struct {
		input *ingestv1.IngestObjectInput
	}

	MetaClientProvider func() (metav1.MetaServiceClient, error)
)

var (
	_ ingestv1.IngestObjectWorkflow = (*IngestObjectWorkflow)(nil)
	_ ingestv1.IngestActivities     = (*Activities)(nil)
)

const (
	heartbeatInterval = 5 * time.Second
)

func (w *Workflows) CreateIngestObjectWorkflow(input *ingestv1.IngestObjectInput) *IngestObjectWorkflow {
	return &IngestObjectWorkflow{input: input}
}

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

	var columns []*metav1.Column
	for _, column := range input.Partition.Columns {
		columns = append(columns, &metav1.Column{
			Name: column.Name,
			Type: metav1.ColumnType(column.Type),
		})
	}

	_, err = client.RecordNewPartition(ctx, &metav1.RecordNewPartitionRequest{
		TenantId:  input.TenantId,
		TableName: input.Table,
		Partition: &metav1.Partition{
			Name: input.Partition.Name,
			Size: int64(input.Partition.Size),
			TimeRange: &metav1.TimeRange{
				StartTime: input.Partition.MinTimestamp,
				EndTime:   input.Partition.MaxTimestamp,
			},
		},
		Columns: columns,
	})
	return err

}

func (a *Activities) TransformToParquet(ctx context.Context, input *ingestv1.TransformToParquetRequest) (*ingestv1.TransformToParquetResponse, error) {
	cmd := exec.CommandContext(ctx, a.ingestConfig.IngestBinaryPath,
		"--source", input.Source,
		"--location", input.Location,
		"--destination", input.Destination,
		"--tenant-id", input.TenantId,
		"--table", input.Table,
	)

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
