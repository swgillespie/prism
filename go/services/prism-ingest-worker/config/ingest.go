package config

type Ingest struct {
	IngestBinaryPath string `yaml:"ingest_binary_path"`
	S3Endpoint       string `yaml:"s3_endpoint"`
}
