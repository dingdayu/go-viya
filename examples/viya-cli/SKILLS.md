---
name: sas-viya
description: Use this skill when you need to execute SAS code, discover or manage CAS data, use Viya files, or submit Job Execution jobs on SAS Viya through the Go viya-cli example. It replaces SAS MCP server style workflows with direct CLI calls and supports text or JSON output.
---

# viya-cli Agent Skill

Use `viya-cli` when an agent needs to execute SAS code, discover or manage CAS
data assets, exchange files, or submit asynchronous jobs on SAS Viya. It is a
Go-only CLI; do not start a Python MCP server for this workflow.

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

## CAS Data Operations

Upload CSV data into CAS:

```bash
viya-cli data upload-csv --server cas-shared-default --caslib Public --table WORK_UPLOAD --file ./data.csv
```

Stream CSV data:

```bash
cat ./data.csv | viya-cli data upload-csv --server cas-shared-default --caslib Public --table WORK_UPLOAD --file -
```

Promote a table to global scope:

```bash
viya-cli data promote --server cas-shared-default --caslib Public --table WORK_UPLOAD
```

## Files Service

List files:

```bash
viya-cli files list --limit 50
viya-cli files list --filter-name report
```

Upload and download files:

```bash
viya-cli files upload --name report.txt --file ./report.txt --content-type text/plain
viya-cli files download --id file-id
```

Use `-o json` for file metadata. File downloads write raw content in text mode.

## Job Execution

Submit SAS code asynchronously:

```bash
viya-cli jobs submit --code "proc options; run;" --name options-check
viya-cli jobs submit --file ./program.sas --context-name "SAS Job Execution compute context"
```

Inspect and manage jobs:

```bash
viya-cli jobs list --limit 20
viya-cli jobs status --id job-id
viya-cli jobs log --id job-id
viya-cli jobs cancel --id job-id
```

Prefer `viya-cli run` for immediate Compute execution with listing/log output.
Use `viya-cli jobs submit` when the user asks for asynchronous Job Execution
service behavior.

## Prompt Workflows

Use these workflows directly in the conversation. Generate SAS code or analysis
text first, then use `viya-cli run` only when execution is needed.

- **Debug a SAS log**: identify errors, warnings, and notable notes; for each issue, explain the likely root cause and suggest a concrete fix. If the user names a severity, focus on that level.
- **Explore a dataset**: generate production-ready profiling code for `library.dataset` using `PROC CONTENTS`, `PROC MEANS` for numeric variables, `PROC FREQ` for categorical variables, and `PROC UNIVARIATE` for distributions. Honor any focus variables.
- **Check data quality**: generate SAS code that checks completeness, duplicate keys, validity or range rules, and consistency. Include user-provided key variables and business rules, and produce a summary report with quality scores.
- **Build statistical analysis**: generate a complete analysis workflow for the requested analysis type, dataset, response variable, and predictors. Include data preparation, model fitting, diagnostics, assumption checks, and interpretation comments.
- **Optimize SAS code**: review code for the requested focus, defaulting to performance and readability. Explain what the current code does, the issue, an optimized replacement, and expected improvement.
- **Explain SAS code**: provide a block-by-block explanation tailored to the requested audience level, defaulting to intermediate. Cover what each block does, key SAS concepts, and potential issues or improvements.
- **Build a SAS macro**: create a production-quality `%macro` with parameter validation, helpful errors, a header comment with usage examples, `%LOCAL` internal variables, and SAS macro best practices.
- **Generate a report**: produce ODS and `PROC REPORT` or `PROC TABULATE` code for the requested dataset, report type, and output format, defaulting to a summary HTML report. Include titles, footnotes, formatting, summary statistics, and proper ODS open/close statements.

Keep generated SAS short and deterministic. Use explicit libraries, table names,
titles, and ODS destinations when producing report code. Do not embed secrets or
credentials in generated SAS.

## Safety

- Never include credentials in generated SAS code or logs.
- Never print credentials, tokens, passwords, or customer data unless the user explicitly provided that exact public-safe output.
- Use `-timeout` for long-running tasks.
- Use `-context-name` or `-context-id` when the deployment has multiple Compute contexts.
- Use `-o json` before parsing command output in scripts, except when intentionally downloading raw file content or raw job logs.

Remaining sas-mcp-server replacement work is tracked in `todo.md`. Do not
depend on Python runtime code for the Go-only skill.
