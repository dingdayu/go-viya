# Architecture

The architecture is optimized for a small, hand-written public Go module. The
harness exists to keep AI-assisted changes narrow, observable, and compatible.

## Harness Engineering Control Model

Separate the model from the harness: the model proposes changes, while
repository tools, tests, review gates, and documented invariants decide whether
those changes are acceptable.

- **Intent gate**: classify work before editing. If a request is ambiguous,
  inspect nearby docs, examples, and tests before asking for clarification.
- **Small-batch execution**: prefer narrow, reversible edits. Do not combine
  unrelated API changes, cleanup, release work, and documentation rewrites.
- **Closed-loop verification**: every material change must be followed by a
  matching sensor: formatter, linter, unit test, race test, compatibility check,
  generated diff review, or documented manual inspection.
- **Traceability**: summaries and pull request descriptions must explain what
  changed, why it changed, which files were touched, and which checks passed.
- **Safe termination**: never leave the workspace in a known-broken state. If a
  fix cannot be completed, revert partial edits or report the remaining blocker
  with exact failing evidence.

## Architectural Invariants

Preserve these fitness functions unless a change intentionally updates the
architecture and includes tests and docs that make the new invariant explicit.

- Public I/O methods must accept and propagate `context.Context`.
- Shared HTTP, transport, token, and instrumentation code must preserve Resty
  behavior and OpenTelemetry spans.
- Authentication failures must remain wrapped or mapped so callers can check
  `ErrViyaAuthFailed`.
- Stable SAS Viya responses should use typed structs; reserve `map[string]any`
  for dynamic configuration payloads.
- Package-level mutable state is allowed only for the existing default-client
  pattern; prefer constructor injection and functional options elsewhere.
- Tests for endpoint helpers must assert HTTP method, path, authentication
  behavior, request body, status and error handling, and response decoding.

## Risk Gates

Use higher verification for changes that touch high-risk harness boundaries.

- **Authentication, token cache, shared client, transport, middleware, or default
  client**: run `go test -race ./...` in addition to standard checks.
- **Public exported API or documented behavior**: update README, examples, or
  package docs when user-facing behavior changes; consider `gorelease` once a
  public base tag exists.
- **Error mapping or status handling**: add tests for the exact error value,
  wrapping behavior, and HTTP status context.
- **CLI or agent-facing workflow**: keep text output human-readable, provide JSON
  output for automation, and document parsing expectations in the relevant
  `SKILLS.md` or README.

## Dependency Boundaries

- Keep the root package focused on the public Go client library.
- Do not make the library depend on example-only workflow code.
- Preserve the current dependency profile unless a new dependency is justified by
  a narrow API or reliability gain.
- Do not require live SAS Viya access for normal unit tests or examples.
