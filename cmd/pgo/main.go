package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
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
	forceRefresh := flag.Bool("force-refresh", false, "Force refresh tags cache, bypassing any cached data")
	inMemoryCacheFlag := flag.Bool("memory", false, "Use in-memory cache only, do not write to disk")
	flag.Parse()

	// Set the global in-memory cache flag
	useInMemoryCache = *inMemoryCacheFlag

	// Parse command
	args := flag.Args()
	if len(args) == 0 {
		return fmt.Errorf("usage: pgo <command> [args]\nAvailable commands:\n  get docs - List documents\n  get docs <id> - Get specific document\n  get tags - List tags\n  get tags <id> - Get specific tag\n  tagcache - Print the cache file path")
	}

	command := args[0]

	// Handle tagcache command (no auth required)
	if command == "tagcache" {
		cachePath, err := getCacheFilePath()
		if err != nil {
			return fmt.Errorf("failed to get cache file path: %w", err)
		}
		fmt.Println(cachePath)
		return nil
	}

	// Check for required arguments for API commands
	if *baseURL == "" {
		return fmt.Errorf("paperless URL is required (use -url flag or PAPERLESS_URL env var)")
	}
	if *token == "" {
		return fmt.Errorf("API token is required (use -token flag or PAPERLESS_TOKEN env var)")
	}

	if command != "get" {
		return fmt.Errorf("unknown command: %s", command)
	}

	if len(args) < 2 {
		return fmt.Errorf("usage: pgo get <resource> [id]\nAvailable resources:\n  docs - List documents\n  docs <id> - Get specific document\n  tags - List tags\n  tags <id> - Get specific tag")
	}

	resource := args[1]
	if resource != "docs" && resource != "tags" {
		return fmt.Errorf("unknown resource: %s", resource)
	}

	// Check if an ID was provided
	var id int
	var hasID bool
	if len(args) > 2 {
		// Parse the ID argument
		if _, err := fmt.Sscanf(args[2], "%d", &id); err != nil {
			return fmt.Errorf("invalid ID format: %s", args[2])
		}
		hasID = true
	}

	// Create client
	client := paperless.NewClient(*baseURL, *token)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if resource == "docs" {
		if hasID {
			// Get specific document
			doc, err := client.GetDocument(ctx, id)
			if err != nil {
				return fmt.Errorf("failed to get document %d: %w", id, err)
			}

			// Fetch tag names for resolution (with caching)
			tagNames, err := getTagNamesWithCache(ctx, client, *forceRefresh, DefaultCacheTTL)
			if err != nil {
				// If tag fetching fails, continue but warn
				fmt.Fprintf(os.Stderr, "Warning: Could not fetch tags for name resolution: %v\n", err)
				tagNames = make(map[int]string) // Empty map as fallback
			}

			fmt.Printf("Document %d:\n", doc.ID)
			fmt.Printf("Title: %s\n", doc.Title)
			fmt.Printf("Created: %s\n", doc.Created.Time().Format(time.RFC3339))

			// Convert tag IDs to names
			tagNamesList := make([]string, len(doc.Tags))
			for i, tagID := range doc.Tags {
				if name, ok := tagNames[tagID]; ok {
					tagNamesList[i] = fmt.Sprintf("\"%s\"", name)
				} else {
					tagNamesList[i] = fmt.Sprintf("\"unknown(%d)\"", tagID)
				}
			}
			fmt.Printf("Tags: [%s]\n", strings.Join(tagNamesList, ", "))
			fmt.Printf("Content: %s\n", doc.Content)
		} else {
			// Fetch tag names for resolution (with caching)
			tagNames, err := getTagNamesWithCache(ctx, client, *forceRefresh, DefaultCacheTTL)
			if err != nil {
				// If tag fetching fails, continue but warn
				fmt.Fprintf(os.Stderr, "Warning: Could not fetch tags for name resolution: %v\n", err)
				tagNames = make(map[int]string) // Empty map as fallback
			}

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

				// Convert tag IDs to names
				tagNamesList := make([]string, len(doc.Tags))
				for i, tagID := range doc.Tags {
					if name, ok := tagNames[tagID]; ok {
						tagNamesList[i] = fmt.Sprintf("\"%s\"", name)
					} else {
						tagNamesList[i] = fmt.Sprintf("\"unknown(%d)\"", tagID)
					}
				}
				fmt.Printf("Tags: [%s]\n", strings.Join(tagNamesList, ", "))
				fmt.Println("---")
			}
		}
	} else if resource == "tags" {
		if hasID {
			// Get specific tag
			tag, err := client.GetTag(ctx, id)
			if err != nil {
				return fmt.Errorf("failed to get tag %d: %w", id, err)
			}

			fmt.Printf("Tag %d:\n", tag.ID)
			fmt.Printf("Name: %s\n", tag.Name)
			fmt.Printf("Slug: %s\n", tag.Slug)
			fmt.Printf("Color: %s\n", tag.Color)
			fmt.Printf("Document Count: %d\n", tag.DocumentCount)
		} else {
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
	}

	return nil
}
