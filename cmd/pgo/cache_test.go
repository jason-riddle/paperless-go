package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetCacheDir_Shared(t *testing.T) {
	t.Run("uses XDG_CACHE_HOME when set", func(t *testing.T) {
		// Save original env
		orig := os.Getenv("XDG_CACHE_HOME")
		defer func() {
			if orig != "" {
				_ = os.Setenv("XDG_CACHE_HOME", orig)
			} else {
				_ = os.Unsetenv("XDG_CACHE_HOME")
			}
		}()

		testPath := "/tmp/test-cache-shared"
		_ = os.Setenv("XDG_CACHE_HOME", testPath)

		cacheDir, err := getCacheDir()
		if err != nil {
			t.Fatalf("getCacheDir failed: %v", err)
		}

		expected := filepath.Join(testPath, "paperless-go")
		if cacheDir != expected {
			t.Errorf("cacheDir = %v, want %v", cacheDir, expected)
		}
	})

	t.Run("falls back to ~/.cache when XDG_CACHE_HOME not set", func(t *testing.T) {
		// Save original env
		orig := os.Getenv("XDG_CACHE_HOME")
		defer func() {
			if orig != "" {
				_ = os.Setenv("XDG_CACHE_HOME", orig)
			} else {
				_ = os.Unsetenv("XDG_CACHE_HOME")
			}
		}()

		_ = os.Unsetenv("XDG_CACHE_HOME")

		cacheDir, err := getCacheDir()
		if err != nil {
			t.Fatalf("getCacheDir failed: %v", err)
		}

		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".cache", "paperless-go")
		if cacheDir != expected {
			t.Errorf("cacheDir = %v, want %v", cacheDir, expected)
		}
	})
}

func TestDefaultCacheTTL_Shared(t *testing.T) {
	// Verify default TTL is 12 hours
	if DefaultCacheTTL != 12*time.Hour {
		t.Errorf("DefaultCacheTTL = %v, want %v", DefaultCacheTTL, 12*time.Hour)
	}
}
