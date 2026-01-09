# paperless-go

A minimal, well-designed Go client library for the [Paperless-ngx](https://github.com/paperless-ngx/paperless-ngx) API.

[![Go Reference](https://pkg.go.dev/badge/github.com/jason-riddle/paperless-go.svg)](https://pkg.go.dev/github.com/jason-riddle/paperless-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/jason-riddle/paperless-go)](https://goreportcard.com/report/github.com/jason-riddle/paperless-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- **Zero external dependencies** - Uses only the Go standard library
- **Context-first API** - All methods accept `context.Context` for cancellation and timeouts
- **Functional options** - Configure clients using the functional options pattern
- **Type-safe** - Leverages Go generics for paginated responses
- **Well-tested** - Comprehensive unit tests and integration tests
- **Go stdlib patterns** - Follows design principles from Go's standard library

## Installation

```bash
go get github.com/jason-riddle/paperless-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/jason-riddle/paperless-go"
)

func main() {
    // Create a new client
    client := paperless.NewClient(
        "http://localhost:8000",
        "your-api-token",
    )

    // List documents
    docs, err := client.ListDocuments(context.Background(), nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d documents\n", docs.Count)
    for _, doc := range docs.Results {
        fmt.Printf("- %s\n", doc.Title)
    }
}
```

## Usage

### Creating a Client

```go
// Basic client with default settings
client := paperless.NewClient("http://localhost:8000", "your-api-token")

// Client with custom timeout
client := paperless.NewClient(
    "http://localhost:8000",
    "your-api-token",
    paperless.WithTimeout(30*time.Second),
)

// Client with custom HTTP client (for connection pooling, proxies, etc.)
httpClient := &http.Client{
    Timeout: 60 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns: 10,
    },
}
client := paperless.NewClient(
    "http://localhost:8000",
    "your-api-token",
    paperless.WithHTTPClient(httpClient),
)
```

### Documents

#### List Documents

```go
// List all documents
docs, err := client.ListDocuments(context.Background(), nil)

// List with pagination
docs, err := client.ListDocuments(context.Background(), &paperless.ListOptions{
    Page:     2,
    PageSize: 50,
})

// Search documents
docs, err := client.ListDocuments(context.Background(), &paperless.ListOptions{
    Query: "invoice",
})

// Sort documents
docs, err := client.ListDocuments(context.Background(), &paperless.ListOptions{
    Ordering: "-created", // Sort by created date, descending
})

// Combine options
docs, err := client.ListDocuments(context.Background(), &paperless.ListOptions{
    Query:    "important",
    Ordering: "-added",
    PageSize: 25,
})
```

#### Get a Single Document

```go
doc, err := client.GetDocument(context.Background(), 123)
if err != nil {
    if paperless.IsNotFound(err) {
        fmt.Println("Document not found")
        return
    }
    log.Fatal(err)
}

fmt.Printf("Title: %s\n", doc.Title)
fmt.Printf("Created: %s\n", doc.Created)
fmt.Printf("Tags: %v\n", doc.Tags)
```

#### Rename a Document

```go
// Rename document with ID 123
doc, err := client.RenameDocument(context.Background(), 123, "New Document Title")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Renamed to: %s\n", doc.Title)
```

#### Update Document Tags

```go
// Update document tags (adds or replaces existing tags)
doc, err := client.UpdateDocumentTags(context.Background(), 123, []int{1, 2, 3})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Tags: %v\n", doc.Tags)

// Remove all tags from a document
doc, err = client.UpdateDocumentTags(context.Background(), 123, []int{})
if err != nil {
    log.Fatal(err)
}
```

### Tags

#### List Tags

```go
// List all tags
tags, err := client.ListTags(context.Background(), nil)

// Sort tags by name
tags, err := client.ListTags(context.Background(), &paperless.ListOptions{
    Ordering: "name",
})
```

#### Get a Single Tag

```go
tag, err := client.GetTag(context.Background(), 1)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Tag: %s\n", tag.Name)
fmt.Printf("Color: %s\n", tag.Color)
fmt.Printf("Documents: %d\n", tag.DocumentCount)
```

### Error Handling

The library provides structured error types and helper functions:

```go
doc, err := client.GetDocument(context.Background(), 999)
if err != nil {
    // Check for 404 errors
    if paperless.IsNotFound(err) {
        fmt.Println("Document not found")
        return
    }
    
    // Access error details
    if apiErr, ok := err.(*paperless.Error); ok {
        fmt.Printf("API Error: %d %s (operation: %s)\n", 
            apiErr.StatusCode, apiErr.Message, apiErr.Op)
        return
    }
    
    // Other errors (network, timeout, etc.)
    log.Fatal(err)
}
```

### Context Usage

All API methods accept a `context.Context` for cancellation and timeouts:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

docs, err := client.ListDocuments(ctx, nil)

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Cancel from another goroutine when needed
go func() {
    <-someChannel
    cancel()
}()

doc, err := client.GetDocument(ctx, 123)
```

## API Coverage

This library currently implements core operations:

- ✅ Documents (list, get, update, rename, update tags)
- ✅ Tags (list, get, create)

Future versions may include:

- ⏳ Document creation, deletion
- ⏳ Tag update, deletion
- ⏳ Correspondents (list, get, create, update, delete)
- ⏳ Document Types (list, get, create, update, delete)
- ⏳ Storage Paths (list, get, create, update, delete)
- ⏳ Saved Views (list, get, create, update, delete)
- ⏳ Tasks (list, get)
- ⏳ File upload and download
- ⏳ Bulk operations

## CLI (pgo)

Build the CLI tool:

```bash
go build -o pgo ./cmd/pgo
```

The CLI uses `PAPERLESS_URL` and `PAPERLESS_TOKEN` (or the `-url`/`-token` flags).

### Output Format

All CLI commands return JSON by default. The `-output-format` flag can be used to specify the output format (currently only `json` is supported):

```bash
# Get tags (returns JSON)
./pgo get tags

# Explicitly specify JSON format
./pgo -output-format=json get tags

# Example output:
# {
#   "count": 2,
#   "next": null,
#   "previous": null,
#   "results": [
#     {
#       "id": 1,
#       "name": "Finance",
#       "slug": "finance",
#       "color": "#a6cee3",
#       "document_count": 5
#     }
#   ]
# }
```

### Document Output

Documents include both tag IDs and resolved tag names for convenience:

```bash
./pgo get docs 123

# Example output:
# {
#   "id": 123,
#   "title": "Invoice 2023-001",
#   "content": "...",
#   "created": "2023-01-15T10:30:00Z",
#   "modified": "2023-01-15T10:30:00Z",
#   "added": "2023-01-15T10:30:00Z",
#   "archive_serial_number": null,
#   "original_file_name": "invoice.pdf",
#   "tags": [1, 2],
#   "tag_names": ["Finance", "Important"]
# }
```

### Search Examples

```bash
# Search document titles and content
./pgo search docs "invoice"

# Search document titles only
./pgo search docs -title-only "invoice"

# Search tags
./pgo search tags "finance"
```

## Testing

### Unit Tests

Run unit tests with:

```bash
make test
```

Or directly:

```bash
go test -v -race ./...
```

### Integration Tests

Integration tests run against a real Paperless-ngx instance using Docker Compose.

1. Start Paperless-ngx:

```bash
make integration-setup
```

2. Get the API token (displayed by the setup script) and export it:

```bash
export PAPERLESS_TOKEN=your-token-here
```

3. Run integration tests:

```bash
make integration-test
```

4. Clean up:

```bash
make integration-teardown
```

Or run everything in one command:

```bash
export PAPERLESS_TOKEN=your-token-here
make integration-test-full
```

### Linting

```bash
make lint
```

Or individually:

```bash
make vet    # Run go vet
make fmt    # Check formatting
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### Guidelines

- Follow Go standard library design principles
- Write tests for new functionality
- Run `make lint` and `make test` before submitting
- Keep the library minimal - no external dependencies

## Documentation

- [Paperless-ngx API Documentation](https://docs.paperless-ngx.com/api/)
- [Go Package Documentation](https://pkg.go.dev/github.com/jason-riddle/paperless-go)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Paperless-ngx](https://github.com/paperless-ngx/paperless-ngx) - The amazing document management system this library connects to
