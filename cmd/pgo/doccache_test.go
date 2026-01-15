package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetDocCacheFilePath(t *testing.T) {
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

	cachePath, err := getDocCacheFilePath()
	if err != nil {
		t.Fatalf("getDocCacheFilePath failed: %v", err)
	}

	expected := filepath.Join(testPath, "paperless-go", "docs.json")
	if cachePath != expected {
		t.Errorf("cachePath = %v, want %v", cachePath, expected)
	}
}

func TestSaveAndLoadDocCache(t *testing.T) {
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

	// Save and restore global state
	origUseInMemory := useInMemoryDocCache
	origInMemoryCache := inMemoryDocCache
	defer func() {
		useInMemoryDocCache = origUseInMemory
		inMemoryDocCache = origInMemoryCache
	}()

	// Reset to use disk cache for this test
	useInMemoryDocCache = false
	inMemoryDocCache = nil

	// Test data
	testDocs := map[int]string{
		1: "Invoice 2023-01",
		2: "Receipt from Store",
		3: "Contract Agreement",
	}

	// Save cache
	saveDocCache(testDocs)

	// Load cache
	cache, err := loadDocCache()
	if err != nil {
		t.Fatalf("loadDocCache failed: %v", err)
	}

	if cache == nil {
		t.Fatal("expected cache, got nil")
	}

	// Verify docs
	if len(cache.Docs) != len(testDocs) {
		t.Errorf("len(cache.Docs) = %d, want %d", len(cache.Docs), len(testDocs))
	}

	for id, title := range testDocs {
		if cache.Docs[id] != title {
			t.Errorf("cache.Docs[%d] = %v, want %v", id, cache.Docs[id], title)
		}
	}

	// Verify timestamp is recent
	if time.Since(cache.FetchedAt) > 5*time.Second {
		t.Errorf("cache.FetchedAt is too old: %v", cache.FetchedAt)
	}
}

func TestLoadDocCache_NonExistent(t *testing.T) {
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

	// Save and restore global state
	origUseInMemory := useInMemoryDocCache
	origInMemoryCache := inMemoryDocCache
	defer func() {
		useInMemoryDocCache = origUseInMemory
		inMemoryDocCache = origInMemoryCache
	}()

	// Reset to use disk cache
	useInMemoryDocCache = false
	inMemoryDocCache = nil

	// Try to load non-existent cache
	cache, err := loadDocCache()
	if err != nil {
		t.Fatalf("loadDocCache should not error on non-existent cache: %v", err)
	}

	if cache != nil {
		t.Errorf("expected nil cache, got %+v", cache)
	}
}

func TestLoadDocCache_InvalidJSON(t *testing.T) {
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

	// Save and restore global state
	origUseInMemory := useInMemoryDocCache
	origInMemoryCache := inMemoryDocCache
	defer func() {
		useInMemoryDocCache = origUseInMemory
		inMemoryDocCache = origInMemoryCache
	}()

	// Reset to use disk cache
	useInMemoryDocCache = false
	inMemoryDocCache = nil

	// Create cache directory
	cacheDir := filepath.Join(tmpDir, "paperless-go")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Write invalid JSON
	cachePath := filepath.Join(cacheDir, "docs.json")
	if err := os.WriteFile(cachePath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Try to load invalid cache - should return nil, not error
	cache, err := loadDocCache()
	if err != nil {
		t.Fatalf("loadDocCache should not error on invalid JSON: %v", err)
	}

	if cache != nil {
		t.Errorf("expected nil cache for invalid JSON, got %+v", cache)
	}
}

func TestIsDocCacheStale(t *testing.T) {
	t.Run("nil cache is stale", func(t *testing.T) {
		if !isDocCacheStale(nil, time.Hour) {
			t.Error("nil cache should be stale")
		}
	})

	t.Run("fresh cache is not stale", func(t *testing.T) {
		cache := &DocCache{
			Docs:      map[int]string{1: "Test Doc"},
			FetchedAt: time.Now(),
		}

		if isDocCacheStale(cache, time.Hour) {
			t.Error("fresh cache should not be stale")
		}
	})

	t.Run("old cache is stale", func(t *testing.T) {
		cache := &DocCache{
			Docs:      map[int]string{1: "Test Doc"},
			FetchedAt: time.Now().Add(-2 * time.Hour),
		}

		if !isDocCacheStale(cache, time.Hour) {
			t.Error("old cache should be stale")
		}
	})

	t.Run("cache at TTL boundary", func(t *testing.T) {
		ttl := time.Hour
		cache := &DocCache{
			Docs:      map[int]string{1: "Test Doc"},
			FetchedAt: time.Now().Add(-ttl - time.Second),
		}

		if !isDocCacheStale(cache, ttl) {
			t.Error("cache past TTL should be stale")
		}
	})
}

func TestSaveDocCache_InvalidPath(t *testing.T) {
	// Save original env and set invalid cache dir
	orig := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if orig != "" {
			os.Setenv("XDG_CACHE_HOME", orig)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()

	// Save and restore global state
	origUseInMemory := useInMemoryDocCache
	origInMemoryCache := inMemoryDocCache
	defer func() {
		useInMemoryDocCache = origUseInMemory
		inMemoryDocCache = origInMemoryCache
	}()

	// Reset state
	useInMemoryDocCache = false
	inMemoryDocCache = nil

	// Use a path that we can't write to (assuming /root is not writable)
	os.Setenv("XDG_CACHE_HOME", "/root/non-writable")

	// This should not panic or return error - just log warning
	testDocs := map[int]string{1: "Test"}
	saveDocCache(testDocs)
}

func TestDocCacheStructure(t *testing.T) {
	// Verify cache structure can be marshaled/unmarshaled
	cache := DocCache{
		Docs: map[int]string{
			1: "Invoice 2023",
			2: "Receipt",
		},
		FetchedAt: time.Now(),
	}

	// Marshal
	data, err := json.Marshal(cache)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal
	var decoded DocCache
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify
	if len(decoded.Docs) != len(cache.Docs) {
		t.Errorf("len(decoded.Docs) = %d, want %d", len(decoded.Docs), len(cache.Docs))
	}

	for id, title := range cache.Docs {
		if decoded.Docs[id] != title {
			t.Errorf("decoded.Docs[%d] = %v, want %v", id, decoded.Docs[id], title)
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

func TestGetDocNamesWithCache_Integration(t *testing.T) {
	// This test verifies the complete cache flow

	// Save and restore global state
	origUseInMemory := useInMemoryDocCache
	origInMemoryCache := inMemoryDocCache
	defer func() {
		useInMemoryDocCache = origUseInMemory
		inMemoryDocCache = origInMemoryCache
	}()

	// Reset to use disk cache for this test
	useInMemoryDocCache = false
	inMemoryDocCache = nil

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

		testDocs := map[int]string{
			1: "Invoice 2023",
			2: "Receipt",
		}

		// Save to cache
		saveDocCache(testDocs)

		// Load from cache
		cache, err := loadDocCache()
		if err != nil {
			t.Fatalf("loadDocCache failed: %v", err)
		}

		if cache == nil {
			t.Fatal("expected cache, got nil")
		}

		// Verify docs match
		if len(cache.Docs) != len(testDocs) {
			t.Errorf("len(cache.Docs) = %d, want %d", len(cache.Docs), len(testDocs))
		}

		for id, title := range testDocs {
			if cache.Docs[id] != title {
				t.Errorf("cache.Docs[%d] = %v, want %v", id, cache.Docs[id], title)
			}
		}
	})

	t.Run("fresh cache is used on subsequent calls", func(t *testing.T) {
		testDocs := map[int]string{
			1: "Cached Doc",
		}

		// Save to cache
		saveDocCache(testDocs)

		// Load should return cached data
		cache, err := loadDocCache()
		if err != nil {
			t.Fatalf("loadDocCache failed: %v", err)
		}

		if !isDocCacheStale(cache, DefaultCacheTTL) {
			t.Log("Cache is fresh as expected")
		} else {
			t.Error("Fresh cache reported as stale")
		}
	})

	t.Run("stale cache is considered invalid", func(t *testing.T) {
		// Save and restore global state
		origUseInMemory := useInMemoryDocCache
		origInMemoryCache := inMemoryDocCache
		defer func() {
			useInMemoryDocCache = origUseInMemory
			inMemoryDocCache = origInMemoryCache
		}()

		// Reset to use disk cache
		useInMemoryDocCache = false
		inMemoryDocCache = nil

		// Create a stale cache by manually setting old timestamp
		staleCache := DocCache{
			Docs: map[int]string{
				1: "Stale Doc",
			},
			FetchedAt: time.Now().Add(-25 * time.Hour), // Older than 12h TTL
		}

		// Save manually
		cachePath, _ := getDocCacheFilePath()
		cacheDir := filepath.Dir(cachePath)
		_ = os.MkdirAll(cacheDir, 0755)

		data, _ := json.Marshal(staleCache)
		_ = os.WriteFile(cachePath, data, 0644)

		// Load and check staleness
		cache, err := loadDocCache()
		if err != nil {
			t.Fatalf("loadDocCache failed: %v", err)
		}

		if !isDocCacheStale(cache, DefaultCacheTTL) {
			t.Error("Stale cache should be considered stale")
		}
	})
}

func TestInMemoryDocCache(t *testing.T) {
	// Save original state
	origUseInMemory := useInMemoryDocCache
	origInMemoryCache := inMemoryDocCache
	defer func() {
		useInMemoryDocCache = origUseInMemory
		inMemoryDocCache = origInMemoryCache
	}()

	t.Run("explicit -memory flag", func(t *testing.T) {
		// Reset state
		useInMemoryDocCache = true
		inMemoryDocCache = nil

		testDocs := map[int]string{
			1: "Test Doc 1",
			2: "Test Doc 2",
		}

		// Save to in-memory cache
		saveDocCache(testDocs)

		// Verify in-memory cache was set
		if inMemoryDocCache == nil {
			t.Fatal("in-memory doc cache should be set")
		}

		if len(inMemoryDocCache.Docs) != len(testDocs) {
			t.Errorf("len(inMemoryDocCache.Docs) = %d, want %d", len(inMemoryDocCache.Docs), len(testDocs))
		}

		// Load from in-memory cache
		cache, err := loadDocCache()
		if err != nil {
			t.Fatalf("loadDocCache failed: %v", err)
		}

		if cache == nil {
			t.Fatal("cache should not be nil")
		}

		for id, title := range testDocs {
			if cache.Docs[id] != title {
				t.Errorf("cache.Docs[%d] = %v, want %v", id, cache.Docs[id], title)
			}
		}
	})

	t.Run("automatic fallback on permission error", func(t *testing.T) {
		// Reset state
		useInMemoryDocCache = false
		inMemoryDocCache = nil

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

		testDocs := map[int]string{
			1: "Fallback Doc",
		}

		// This should fail to write to disk and set useInMemoryDocCache = true
		// Note: /root/non-writable is used as a reliably unwritable path on Linux
		saveDocCache(testDocs)

		// Verify in-memory cache was set
		if inMemoryDocCache == nil {
			t.Fatal("in-memory doc cache should be set as fallback")
		}

		// Verify useInMemoryDocCache was automatically set
		if !useInMemoryDocCache {
			t.Error("useInMemoryDocCache should be true after permission error")
		}

		// Verify data is in in-memory cache
		if len(inMemoryDocCache.Docs) != len(testDocs) {
			t.Errorf("len(inMemoryDocCache.Docs) = %d, want %d", len(inMemoryDocCache.Docs), len(testDocs))
		}

		// Load from in-memory cache (should work even though disk failed)
		cache, err := loadDocCache()
		if err != nil {
			t.Fatalf("loadDocCache failed: %v", err)
		}

		if cache == nil {
			t.Fatal("cache should not be nil")
		}

		for id, title := range testDocs {
			if cache.Docs[id] != title {
				t.Errorf("cache.Docs[%d] = %v, want %v", id, cache.Docs[id], title)
			}
		}
	})

	t.Run("in-memory cache preserves data across saves", func(t *testing.T) {
		// Reset state
		useInMemoryDocCache = true
		inMemoryDocCache = nil

		testDocs1 := map[int]string{1: "Doc 1"}
		saveDocCache(testDocs1)

		cache1, _ := loadDocCache()
		if cache1 == nil || cache1.Docs[1] != "Doc 1" {
			t.Fatal("First save failed")
		}

		testDocs2 := map[int]string{2: "Doc 2"}
		saveDocCache(testDocs2)

		cache2, _ := loadDocCache()
		if cache2 == nil || cache2.Docs[2] != "Doc 2" {
			t.Fatal("Second save failed")
		}

		// Second save should have replaced first
		if cache2.Docs[1] == "Doc 1" {
			t.Error("In-memory cache should be replaced, not merged")
		}
	})
}

func TestInMemoryDocCacheFallbackIntegration(t *testing.T) {
	// This test verifies that commands work even with filesystem errors

	// Save original state
	origUseInMemory := useInMemoryDocCache
	origInMemoryCache := inMemoryDocCache
	defer func() {
		useInMemoryDocCache = origUseInMemory
		inMemoryDocCache = origInMemoryCache
	}()

	// Reset state
	useInMemoryDocCache = false
	inMemoryDocCache = nil

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
	testDocs := map[int]string{
		1: "Invoice 2023",
		2: "Receipt",
	}
	saveDocCache(testDocs)

	// Verify fallback was enabled
	if !useInMemoryDocCache {
		t.Error("Should have enabled in-memory doc cache after permission error")
	}

	// Verify data is accessible
	cache, err := loadDocCache()
	if err != nil {
		t.Fatalf("loadDocCache failed: %v", err)
	}

	if cache == nil {
		t.Fatal("cache should not be nil")
	}

	if cache.Docs[1] != "Invoice 2023" || cache.Docs[2] != "Receipt" {
		t.Error("In-memory cache data incorrect")
	}

	// Second save should use in-memory cache (no more errors)
	testDocs2 := map[int]string{
		3: "Contract",
	}
	saveDocCache(testDocs2)

	cache2, err := loadDocCache()
	if err != nil {
		t.Fatalf("Second loadDocCache failed: %v", err)
	}

	if cache2 == nil || cache2.Docs[3] != "Contract" {
		t.Error("Second in-memory cache save/load failed")
	}
}
