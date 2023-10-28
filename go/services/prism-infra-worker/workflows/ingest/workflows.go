package ingest

import (
	"go.temporal.io/sdk/worker"

	"code.prism.io/go/proto"
)

type (
	workflows struct {
	}

	activities struct {
		metaClientProvider MetaClientProvider
	}

	MetaClientProvider func() (proto.MetaServiceClient, error)
)

func Register(w worker.Worker, wf *workflows, a *activities) {
}

func NewWorkflows() *workflows {
	return &workflows{}
}

func NewActivities(metaClientProvider MetaClientProvider) *activities {
	return &activities{
		metaClientProvider: metaClientProvider,
	}
}
