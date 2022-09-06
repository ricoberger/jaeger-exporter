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
