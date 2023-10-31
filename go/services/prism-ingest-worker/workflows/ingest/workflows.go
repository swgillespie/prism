package ingest

import (
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

	MetaClientProvider func() (metav1.MetaServiceClient, error)
)

var (
	_ ingestv1.IngestWorkflows = (*Workflows)(nil)
)

func (w *Workflows) IngestObject(ctx workflow.Context, input *ingestv1.IngestObjectInput) (ingestv1.IngestObjectWorkflow, error) {
	return &IngestObjectWorkflow{input: input}, nil
}

func NewWorkflows() *Workflows {
	return &Workflows{}
}

func NewActivities(
	ingestConfig *config.Ingest,
	metaClientProvider MetaClientProvider,
) *Activities {
	return &Activities{
		ingestConfig:       ingestConfig,
		metaClientProvider: metaClientProvider,
	}
}
