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

1. **Format code**: `go fmt ./...`
2. **Run tests**: `go test -v -race ./...`
3. **Run linters**: `make lint`
4. **For integration tests**: `make integration-test-full`

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

Document output includes tag names (not just IDs) for better readability.

## API Coverage

Current implementation:
- ✅ Documents (list, get)
- ✅ Tags (list, get)

Future considerations:
- Document creation, update, deletion
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
