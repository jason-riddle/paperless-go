# Copilot Instructions for paperless-go

> **Note**: Please read [AGENTS.md](../AGENTS.md) for comprehensive documentation about this project, including project structure, key principles, and development workflow.

## Code Quality Standards

### Always Run Before Committing

1. **Format Code**: Run `go fmt ./...` on all Go code before committing
2. **Run Tests**: Run `go test -v -race ./...` to ensure all tests pass
3. **Run Linters**: Run `make lint` to check for code quality issues

### Development Workflow

When making changes to the codebase:

1. Format your code:
   ```bash
   go fmt ./...
   ```

2. Run unit tests with race detector:
   ```bash
   go test -v -race ./...
   ```

3. Run all quality checks:
   ```bash
   make lint
   ```

4. For integration tests:
   ```bash
   make integration-test-full
   ```

### Go Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting (enforced by CI)
- Keep functions small and focused
- Write tests for all public APIs
- Use meaningful variable and function names
- Document all exported functions, types, and constants

### Testing Requirements

- All new code must have unit tests
- Maintain or improve code coverage (currently 91.8%)
- Use `httptest.Server` for mocking HTTP calls in unit tests
- Integration tests should use the `//go:build integration` tag
- Tests must pass with race detector enabled

### Error Handling

- Use structured error types from `errors.go`
- Wrap errors with context using `fmt.Errorf` with `%w`
- Set operation names on API errors for better debugging
- Use `IsNotFound()` helper for checking 404 errors

### API Design

- All API methods must accept `context.Context` as the first parameter
- Use functional options pattern for configuration
- Return pointers for struct types to allow nil values
- Use generics where appropriate for type safety

### Dependencies

- **Zero external dependencies policy** - only use Go standard library
- Do not add external packages without explicit approval
- Prefer stdlib solutions over third-party libraries

### Documentation

- Update README.md for any API changes
- Add godoc comments for all exported types and functions
- Include examples in example_test.go for new features
- Keep documentation clear and concise
