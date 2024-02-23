VERSION := $(shell git describe --tags --always --dirty="-dev" --match "v*.*.*")
VERSION := $(VERSION:v%=%)

default: build

.PHONY: build
build:
	CGO_ENABLED=0 \
	@go build \
			-ldflags "-X main.version=${VERSION}" \
			-o ./bin/prometheus-sns-lambda-webhook \
		github.com/flashbots/prometheus-sns-lambda-webhook/cmd

.PHONY: snapshot
snapshot:
	@goreleaser release --snapshot --clean

.PHONY: release
release:
	@rm -rf ./dist
	@if [[ -z $${GITHUB_TOKEN} ]]; then GITHUB_TOKEN=$$( gh auth token ) goreleaser release; else goreleaser release; fi
