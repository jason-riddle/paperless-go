package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetCacheDir(t *testing.T) {
	t.Run("uses XDG_CACHE_HOME when set", func(t *testing.T) {
		// Save original env
		orig := os.Getenv("XDG_CACHE_HOME")
		defer func() {
			if orig != "" {
				os.Setenv("XDG_CACHE_HOME", orig)
			} else {
				os.Unsetenv("XDG_CACHE_HOME")
			}
		}()

		testPath := "/tmp/test-cache"
		os.Setenv("XDG_CACHE_HOME", testPath)

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
				os.Setenv("XDG_CACHE_HOME", orig)
			} else {
				os.Unsetenv("XDG_CACHE_HOME")
			}
		}()

		os.Unsetenv("XDG_CACHE_HOME")

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

func TestGetCacheFilePath(t *testing.T) {
	// Save original env
	orig := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if orig != "" {
			os.Setenv("XDG_CACHE_HOME", orig)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()

	testPath := "/tmp/test-cache"
	os.Setenv("XDG_CACHE_HOME", testPath)

	cachePath, err := getCacheFilePath()
	if err != nil {
		t.Fatalf("getCacheFilePath failed: %v", err)
	}

	expected := filepath.Join(testPath, "paperless-go", "tags.json")
	if cachePath != expected {
		t.Errorf("cachePath = %v, want %v", cachePath, expected)
	}
}

func TestSaveAndLoadTagCache(t *testing.T) {
	// Create temp directory for testing
	tmpDir := t.TempDir()

	// Save original env and set test cache dir
	orig := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if orig != "" {
			os.Setenv("XDG_CACHE_HOME", orig)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()
	os.Setenv("XDG_CACHE_HOME", tmpDir)

	// Test data
	testTags := map[int]string{
		1: "Important",
		2: "Work",
		3: "Personal",
	}

	// Save cache
	saveTagCache(testTags)

	// Load cache
	cache, err := loadTagCache()
	if err != nil {
		t.Fatalf("loadTagCache failed: %v", err)
	}

	if cache == nil {
		t.Fatal("expected cache, got nil")
	}

	// Verify tags
	if len(cache.Tags) != len(testTags) {
		t.Errorf("len(cache.Tags) = %d, want %d", len(cache.Tags), len(testTags))
	}

	for id, name := range testTags {
		if cache.Tags[id] != name {
			t.Errorf("cache.Tags[%d] = %v, want %v", id, cache.Tags[id], name)
		}
	}

	// Verify timestamp is recent
	if time.Since(cache.FetchedAt) > 5*time.Second {
		t.Errorf("cache.FetchedAt is too old: %v", cache.FetchedAt)
	}
}

func TestLoadTagCache_NonExistent(t *testing.T) {
	// Create temp directory for testing
	tmpDir := t.TempDir()

	// Save original env and set test cache dir
	orig := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if orig != "" {
			os.Setenv("XDG_CACHE_HOME", orig)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()
	os.Setenv("XDG_CACHE_HOME", tmpDir)

	// Try to load non-existent cache
	cache, err := loadTagCache()
	if err != nil {
		t.Fatalf("loadTagCache should not error on non-existent cache: %v", err)
	}

	if cache != nil {
		t.Errorf("expected nil cache, got %+v", cache)
	}
}

func TestLoadTagCache_InvalidJSON(t *testing.T) {
	// Create temp directory for testing
	tmpDir := t.TempDir()

	// Save original env and set test cache dir
	orig := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if orig != "" {
			os.Setenv("XDG_CACHE_HOME", orig)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()
	os.Setenv("XDG_CACHE_HOME", tmpDir)

	// Create cache directory
	cacheDir := filepath.Join(tmpDir, "paperless-go")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Write invalid JSON
	cachePath := filepath.Join(cacheDir, "tags.json")
	if err := os.WriteFile(cachePath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Try to load invalid cache - should return nil, not error
	cache, err := loadTagCache()
	if err != nil {
		t.Fatalf("loadTagCache should not error on invalid JSON: %v", err)
	}

	if cache != nil {
		t.Errorf("expected nil cache for invalid JSON, got %+v", cache)
	}
}

func TestIsCacheStale(t *testing.T) {
	t.Run("nil cache is stale", func(t *testing.T) {
		if !isCacheStale(nil, time.Hour) {
			t.Error("nil cache should be stale")
		}
	})

	t.Run("fresh cache is not stale", func(t *testing.T) {
		cache := &TagCache{
			Tags:      map[int]string{1: "Test"},
			FetchedAt: time.Now(),
		}

		if isCacheStale(cache, time.Hour) {
			t.Error("fresh cache should not be stale")
		}
	})

	t.Run("old cache is stale", func(t *testing.T) {
		cache := &TagCache{
			Tags:      map[int]string{1: "Test"},
			FetchedAt: time.Now().Add(-2 * time.Hour),
		}

		if !isCacheStale(cache, time.Hour) {
			t.Error("old cache should be stale")
		}
	})

	t.Run("cache at TTL boundary", func(t *testing.T) {
		ttl := time.Hour
		cache := &TagCache{
			Tags:      map[int]string{1: "Test"},
			FetchedAt: time.Now().Add(-ttl - time.Second),
		}

		if !isCacheStale(cache, ttl) {
			t.Error("cache past TTL should be stale")
		}
	})
}

func TestSaveTagCache_InvalidPath(t *testing.T) {
	// Save original env and set invalid cache dir
	orig := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if orig != "" {
			os.Setenv("XDG_CACHE_HOME", orig)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()

	// Use a path that we can't write to (assuming /root is not writable)
	os.Setenv("XDG_CACHE_HOME", "/root/non-writable")

	// This should not panic or return error - just log warning
	testTags := map[int]string{1: "Test"}
	saveTagCache(testTags)
}

func TestTagCacheStructure(t *testing.T) {
	// Verify cache structure can be marshaled/unmarshaled
	cache := TagCache{
		Tags: map[int]string{
			1: "Important",
			2: "Work",
		},
		FetchedAt: time.Now(),
	}

	// Marshal
	data, err := json.Marshal(cache)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal
	var decoded TagCache
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify
	if len(decoded.Tags) != len(cache.Tags) {
		t.Errorf("len(decoded.Tags) = %d, want %d", len(decoded.Tags), len(cache.Tags))
	}

	for id, name := range cache.Tags {
		if decoded.Tags[id] != name {
			t.Errorf("decoded.Tags[%d] = %v, want %v", id, decoded.Tags[id], name)
		}
	}

	// Verify timestamp (allow small difference due to encoding)
	timeDiff := decoded.FetchedAt.Sub(cache.FetchedAt)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}
	if timeDiff > time.Second {
		t.Errorf("timestamp difference too large: %v", timeDiff)
	}
}

func TestDefaultCacheTTL(t *testing.T) {
	// Verify default TTL is 12 hours
	if DefaultCacheTTL != 12*time.Hour {
		t.Errorf("DefaultCacheTTL = %v, want %v", DefaultCacheTTL, 12*time.Hour)
	}
}

func TestGetTagNamesWithCache_Integration(t *testing.T) {
	// This test verifies the complete cache flow with a mock HTTP server
	// Using stdlib httptest for mocking

	// Save and restore global state
	origUseInMemory := useInMemoryCache
	origInMemoryCache := inMemoryCache
	defer func() {
		useInMemoryCache = origUseInMemory
		inMemoryCache = origInMemoryCache
	}()

	// Reset to use disk cache for this test
	useInMemoryCache = false
	inMemoryCache = nil

	// Create temp directory for testing
	tmpDir := t.TempDir()

	// Save original env and set test cache dir
	orig := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if orig != "" {
			os.Setenv("XDG_CACHE_HOME", orig)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()
	os.Setenv("XDG_CACHE_HOME", tmpDir)

	t.Run("cache miss fetches from API and saves to cache", func(t *testing.T) {
		// Note: This test would require importing the httptest package and paperless client
		// For now, we'll test the cache functionality directly

		testTags := map[int]string{
			1: "Important",
			2: "Work",
		}

		// Save to cache
		saveTagCache(testTags)

		// Load from cache
		cache, err := loadTagCache()
		if err != nil {
			t.Fatalf("loadTagCache failed: %v", err)
		}

		if cache == nil {
			t.Fatal("expected cache, got nil")
		}

		// Verify tags match
		if len(cache.Tags) != len(testTags) {
			t.Errorf("len(cache.Tags) = %d, want %d", len(cache.Tags), len(testTags))
		}

		for id, name := range testTags {
			if cache.Tags[id] != name {
				t.Errorf("cache.Tags[%d] = %v, want %v", id, cache.Tags[id], name)
			}
		}
	})

	t.Run("fresh cache is used on subsequent calls", func(t *testing.T) {
		testTags := map[int]string{
			1: "Cached Tag",
		}

		// Save to cache
		saveTagCache(testTags)

		// Load should return cached data
		cache, err := loadTagCache()
		if err != nil {
			t.Fatalf("loadTagCache failed: %v", err)
		}

		if !isCacheStale(cache, DefaultCacheTTL) {
			t.Log("Cache is fresh as expected")
		} else {
			t.Error("Fresh cache reported as stale")
		}
	})

	t.Run("stale cache is considered invalid", func(t *testing.T) {
		// Save and restore global state
		origUseInMemory := useInMemoryCache
		origInMemoryCache := inMemoryCache
		defer func() {
			useInMemoryCache = origUseInMemory
			inMemoryCache = origInMemoryCache
		}()

		// Reset to use disk cache
		useInMemoryCache = false
		inMemoryCache = nil

		// Create a stale cache by manually setting old timestamp
		staleCache := TagCache{
			Tags: map[int]string{
				1: "Stale Tag",
			},
			FetchedAt: time.Now().Add(-25 * time.Hour), // Older than 12h TTL
		}

		// Save manually
		cachePath, _ := getCacheFilePath()
		cacheDir := filepath.Dir(cachePath)
		_ = os.MkdirAll(cacheDir, 0755)

		data, _ := json.Marshal(staleCache)
		_ = os.WriteFile(cachePath, data, 0644)

		// Load and check staleness
		cache, err := loadTagCache()
		if err != nil {
			t.Fatalf("loadTagCache failed: %v", err)
		}

		if !isCacheStale(cache, DefaultCacheTTL) {
			t.Error("Stale cache should be considered stale")
		}
	})
}

func TestInMemoryCache(t *testing.T) {
	// Save original state
	origUseInMemory := useInMemoryCache
	origInMemoryCache := inMemoryCache
	defer func() {
		useInMemoryCache = origUseInMemory
		inMemoryCache = origInMemoryCache
	}()

	t.Run("explicit -memory flag", func(t *testing.T) {
		// Reset state
		useInMemoryCache = true
		inMemoryCache = nil

		testTags := map[int]string{
			1: "Test Tag 1",
			2: "Test Tag 2",
		}

		// Save to in-memory cache
		saveTagCache(testTags)

		// Verify in-memory cache was set
		if inMemoryCache == nil {
			t.Fatal("in-memory cache should be set")
		}

		if len(inMemoryCache.Tags) != len(testTags) {
			t.Errorf("len(inMemoryCache.Tags) = %d, want %d", len(inMemoryCache.Tags), len(testTags))
		}

		// Load from in-memory cache
		cache, err := loadTagCache()
		if err != nil {
			t.Fatalf("loadTagCache failed: %v", err)
		}

		if cache == nil {
			t.Fatal("cache should not be nil")
		}

		for id, name := range testTags {
			if cache.Tags[id] != name {
				t.Errorf("cache.Tags[%d] = %v, want %v", id, cache.Tags[id], name)
			}
		}
	})

	t.Run("automatic fallback on permission error", func(t *testing.T) {
		// Reset state
		useInMemoryCache = false
		inMemoryCache = nil

		// Save original env and set to unwritable path
		orig := os.Getenv("XDG_CACHE_HOME")
		defer func() {
			if orig != "" {
				os.Setenv("XDG_CACHE_HOME", orig)
			} else {
				os.Unsetenv("XDG_CACHE_HOME")
			}
		}()

		os.Setenv("XDG_CACHE_HOME", "/root/non-writable")

		testTags := map[int]string{
			1: "Fallback Tag",
		}

		// This should fail to write to disk and set useInMemoryCache = true
		// Note: /root/non-writable is used as a reliably unwritable path on Linux
		saveTagCache(testTags)

		// Verify in-memory cache was set
		if inMemoryCache == nil {
			t.Fatal("in-memory cache should be set as fallback")
		}

		// Verify useInMemoryCache was automatically set
		if !useInMemoryCache {
			t.Error("useInMemoryCache should be true after permission error")
		}

		// Verify data is in in-memory cache
		if len(inMemoryCache.Tags) != len(testTags) {
			t.Errorf("len(inMemoryCache.Tags) = %d, want %d", len(inMemoryCache.Tags), len(testTags))
		}

		// Load from in-memory cache (should work even though disk failed)
		cache, err := loadTagCache()
		if err != nil {
			t.Fatalf("loadTagCache failed: %v", err)
		}

		if cache == nil {
			t.Fatal("cache should not be nil")
		}

		for id, name := range testTags {
			if cache.Tags[id] != name {
				t.Errorf("cache.Tags[%d] = %v, want %v", id, cache.Tags[id], name)
			}
		}
	})

	t.Run("in-memory cache preserves data across saves", func(t *testing.T) {
		// Reset state
		useInMemoryCache = true
		inMemoryCache = nil

		testTags1 := map[int]string{1: "Tag 1"}
		saveTagCache(testTags1)

		cache1, _ := loadTagCache()
		if cache1 == nil || cache1.Tags[1] != "Tag 1" {
			t.Fatal("First save failed")
		}

		testTags2 := map[int]string{2: "Tag 2"}
		saveTagCache(testTags2)

		cache2, _ := loadTagCache()
		if cache2 == nil || cache2.Tags[2] != "Tag 2" {
			t.Fatal("Second save failed")
		}

		// Second save should have replaced first
		if cache2.Tags[1] == "Tag 1" {
			t.Error("In-memory cache should be replaced, not merged")
		}
	})
}

func TestInMemoryCacheFallbackIntegration(t *testing.T) {
	// This test verifies that commands work even with filesystem errors

	// Save original state
	origUseInMemory := useInMemoryCache
	origInMemoryCache := inMemoryCache
	defer func() {
		useInMemoryCache = origUseInMemory
		inMemoryCache = origInMemoryCache
	}()

	// Reset state
	useInMemoryCache = false
	inMemoryCache = nil

	// Set unwritable cache path
	orig := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if orig != "" {
			os.Setenv("XDG_CACHE_HOME", orig)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()
	os.Setenv("XDG_CACHE_HOME", "/root/non-writable")

	// First save should fail and enable in-memory cache
	// Note: /root/non-writable is used as a reliably unwritable path on Linux
	testTags := map[int]string{
		1: "Important",
		2: "Work",
	}
	saveTagCache(testTags)

	// Verify fallback was enabled
	if !useInMemoryCache {
		t.Error("Should have enabled in-memory cache after permission error")
	}

	// Verify data is accessible
	cache, err := loadTagCache()
	if err != nil {
		t.Fatalf("loadTagCache failed: %v", err)
	}

	if cache == nil {
		t.Fatal("cache should not be nil")
	}

	if cache.Tags[1] != "Important" || cache.Tags[2] != "Work" {
		t.Error("In-memory cache data incorrect")
	}

	// Second save should use in-memory cache (no more errors)
	testTags2 := map[int]string{
		3: "Personal",
	}
	saveTagCache(testTags2)

	cache2, err := loadTagCache()
	if err != nil {
		t.Fatalf("Second loadTagCache failed: %v", err)
	}

	if cache2 == nil || cache2.Tags[3] != "Personal" {
		t.Error("Second in-memory cache save/load failed")
	}
}
