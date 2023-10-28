package config

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(func(p Provider) *Meta {
		return p.GetMeta()
	}),
)
