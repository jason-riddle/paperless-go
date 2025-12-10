.PHONY: test integration-test integration-setup integration-teardown lint fmt vet help

# Default target
.DEFAULT_GOAL := help

## test: Run unit tests
test:
	go test -v -race -cover ./...

## integration-setup: Start Paperless-ngx for integration testing
integration-setup:
	docker-compose up -d
	@echo "Waiting for Paperless-ngx to be ready..."
	@./scripts/wait-for-paperless.sh

## integration-test: Run integration tests (requires running Paperless instance)
integration-test:
	go test -v -tags=integration ./...

## integration-test-full: Setup, run integration tests, and teardown
integration-test-full: integration-setup integration-test integration-teardown

## integration-teardown: Stop and remove Paperless-ngx containers
integration-teardown:
	docker-compose down -v

## lint: Run all linters
lint: vet fmt
	@echo "All linting passed!"

## vet: Run go vet
vet:
	go vet ./...

## fmt: Check code formatting
fmt:
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "The following files are not properly formatted:"; \
		gofmt -l .; \
		echo "Please run 'gofmt -w .' to format your code."; \
		exit 1; \
	fi

## fmt-fix: Fix code formatting
fmt-fix:
	gofmt -w .

## coverage: Generate test coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## clean: Clean up generated files
clean:
	rm -f coverage.out coverage.html
	go clean -testcache

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
