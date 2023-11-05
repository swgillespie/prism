package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type (
	Provider interface {
		GetIngestEventListener() *IngestEventListener
		GetTemporal() *Temporal
	}

	yamlProvider struct {
		IngestEventListener *IngestEventListener `yaml:"ingest_event_listener"`
		Temporal            *Temporal            `yaml:"temporal"`
	}
)

func (y *yamlProvider) GetIngestEventListener() *IngestEventListener {
	return y.IngestEventListener
}

func (y *yamlProvider) GetTemporal() *Temporal {
	return y.Temporal
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
