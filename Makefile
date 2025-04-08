BRANCH    ?= $(shell git rev-parse --abbrev-ref HEAD)
BUILDTIME ?= $(shell date '+%Y-%m-%d@%H:%M:%S')
BUILDUSER ?= $(shell id -un)
REVISION  ?= $(shell git rev-parse HEAD)
VERSION   ?= $(shell git describe --tags)

.PHONY: build
build:
	@go build -ldflags "-X github.com/ricoberger/jaeger-exporter/pkg/version.Version=${VERSION} \
		-X github.com/ricoberger/jaeger-exporter/pkg/version.Revision=${REVISION} \
		-X github.com/ricoberger/jaeger-exporter/pkg/version.Branch=${BRANCH} \
		-X github.com/ricoberger/jaeger-exporter/pkg/version.BuildUser=${BUILDUSER} \
		-X github.com/ricoberger/jaeger-exporter/pkg/version.BuildDate=${BUILDTIME}" \
		-o ./bin/exporter ./cmd/exporter;

.PHONY: run
run: ## Run a controller from your host.
	go run ./cmd/exporter/exporter.go

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt vet ## Run tests.
	go test -covermode=atomic -coverpkg=./... -coverprofile=coverage.out -v ./...

.PHONY: lint
lint: ## Run golangci-lint linter
	golangci-lint run

.PHONY: lint-fix
lint-fix: ## Run golangci-lint linter and perform fixes
	golangci-lint run --fix
