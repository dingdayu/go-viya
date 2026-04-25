# AGENTS.md

This file gives coding agents and maintainers the working rules for this repository.

## Project Shape

- Module: `github.com/dingdayu/go-viya`
- Package name: `viya`
- Language: Go
- Public surface: exported types, constructors, and methods in the root package
- Scope: a small SAS Viya REST API client library

## Ground Rules

- Keep the package name as `viya`; do not derive it mechanically from the module suffix `go-viya`.
- Prefer small, explicit API additions over broad generated clients.
- Do not introduce global mutable state except where the existing default-client pattern requires it.
- Preserve context propagation on all request paths.
- Return useful errors that include operation context and HTTP status when possible.
- Keep authentication errors wrapped or mapped consistently with `ErrViyaAuthFailed`.
- Use structured JSON types when response shapes are stable; use `map[string]any` only for dynamic configuration payloads.

## Required Checks

Run these before considering a change complete:

```bash
gofmt -w .
go vet ./...
go test ./...
```

For changes touching authentication, HTTP middleware, or shared client behavior, also run:

```bash
go test -race ./...
```

## Release Notes

User-facing changes should be written so they can be reused in GitHub Releases:

- Added: new API or capability.
- Changed: behavior changes that users may notice.
- Fixed: bug fixes.
- Deprecated: API that remains available but should be avoided.
- Removed: breaking changes, only for major versions.

## Compatibility

This is a Go module intended for public use. Avoid breaking exported names or behavior in minor and patch releases. Before tagging a release after the first public version exists, run:

```bash
go run golang.org/x/exp/cmd/gorelease@latest -base=latest
```
