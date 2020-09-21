.PHONY: setup
setup: ## Install all the build and lint dependencies
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(GOPATH)/bin latest

.PHONY: dep
dep: ## Go get for entire project
	go get ./...

.PHONY: fmt
fmt: ## Run goimports on all go files
	find . -name '*.go' | while read -r file; do goimports -w "$$file"; done

.PHONY: lint
lint: ## Run all the linters
	golangci-lint run ./...

.PHONY: ci
ci: lint  ## Run all the tests and code checks

.PHONY: build
build: ## Build a version
	go build -v ./...

.DEFAULT_GOAL := build

.PHONY: clean
clean: ## Remove temporary files
	go clean

install: ## Install binary
	go install ./...

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'