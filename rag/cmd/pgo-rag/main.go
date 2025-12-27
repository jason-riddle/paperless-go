package main

import (
	"context"
	"flag"
	"fmt"
	"os"
)

const usage = `pgo-rag: local RAG indexing and search for Paperless

Usage:
  pgo-rag build   -db <path> -url <paperless-url> -token <api-token>
  pgo-rag search  -db <path> -query <text> [-limit 10] [-threshold 0.7]

Global flags:
  -url    Paperless instance URL (or PAPERLESS_URL)
  -token  Paperless API token (or PAPERLESS_TOKEN)
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

func runBuild(_ context.Context, args []string) error {
	flags := flag.NewFlagSet("build", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	dbPath := flags.String("db", "", "SQLite database path")
	url := flags.String("url", os.Getenv("PAPERLESS_URL"), "Paperless URL")
	token := flags.String("token", os.Getenv("PAPERLESS_TOKEN"), "Paperless token")

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

	return fmt.Errorf("rag build not implemented yet")
}

func runSearch(_ context.Context, args []string) error {
	flags := flag.NewFlagSet("search", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	dbPath := flags.String("db", "", "SQLite database path")
	query := flags.String("query", "", "Search query")
	limit := flags.Int("limit", 10, "Max results")
	threshold := flags.Float64("threshold", 0.7, "Similarity threshold")

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

	return fmt.Errorf("rag search not implemented yet")
}
