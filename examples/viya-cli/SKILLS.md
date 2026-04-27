---
name: sas-viya
description: Use this skill when you need to execute SAS code or discover CAS data assets on SAS Viya through the Go viya-cli example. It replaces SAS MCP server style workflows with direct CLI calls and supports text or JSON output.
---

# viya-cli Agent Skill

Use `viya-cli` when an agent needs to execute SAS code or discover CAS data
assets on SAS Viya. It is a Go-only CLI; do not start a Python MCP server for
this workflow.

## Setup

Install from the repository root when needed:

```bash
go install ./examples/viya-cli
```

Configuration can come from flags, environment variables, or
`~/.sas/config.json` and `~/.sas/credentials.json`.

Common environment variables:

```bash
VIYA_BASE_URL=https://example.viya.sas.com
VIYA_ACCESS_TOKEN=...
VIYA_COMPUTE_CONTEXT_NAME="SAS Job Execution compute context"
```

## Execute SAS Code

Run short generated code:

```bash
viya-cli run --code "proc options; run;"
```

Run a file:

```bash
viya-cli run --file ./program.sas
```

Stream generated code:

```bash
printf '%s\n' "proc contents data=sashelp.class; run;" | viya-cli run --file -
```

Default output is console-friendly text. Use `-o json` when a script or agent
needs structured output. In JSON output, check `ok` first. On success, use
`listing` for report output and `log` for diagnostics. On failure, inspect
`error`, `log`, `state`, and `jobConditionCode`.

## CAS Discovery

All CAS discovery commands default to table-like text. Use `-o json` for JSON
shaped like:

```json
{ "ok": true, "data": {} }
```

Failures use:

```json
{ "ok": false, "error": "..." }
```

List CAS servers:

```bash
viya-cli cas servers
```

Machine-readable form:

```bash
viya-cli cas -o json servers
```

List caslibs:

```bash
viya-cli cas caslibs --server cas-shared-default --limit 50
```

List tables:

```bash
viya-cli cas tables --server cas-shared-default --caslib Public --limit 50
```

Get table metadata:

```bash
viya-cli cas table-info --server cas-shared-default --caslib Public --table HMEQ
```

Get columns:

```bash
viya-cli cas columns --server cas-shared-default --caslib Public --table HMEQ --limit 200
```

Fetch sample rows:

```bash
viya-cli cas rows --server cas-shared-default --caslib Public --table HMEQ --start 0 --limit 100
```

Prefer discovery commands before generating SAS that references unknown caslibs,
tables, or columns.

## Prompt Workflows

For SAS log debugging, data exploration, quality checks, statistical analysis,
optimization, macro building, and report generation, produce the needed SAS code
or analysis text directly in the conversation, then use `viya-cli run` only when
execution is needed.

Keep generated SAS short and deterministic. Use explicit libraries, table names,
titles, and ODS destinations when producing report code.

## Safety

- Never include credentials in generated SAS code or logs.
- Never print credentials, tokens, passwords, or customer data unless the user explicitly provided that exact public-safe output.
- Use `-timeout` for long-running tasks.
- Use `-context-name` or `-context-id` when the deployment has multiple Compute contexts.

Remaining sas-mcp-server replacement work is tracked in `todo.md`. Do not
depend on Python runtime code for the Go-only skill.
