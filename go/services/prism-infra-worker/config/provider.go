package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type (
	Provider interface {
		GetMeta() *Meta
	}

	yamlProvider struct {
		Meta *Meta `yaml:"meta"`
	}
)

func (y *yamlProvider) GetMeta() *Meta {
	return y.Meta
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
