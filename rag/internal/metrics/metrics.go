package metrics

import (
	"expvar"
)

var (
	// SearchRequestsTotal counts total search requests
	SearchRequestsTotal = expvar.NewInt("search_requests_total")

	// SyncOperationsTotal counts total sync operations (attempts)
	SyncOperationsTotal = expvar.NewInt("sync_operations_total")

	// SyncOperationsSuccessful counts successful sync operations
	SyncOperationsSuccessful = expvar.NewInt("sync_operations_successful")

	// SyncOperationsFailed counts failed sync operations
	SyncOperationsFailed = expvar.NewInt("sync_operations_failed")

	// EmbeddingsGeneratedTotal counts successful embedding generations
	EmbeddingsGeneratedTotal = expvar.NewInt("embeddings_generated_total")

	// EmbeddingsFailedTotal counts failed embedding generations
	EmbeddingsFailedTotal = expvar.NewInt("embeddings_failed_total")

	// DocumentsCount tracks the current number of documents
	DocumentsCount = expvar.NewInt("documents_count")

	// APIErrorsTotal counts total API errors
	APIErrorsTotal = expvar.NewInt("api_errors_total")
)
