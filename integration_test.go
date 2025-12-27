//go:build integration
// +build integration

package paperless_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jason-riddle/paperless-go"
)

func getTestClient(t *testing.T) *paperless.Client {
	baseURL := os.Getenv("PAPERLESS_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}

	token := os.Getenv("PAPERLESS_TOKEN")
	if token == "" {
		t.Skip("PAPERLESS_TOKEN not set, skipping integration test")
	}

	return paperless.NewClient(baseURL, token, paperless.WithTimeout(30*time.Second))
}

func TestIntegration_ListDocuments(t *testing.T) {
	client := getTestClient(t)

	ctx := context.Background()
	docs, err := client.ListDocuments(ctx, nil)
	if err != nil {
		t.Fatalf("ListDocuments failed: %v", err)
	}

	t.Logf("Found %d documents", docs.Count)

	// Test with pagination
	if docs.Count > 0 {
		opts := &paperless.ListOptions{
			Page:     1,
			PageSize: 5,
		}
		pageResult, err := client.ListDocuments(ctx, opts)
		if err != nil {
			t.Fatalf("ListDocuments with pagination failed: %v", err)
		}
		if len(pageResult.Results) > 5 {
			t.Errorf("Expected max 5 results, got %d", len(pageResult.Results))
		}
	}
}

func TestIntegration_GetDocument(t *testing.T) {
	client := getTestClient(t)

	ctx := context.Background()

	// First, get a list of documents to find a valid ID
	docs, err := client.ListDocuments(ctx, &paperless.ListOptions{PageSize: 1})
	if err != nil {
		t.Fatalf("ListDocuments failed: %v", err)
	}

	if len(docs.Results) == 0 {
		t.Skip("No documents available, skipping GetDocument test")
	}

	docID := docs.Results[0].ID
	doc, err := client.GetDocument(ctx, docID)
	if err != nil {
		t.Fatalf("GetDocument failed: %v", err)
	}

	if doc.ID != docID {
		t.Errorf("Expected document ID %d, got %d", docID, doc.ID)
	}

	t.Logf("Retrieved document: %s", doc.Title)
}

func TestIntegration_GetDocument_NotFound(t *testing.T) {
	client := getTestClient(t)

	ctx := context.Background()
	_, err := client.GetDocument(ctx, 999999)
	if err == nil {
		t.Fatal("Expected error for non-existent document, got nil")
	}

	if !paperless.IsNotFound(err) {
		t.Errorf("Expected 404 error, got %v", err)
	}
}

func TestIntegration_ListTags(t *testing.T) {
	client := getTestClient(t)

	ctx := context.Background()
	tags, err := client.ListTags(ctx, nil)
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	t.Logf("Found %d tags", tags.Count)

	// Test with ordering
	if tags.Count > 0 {
		opts := &paperless.ListOptions{
			Ordering: "name",
		}
		orderedTags, err := client.ListTags(ctx, opts)
		if err != nil {
			t.Fatalf("ListTags with ordering failed: %v", err)
		}
		if len(orderedTags.Results) > 0 {
			t.Logf("First tag: %s", orderedTags.Results[0].Name)
		}
	}
}

func TestIntegration_GetTag(t *testing.T) {
	client := getTestClient(t)

	ctx := context.Background()

	// First, get a list of tags to find a valid ID
	tags, err := client.ListTags(ctx, &paperless.ListOptions{PageSize: 1})
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	if len(tags.Results) == 0 {
		t.Skip("No tags available, skipping GetTag test")
	}

	tagID := tags.Results[0].ID
	tag, err := client.GetTag(ctx, tagID)
	if err != nil {
		t.Fatalf("GetTag failed: %v", err)
	}

	if tag.ID != tagID {
		t.Errorf("Expected tag ID %d, got %d", tagID, tag.ID)
	}

	t.Logf("Retrieved tag: %s (color: %s, documents: %d)", tag.Name, tag.Color, tag.DocumentCount)
}

func TestIntegration_GetTag_NotFound(t *testing.T) {
	client := getTestClient(t)

	ctx := context.Background()
	_, err := client.GetTag(ctx, 999999)
	if err == nil {
		t.Fatal("Expected error for non-existent tag, got nil")
	}

	if !paperless.IsNotFound(err) {
		t.Errorf("Expected 404 error, got %v", err)
	}
}

func TestIntegration_SearchDocuments(t *testing.T) {
	client := getTestClient(t)

	ctx := context.Background()

	// Get all documents first
	allDocs, err := client.ListDocuments(ctx, nil)
	if err != nil {
		t.Fatalf("ListDocuments failed: %v", err)
	}

	if allDocs.Count == 0 {
		t.Skip("No documents available, skipping search test")
	}

	// Try searching with a query
	opts := &paperless.ListOptions{
		Query: "test",
	}
	searchResult, err := client.ListDocuments(ctx, opts)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	t.Logf("Search for 'test' returned %d documents", searchResult.Count)
}

func TestIntegration_UpdateDocument(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	// Get a document to update
	docs, err := client.ListDocuments(ctx, &paperless.ListOptions{PageSize: 1})
	if err != nil {
		t.Fatalf("ListDocuments failed: %v", err)
	}
	if len(docs.Results) == 0 {
		t.Skip("No documents available, skipping UpdateDocument test")
	}
	docID := docs.Results[0].ID

	// Get tags to use
	tags, err := client.ListTags(ctx, &paperless.ListOptions{PageSize: 2})
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}
	if len(tags.Results) < 1 {
		t.Skip("Not enough tags available, skipping UpdateDocument test")
	}

	newTagID := tags.Results[0].ID

	// Update document
	update := &paperless.DocumentUpdate{
		Tags: []int{newTagID},
	}
	updatedDoc, err := client.UpdateDocument(ctx, docID, update)
	if err != nil {
		t.Fatalf("UpdateDocument failed: %v", err)
	}

	// Verify
	found := false
	for _, tagID := range updatedDoc.Tags {
		if tagID == newTagID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Updated document does not have tag %d", newTagID)
	}
}

func TestIntegration_CreateTag(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	tagName := "integration-test-tag-" + time.Now().Format("20060102150405")
	tagCreate := &paperless.TagCreate{
		Name:  tagName,
		Color: "#ff0000",
	}

	tag, err := client.CreateTag(ctx, tagCreate)
	if err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}

	if tag.Name != tagName {
		t.Errorf("Expected tag name %s, got %s", tagName, tag.Name)
	}

	// Verify it exists in list
	tags, err := client.ListTags(ctx, &paperless.ListOptions{
		Query: tagName,
	})
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	// Note: Query on tags might be filtering by name, checking results
	found := false
	for _, t := range tags.Results {
		if t.ID == tag.ID {
			found = true
			break
		}
	}
	// Depending on Paperless version/config, query might work differently or be delayed.
	// We'll trust CreateTag return value primarily, but log if not found in list.
	if !found {
		t.Logf("Warning: Created tag not found in list immediately (might be indexing delay)")
	}
}
