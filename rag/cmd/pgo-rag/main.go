package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	paperless "github.com/jason-riddle/paperless-go"
	"github.com/jason-riddle/paperless-go/rag/internal/embedding"
	"github.com/jason-riddle/paperless-go/rag/internal/indexer"
	"github.com/jason-riddle/paperless-go/rag/internal/storage"
)

const usage = `pgo-rag: local RAG indexing and search for Paperless

Usage:
  pgo-rag build   -db <path> -url <paperless-url> -token <api-token>
  pgo-rag search  -db <path> -query <text> [-limit 10] [-threshold 0.7]

Global flags:
  -url           Paperless instance URL (or PAPERLESS_URL)
  -token         Paperless API token (or PAPERLESS_TOKEN)
  -embedder-url  Embeddings API base URL (or PGO_RAG_EMBEDDER_URL)
  -embedder-key  Embeddings API key (or PGO_RAG_EMBEDDER_KEY)
  -embedder-model Embeddings model name (or PGO_RAG_EMBEDDER_MODEL)
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	ctx := context.Background()
	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "build":
		if err := runBuild(ctx, args); err != nil {
			fmt.Fprintln(os.Stderr, "build error:", err)
			os.Exit(1)
		}
	case "search":
		if err := runSearch(ctx, args); err != nil {
			fmt.Fprintln(os.Stderr, "search error:", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		fmt.Fprint(os.Stdout, usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
}

func runBuild(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("build", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	dbPath := flags.String("db", "", "SQLite database path")
	url := flags.String("url", os.Getenv("PAPERLESS_URL"), "Paperless URL")
	token := flags.String("token", os.Getenv("PAPERLESS_TOKEN"), "Paperless token")
	pageSize := flags.Int("page-size", 100, "Paperless page size")
	embedderURL := flags.String("embedder-url", getenvDefault("PGO_RAG_EMBEDDER_URL", "http://localhost:11434/v1"), "Embeddings API base URL")
	embedderKey := flags.String("embedder-key", os.Getenv("PGO_RAG_EMBEDDER_KEY"), "Embeddings API key")
	embedderModel := flags.String("embedder-model", getenvDefault("PGO_RAG_EMBEDDER_MODEL", "nomic-embed-text"), "Embeddings model")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *dbPath == "" {
		return fmt.Errorf("-db is required")
	}
	if *url == "" {
		return fmt.Errorf("-url is required")
	}
	if *token == "" {
		return fmt.Errorf("-token is required")
	}
	if *embedderURL == "" {
		return fmt.Errorf("-embedder-url is required")
	}
	if *embedderModel == "" {
		return fmt.Errorf("-embedder-model is required")
	}

	db, err := storage.NewDB(*dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	client := paperless.NewClient(*url, *token)
	embedder := newEmbedder(*embedderURL, *embedderKey, *embedderModel)

	start := time.Now()
	summary, err := indexer.BuildIndex(ctx, client, db, embedder, indexer.BuildOptions{PageSize: *pageSize})
	if err != nil {
		return err
	}

	resp := struct {
		indexer.BuildSummary
		DurationMs int64 `json:"duration_ms"`
	}{
		BuildSummary: summary,
		DurationMs:   time.Since(start).Milliseconds(),
	}

	return writeJSON(resp)
}

func runSearch(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("search", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	dbPath := flags.String("db", "", "SQLite database path")
	query := flags.String("query", "", "Search query")
	limit := flags.Int("limit", 10, "Max results")
	threshold := flags.Float64("threshold", 0.7, "Similarity threshold")
	embedderURL := flags.String("embedder-url", getenvDefault("PGO_RAG_EMBEDDER_URL", "http://localhost:11434/v1"), "Embeddings API base URL")
	embedderKey := flags.String("embedder-key", os.Getenv("PGO_RAG_EMBEDDER_KEY"), "Embeddings API key")
	embedderModel := flags.String("embedder-model", getenvDefault("PGO_RAG_EMBEDDER_MODEL", "nomic-embed-text"), "Embeddings model")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *dbPath == "" {
		return fmt.Errorf("-db is required")
	}
	if *query == "" {
		return fmt.Errorf("-query is required")
	}
	if *limit <= 0 {
		return fmt.Errorf("-limit must be > 0")
	}
	if *threshold < 0 || *threshold > 1 {
		return fmt.Errorf("-threshold must be between 0 and 1")
	}
	if *embedderURL == "" {
		return fmt.Errorf("-embedder-url is required")
	}
	if *embedderModel == "" {
		return fmt.Errorf("-embedder-model is required")
	}

	db, err := storage.NewDB(*dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	embedder := newEmbedder(*embedderURL, *embedderKey, *embedderModel)

	summary, err := indexer.SearchIndex(ctx, db, embedder, *query, *limit, *threshold)
	if err != nil {
		return err
	}

	return writeJSON(summary)
}

func newEmbedder(baseURL, apiKey, model string) indexer.Embedder {
	if apiKey == "" {
		return embedding.NewOllamaClient(baseURL, model)
	}
	return embedding.NewClientWithBaseURL(apiKey, model, baseURL)
}

func writeJSON(value interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
