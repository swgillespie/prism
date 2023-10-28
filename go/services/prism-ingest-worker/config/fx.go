package config

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(func(p Provider) *Meta {
		return p.GetMeta()
	}),
	fx.Provide(func(p Provider) *Temporal {
		return p.GetTemporal()
	}),
)
