# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

export PATH := $(abspath bin/):${PATH}
OS = $(shell uname | tr A-Z a-z)

.PHONY: build
build: ## Build all binaries
	@mkdir -p build
	dagger call build
	# go build -trimpath -o build/app .

.PHONY: run
run: build ## Build and run the application
	build/app

.PHONY: check
check: test lint ## Run checks (tests and linters)

.PHONY: test
test: ## Run tests
	dagger call test

.PHONY: lint
lint: ## Run linter
	dagger call lint

# Dependency versions
GOLANGCI_VERSION ?= 1.52.2
DAGGER_VERSION ?= 0.9.3

deps: bin/golangci-lint

bin/golangci-lint:
	@mkdir -p bin
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | BINARY=golangci-lint bash -s -- v${GOLANGCI_VERSION}

bin/dagger:
	@mkdir -p bin
	curl -L https://dl.dagger.io/dagger/install.sh | sh
	@echo ${HELLO}

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-10s\033[0m %s\n", $$1, $$2}'







































HELLO := "🦄 🌈 🦄 🌈 🦄 🌈"
