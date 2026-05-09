# AGENTS.md

Working rules for coding agents and maintainers in this repository.

This file is the repository harness entry point. Keep it small enough for every
agent session to read first. Load the linked documents only when the task touches
that area.

## Project Snapshot

- Module: `github.com/dingdayu/go-viya`
- Package name: `viya`; do not derive it from the module suffix `go-viya`.
- Language: Go.
- Scope: a small, hand-written SAS Viya REST API client library.
- Public surface: exported types, constructors, functions, methods, and
  documented behavior in the root package.
- Main dependencies: Resty for HTTP, OAuth2 for SAS Logon token flows, and
  OpenTelemetry for instrumentation.

## Progressive Disclosure Map

Read these documents according to the task scope:

- `SKILLS.md`: project skills, technology concepts, and tool guidance needed to
  keep open-source evolution stable.
- `docs/project-structure.md`: repository layout, ownership boundaries, and
  where new files belong.
- `docs/architecture.md`: architectural invariants, harness control model, and
  risk gates.
- `docs/go-development.md`: Go implementation constraints, testing patterns,
  observability rules, and local verification.
- `docs/api-design.md`: exported API style, compatibility rules, endpoint helper
  conventions, and SAS Viya REST modeling.
- `docs/release.md`: release notes, Conventional Commits, release-please,
  GoReleaser, and Go module compatibility checks.
- `examples/viya-cli/SKILLS.md`: agent workflow for the `viya-cli` example.

## Open Source Quality Bar

- Keep the library easy to evaluate from README examples, package docs, and
  tests.
- Prefer narrow, tested API additions over broad generated clients or
  speculative abstractions.
- Make changes friendly to downstream users: clear errors, stable exported
  names, useful docs, and minimal dependency churn.
- Keep pull requests focused on one behavior change or one coherent feature.
- Do not include real SAS Viya credentials, tenant URLs, tokens, customer data,
  or other secrets in code, tests, examples, logs, or fixtures.

## Agent Operating Contract

- Classify the request before editing: docs, test, bug fix, feature, refactor,
  release, or investigation.
- Inspect existing code, README examples, package docs, and tests before adding
  public API, new options, new endpoint helpers, or new error behavior.
- Preserve user work. Do not overwrite unrelated local changes, generated files,
  or untracked files unless the task explicitly owns them.
- Do not commit, tag, publish, or push unless the user explicitly requests that
  operation.
- Prefer deterministic repository-local tooling over live services. Live SAS
  Viya access is integration work and must not be required for normal unit tests.
- Treat secrets as toxic data. Do not print, transform, store, or fixture real
  credentials, tokens, tenant URLs, customer data, or logs containing them.

## Required Checks

Run before marking code changes complete:

```bash
gofmt -w .
go mod tidy
go vet ./...
go test ./...
```

For documentation-only changes, inspect the rendered Markdown or diff and run
`git diff --check` for changed Markdown files. That is sufficient when no code,
module metadata, examples, or generated artifacts changed. Do not run live SAS
Viya integration workflows unless the task explicitly requires live credentials.

For authentication, HTTP middleware, default-client, token cache, transport, or
shared client changes, also run:

```bash
go test -race ./...
```

For release preparation, also validate release configuration when GoReleaser is
available:

```bash
goreleaser check
```

## Completion Report

Every final summary or pull request description must include:

- What changed and why.
- Which files were touched.
- User-visible behavior or documentation impact.
- Verification commands that passed, or the exact reason a check did not apply.
