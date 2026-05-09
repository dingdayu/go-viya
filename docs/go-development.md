# Go Development

This guide captures implementation constraints and common Go techniques for the
`go-viya` codebase.

## Core Constraints

- Use package name `viya` in the root package.
- Keep request paths context-aware. Public methods that perform I/O should accept
  and propagate `context.Context`.
- Return errors with operation context and HTTP status where possible.
- Keep authentication failures wrapped or mapped consistently so callers can
  check `ErrViyaAuthFailed`.
- Do not introduce global mutable state except where the existing default-client
  pattern requires it.
- Preserve Resty and OpenTelemetry behavior when changing shared client,
  transport, or token code.

## Common Go Patterns

- Use table-driven tests for request construction, response decoding, and error
  mapping.
- Use `httptest.Server` to model SAS Viya HTTP behavior without live credentials.
- Prefer small helper functions over broad abstractions when the behavior is
  specific to one service area.
- Use functional options for configurable client or token-provider behavior.
- Keep examples buildable and aligned with the exported API.
- Prefer typed structs for stable JSON payloads and reserve `map[string]any` for
  dynamic configuration payloads.

## Testing Expectations

- Add or update tests for user-visible behavior, error handling, request
  construction, response decoding, and concurrency-sensitive code.
- Endpoint helper tests should assert method, path, authentication behavior,
  request body, status and error handling, and response decoding.
- Avoid tests that require live SAS Viya credentials unless they are explicitly
  marked as integration tests.
- Treat failing tests as harness feedback. Fix the implementation or the test
  contract; do not delete tests merely to make the suite pass.

## Observability and Secrets

- Preserve OpenTelemetry spans around networked operations.
- Include enough span status information to debug failed HTTP calls.
- Never capture bearer tokens, passwords, tenant secrets, customer data, or full
  sensitive payloads in logs, spans, fixtures, examples, or tests.

## Standard Verification

Run before marking code changes complete:

```bash
gofmt -w .
go mod tidy
go vet ./...
go test ./...
```

For authentication, HTTP middleware, default-client, token cache, transport, or
shared client changes, also run:

```bash
go test -race ./...
```
