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
	forceRefresh := flag.Bool("force-refresh", false, "Force refresh caches, bypassing any cached data")
	inMemoryCacheFlag := flag.Bool("memory", false, "Use in-memory cache only for tags and docs, do not write to disk")
	flag.Parse()

	// Set the global in-memory cache flags for both tag and doc caches
	useInMemoryCache = *inMemoryCacheFlag
	useInMemoryDocCache = *inMemoryCacheFlag

	// Parse command
	args := flag.Args()
	if len(args) == 0 {
		return fmt.Errorf("usage: pgo <command> [args]\nAvailable commands:\n  get docs - List documents\n  get docs <id> - Get specific document\n  get tags - List tags\n  get tags <id> - Get specific tag\n  search docs <query> - Search documents (use -title-only to search titles only)\n  search tags <query> - Search tags\n  tagcache - Print the tag cache file path\n  doccache - Print the doc cache file path")
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

	// Handle doccache command (no auth required)
	if command == "doccache" {
		cachePath, err := getDocCacheFilePath()
		if err != nil {
			return fmt.Errorf("failed to get doc cache file path: %w", err)
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

	if command != "get" && command != "search" {
		return fmt.Errorf("unknown command: %s", command)
	}

	if len(args) < 2 {
		return fmt.Errorf("usage: pgo %s <resource> [args]\nAvailable resources:\n  docs - Documents\n  tags - Tags", command)
	}

	resource := args[1]
	if resource != "docs" && resource != "tags" {
		return fmt.Errorf("unknown resource: %s", resource)
	}

	// Check if an ID was provided
	var id int
	var hasID bool
	if command == "get" && len(args) > 2 {
		// Parse the ID argument
		if _, err := fmt.Sscanf(args[2], "%d", &id); err != nil {
			return fmt.Errorf("invalid ID format: %s", args[2])
		}
		hasID = true
	}

	var searchQuery string
	var titleOnly bool
	if command == "search" {
		switch resource {
		case "docs":
			searchFlags := flag.NewFlagSet("search docs", flag.ContinueOnError)
			titleOnlyFlag := searchFlags.Bool("title-only", false, "Search only document titles")
			if err := searchFlags.Parse(args[2:]); err != nil {
				return err
			}
			remaining := searchFlags.Args()
			if len(remaining) == 0 {
				return fmt.Errorf("usage: pgo search docs [-title-only] <query>")
			}
			searchQuery = strings.Join(remaining, " ")
			titleOnly = *titleOnlyFlag
		case "tags":
			if len(args) < 3 {
				return fmt.Errorf("usage: pgo search tags <query>")
			}
			searchQuery = strings.Join(args[2:], " ")
		}
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
			var opts *paperless.ListOptions
			if command == "search" {
				opts = &paperless.ListOptions{
					Query:     searchQuery,
					TitleOnly: titleOnly,
				}
			}
			docs, err := client.ListDocuments(ctx, opts)
			if err != nil {
				return fmt.Errorf("failed to %s documents: %w", command, err)
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
			var opts *paperless.ListOptions
			if command == "search" {
				opts = &paperless.ListOptions{
					Query: searchQuery,
				}
			}
			tags, err := client.ListTags(ctx, opts)
			if err != nil {
				return fmt.Errorf("failed to %s tags: %w", command, err)
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
