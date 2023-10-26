package ingest

import "go.temporal.io/sdk/worker"

type (
	workflows struct {
	}

	activities struct {
	}
)

func Register(w worker.Worker, wf *workflows, a *activities) {
}

func NewWorkflows() *workflows {
	return &workflows{}
}

func NewActivities() *activities {
	return &activities{}
}
