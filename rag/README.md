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

`pgo-rag` uses an OpenAI-compatible embeddings endpoint (for example, Ollama).
Defaults target a local Ollama instance:

- `PGO_RAG_EMBEDDER_URL` (default: `http://localhost:11434/v1`)
- `PGO_RAG_EMBEDDER_MODEL` (default: `nomic-embed-text`)
- `PGO_RAG_EMBEDDER_KEY` (optional API key)
