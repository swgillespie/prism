package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type (
	Provider interface {
		GetMeta() *Meta
		GetTemporal() *Temporal
		GetIngest() *Ingest
	}

	yamlProvider struct {
		Meta     *Meta     `yaml:"meta"`
		Temporal *Temporal `yaml:"temporal"`
		Ingest   *Ingest   `yaml:"ingest"`
	}
)

func (y *yamlProvider) GetMeta() *Meta {
	return y.Meta
}

func (y *yamlProvider) GetTemporal() *Temporal {
	return y.Temporal
}

func (y *yamlProvider) GetIngest() *Ingest {
	return y.Ingest
}

func NewYAMLProvider(filepath string) (Provider, error) {
	contents, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var p yamlProvider
	if err := yaml.Unmarshal(contents, &p); err != nil {
		return nil, err
	}

	return &p, nil
}
