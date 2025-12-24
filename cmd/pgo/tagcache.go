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

// TagCache represents cached tag data with timestamp.
// This cache stores only tag ID to name mappings for efficient tag name resolution
// when displaying documents. The 'pgo get tags' command does not use this cache
// as it needs full Tag objects with all fields (Slug, Color, DocumentCount, etc.),
// not just the name mapping.
type TagCache struct {
	Tags      map[int]string `json:"tags"`
	FetchedAt time.Time      `json:"fetched_at"`
}

// DefaultCacheTTL is the default time-to-live for cached tags (12 hours)
const DefaultCacheTTL = 12 * time.Hour

// inMemoryCache holds the in-memory cache state
var inMemoryCache *TagCache

// useInMemoryCache tracks whether to use in-memory cache only
var useInMemoryCache bool

// getCacheDir returns the cache directory path, preferring XDG_CACHE_HOME
func getCacheDir() (string, error) {
	// Try XDG_CACHE_HOME first
	if cacheHome := os.Getenv("XDG_CACHE_HOME"); cacheHome != "" {
		return filepath.Join(cacheHome, "paperless-go"), nil
	}

	// Fall back to ~/.cache
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}

	return filepath.Join(home, ".cache", "paperless-go"), nil
}

// getCacheFilePath returns the full path to the tags cache file
func getCacheFilePath() (string, error) {
	dir, err := getCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "tags.json"), nil
}

// loadTagCache loads cached tags from disk or in-memory cache
// Returns nil if cache doesn't exist or is invalid (non-fatal)
func loadTagCache() (*TagCache, error) {
	// If using in-memory cache, return it directly
	if useInMemoryCache {
		return inMemoryCache, nil
	}

	cachePath, err := getCacheFilePath()
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

	var cache TagCache
	if err := json.Unmarshal(data, &cache); err != nil {
		// Invalid cache file - treat as non-existent
		return nil, nil
	}

	return &cache, nil
}

// saveTagCache saves tags to disk cache or in-memory cache
// Errors are non-fatal - logged but not returned
// If filesystem errors occur, automatically falls back to in-memory cache
func saveTagCache(tags map[int]string) {
	cache := TagCache{
		Tags:      tags,
		FetchedAt: time.Now(),
	}

	// Always update in-memory cache
	inMemoryCache = &cache

	// If using in-memory cache only, skip disk write
	if useInMemoryCache {
		return
	}

	cachePath, err := getCacheFilePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not determine cache path: %v\n", err)
		fmt.Fprintf(os.Stderr, "Info: Using in-memory cache as fallback\n")
		useInMemoryCache = true
		return
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not marshal cache data: %v\n", err)
		return
	}

	// Ensure cache directory exists
	cacheDir := filepath.Dir(cachePath)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not create cache directory: %v\n", err)
		fmt.Fprintf(os.Stderr, "Info: Using in-memory cache as fallback\n")
		useInMemoryCache = true
		return
	}

	// Write cache file
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not write cache file: %v\n", err)
		fmt.Fprintf(os.Stderr, "Info: Using in-memory cache as fallback\n")
		useInMemoryCache = true
		return
	}
}

// isCacheStale checks if cached data has exceeded TTL
func isCacheStale(cache *TagCache, ttl time.Duration) bool {
	if cache == nil {
		return true
	}
	return time.Since(cache.FetchedAt) > ttl
}

// getTagNamesWithCache fetches tags with caching support
func getTagNamesWithCache(ctx context.Context, client *paperless.Client, forceRefresh bool, ttl time.Duration) (map[int]string, error) {
	// Check cache first (unless force refresh)
	if !forceRefresh {
		cache, err := loadTagCache()
		if err != nil {
			// Log error but continue with fresh fetch
			fmt.Fprintf(os.Stderr, "Warning: Could not load cache: %v\n", err)
		} else if !isCacheStale(cache, ttl) {
			// Cache is fresh - use it
			return cache.Tags, nil
		}
	}

	// Cache miss or stale - fetch from remote
	tagNames := make(map[int]string)

	// Fetch all pages of tags
	opts := &paperless.ListOptions{PageSize: 100} // Large page size to minimize requests
	for {
		tags, err := client.ListTags(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tags: %w", err)
		}

		// Add tags from this page
		for _, tag := range tags.Results {
			tagNames[tag.ID] = tag.Name
		}

		// Check if there are more pages
		if tags.Next == nil || *tags.Next == "" {
			break
		}

		// For simplicity, just increase page number (this assumes consistent ordering)
		if opts.Page == 0 {
			opts.Page = 1
		}
		opts.Page++
	}

	// Update cache (non-fatal on error)
	saveTagCache(tagNames)

	return tagNames, nil
}
