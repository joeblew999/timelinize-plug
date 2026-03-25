# Tests

This directory contains all tests for timelinize-plug.

## Structure

```
tests/
├── unit/                              # Unit tests for individual components
│   └── sse_formatter_test.go
├── integration/                       # Integration tests for full workflows
│   ├── oauth_flow_test.go            # Real Google auth flow (guided, requires manual consent)
│   └── SETUP_OAUTH.md                # OAuth setup guide
└── README.md                          # This file
```

## Running Tests

### Unit Tests (Go)

```bash
# Run all unit tests
go test ./tests/unit/...

# Run with verbose output
go test -v ./tests/unit/...

# Run with coverage
go test -cover ./tests/unit/...
```

### Integration Tests (Go, guided)

```bash
# Requires manual consent in the browser (opens OAuth for gedw99@gmail.com)
E2E_REAL=1 GOOGLE_CLIENT_ID=... GOOGLE_CLIENT_SECRET=... \
  go test ./tests/integration -run TestGoogleOAuthFlow -v
```

This will:
1. Build and start `tplug serve` in-process.
2. Prompt you to complete the Google OAuth flow in your browser.
3. Persist the encrypted token in PocketBase and verify it.

**First time setup:** See `tests/integration/SETUP_OAUTH.md` for OAuth configuration.

### Run All Tests

```bash
# Unit tests
go test ./tests/unit/...

# Guided integration test (requires manual consent)
E2E_REAL=1 GOOGLE_CLIENT_ID=... GOOGLE_CLIENT_SECRET=... \
  go test ./tests/integration -run TestGoogleOAuthFlow -v
```

## Test Coverage

Current test coverage:

- **Unit Tests**:
  - ✅ SSE formatter (basic structure tests)
  - ⏳ NATS integration (TODO)
  - ⏳ Progress event emission (TODO)
  - ⏳ Server routes (TODO)

- **Integration Tests**:
  - ✅ Google OAuth flow & token persistence (gedw99@gmail.com)
  - ⏳ Automated import progress simulation (TODO)
  - ⏳ GitHub import (TODO)
  - ⏳ Apple import (TODO)

## Adding New Tests

### Adding Unit Tests

Create a new file in `tests/unit/` with the pattern `*_test.go`:

```go
package unit

import "testing"

func TestMyComponent(t *testing.T) {
    // Your test here
}
```

### Adding Integration Tests

Integration tests live in Go under `tests/integration`. Prefer Go tests that orchestrate the server and call into the CLI or HTTP endpoints using helper functions.

## CI/CD Integration

These tests are designed to run in CI/CD pipelines:

```yaml
# Example GitHub Actions
- name: Run unit tests
  run: go test -v ./tests/unit/...

- name: Run integration tests (requires manual consent)
  run: |
    E2E_REAL=1 GOOGLE_CLIENT_ID=*** GOOGLE_CLIENT_SECRET=*** \
      go test ./tests/integration -run TestGoogleOAuthFlow -v
```

## Test Dependencies

- Go 1.25+
- `nats` CLI (optional, for integration tests)
- `jq` (optional, for JSON parsing in shell tests)
- `curl` (for HTTP endpoint testing)

Install optional dependencies:
```bash
# NATS CLI
go install github.com/nats-io/natscli/nats@latest

# jq (macOS)
brew install jq
```

## Troubleshooting

### Port Already in Use

If tests fail with "address already in use":

```bash
# Kill any running tplug processes
killall tplug

# Kill processes on NATS port
lsof -ti:4222 | xargs kill -9

# Kill processes on HTTP port
lsof -ti:12002 | xargs kill -9
```

### Tests Hanging

Integration tests may hang if server doesn't start properly. Check:

1. NATS port is available (4222)
2. HTTP port is available (12002)
3. Check `server.log` for errors

## Future Test Additions

- [ ] Unit tests for all internal packages
- [ ] Integration test for OAuth flows
- [ ] Load testing for SSE streaming
- [ ] Multi-tenant isolation tests
- [ ] PocketBase integration tests
- [ ] End-to-end UI tests (with Playwright/Cypress)
