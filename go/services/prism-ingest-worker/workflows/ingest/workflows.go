package ingest

import (
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"code.prism.io/go/services/prism-ingest-worker/config"
	metav1 "code.prism.io/proto/rpc/gen/go/prism/meta/v1"
)

type (
	workflows struct {
	}

	activities struct {
		ingestConfig       *config.Ingest
		metaClientProvider MetaClientProvider
	}

	MetaClientProvider func() (metav1.MetaServiceClient, error)
)

func Register(w worker.Worker, wf *workflows, a *activities) {
	w.RegisterWorkflowWithOptions(wf.Ingest, workflow.RegisterOptions{Name: "ingest"})
	w.RegisterActivity(a)
}

func NewWorkflows() *workflows {
	return &workflows{}
}

func NewActivities(ingestConfig *config.Ingest, metaClientProvider MetaClientProvider) *activities {
	return &activities{
		ingestConfig:       ingestConfig,
		metaClientProvider: metaClientProvider,
	}
}
