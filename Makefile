GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

## Dev
install-dependencies: ## Installs all compile-time dependencies.
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	go install -tags 'cockroachdb' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.16.2
	go install github.com/golang/mock/mockgen@v1.6.0
	go install github.com/bufbuild/buf/cmd/buf@v1.27.2
	go install github.com/cludden/protoc-gen-go-temporal/cmd/protoc-gen-go_temporal@v1.0.2

run: ## Runs all services locally.
	overmind start

## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)