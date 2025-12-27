# pgo-rag

`pgo-rag` is a separate binary that provides local RAG indexing and search for Paperless.

This module intentionally lives under `rag/` so the root `paperless-go` module stays
zero-dependency. The RAG implementation can depend on SQLite and local embedding
models without polluting the core library.

## Development

From the repository root:

```
go work init
```

Or use the included `go.work` file (if present) to work across modules.

Build the RAG CLI:

```
cd rag
go build ./cmd/pgo-rag
```

## Planned commands

- `pgo-rag build` — build or refresh the local SQLite index
- `pgo-rag search` — run a similarity search against the local index

## Embeddings configuration

`pgo-rag` uses an OpenAI-compatible embeddings endpoint.

- `PGO_RAG_EMBEDDINGS_URL` (required)
- `PGO_RAG_EMBEDDINGS_KEY` (required)
- `PGO_RAG_EMBEDDINGS_MODEL` (required)
- `PGO_RAG_MAX_DOCS` (optional; limit indexed documents for testing, default: 5)
- `PGO_RAG_TAG` or `PAPERLESS_RAG_INDEX_TAG` (optional; tag name filter, exact match)
