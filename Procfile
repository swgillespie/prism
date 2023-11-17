ingest-worker: go run ./go/services/prism-ingest-worker worker
ingest-event-listener: go run ./go/services/prism-ingest-event-listener
meta: go run ./go/services/prism-meta server
vector: vector --config ./misc/vector.json

# select count(*) from prism.demo.http_requests;