package indexer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	paperless "github.com/jason-riddle/paperless-go"
	"github.com/jason-riddle/paperless-go/cmd/pgo-rag/internal/embedding"
	"github.com/jason-riddle/paperless-go/cmd/pgo-rag/internal/storage"
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
	MaxDocs  int
	TagName  string
}

// BuildSummary describes the result of an index build.
type BuildSummary struct {
	DocumentsFetched    int `json:"documents_fetched"`
	DocumentsIndexed    int `json:"documents_indexed"`
	DocumentsSkipped    int `json:"documents_skipped"`
	DocumentsFailed     int `json:"documents_failed"`
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

	state, err := db.GetIndexState()
	if err != nil {
		return summary, err
	}
	if state.LastPaperlessID > 0 {
		slog.Info("Resuming index build",
			"last_paperless_id", state.LastPaperlessID,
			"last_updated_at", state.UpdatedAt,
		)
	}

	page := 1
	for {
		if opts.MaxDocs > 0 && summary.DocumentsFetched >= opts.MaxDocs {
			break
		}

		select {
		case <-ctx.Done():
			return summary, ctx.Err()
		default:
		}

		effectivePageSize := pageSize
		if opts.MaxDocs > 0 {
			remaining := opts.MaxDocs - summary.DocumentsFetched
			if remaining <= 0 {
				break
			}
			if remaining < effectivePageSize {
				effectivePageSize = remaining
			}
		}

		list, err := client.ListDocuments(ctx, &paperless.ListOptions{
			Page:     page,
			PageSize: effectivePageSize,
			Ordering: "id",
		})
		if err != nil {
			return summary, err
		}
		if len(list.Results) == 0 {
			break
		}

		for _, doc := range list.Results {
			if opts.MaxDocs > 0 && summary.DocumentsFetched >= opts.MaxDocs {
				break
			}

			summary.DocumentsFetched++

			if err := processDocument(ctx, db, embedder, tagsByID, opts, doc, &summary); err != nil {
				return summary, err
			}

			if err := db.UpdateIndexState(doc.ID); err != nil {
				return summary, err
			}
		}

		if list.Next == nil {
			break
		}
		page++
	}

	return summary, nil
}

func processDocument(ctx context.Context, db *storage.DB, embedder Embedder, tagsByID map[int]string, opts BuildOptions, doc paperless.Document, summary *BuildSummary) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if opts.TagName != "" && !documentHasTag(doc, tagsByID, opts.TagName) {
		slog.Info("Skipping document without tag",
			"paperless_id", doc.ID,
			"required_tag", opts.TagName,
		)
		summary.DocumentsSkipped++
		return nil
	}

	tags := formatTags(doc.Tags, tagsByID)
	text := buildEmbeddingText(doc.Title, tags, doc.Content)
	if text == "" {
		slog.Info("Skipping document with empty embedding text",
			"paperless_id", doc.ID,
			"tags", tags,
		)
		summary.DocumentsSkipped++
		return nil
	}

	modified := doc.Modified.Time()
	existing, err := db.GetDocumentByPaperlessID(doc.ID)
	if err != nil {
		return err
	}
	if existing != nil && existing.LastModified.Equal(modified) && !existing.EmbeddedAt.IsZero() {
		slog.Info("Skipping unchanged document",
			"paperless_id", doc.ID,
			"last_modified", modified,
		)
		summary.DocumentsSkipped++
		return nil
	}

	vector, err := embedder.GenerateEmbedding(text)
	if err != nil {
		return recordDocumentFailure(db, summary, doc.ID, fmt.Errorf("generate embedding for document %d: %w", doc.ID, err))
	}

	slog.Info("Embedded document",
		"paperless_id", doc.ID,
		"tags", tags,
		"embedding_text_len", len(text),
	)

	if err := db.UpsertDocumentWithEmbedding(storage.Document{
		PaperlessID:  doc.ID,
		PaperlessURL: docURL(doc),
		Title:        doc.Title,
		Tags:         tags,
		LastModified: modified,
	}, text, vector); err != nil {
		return recordDocumentFailure(db, summary, doc.ID, fmt.Errorf("update index for document %d: %w", doc.ID, err))
	}

	if err := db.ClearIndexFailure(doc.ID); err != nil {
		return err
	}

	summary.DocumentsIndexed++
	summary.EmbeddingsGenerated++
	return nil
}

func recordDocumentFailure(db *storage.DB, summary *BuildSummary, paperlessID int, err error) error {
	slog.Error("Failed to index document",
		"paperless_id", paperlessID,
		"error", err,
	)
	if recordErr := db.RecordIndexFailure(paperlessID, err); recordErr != nil {
		return recordErr
	}
	summary.DocumentsFailed++
	return nil
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

func documentHasTag(doc paperless.Document, tagsByID map[int]string, tagName string) bool {
	for _, id := range doc.Tags {
		if tagsByID[id] == tagName {
			return true
		}
	}
	return false
}
