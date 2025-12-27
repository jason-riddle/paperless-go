package indexer

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	paperless "github.com/jason-riddle/paperless-go"
	"github.com/jason-riddle/paperless-go/rag/internal/embedding"
	"github.com/jason-riddle/paperless-go/rag/internal/storage"
)

// Embedder generates vector embeddings for text.
type Embedder interface {
	GenerateEmbedding(text string) ([]float32, error)
}

// PaperlessClient provides the Paperless API calls needed for indexing.
type PaperlessClient interface {
	ListDocuments(ctx context.Context, opts *paperless.ListOptions) (*paperless.DocumentList, error)
	ListTags(ctx context.Context, opts *paperless.ListOptions) (*paperless.TagList, error)
}

// BuildOptions configures the indexing process.
type BuildOptions struct {
	PageSize int
}

// BuildSummary describes the result of an index build.
type BuildSummary struct {
	DocumentsFetched    int `json:"documents_fetched"`
	DocumentsIndexed    int `json:"documents_indexed"`
	DocumentsSkipped    int `json:"documents_skipped"`
	EmbeddingsGenerated int `json:"embeddings_generated"`
}

// SearchSummary includes the results and timing for a search.
type SearchSummary struct {
	Results      []storage.SearchResult `json:"results"`
	QueryTimeMs  int64                  `json:"query_time_ms"`
	TotalResults int                    `json:"total_results"`
}

// BuildIndex fetches documents from Paperless and updates the local SQLite index.
func BuildIndex(ctx context.Context, client PaperlessClient, db *storage.DB, embedder Embedder, opts BuildOptions) (BuildSummary, error) {
	var summary BuildSummary

	if client == nil {
		return summary, errors.New("paperless client is required")
	}
	if db == nil {
		return summary, errors.New("storage database is required")
	}
	if embedder == nil {
		return summary, errors.New("embedder is required")
	}

	pageSize := opts.PageSize
	if pageSize <= 0 {
		pageSize = 100
	}

	tagsByID, err := listAllTags(ctx, client, pageSize)
	if err != nil {
		return summary, err
	}

	documents, err := listAllDocuments(ctx, client, pageSize)
	if err != nil {
		return summary, err
	}

	summary.DocumentsFetched = len(documents)

	for _, doc := range documents {
		select {
		case <-ctx.Done():
			return summary, ctx.Err()
		default:
		}

		tags := formatTags(doc.Tags, tagsByID)
		text := buildEmbeddingText(doc.Title, tags, doc.Content)
		if text == "" {
			summary.DocumentsSkipped++
			continue
		}

		modified := doc.Modified.Time()
		existing, err := db.GetDocumentByPaperlessID(doc.ID)
		if err != nil {
			return summary, err
		}
		if existing != nil && existing.LastModified.Equal(modified) {
			summary.DocumentsSkipped++
			continue
		}

		var docID int
		if existing == nil {
			newID, err := db.InsertDocument(storage.Document{
				PaperlessID:  doc.ID,
				PaperlessURL: docURL(doc),
				Title:        doc.Title,
				Tags:         tags,
				LastModified: modified,
			})
			if err != nil {
				return summary, err
			}
			docID = int(newID)
		} else {
			docID = existing.ID
			if err := db.UpdateDocument(storage.Document{
				PaperlessID:  doc.ID,
				PaperlessURL: docURL(doc),
				Title:        doc.Title,
				Tags:         tags,
				LastModified: modified,
			}); err != nil {
				return summary, err
			}

			if err := db.DeleteEmbeddingsByDocumentID(docID); err != nil {
				return summary, err
			}
		}

		vector, err := embedder.GenerateEmbedding(text)
		if err != nil {
			return summary, fmt.Errorf("generate embedding for document %d: %w", doc.ID, err)
		}
		if err := db.InsertEmbedding(docID, text, vector); err != nil {
			return summary, err
		}

		summary.DocumentsIndexed++
		summary.EmbeddingsGenerated++
	}

	return summary, nil
}

// SearchIndex runs a similarity search against the local index.
func SearchIndex(ctx context.Context, db *storage.DB, embedder Embedder, query string, limit int, threshold float64) (SearchSummary, error) {
	var summary SearchSummary

	if db == nil {
		return summary, errors.New("storage database is required")
	}
	if embedder == nil {
		return summary, errors.New("embedder is required")
	}
	if strings.TrimSpace(query) == "" {
		return summary, errors.New("query is required")
	}
	if limit <= 0 {
		limit = 10
	}
	if threshold <= 0 {
		threshold = 0.7
	}

	select {
	case <-ctx.Done():
		return summary, ctx.Err()
	default:
	}

	start := time.Now()
	vector, err := embedder.GenerateEmbedding(query)
	if err != nil {
		return summary, fmt.Errorf("generate embedding for query: %w", err)
	}

	results, err := db.SearchSimilar(vector, limit, threshold)
	if err != nil {
		return summary, err
	}

	summary.Results = results
	summary.TotalResults = len(results)
	summary.QueryTimeMs = time.Since(start).Milliseconds()

	return summary, nil
}

func listAllTags(ctx context.Context, client PaperlessClient, pageSize int) (map[int]string, error) {
	page := 1
	tagsByID := make(map[int]string)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		list, err := client.ListTags(ctx, &paperless.ListOptions{Page: page, PageSize: pageSize})
		if err != nil {
			return nil, err
		}

		for _, tag := range list.Results {
			tagsByID[tag.ID] = tag.Name
		}

		if list.Next == nil || len(list.Results) == 0 {
			break
		}
		page++
	}

	return tagsByID, nil
}

func listAllDocuments(ctx context.Context, client PaperlessClient, pageSize int) ([]paperless.Document, error) {
	page := 1
	var documents []paperless.Document

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		list, err := client.ListDocuments(ctx, &paperless.ListOptions{Page: page, PageSize: pageSize})
		if err != nil {
			return nil, err
		}

		documents = append(documents, list.Results...)
		if list.Next == nil || len(list.Results) == 0 {
			break
		}
		page++
	}

	return documents, nil
}

func formatTags(tagIDs []int, tagsByID map[int]string) string {
	if len(tagIDs) == 0 {
		return ""
	}

	parts := make([]string, 0, len(tagIDs))
	for _, id := range tagIDs {
		name := tagsByID[id]
		if name == "" {
			name = fmt.Sprintf("tag-%d", id)
		}
		parts = append(parts, name)
	}

	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

func buildEmbeddingText(title string, tags string, content string) string {
	base := embedding.FormatDocumentText(strings.TrimSpace(title), strings.TrimSpace(tags))
	content = strings.TrimSpace(content)

	if content == "" {
		return base
	}
	if base == "" {
		return content
	}

	return base + "\n\n" + content
}

func docURL(doc paperless.Document) string {
	return fmt.Sprintf("/api/documents/%d/", doc.ID)
}
