GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

OUT := ./out
GOPATH := $(shell go env GOPATH)

## Generate

generate: generate-proto generate-go ## Generates all code.

generate-proto: ## Generates all protobuf code.
	cd ./proto/common && buf generate
	cd ./proto/rpc && buf generate
	cd ./proto/workflow && buf generate

generate-go: ## Runs `go generate` over all Go source code.
	go generate ./go/...

## Build

build: generate build-go build-rust ## Builds all targets.

$(OUT):
	mkdir -p $(OUT)

build-go: build-meta build-ingest-worker build-ingest-event-listener build-api ## Builds all Go targets.

build-rust: build-ingest build-query ## Builds all Rust targets.

build-meta: $(OUT) ## Builds the Meta service.
	go build -o $(OUT)/prism-meta ./go/services/prism-meta

build-ingest-worker: $(OUT) ## Builds the ingest worker.
	go build -o $(OUT)/prism-ingest-worker ./go/services/prism-ingest-worker

build-ingest-event-listener: $(OUT) ## Builds the ingest event listener.
	go build -o $(OUT)/prism-ingest-event-listener ./go/services/prism-ingest-event-listener

build-api: $(OUT) ## Builds the API server.
	go build -o $(OUT)/prism-api ./go/services/prism-api

build-ingest:
	cargo build --bin prism-ingest
	cp ./target/debug/prism-ingest $(OUT)/prism-ingest

build-query:
	cargo build --bin prism-query
	cp ./target/debug/prism-query $(OUT)/prism-query

clean:
	rm -rf $(OUT)

## Lint
lint: lint-go lint-rust ## Run all linters

lint-go: ## Lint all Go code
	golangci-lint run --timeout 120s ./go/...

lint-rust: ## Lint all Rust code
	cargo clippy --all-targets --all-features -- -D warnings

## Test

test: generate test-go test-rust ## Run all tests

test-go: ## Run all Go tests
	go test -cover ./go/...

test-rust: ## Run all Rust tests
	cargo test

## Dev

install-dependencies: ## Installs all compile-time dependencies.
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	go install -tags 'cockroachdb' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.16.2
	go install github.com/golang/mock/mockgen@v1.6.0
	go install github.com/bufbuild/buf/cmd/buf@v1.27.2
	go install github.com/cludden/protoc-gen-go-temporal/cmd/protoc-gen-go_temporal@v1.0.2
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.55.2


## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-30s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)