package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	paperless "github.com/jason-riddle/paperless-go"
	"github.com/jason-riddle/paperless-go/cmd/pgo-rag/internal/embedding"
	"github.com/jason-riddle/paperless-go/cmd/pgo-rag/internal/indexer"
	"github.com/jason-riddle/paperless-go/cmd/pgo-rag/internal/storage"
)

const usage = `pgo-rag: local RAG indexing and search for Paperless

Usage:
  pgo-rag build   -db <path> -url <paperless-url> -token <api-token>
  pgo-rag search  -db <path> -query <text> [-limit 10] [-threshold 0.7]

Global flags:
  -url             Paperless instance URL (or PAPERLESS_URL)
  -token           Paperless API token (or PAPERLESS_TOKEN)
  -embeddings-url  Embeddings API base URL (or PGO_RAG_EMBEDDINGS_URL)
  -embeddings-key  Embeddings API key (or PGO_RAG_EMBEDDINGS_KEY)
  -embeddings-model Embeddings model name (or PGO_RAG_EMBEDDINGS_MODEL)
  -max-docs        Maximum documents to index (or PGO_RAG_MAX_DOCS)
  -tag             Tag name filter (or PGO_RAG_TAG)
`

func main() {
	loaded, err := loadDotEnv(".env")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load .env:", err)
	} else if loaded {
		fmt.Fprintln(os.Stderr, "loaded .env")
	} else {
		fmt.Fprintln(os.Stderr, "no .env found, using flags/env")
	}

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
	maxDocs := flags.Int("max-docs", getenvIntDefault("PGO_RAG_MAX_DOCS", 5), "Maximum documents to index (0 = no limit)")
	tagName := flags.String("tag", strings.TrimSpace(os.Getenv("PGO_RAG_TAG")), "Tag name filter (exact match)")
	embeddingsURL := flags.String("embeddings-url", os.Getenv("PGO_RAG_EMBEDDINGS_URL"), "Embeddings API base URL")
	embeddingsKey := flags.String("embeddings-key", os.Getenv("PGO_RAG_EMBEDDINGS_KEY"), "Embeddings API key")
	embeddingsModel := flags.String("embeddings-model", os.Getenv("PGO_RAG_EMBEDDINGS_MODEL"), "Embeddings model")

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
	if *embeddingsURL == "" {
		return fmt.Errorf("-embeddings-url is required")
	}
	if *embeddingsKey == "" {
		return fmt.Errorf("-embeddings-key is required")
	}
	if *embeddingsModel == "" {
		return fmt.Errorf("-embeddings-model is required")
	}

	db, err := storage.NewDB(*dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	client := paperless.NewClient(*url, *token)
	embedder := embedding.NewClient(*embeddingsURL, *embeddingsKey, *embeddingsModel)

	start := time.Now()
	summary, err := indexer.BuildIndex(ctx, client, db, embedder, indexer.BuildOptions{
		PageSize: *pageSize,
		MaxDocs:  *maxDocs,
		TagName:  *tagName,
	})
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
	embeddingsURL := flags.String("embeddings-url", os.Getenv("PGO_RAG_EMBEDDINGS_URL"), "Embeddings API base URL")
	embeddingsKey := flags.String("embeddings-key", os.Getenv("PGO_RAG_EMBEDDINGS_KEY"), "Embeddings API key")
	embeddingsModel := flags.String("embeddings-model", os.Getenv("PGO_RAG_EMBEDDINGS_MODEL"), "Embeddings model")

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
	if *embeddingsURL == "" {
		return fmt.Errorf("-embeddings-url is required")
	}
	if *embeddingsKey == "" {
		return fmt.Errorf("-embeddings-key is required")
	}
	if *embeddingsModel == "" {
		return fmt.Errorf("-embeddings-model is required")
	}

	db, err := storage.NewDB(*dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	embedder := embedding.NewClient(*embeddingsURL, *embeddingsKey, *embeddingsModel)

	summary, err := indexer.SearchIndex(ctx, db, embedder, *query, *limit, *threshold)
	if err != nil {
		return err
	}

	return writeJSON(summary)
}

func writeJSON(value interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func getenvIntDefault(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}

func loadDotEnv(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if info.IsDir() {
		return false, fmt.Errorf("%s is a directory", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		idx := strings.IndexByte(line, '=')
		if idx <= 0 {
			return false, fmt.Errorf("invalid .env line %d", lineNum)
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		if key == "" {
			return false, fmt.Errorf("invalid .env line %d", lineNum)
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return false, err
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return true, nil
}
