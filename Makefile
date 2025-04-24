.DEFAULT_GOAL := help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

deps: ## Download dependencies
	${call print, "Downloading dependencies"}
	@go mod vendor && go mod tidy

$(GO_BIN)/golangci-lint:
	${call print, "Installing golangci-lint within ${GO_BIN}"}
	@go install -v github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

lint: $(GO_BIN)/golangci-lint ## Lint Go source files
	${call print, "Linting Go source files"}
	@golangci-lint run -v --fix -c .golangci.yaml ./...

test:  ## Run all tests.
	${call print, "Running tests"}
	@go test -race \
			-run "$(FILTER)" \
			-coverpkg=./... \
			-coverprofile=coverage.out \
			-covermode=atomic \
			-count=1 \
			-timeout=10m \
			${GO_PACKAGES}