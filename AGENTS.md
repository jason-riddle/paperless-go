# Agent Documentation

This file contains documentation for AI agents working on this repository.

## Project Overview

paperless-go is a minimal, well-designed Go client library for the Paperless-ngx API. It follows Go standard library design principles and maintains zero external dependencies.

## Key Principles

### Zero Dependencies
- **Only use Go standard library** - No external dependencies are allowed
- This is a core principle of the project
- Prefer stdlib solutions over third-party libraries

### Code Quality Standards
- All code must be formatted with `go fmt`
- All code must pass `go vet`
- All code must pass `make lint`
- Tests must pass with race detector enabled: `go test -v -race ./...`
- Maintain or improve code coverage (currently 91.8%)

### Design Patterns
- **Context-first**: All API methods accept `context.Context` as the first parameter
- **Functional options**: Use functional options pattern for configuration
- **Type-safe**: Leverage Go generics for paginated responses
- **Error handling**: Use structured error types from `errors.go`

## Project Structure

```
.
├── cmd/
│   └── pgo/          # CLI tool for interacting with Paperless
│   └── pgo-rag/      # RAG CLI tool (separate module, has deps)
├── .github/
│   ├── workflows/    # GitHub Actions workflows
│   └── copilot-instructions.md
├── client.go         # Main client implementation
├── documents.go      # Document-related API methods
├── tags.go           # Tag-related API methods
├── types.go          # Type definitions
├── errors.go         # Error handling
└── *_test.go         # Test files
```

## Development Workflow

When making changes:

1. **Format code**: `go fmt ./...` (ALWAYS run this after modifying Go files)
2. **Vet code**: `go vet ./...` (ALWAYS run this after modifying Go files)
3. **Run tests**: `go test -v -race ./...`
4. **Run linters**: `make lint`
5. **For integration tests**: `make integration-test-full`

## Testing

- Unit tests mock HTTP calls using `httptest.Server`
- Integration tests use the `//go:build integration` tag
- Integration tests require a running Paperless-ngx instance via Docker Compose

## CLI Tool (pgo)

The `pgo` CLI tool provides command-line access to Paperless-ngx:

- Built in `cmd/pgo/`
- Uses the paperless-go library
- Configuration via environment variables or flags:
  - `PAPERLESS_URL` or `-url`: Paperless instance URL
  - `PAPERLESS_TOKEN` or `-token`: API authentication token

### CLI Commands

- `pgo get docs` - List all documents
- `pgo get docs <id>` - Get a specific document by ID
- `pgo get tags` - List all tags
- `pgo get tags <id>` - Get a specific tag by ID
- `pgo search docs <query>` - Search documents (use `-title-only` to search titles only)
- `pgo search tags <query>` - Search tags
- `pgo apply docs <id> --tags=<id1>,<id2>...` - Update tags for a document
- `pgo add tag "<name>"` - Create a new tag
- `pgo tagcache [path|build]` - Print or build the tag cache (`path` requires no authentication)
- `pgo doccache [path|build]` - Print or build the doc cache (`path` requires no authentication)
- `pgo rag <args>` - Invoke `pgo-rag` if installed in PATH

All commands return JSON output by default. Document output includes both tag IDs and resolved tag names for convenience.

## CLI Tool (pgo-rag)

The `pgo-rag` CLI tool provides local RAG indexing and search:

- Built in `cmd/pgo-rag/` (separate module with its own `go.mod`)
- Uses the paperless-go library
- Uses SQLite + embeddings client (dependencies are isolated from the root module)

### CLI Flags

- `-url` - Paperless instance URL (default: `$PAPERLESS_URL`)
- `-token` - API authentication token (default: `$PAPERLESS_TOKEN`)
- `-output-format` - Output format, only `json` is supported (default: `json`)
- `-force-refresh` - Force refresh tags cache, bypassing any cached data
- `-memory` - Use in-memory cache only, do not write to disk

### Tag Caching

The CLI includes a tag cache to reduce API calls when fetching tags for document display:

- **Cache Location**: `$XDG_CACHE_HOME/paperless-go/tags.json` (or `~/.cache/paperless-go/tags.json`)
- **TTL**: 12 hours (tags are auto-refreshed when stale)
- **Scope**: Cache is used by `pgo get docs` commands for tag name resolution
- **In-Memory Fallback**: If filesystem permissions prevent cache writes, the CLI automatically falls back to an in-memory cache that persists for the duration of the command
- **Explicit In-Memory Mode**: Use `-memory` flag to skip disk caching entirely
- **Force Refresh**: Use `-force-refresh` flag to bypass cache and fetch fresh data
- **Cache Inspection**: Use `pgo tagcache path` (or `pgo tagcache`) to print the full path to the cache file
- **Cache Build**: Use `pgo tagcache build` to fetch fresh tags and write the cache

The cache ensures that:
1. Commands work even with filesystem permission issues (automatic in-memory fallback)
2. Reduced API calls improve performance for read-heavy workflows
3. Tag data remains reasonably fresh (12-hour TTL)

## API Coverage

Current implementation:
- ✅ Documents (list, get, update, rename, update tags)
- ✅ Tags (list, get, create)

Future considerations:
- Document creation, deletion
- Tag update, deletion
- Correspondents, Document Types, Storage Paths
- Saved Views, Tasks
- File upload and download
- Bulk operations

## Important Notes

- Follow Go standard library design principles
- Keep the library minimal
- Write tests for all new functionality
- Update documentation for API changes
- Do not break existing behavior
