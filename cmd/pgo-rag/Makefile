SHELL := /bin/bash

GO ?= go
BINARY ?= pgo-rag
OUT_DIR ?= bin

-include .env

PAPERLESS_URL ?= $(PAPERLESS_URL)
PAPERLESS_TOKEN ?= $(PAPERLESS_TOKEN)
PGO_RAG_EMBEDDINGS_URL ?= $(PGO_RAG_EMBEDDINGS_URL)
PGO_RAG_EMBEDDINGS_KEY ?= $(PGO_RAG_EMBEDDINGS_KEY)
PGO_RAG_EMBEDDINGS_MODEL ?= $(PGO_RAG_EMBEDDINGS_MODEL)
PGO_RAG_MAX_DOCS ?= 5
PGO_RAG_TAG ?= $(PGO_RAG_TAG)
RAG_PAGE_SIZE ?=
RAG_DB ?= rag.db
RAG_QUERY ?=
RAG_LIMIT ?=
RAG_THRESHOLD ?=

.PHONY: all build rag rag-build rag-search test test-race fmt vet tidy clean help

all: build

build:
	@mkdir -p "$(OUT_DIR)"
	$(GO) build -o "$(OUT_DIR)/$(BINARY)" .

rag: rag-build

rag-build:
	$(GO) run . build -db "$(RAG_DB)" -url "$(PAPERLESS_URL)" -token "$(PAPERLESS_TOKEN)" -page-size "$(RAG_PAGE_SIZE)" -max-docs "$(PGO_RAG_MAX_DOCS)" -tag "$(PGO_RAG_TAG)" -embeddings-url "$(PGO_RAG_EMBEDDINGS_URL)" -embeddings-key "$(PGO_RAG_EMBEDDINGS_KEY)" -embeddings-model "$(PGO_RAG_EMBEDDINGS_MODEL)" $(RAG_ARGS)

rag-search:
	$(GO) run . search -db "$(RAG_DB)" -query "$(RAG_QUERY)" -limit "$(RAG_LIMIT)" -threshold "$(RAG_THRESHOLD)" -embeddings-url "$(PGO_RAG_EMBEDDINGS_URL)" -embeddings-key "$(PGO_RAG_EMBEDDINGS_KEY)" -embeddings-model "$(PGO_RAG_EMBEDDINGS_MODEL)" $(RAG_ARGS)

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

clean:
	@rm -rf "$(OUT_DIR)"

help:
	@printf '%s\n' \
	  'Targets:' \
	  '  build      Build ./$(OUT_DIR)/$(BINARY)' \
	  '  rag        Run build via go run .' \
	  '  rag-build  Run build subcommand via go run .' \
	  '  rag-search Run search subcommand via go run .' \
	  '  test       Run unit tests' \
	  '  test-race  Run tests with -race' \
	  '  fmt        Run go fmt ./...' \
	  '  vet        Run go vet ./...' \
	  '  tidy       Run go mod tidy' \
	  '  clean      Remove $(OUT_DIR)'
