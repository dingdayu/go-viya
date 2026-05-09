# SKILLS.md

Skills, technology concepts, and tool guidance needed to evolve `go-viya` as a
stable open-source Go client library.

Use this file to decide which domain knowledge to load for a task. Operational
rules live in `AGENTS.md`; detailed constraints live in `docs/`.

## Open-Source Library Stewardship

- Keep the public API small, documented, and compatible with downstream users.
- Prefer incremental service coverage over generated clients or broad
  speculative abstractions.
- Evaluate changes through user-visible behavior: examples, README snippets,
  package docs, errors, and tests.
- Keep pull requests reviewable by separating API changes, refactors,
  documentation, release work, and tooling changes.
- Understand semantic versioning, Go module compatibility, release notes, and
  release-please-compatible Conventional Commits. Keep commits atomic and commit
  subjects short. See `docs/release.md`.

## Go Public API Design

- Design exported names, constructors, functions, and methods for long-term
  compatibility.
- Use `context.Context` for networked operations and token acquisition.
- Use functional options for configurable client and token-provider behavior.
- Keep package-level mutable state minimal and limited to documented default
  client behavior.
- Return errors with operation context and preserve checkable sentinels such as
  `ErrViyaAuthFailed`.
- Use table-driven tests and `httptest.Server` to verify behavior without live
  services. See `docs/go-development.md` and `docs/api-design.md`.

## SAS Viya REST Concepts

- Understand SAS Viya REST service boundaries, including identities,
  configuration, batch, compute, CAS, files, and Job Execution.
- Understand SAS Logon OAuth2 flows used by this library: client credentials,
  password grant, authorization code, and custom bearer-token providers.
- Model stable REST responses with typed structs and use dynamic maps only for
  configuration payloads or intentionally open-ended data.
- Validate API behavior with representative local HTTP tests by default. Use
  real Viya environments only for explicitly requested or marked integration
  work.
- Document service-specific assumptions such as media types, CAS server names,
  compute contexts, status fields, and pagination behavior.

## HTTP Client, Authentication, and Concurrency

- Understand Resty request construction, middleware behavior, response handling,
  and context propagation.
- Preserve token-provider cache semantics and concurrency safety when changing
  authentication code.
- Use `go test -race ./...` for authentication, shared client, default-client,
  token cache, transport, or middleware changes.
- Keep refresh-token storage, rotation, revocation, encryption, auditing, tenant
  isolation, and distributed locking outside this library unless explicitly
  modeled through a user-provided `TokenProvider`.

## Observability and Security

- Preserve OpenTelemetry spans around outbound token requests and client
  operations.
- Include useful operation and status context in spans while excluding bearer
  tokens, passwords, tenant secrets, customer data, request bodies, and sensitive
  payloads.
- Treat credentials, real tenant URLs, tokens, customer data, and production logs
  as toxic data: do not print, store, fixture, or transform them in repository
  artifacts.
- Keep agent-facing command output safe for logs and automation.

## Testing and Verification Tools

- `gofmt -w .`: normalize Go formatting after code edits.
- `go mod tidy`: keep module metadata deterministic after dependency or import
  changes.
- `go vet ./...`: catch common Go correctness issues.
- `go test ./...`: required unit-test baseline for code changes.
- `go test -race ./...`: required for concurrency-sensitive client,
  authentication, token, middleware, transport, and default-client changes.
- `git diff --check`: required whitespace check for documentation-only changes.
- `goreleaser check`: validate release configuration when preparing releases
  and GoReleaser is available.
- `go run golang.org/x/exp/cmd/gorelease@latest -base=latest`: check public Go
  API compatibility after the first public tag exists.

## Agent-Facing Tooling and Examples

- Maintain examples that are deterministic, minimal, and aligned with exported
  API behavior.
- Use JSON output for automation and human-readable output for interactive CLI
  workflows.
- For `examples/viya-cli`, load `examples/viya-cli/SKILLS.md` before changing or
  using the CLI workflow.
- Do not require Python MCP servers or live SAS Viya credentials for ordinary
  Go-only development workflows.

## Documentation Skills

- Keep `AGENTS.md` short as the mandatory entry point.
- Put long-form project structure, architecture, Go development, API design, and
  release guidance under `docs/` for progressive disclosure.
- Write agent instructions as concrete controls with observable checks.
- Update README, examples, package docs, or `llms.txt` when public behavior or
  agent onboarding changes.
