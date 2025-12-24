package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jason-riddle/paperless-go"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse command line flags
	baseURL := flag.String("url", os.Getenv("PAPERLESS_URL"), "Paperless instance URL (default: $PAPERLESS_URL)")
	token := flag.String("token", os.Getenv("PAPERLESS_TOKEN"), "API authentication token (default: $PAPERLESS_TOKEN)")
	flag.Parse()

	// Check for required arguments
	if *baseURL == "" {
		return fmt.Errorf("paperless URL is required (use -url flag or PAPERLESS_URL env var)")
	}
	if *token == "" {
		return fmt.Errorf("API token is required (use -token flag or PAPERLESS_TOKEN env var)")
	}

	// Parse command
	args := flag.Args()
	if len(args) == 0 {
		return fmt.Errorf("usage: pgo <command> [args]\nAvailable commands:\n  get docs - List documents\n  get tags - List tags")
	}

	command := args[0]
	if command != "get" {
		return fmt.Errorf("unknown command: %s", command)
	}

	if len(args) < 2 {
		return fmt.Errorf("usage: pgo get <resource>\nAvailable resources:\n  docs - List documents\n  tags - List tags")
	}

	resource := args[1]
	if resource != "docs" && resource != "tags" {
		return fmt.Errorf("unknown resource: %s", resource)
	}

	// Create client
	client := paperless.NewClient(*baseURL, *token)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if resource == "docs" {
		// Fetch documents
		docs, err := client.ListDocuments(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to list documents: %w", err)
		}

		// Display results
		fmt.Printf("Found %d documents\n\n", docs.Count)
		for _, doc := range docs.Results {
			fmt.Printf("ID: %d\n", doc.ID)
			fmt.Printf("Title: %s\n", doc.Title)
			fmt.Printf("Created: %s\n", doc.Created.Time().Format(time.RFC3339))
			fmt.Printf("Tags: %v\n", doc.Tags)
			fmt.Println("---")
		}
	} else if resource == "tags" {
		// Fetch tags
		tags, err := client.ListTags(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to list tags: %w", err)
		}

		// Display results
		fmt.Printf("Found %d tags\n\n", tags.Count)
		for _, tag := range tags.Results {
			fmt.Printf("ID: %d\n", tag.ID)
			fmt.Printf("Name: %s\n", tag.Name)
			fmt.Printf("Slug: %s\n", tag.Slug)
			fmt.Printf("Color: %s\n", tag.Color)
			fmt.Printf("Document Count: %d\n", tag.DocumentCount)
			fmt.Println("---")
		}
	}

	return nil
}
