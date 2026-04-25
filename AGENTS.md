# AGENTS.md

Working rules for coding agents and maintainers in this repository.

## Project Shape

- Module: `github.com/dingdayu/go-viya`
- Package name: `viya`; do not derive it from the module suffix `go-viya`.
- Language: Go.
- Scope: a small, hand-written SAS Viya REST API client library.
- Public surface: exported types, constructors, functions, methods, and documented behavior in the root package.
- Main dependencies: Resty for HTTP, OAuth2 for SAS Logon token flows, and OpenTelemetry for instrumentation.

## Open Source Quality Bar

- Keep the library easy to evaluate from README examples, package docs, and tests.
- Prefer narrow, tested API additions over broad generated clients or speculative abstractions.
- Make changes friendly to downstream users: clear errors, stable exported names, useful docs, and minimal dependency churn.
- Keep pull requests focused on one behavior change or one coherent feature.
- Do not include real SAS Viya credentials, tenant URLs, tokens, customer data, or other secrets in code, tests, examples, logs, or fixtures.

## Ground Rules

- Keep request paths context-aware. Public methods that perform I/O should accept and propagate `context.Context`.
- Do not introduce global mutable state except where the existing default-client pattern requires it.
- Return errors with operation context and HTTP status where possible.
- Keep authentication failures wrapped or mapped consistently so callers can check `ErrViyaAuthFailed`.
- Use structured JSON types for stable SAS Viya responses; use `map[string]any` only for dynamic configuration payloads.
- Preserve Resty and OpenTelemetry behavior when changing shared client, transport, or token code.
- Add or update tests for user-visible behavior, error handling, request construction, and concurrency-sensitive code.
- Keep examples and README snippets aligned with the exported API.

## API Design

- Follow semantic versioning expectations for a public Go module.
- Avoid breaking exported names, method signatures, error sentinels, JSON field semantics, and documented behavior in minor or patch releases.
- Add options with functional options when they configure client or token-provider behavior.
- Prefer typed request and response structs for documented SAS Viya APIs.
- Keep SAS Viya service coverage incremental: identities, configuration, batch, CAS, authentication, and observability should grow around tested workflows.
- When adding endpoint helpers, include tests that assert method, path, authentication behavior, request body, and response decoding.

## Required Checks

Run before marking changes complete:

```bash
gofmt -w .
go mod tidy
go vet ./...
go test ./...
```

For authentication, HTTP middleware, default-client, token cache, transport, or shared client changes, also run:

```bash
go test -race ./...
```

For release preparation, also validate release configuration when GoReleaser is available:

```bash
goreleaser check
```

## Release Notes

Format user-facing changes for GitHub Releases using these categories:

- Added: new API or capability.
- Changed: behavior changes that users may notice.
- Fixed: bug fixes.
- Deprecated: API that remains available but should be avoided.
- Removed: breaking changes, only for major versions.

## Compatibility

This is a public Go module. After the first public tag exists, run the compatibility check before tagging a release:

```bash
go run golang.org/x/exp/cmd/gorelease@latest -base=latest
```

`gorelease` may fail before the first public tag exists; do not treat that as a project failure before an initial release.
