.PHONY: help build test generate prerender tidy lint clean

help:
	@echo "Targets:"
	@echo "  build       Build the ailloy-docs binary into ./bin"
	@echo "  test        Run go test -race ./..."
	@echo "  generate    Regenerate pre-rendered glamour artifacts"
	@echo "  prerender   Alias for generate"
	@echo "  tidy        go mod tidy"
	@echo "  lint        Run golangci-lint"
	@echo "  clean       Remove build artifacts"

DOCS_VERSION := $(shell tr -d '\n[:space:]' < ailloy-docs-version.txt)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

build:
	@mkdir -p bin
	go build \
		-trimpath \
		-ldflags "-s -w -X main.version=$(VERSION) -X main.ailloyDocsVersion=$(DOCS_VERSION)" \
		-o bin/ailloy-docs \
		./cmd/ailloy-docs

test:
	go test -race ./...

generate prerender:
	go run ./cmd/prerender -out internal/embedded -v

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

clean:
	rm -rf bin dist
