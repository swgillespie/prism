ingest-build: cargo build --bin prism-ingest
ingest-worker: go run ./go/services/prism-ingest-worker --config ./misc/ingest-worker.yaml
meta: go run ./go/services/prism-meta server
