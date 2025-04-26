.DEFAULT_GOAL := help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

deps: ## Download dependencies
	@go mod vendor && go mod tidy

$(GO_BIN)/golangci-lint:
	@go install -v github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

lint: $(GO_BIN)/golangci-lint ## Lint Go source files
	@golangci-lint run -v --fix -c .golangci.yaml ./...

test:  ## Run all tests.
	@go test -count=1 -v -cover -timeout=10s ./...

build: ## Build.
	@go build -v ./...