package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/jason-riddle/paperless-go"
)

// DocumentWithTagNames represents a document with tag names resolved
type DocumentWithTagNames struct {
	ID                  int      `json:"id"`
	Title               string   `json:"title"`
	Content             string   `json:"content"`
	Created             string   `json:"created"`
	Modified            string   `json:"modified"`
	Added               string   `json:"added"`
	ArchiveSerialNumber *int     `json:"archive_serial_number"`
	OriginalFileName    string   `json:"original_file_name"`
	Tags                []int    `json:"tags"`
	TagNames            []string `json:"tag_names"`
}

// DocumentListOutput represents the output for list documents command
type DocumentListOutput struct {
	Count   int                    `json:"count"`
	Results []DocumentWithTagNames `json:"results"`
}

// CacheBuildOutput represents the output for cache build commands
type CacheBuildOutput struct {
	Path      string `json:"path"`
	Entries   int    `json:"entries"`
	FetchedAt string `json:"fetched_at"`
	InMemory  bool   `json:"in_memory"`
}

// convertDocToOutput converts a paperless.Document to DocumentWithTagNames
func convertDocToOutput(doc *paperless.Document, tagNames map[int]string) DocumentWithTagNames {
	tagNamesList := make([]string, len(doc.Tags))
	for i, tagID := range doc.Tags {
		if name, ok := tagNames[tagID]; ok {
			tagNamesList[i] = name
		} else {
			tagNamesList[i] = fmt.Sprintf("unknown(%d)", tagID)
		}
	}

	return DocumentWithTagNames{
		ID:                  doc.ID,
		Title:               doc.Title,
		Content:             doc.Content,
		Created:             doc.Created.Time().Format(time.RFC3339),
		Modified:            doc.Modified.Time().Format(time.RFC3339),
		Added:               doc.Added.Time().Format(time.RFC3339),
		ArchiveSerialNumber: doc.ArchiveSerialNumber,
		OriginalFileName:    doc.OriginalFileName,
		Tags:                doc.Tags,
		TagNames:            tagNamesList,
	}
}

// outputJSON outputs data as JSON to stdout
func outputJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

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
	outputFormat := flag.String("output-format", "json", "Output format (only 'json' is supported)")
	flag.Parse()

	// Set the global in-memory cache flags for both tag and doc caches
	useInMemoryCache = *inMemoryCacheFlag
	useInMemoryDocCache = *inMemoryCacheFlag

	// Validate output format
	if *outputFormat != "json" {
		return fmt.Errorf("unsupported output format: %s (only 'json' is supported)", *outputFormat)
	}

	// Parse command
	args := flag.Args()
	if len(args) == 0 {
		return fmt.Errorf("usage: pgo <command> [args]\nAvailable commands:\n  get docs - List documents\n  get docs <id> - Get specific document\n  get tags - List tags\n  get tags <id> - Get specific tag\n  search docs <query> - Search documents (use -title-only to search titles only)\n  search tags <query> - Search tags\n  apply docs <id> --tags=<id1>,<id2>... - Update tags for a document\n  add tag \"<name>\" - Create a new tag\n  rag <args> - Run pgo-rag (RAG indexing/search)\n  tagcache [path|build] - Print or build the tag cache\n  doccache [path|build] - Print or build the doc cache")
	}

	command := args[0]

	// Handle tagcache command
	if command == "tagcache" {
		subcommand := ""
		if len(args) > 1 {
			subcommand = args[1]
		}

		switch subcommand {
		case "", "path":
			if len(args) > 2 {
				return fmt.Errorf("usage: pgo tagcache [path|build]")
			}
			cachePath, err := getCacheFilePath()
			if err != nil {
				return fmt.Errorf("failed to get cache file path: %w", err)
			}
			fmt.Println(cachePath)
			return nil
		case "build":
			if len(args) > 2 {
				return fmt.Errorf("usage: pgo tagcache [path|build]")
			}
			if *baseURL == "" {
				return fmt.Errorf("paperless URL is required (use -url flag or PAPERLESS_URL env var)")
			}
			if *token == "" {
				return fmt.Errorf("API token is required (use -token flag or PAPERLESS_TOKEN env var)")
			}

			client := paperless.NewClient(*baseURL, *token)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			tagNames, err := getTagNamesWithCache(ctx, client, true, DefaultCacheTTL)
			if err != nil {
				return fmt.Errorf("failed to build tag cache: %w", err)
			}

			cachePath, err := getCacheFilePath()
			if err != nil {
				return fmt.Errorf("failed to get cache file path: %w", err)
			}

			fetchedAt := time.Now()
			if cache, err := loadTagCache(); err == nil && cache != nil {
				fetchedAt = cache.FetchedAt
			}

			output := CacheBuildOutput{
				Path:      cachePath,
				Entries:   len(tagNames),
				FetchedAt: fetchedAt.Format(time.RFC3339),
				InMemory:  useInMemoryCache,
			}
			if err := outputJSON(output); err != nil {
				return fmt.Errorf("failed to output JSON: %w", err)
			}
			return nil
		default:
			return fmt.Errorf("usage: pgo tagcache [path|build]")
		}
	}

	// Handle doccache command
	if command == "doccache" {
		subcommand := ""
		if len(args) > 1 {
			subcommand = args[1]
		}

		switch subcommand {
		case "", "path":
			if len(args) > 2 {
				return fmt.Errorf("usage: pgo doccache [path|build]")
			}
			cachePath, err := getDocCacheFilePath()
			if err != nil {
				return fmt.Errorf("failed to get doc cache file path: %w", err)
			}
			fmt.Println(cachePath)
			return nil
		case "build":
			if len(args) > 2 {
				return fmt.Errorf("usage: pgo doccache [path|build]")
			}
			if *baseURL == "" {
				return fmt.Errorf("paperless URL is required (use -url flag or PAPERLESS_URL env var)")
			}
			if *token == "" {
				return fmt.Errorf("API token is required (use -token flag or PAPERLESS_TOKEN env var)")
			}

			client := paperless.NewClient(*baseURL, *token)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			docNames, err := getDocNamesWithCache(ctx, client, true, DefaultCacheTTL)
			if err != nil {
				return fmt.Errorf("failed to build doc cache: %w", err)
			}

			cachePath, err := getDocCacheFilePath()
			if err != nil {
				return fmt.Errorf("failed to get doc cache file path: %w", err)
			}

			fetchedAt := time.Now()
			if cache, err := loadDocCache(); err == nil && cache != nil {
				fetchedAt = cache.FetchedAt
			}

			output := CacheBuildOutput{
				Path:      cachePath,
				Entries:   len(docNames),
				FetchedAt: fetchedAt.Format(time.RFC3339),
				InMemory:  useInMemoryDocCache,
			}
			if err := outputJSON(output); err != nil {
				return fmt.Errorf("failed to output JSON: %w", err)
			}
			return nil
		default:
			return fmt.Errorf("usage: pgo doccache [path|build]")
		}
	}

	if command == "rag" {
		return runRag(args[1:])
	}

	// Check for required arguments for API commands
	if *baseURL == "" {
		return fmt.Errorf("paperless URL is required (use -url flag or PAPERLESS_URL env var)")
	}
	if *token == "" {
		return fmt.Errorf("API token is required (use -token flag or PAPERLESS_TOKEN env var)")
	}

	if command == "apply" {
		if len(args) < 3 {
			return fmt.Errorf("usage: pgo apply docs <id> --tags=<id1>,<id2>")
		}

		resource := args[1]
		if resource != "docs" {
			return fmt.Errorf("unknown resource for apply: %s", resource)
		}

		// Parse ID and flags
		var id int
		var tagsStr string

		// First argument after resource MUST be ID
		if _, err := fmt.Sscanf(args[2], "%d", &id); err != nil {
			return fmt.Errorf("invalid ID format: %s", args[2])
		}

		// Loop through remaining args to find flags
		for _, arg := range args[3:] {
			if strings.HasPrefix(arg, "--tags=") {
				tagsStr = strings.TrimPrefix(arg, "--tags=")
			}
		}

		if tagsStr == "" {
			return fmt.Errorf("missing required flag: --tags")
		}

		// Parse tags
		var tagIDs []int
		if tagsStr != "" {
			parts := strings.Split(tagsStr, ",")
			for _, p := range parts {
				tid, err := strconv.Atoi(strings.TrimSpace(p))
				if err != nil {
					return fmt.Errorf("invalid tag ID: %s", p)
				}
				tagIDs = append(tagIDs, tid)
			}
		}

		// Create client
		client := paperless.NewClient(*baseURL, *token)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Call update
		update := &paperless.DocumentUpdate{
			Tags: &tagIDs,
		}

		doc, err := client.UpdateDocument(ctx, id, update)
		if err != nil {
			return fmt.Errorf("failed to update document: %w", err)
		}

		tagNames, err := getTagNamesWithCache(ctx, client, *forceRefresh, DefaultCacheTTL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not fetch tags for name resolution: %v\n", err)
			tagNames = make(map[int]string)
		}

		output := convertDocToOutput(doc, tagNames)
		if err := outputJSON(output); err != nil {
			return fmt.Errorf("failed to output JSON: %w", err)
		}
		return nil
	}

	if command == "add" {
		if len(args) < 2 {
			return fmt.Errorf("usage: pgo add <resource> [args]\nAvailable resources:\n  tag \"<name>\" - Create a new tag")
		}

		resource := args[1]
		if resource != "tag" {
			return fmt.Errorf("unknown resource for add: %s", resource)
		}

		if len(args) < 3 {
			return fmt.Errorf("usage: pgo add tag \"<name>\"")
		}
		tagName := args[2]

		// Create client
		client := paperless.NewClient(*baseURL, *token)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Create tag
		tagCreate := &paperless.TagCreate{
			Name: tagName,
		}

		tag, err := client.CreateTag(ctx, tagCreate)
		if err != nil {
			return fmt.Errorf("failed to create tag: %w", err)
		}

		if err := outputJSON(tag); err != nil {
			return fmt.Errorf("failed to output JSON: %w", err)
		}
		return nil
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
				return fmt.Errorf("parse search docs flags: %w", err)
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

			// Convert to output format and display as JSON
			output := convertDocToOutput(doc, tagNames)
			if err := outputJSON(output); err != nil {
				return fmt.Errorf("failed to output JSON: %w", err)
			}
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

			// Convert documents to output format
			results := make([]DocumentWithTagNames, len(docs.Results))
			for i, doc := range docs.Results {
				results[i] = convertDocToOutput(&doc, tagNames)
			}

			// Output as JSON
			output := DocumentListOutput{
				Count:   docs.Count,
				Results: results,
			}
			if err := outputJSON(output); err != nil {
				return fmt.Errorf("failed to output JSON: %w", err)
			}
		}
	} else if resource == "tags" {
		if hasID {
			// Get specific tag
			tag, err := client.GetTag(ctx, id)
			if err != nil {
				return fmt.Errorf("failed to get tag %d: %w", id, err)
			}

			// Output as JSON
			if err := outputJSON(tag); err != nil {
				return fmt.Errorf("failed to output JSON: %w", err)
			}
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

			// Output as JSON
			if err := outputJSON(tags); err != nil {
				return fmt.Errorf("failed to output JSON: %w", err)
			}
		}
	}

	return nil
}

func runRag(args []string) error {
	path, err := exec.LookPath("pgo-rag")
	if err != nil {
		return fmt.Errorf("pgo-rag not found in PATH; build it with: (cd cmd/pgo-rag && go build)")
	}

	cmd := exec.Command(path, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
