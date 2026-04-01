default: help

.PHONY: help
help: ## Show this help.
	@fgrep -h "##" $(MAKEFILE_LIST)  | fgrep -v fgrep | sed -e 's/:.*##/:##/' | awk -F':##' '{printf "%-12s %s\n",$$1, $$2}'

.PHONY: lint
lint: ## Run golangci-lint.
	golangci-lint run ./...

.PHONY: test
test: ## Run tests.
	go test ./...

.PHONY: format
format: ## Format code.
	gofmt -s -w .

.PHONY: build
build: ## Build the binary.
	go build -o bin/bwql ./cmd/bwql

.PHONY: run
run: ## Run bwql.
	go run ./cmd/bwql
