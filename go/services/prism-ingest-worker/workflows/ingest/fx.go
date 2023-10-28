package ingest

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		NewWorkflows,
		NewActivities,
	),
	fx.Invoke(Register),
)
