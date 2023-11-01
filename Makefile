# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

export PATH := $(abspath bin/):${PATH}

.PHONY: build
build: ## Build all binaries
	@mkdir -p build
	go build -trimpath -o build/app .

.PHONY: run
run: build ## Build and run the application
	build/app

.PHONY: check
check: test lint ## Run checks (tests and linters)

.PHONY: test
test: ## Run tests
	go test -v ./...

.PHONY: lint
lint: ## Run linter
	golangci-lint run

# Dependency versions
GOLANGCI_VERSION ?= 1.52.2

deps: bin/golangci-lint

bin/golangci-lint:
	@mkdir -p bin
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | BINARY=golangci-lint bash -s -- v${GOLANGCI_VERSION}

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-10s\033[0m %s\n", $$1, $$2}'
