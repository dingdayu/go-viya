# Project Structure

This repository is intentionally flat at the root because the public package is
the product. Keep new code easy to discover from README examples, package docs,
and tests.

## Root Package

- Go package name: `viya`.
- Module path: `github.com/dingdayu/go-viya`.
- Root `*.go` files define the public client library. Avoid moving public API
  into nested packages unless a future design explicitly introduces that package
  boundary.
- Root `*_test.go` files should sit near the behavior they verify.

## Service Areas

The client grows around tested SAS Viya workflows rather than generated service
coverage.

- `client.go`, `response.go`: shared Resty client behavior and response helpers.
- `provider.go`, `token.go`: token providers and authentication behavior.
- `otel.go`: OpenTelemetry instrumentation.
- `identities.go`, `configuration.go`: identity and configuration APIs.
- `batch_*.go`, `compute_*.go`, `cas_*.go`, `files.go`, `job_execution.go`:
  focused service helpers.
- `examples/`: runnable examples and agent-facing CLI workflows.

## Documentation and Agent Guidance

- `README.md`: user-facing overview, quick start, supported features, and public
  examples.
- `llms.txt`: compact guide for AI agents and tooling.
- `AGENTS.md`: mandatory entry-point controls for all agent work.
- `SKILLS.md`: skills, technology concepts, and tools needed for stable
  open-source iteration.
- `docs/`: progressively disclosed engineering guidance for architecture, Go
  development, API design, and releases.
- `examples/viya-cli/SKILLS.md`: operational skill guide for the CLI example.

## Placement Rules

- Add new public endpoint helpers in the root package next to related service
  files.
- Add tests beside the implementation, using representative local HTTP servers
  rather than live SAS Viya dependencies.
- Add examples only when they clarify exported API usage or agent workflows.
- Add long-form engineering guidance under `docs/` and link it from
  `AGENTS.md` or `SKILLS.md` instead of expanding the entry-point files.
- Do not add generated clients, generated fixtures, or broad service trees unless
  a future change explicitly revises the project scope.
