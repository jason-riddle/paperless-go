package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jason-riddle/paperless-go"
)

// DocCache represents cached document data with timestamp.
// This cache stores only document ID to title mappings for efficient document name resolution.
type DocCache struct {
	Docs      map[int]string `json:"docs"`
	FetchedAt time.Time      `json:"fetched_at"`
}

// inMemoryDocCache holds the in-memory doc cache state
// Note: These global variables are safe for CLI usage as each invocation
// runs in a separate process. They are not safe for concurrent use in
// long-running server applications.
var inMemoryDocCache *DocCache

// useInMemoryDocCache tracks whether to use in-memory doc cache only
var useInMemoryDocCache bool

// getDocCacheFilePath returns the full path to the docs cache file
func getDocCacheFilePath() (string, error) {
	dir, err := getCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "docs.json"), nil
}

// loadDocCache loads cached docs from disk or in-memory cache
// Returns nil if cache doesn't exist or is invalid (non-fatal)
func loadDocCache() (*DocCache, error) {
	// If using in-memory cache, return it directly
	if useInMemoryDocCache {
		return inMemoryDocCache, nil
	}

	cachePath, err := getDocCacheFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Cache doesn't exist - not an error
			return nil, nil
		}
		return nil, fmt.Errorf("read cache file: %w", err)
	}

	var cache DocCache
	if err := json.Unmarshal(data, &cache); err != nil {
		// Invalid cache file - treat as non-existent
		return nil, nil
	}

	return &cache, nil
}

// saveDocCache saves docs to disk cache or in-memory cache
// Errors are non-fatal - logged but not returned
// If filesystem errors occur, automatically falls back to in-memory cache
func saveDocCache(docs map[int]string) {
	cache := DocCache{
		Docs:      docs,
		FetchedAt: time.Now(),
	}

	// If using in-memory cache only, skip disk write
	if useInMemoryDocCache {
		// Update in-memory cache
		inMemoryDocCache = &cache
		return
	}

	cachePath, err := getDocCacheFilePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not determine doc cache path: %v\n", err)
		fmt.Fprintf(os.Stderr, "Info: Using in-memory doc cache as fallback\n")
		useInMemoryDocCache = true
		inMemoryDocCache = &cache
		return
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not marshal doc cache data: %v\n", err)
		return
	}

	// Ensure cache directory exists
	cacheDir := filepath.Dir(cachePath)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not create doc cache directory: %v\n", err)
		fmt.Fprintf(os.Stderr, "Info: Using in-memory doc cache as fallback\n")
		useInMemoryDocCache = true
		inMemoryDocCache = &cache
		return
	}

	// Write cache file
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not write doc cache file: %v\n", err)
		fmt.Fprintf(os.Stderr, "Info: Using in-memory doc cache as fallback\n")
		useInMemoryDocCache = true
		inMemoryDocCache = &cache
		return
	}

	// Successfully wrote to disk, also update in-memory cache as a hot cache
	inMemoryDocCache = &cache
}

// isDocCacheStale checks if cached doc data has exceeded TTL
func isDocCacheStale(cache *DocCache, ttl time.Duration) bool {
	if cache == nil {
		return true
	}
	return time.Since(cache.FetchedAt) > ttl
}

// getDocNamesWithCache fetches document names with caching support
func getDocNamesWithCache(ctx context.Context, client *paperless.Client, forceRefresh bool, ttl time.Duration) (map[int]string, error) {
	// Check cache first (unless force refresh)
	if !forceRefresh {
		cache, err := loadDocCache()
		if err != nil {
			// Log error but continue with fresh fetch
			fmt.Fprintf(os.Stderr, "Warning: Could not load doc cache: %v\n", err)
		} else if !isDocCacheStale(cache, ttl) {
			// Cache is fresh - use it
			return cache.Docs, nil
		}
	}

	// Cache miss or stale - fetch from remote
	docNames := make(map[int]string)

	// Fetch all pages of documents
	opts := &paperless.ListOptions{PageSize: 100} // Large page size to minimize requests
	for {
		docs, err := client.ListDocuments(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch documents: %w", err)
		}

		// Add docs from this page
		for _, doc := range docs.Results {
			docNames[doc.ID] = doc.Title
		}

		// Check if there are more pages
		if docs.Next == nil || *docs.Next == "" {
			break
		}

		// For simplicity, just increase page number (this assumes consistent ordering)
		if opts.Page == 0 {
			opts.Page = 1
		}
		opts.Page++
	}

	// Update cache (non-fatal on error)
	saveDocCache(docNames)

	return docNames, nil
}
