temporal: temporal server start-dev
ingest-build: cargo build --bin prism-ingest
ingest-worker: go run ./go/services/prism-ingest-worker worker
meta: go run ./go/services/prism-meta server
localstack: localstack start
