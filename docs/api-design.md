# API Design

`go-viya` is a public Go module. API additions should be incremental, tested,
and easy for downstream users to evaluate from examples and docs.

## Public API Style

- Follow semantic versioning expectations for a public Go module.
- Avoid breaking exported names, method signatures, error sentinels, JSON field
  semantics, and documented behavior in minor or patch releases.
- Add options with functional options when they configure client or
  token-provider behavior.
- Prefer typed request and response structs for documented SAS Viya APIs.
- Keep service coverage incremental: identities, configuration, batch, CAS,
  authentication, files, compute, Job Execution, and observability should grow
  around tested workflows.

## Endpoint Helper Style

When adding endpoint helpers:

- Accept `context.Context` for I/O methods and propagate it to the request.
- Build paths in the same style as nearby service helpers.
- Use the predefined `Accept*` constants from `response.go` for Content-Type
  negotiation instead of inline media type strings.
- Preserve authentication behavior and shared Resty client behavior.
- Return pure `error` for state-changing operations that have no meaningful
  response body (the boolean pattern was removed in v0.7.0 — callers check
  `err != nil` only). For resource-returning operations, use the standard
  `(resp Type, err error)` named return pattern.
- Return clear operation context in errors and include HTTP status where useful.
- Decode stable responses into typed structs.
- Use `map[string]any` only for dynamic configuration or intentionally open-ended
  payloads.
- Add OpenTelemetry spans (`ctx, span := tracer.Start(ctx, "MethodName")`) to
  every public endpoint method for observability.
- Include tests for method, path, authentication behavior, request body, status
  and error handling, and response decoding.

## Documentation Requirements

- Update README, examples, or package docs when user-facing behavior changes.
- Keep README snippets aligned with the exported API.
- Do not document live credentials, real tenant URLs, tokens, customer data, or
  logs containing secrets.
- Prefer concise examples that show complete setup, context propagation, and
  error handling.

## Compatibility

- After the first public tag exists, run compatibility checks before tagging a
  release:

```bash
go run golang.org/x/exp/cmd/gorelease@latest -base=latest
```

- `gorelease` may fail before the first public version exists; do not treat that
  as a project failure before an initial release.
