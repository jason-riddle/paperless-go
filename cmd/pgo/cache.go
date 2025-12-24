package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DefaultCacheTTL is the default time-to-live for cached data (12 hours)
const DefaultCacheTTL = 12 * time.Hour

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
