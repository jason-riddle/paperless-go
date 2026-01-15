# Security Guidelines

This document outlines security best practices for using and contributing to paperless-go.

## API Token Handling

### For Users

**DO:**
- Store API tokens in environment variables (`PAPERLESS_TOKEN`)
- Use secure credential management tools (e.g., password managers, secrets vaults)
- Rotate tokens periodically
- Use different tokens for different environments (dev, staging, production)

**DON'T:**
- Hardcode tokens in source code
- Commit tokens to version control
- Share tokens via insecure channels (email, chat, etc.)
- Log tokens to stdout/stderr or log files

### For Developers

The paperless-go library is designed with security in mind:

1. **No Credential Logging**: The library never logs API tokens, passwords, or other credentials
2. **Context-Based Auth**: All API methods accept `context.Context` for proper timeout and cancellation handling
3. **Secure Error Messages**: Error messages do not expose sensitive information

Example of secure token usage:

```go
// Good: Token from environment variable
token := os.Getenv("PAPERLESS_TOKEN")
client := paperless.NewClient("https://paperless.example.com", token)

// Bad: Hardcoded token
// client := paperless.NewClient("https://paperless.example.com", "abc123...")
```

## Testing Credentials

The repository includes test credentials in `docker-compose.yml` and test scripts. These are:

- **ONLY for local testing and CI/CD**
- **NEVER for production use**
- Clearly marked with security warnings

These credentials are:
- `POSTGRES_PASSWORD: paperless` (PostgreSQL database)
- `PAPERLESS_ADMIN_PASSWORD: admin` (Paperless admin user)
- `PAPERLESS_SECRET_KEY: test-secret-key-for-integration-tests` (Django secret key)

**Note on Token Output**: The `wait-for-paperless.sh` script outputs API tokens to stderr (not stdout) to reduce the risk of accidental capture in piped output or logs. This is intentional for integration testing purposes, but users should be aware that tokens will still be visible in terminal output.

**Security Best Practices for Test Scripts**:
- Clear your terminal history after running the script: `history -c` (bash/zsh)
- Use a private/secure terminal session when running in shared environments
- Consider using a secrets management tool even for local development
- Rotate test tokens periodically

## Logging Sensitive Information

### CLI Tools (pgo, pgo-rag)

The CLI tools follow these logging practices:

1. **Never log full document content** - Only document IDs and metadata lengths
2. **Never log API tokens** - Tokens are only used for API authentication, never printed
3. **Sanitize outputs** - Document titles and content are not included in structured logs
4. **Debug mode caution** - Even in debug mode, sensitive content is not logged

### Example: Secure Logging

```go
// Good: Log metadata only
slog.Info("Processing document", "document_id", id, "content_length", len(content))

// Bad: Don't log sensitive content
// slog.Info("Processing document", "document_id", id, "content", content)
```

### Tag Names in Logs

Tag names (e.g., "finance", "confidential", "project-x") are logged in the RAG indexer for debugging purposes. While tags are generally non-sensitive metadata, they may occasionally contain sensitive information:

- If your tags contain sensitive project names or classifications, consider using less verbose logging levels
- Production deployments should use `info` level or higher to minimize tag exposure
- Debug logs should only be enabled in secure, controlled environments

To adjust logging verbosity in pgo-rag:
```bash
# Use info level to reduce log output
pgo-rag build -log-level info ...

# Use error level for minimal logging
pgo-rag build -log-level error ...
```

## Environment Variables

The following environment variables may contain sensitive information:

- `PAPERLESS_TOKEN` - API authentication token
- `PAPERLESS_URL` - Paperless instance URL (may expose internal network topology)
- `PGO_RAG_EMBEDDINGS_KEY` - Embeddings API key for RAG functionality

Best practices:
- Use `.env` files for local development (add to `.gitignore`)
- Use secrets management tools in production (e.g., AWS Secrets Manager, HashiCorp Vault)
- Never commit `.env` files to version control

## Reporting Security Issues

If you discover a security vulnerability in paperless-go:

1. **DO NOT** create a public GitHub issue
2. Report the issue privately using one of these methods:
   - Use GitHub's [Security Advisories](https://github.com/jason-riddle/paperless-go/security/advisories) feature
   - Email the maintainer directly (see repository owner's GitHub profile for contact information)
3. Include in your report:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if available)

We take security seriously and will respond promptly to all reports. We appreciate responsible disclosure and will work with you to address any issues.

**Response Timeline**:
- Initial response: Within 48 hours
- Status update: Within 7 days
- Fix timeline: Depends on severity, but critical issues will be prioritized

## Security Checklist for Contributors

Before submitting a PR:

- [ ] No hardcoded credentials in code
- [ ] No sensitive information in log messages
- [ ] Environment variables used for configuration
- [ ] Test credentials clearly marked and isolated
- [ ] Documentation updated for any security-relevant changes
- [ ] No credentials in test files (use environment variables or mocks)

## Third-Party Dependencies

paperless-go follows a **zero external dependencies** policy for the core library:

- Reduces attack surface
- Simplifies security audits
- Makes vulnerability management easier

The only dependencies are:
- Go standard library (core library)
- Limited third-party packages for CLI tools (pgo-rag sub-module only)

## Additional Resources

- [OWASP API Security Top 10](https://owasp.org/www-project-api-security/)
- [Go Security Best Practices](https://golang.org/doc/security/best-practices)
- [Paperless-ngx Security](https://docs.paperless-ngx.com/configuration/#security)
