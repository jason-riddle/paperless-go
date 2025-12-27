# pgo-rag

`pgo-rag` is a separate binary that provides local RAG indexing and search for Paperless.

This command lives under `cmd/pgo-rag` so the root `paperless-go` module remains
zero-dependency. The RAG implementation can depend on SQLite and embedding
clients without polluting the core library.

## Development

From the repository root:

```
cd cmd/pgo-rag
go build
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
- `PGO_RAG_TAG` (optional; tag name filter, exact match; unset = all documents)
