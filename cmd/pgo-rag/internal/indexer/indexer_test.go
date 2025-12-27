package indexer

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	paperless "github.com/jason-riddle/paperless-go"
	"github.com/jason-riddle/paperless-go/cmd/pgo-rag/internal/storage"
)

type fakeEmbedder struct {
	vectors map[string][]float32
}

func (f fakeEmbedder) GenerateEmbedding(text string) ([]float32, error) {
	vector, ok := f.vectors[text]
	if !ok {
		return []float32{0, 0, 1}, nil
	}
	return vector, nil
}

type fakePaperless struct {
	documents []paperless.Document
	tags      []paperless.Tag
}

func (f fakePaperless) ListDocuments(_ context.Context, opts *paperless.ListOptions) (*paperless.DocumentList, error) {
	page, pageSize := normalizePage(opts, len(f.documents))
	start := (page - 1) * pageSize
	if start >= len(f.documents) {
		return &paperless.DocumentList{Count: len(f.documents)}, nil
	}

	end := start + pageSize
	if end > len(f.documents) {
		end = len(f.documents)
	}

	list := &paperless.DocumentList{Count: len(f.documents), Results: f.documents[start:end]}
	if end < len(f.documents) {
		next := "next"
		list.Next = &next
	}
	return list, nil
}

func (f fakePaperless) ListTags(_ context.Context, opts *paperless.ListOptions) (*paperless.TagList, error) {
	page, pageSize := normalizePage(opts, len(f.tags))
	start := (page - 1) * pageSize
	if start >= len(f.tags) {
		return &paperless.TagList{Count: len(f.tags)}, nil
	}

	end := start + pageSize
	if end > len(f.tags) {
		end = len(f.tags)
	}

	list := &paperless.TagList{Count: len(f.tags), Results: f.tags[start:end]}
	if end < len(f.tags) {
		next := "next"
		list.Next = &next
	}
	return list, nil
}

func normalizePage(opts *paperless.ListOptions, total int) (int, int) {
	page := 1
	pageSize := total
	if opts != nil {
		if opts.Page > 0 {
			page = opts.Page
		}
		if opts.PageSize > 0 {
			pageSize = opts.PageSize
		}
	}
	if pageSize == 0 {
		pageSize = 1
	}
	return page, pageSize
}

func TestBuildIndexAndSearch(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "index.db")
	db, err := storage.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	modified := time.Now().UTC().Truncate(time.Second)
	docs := []paperless.Document{
		{
			ID:       101,
			Title:    "Alpha Report",
			Content:  "alpha content",
			Tags:     []int{1},
			Modified: paperless.Date(modified),
		},
		{
			ID:       202,
			Title:    "Beta Memo",
			Content:  "beta content",
			Tags:     []int{2},
			Modified: paperless.Date(modified),
		},
	}

	tags := []paperless.Tag{
		{ID: 1, Name: "finance"},
		{ID: 2, Name: "notes"},
	}

	client := fakePaperless{documents: docs, tags: tags}

	alphaText := buildEmbeddingText("Alpha Report", "finance", "alpha content")
	betaText := buildEmbeddingText("Beta Memo", "notes", "beta content")

	embedder := fakeEmbedder{
		vectors: map[string][]float32{
			alphaText:     {1, 0, 0},
			betaText:      {0, 1, 0},
			"alpha query": {1, 0, 0},
		},
	}

	summary, err := BuildIndex(ctx, client, db, embedder, BuildOptions{PageSize: 1})
	if err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}

	if summary.DocumentsFetched != 2 {
		t.Fatalf("expected 2 documents fetched, got %d", summary.DocumentsFetched)
	}
	if summary.DocumentsIndexed != 2 {
		t.Fatalf("expected 2 documents indexed, got %d", summary.DocumentsIndexed)
	}
	if summary.EmbeddingsGenerated != 2 {
		t.Fatalf("expected 2 embeddings, got %d", summary.EmbeddingsGenerated)
	}

	count, err := db.CountDocuments()
	if err != nil {
		t.Fatalf("CountDocuments failed: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 documents in DB, got %d", count)
	}

	searchSummary, err := SearchIndex(ctx, db, embedder, "alpha query", 5, 0.5)
	if err != nil {
		t.Fatalf("SearchIndex failed: %v", err)
	}

	if searchSummary.TotalResults != 1 {
		t.Fatalf("expected 1 search result, got %d", searchSummary.TotalResults)
	}
	if searchSummary.Results[0].Title != "Alpha Report" {
		t.Fatalf("expected Alpha Report result, got %s", searchSummary.Results[0].Title)
	}
}

func TestBuildIndexSkipsUnchanged(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "index.db")
	db, err := storage.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	modified := time.Now().UTC().Truncate(time.Second)
	client := fakePaperless{
		documents: []paperless.Document{{
			ID:       303,
			Title:    "Gamma",
			Content:  "gamma content",
			Tags:     []int{1},
			Modified: paperless.Date(modified),
		}},
		tags: []paperless.Tag{{ID: 1, Name: "archive"}},
	}

	text := buildEmbeddingText("Gamma", "archive", "gamma content")
	embedder := fakeEmbedder{vectors: map[string][]float32{text: {0.3, 0.3, 0.3}}}

	first, err := BuildIndex(ctx, client, db, embedder, BuildOptions{})
	if err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}
	if first.DocumentsIndexed != 1 {
		t.Fatalf("expected 1 document indexed, got %d", first.DocumentsIndexed)
	}

	second, err := BuildIndex(ctx, client, db, embedder, BuildOptions{})
	if err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}
	if second.DocumentsSkipped != 1 {
		t.Fatalf("expected 1 document skipped, got %d", second.DocumentsSkipped)
	}
}

func TestBuildIndexMaxDocs(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "index.db")
	db, err := storage.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	modified := time.Now().UTC().Truncate(time.Second)
	client := fakePaperless{
		documents: []paperless.Document{
			{ID: 1, Title: "Doc1", Content: "content1", Modified: paperless.Date(modified)},
			{ID: 2, Title: "Doc2", Content: "content2", Modified: paperless.Date(modified)},
			{ID: 3, Title: "Doc3", Content: "content3", Modified: paperless.Date(modified)},
		},
	}

	embedder := fakeEmbedder{vectors: map[string][]float32{
		buildEmbeddingText("Doc1", "", "content1"): {1, 0, 0},
		buildEmbeddingText("Doc2", "", "content2"): {0, 1, 0},
		buildEmbeddingText("Doc3", "", "content3"): {0, 0, 1},
	}}

	summary, err := BuildIndex(ctx, client, db, embedder, BuildOptions{MaxDocs: 2})
	if err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}
	if summary.DocumentsIndexed != 2 {
		t.Fatalf("expected 2 documents indexed, got %d", summary.DocumentsIndexed)
	}
	if summary.DocumentsFetched != 2 {
		t.Fatalf("expected 2 documents fetched, got %d", summary.DocumentsFetched)
	}
}

func TestBuildIndexTagFilter(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "index.db")
	db, err := storage.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	modified := time.Now().UTC().Truncate(time.Second)
	client := fakePaperless{
		documents: []paperless.Document{
			{ID: 1, Title: "Doc1", Content: "content1", Tags: []int{1}, Modified: paperless.Date(modified)},
			{ID: 2, Title: "Doc2", Content: "content2", Tags: []int{2}, Modified: paperless.Date(modified)},
		},
		tags: []paperless.Tag{{ID: 1, Name: "FOO"}, {ID: 2, Name: "BAR"}},
	}

	embedder := fakeEmbedder{vectors: map[string][]float32{
		buildEmbeddingText("Doc1", "FOO", "content1"): {1, 0, 0},
		buildEmbeddingText("Doc2", "BAR", "content2"): {0, 1, 0},
	}}

	summary, err := BuildIndex(ctx, client, db, embedder, BuildOptions{TagName: "FOO"})
	if err != nil {
		t.Fatalf("BuildIndex failed: %v", err)
	}
	if summary.DocumentsIndexed != 1 {
		t.Fatalf("expected 1 document indexed, got %d", summary.DocumentsIndexed)
	}
}

func TestHelpers(t *testing.T) {
	if result := formatTags([]int{2, 1}, map[int]string{1: "alpha", 2: "beta"}); result != "alpha, beta" {
		t.Fatalf("unexpected tags: %s", result)
	}
	if result := formatTags([]int{3}, map[int]string{}); result != "tag-3" {
		t.Fatalf("unexpected missing tag format: %s", result)
	}

	text := buildEmbeddingText("Title", "tag", "content")
	if text != "Title. Tags: tag\n\ncontent" {
		t.Fatalf("unexpected embedding text: %s", text)
	}

	if docURL(paperless.Document{ID: 42}) != "/api/documents/42/" {
		t.Fatalf("unexpected doc URL")
	}
}

func TestValidation(t *testing.T) {
	_, err := BuildIndex(context.Background(), nil, nil, nil, BuildOptions{})
	if err == nil {
		t.Fatalf("expected error for nil inputs")
	}

	db, err := storage.NewDB(filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	_, err = SearchIndex(context.Background(), db, nil, "query", 1, 0.5)
	if err == nil {
		t.Fatalf("expected error for nil embedder")
	}

	_, err = SearchIndex(context.Background(), db, fakeEmbedder{}, "", 1, 0.5)
	if err == nil {
		t.Fatalf("expected error for empty query")
	}
}

func TestPaginationHelpers(t *testing.T) {
	client := fakePaperless{
		documents: []paperless.Document{{ID: 1}, {ID: 2}, {ID: 3}},
		tags:      []paperless.Tag{{ID: 1, Name: "one"}, {ID: 2, Name: "two"}},
	}

	docs, err := listAllDocuments(context.Background(), client, 2, 0)
	if err != nil {
		t.Fatalf("listAllDocuments failed: %v", err)
	}
	if len(docs) != 3 {
		t.Fatalf("expected 3 documents, got %d", len(docs))
	}

	limited, err := listAllDocuments(context.Background(), client, 2, 2)
	if err != nil {
		t.Fatalf("listAllDocuments failed: %v", err)
	}
	if len(limited) != 2 {
		t.Fatalf("expected 2 documents, got %d", len(limited))
	}

	tags, err := listAllTags(context.Background(), client, 1)
	if err != nil {
		t.Fatalf("listAllTags failed: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	if tags[2] != "two" {
		t.Fatalf("expected tag 2 name 'two', got %s", tags[2])
	}
}
