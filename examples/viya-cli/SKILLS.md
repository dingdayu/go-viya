---
name: sas-viya
description: Use this skill when you need to execute SAS code, discover or manage CAS data, use Viya files, inspect reports, create Visual Analytics dashboards, or submit Job Execution jobs on SAS Viya through the Go viya-cli example. It replaces SAS MCP server style workflows with direct CLI calls and supports text or JSON output.
---

# viya-cli Agent Skill

Use `viya-cli` when an agent needs to execute SAS code, discover or manage CAS
data assets, exchange files, inspect reports, create Visual Analytics
dashboards, or submit asynchronous jobs on SAS Viya. It is a Go-only CLI; do
not start a Python MCP server for this workflow.

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

## Reports

In SAS Viya, a dashboard is a Visual Analytics report. Use the word
**report** for the persisted Viya resource and use **dashboard** for the
agent-friendly workflow that creates a report with pages and visual objects.

Use `reports` commands when the report already exists or when you need metadata,
definition, or a rendered preview. Use `dashboard create` when the user asks to
build a new dashboard from data and intent.

List Visual Analytics reports:

```bash
viya-cli reports list --limit 50
viya-cli reports list --filter-name sales
```

Get report metadata and definition:

```bash
viya-cli reports get --id report-id
```

Request report section image rendering:

```bash
viya-cli reports image --id report-id --section-index 0 --size 800x600
```

Use `-o json` for report definitions and report image job details. After
creating a dashboard, call `viya-cli reports image` with the returned
`resultReportId` when the user wants a preview or confirmation that rendering
was requested.

## Dashboards

Create dashboards when the user asks for a chart, report page, scorecard,
summary view, business dashboard, operational dashboard, KPI view, or visual
analytics report based on a CAS table.

Dashboard creation uses a compact JSON spec that agents can generate from chat
intent. The CLI converts this spec into Visual Analytics report operations:

- `addData` for the CAS table.
- `addPage` for each page in the spec.
- `addObject` for each visual object.
- raw `operations`, when present, appended after the generated operations.

Basic command:

```bash
viya-cli dashboard create \
  --server cas-shared-default \
  --caslib Public \
  --table HMEQ \
  --name "HMEQ Dashboard" \
  --folder-uri /folders/folders/@myFolder \
  --spec dashboard.json
```

Use `--result-name-conflict rename` by default. Use `replace` only when the user
explicitly wants to overwrite an existing report name in the target folder.

### Dashboard Creation Workflow

Follow this workflow for chat-driven dashboard requests:

1. Identify the target data.
   If the user gives a table, use it. If not, ask for the CAS server, caslib,
   and table, or discover candidates:

```bash
viya-cli cas servers
viya-cli cas caslibs --server cas-shared-default
viya-cli cas tables --server cas-shared-default --caslib Public
```

2. Inspect columns before choosing visuals:

```bash
viya-cli cas columns --server cas-shared-default --caslib Public --table HMEQ -o json
viya-cli cas rows --server cas-shared-default --caslib Public --table HMEQ --limit 10 -o json
```

Use column metadata and sample rows to infer field roles:

- Categorical fields are good `category` values for bar charts.
- Numeric business metrics are good `measure` values.
- Dates or timestamps are good candidates for trend pages, but only use them
  when the user asks for time analysis or the field is clearly temporal.
- Avoid using IDs, free-text notes, keys, or high-cardinality identifiers as
  chart categories unless the user explicitly asks for them.

3. Translate the user's intent into pages and visuals.
   Keep the first version small and useful. Prefer 1-3 pages and 1-4 objects per
   page. Name pages and titles in business language from the user's request.

4. Generate a dashboard spec. The common object shape is:

```json
{
  "type": "barChart",
  "title": "Bad Rate by Job",
  "category": "JOB",
  "measure": "BAD"
}
```

Use `measures` when a visual should compare multiple numeric fields:

```json
{
  "type": "barChart",
  "title": "Loan and Mortgage by Region",
  "category": "REGION",
  "measures": ["LOAN", "MORTDUE"]
}
```

Use `dataRoles` for advanced Visual Analytics roles that do not fit
`category`/`measure`:

```json
{
  "type": "barChart",
  "title": "Sales by Segment",
  "dataRoles": {
    "category": "SEGMENT",
    "measures": ["SALES"]
  }
}
```

5. Create the dashboard:

```bash
viya-cli dashboard create \
  --server cas-shared-default \
  --caslib Public \
  --table HMEQ \
  --name "HMEQ Risk Dashboard" \
  --folder-uri /folders/folders/@myFolder \
  --spec dashboard.json \
  -o json
```

6. Use the JSON response. On success, read `data.resultReportId`,
   `data.resultReportUri`, and `data.status`. Then list or inspect the report:

```bash
viya-cli reports get --id report-id -o json
viya-cli reports image --id report-id --section-index 0 --size 800x600 -o json
```

### Spec Examples

Single-page dashboard:

```json
{
  "pages": [
    {
      "name": "Overview",
      "objects": [
        {
          "type": "barChart",
          "title": "Bad Rate by Job",
          "category": "JOB",
          "measure": "BAD"
        }
      ]
    }
  ]
}
```

Multi-page dashboard:

```json
{
  "pages": [
    {
      "name": "Executive Summary",
      "objects": [
        {
          "type": "barChart",
          "title": "Bad Rate by Job",
          "category": "JOB",
          "measure": "BAD"
        },
        {
          "type": "barChart",
          "title": "Average Loan by Reason",
          "category": "REASON",
          "measure": "LOAN"
        }
      ]
    },
    {
      "name": "Portfolio Detail",
      "objects": [
        {
          "type": "barChart",
          "title": "Debt-to-Income by Job",
          "category": "JOB",
          "measure": "DEBTINC"
        }
      ]
    }
  ]
}
```

Advanced spec with raw Visual Analytics operations:

```json
{
  "pages": [
    {
      "name": "Overview",
      "objects": []
    }
  ],
  "operations": [
    {
      "operationId": "customText",
      "includeObjectInResponse": true,
      "addObject": {
        "object": {
          "text": {
            "options": {
              "content": "Dashboard generated from a chat request."
            }
          }
        },
        "placement": {
          "page": {
            "target": "Overview",
            "position": "end"
          }
        }
      }
    }
  ]
}
```

### Choosing Visuals From User Requests

- "Show distribution" or "breakdown by X": use a bar chart with `category=X`.
- "Compare metric Y by group X": use a bar chart with `category=X`,
  `measure=Y`.
- "Executive dashboard": create an overview page with the most important
  categorical breakdowns and numeric metrics.
- "Risk dashboard": include target/risk indicators first, then explainers such
  as job, reason, region, or score bands if those fields exist.
- "Sales dashboard": prioritize revenue/sales measures by time, region,
  product, segment, or channel, depending on available columns.
- "Data quality dashboard": prefer counts, missingness indicators, duplicate
  flags, validity flags, and rule-failure measures if available.

If the requested visual cannot be represented by the compact spec, use raw
`operations` only when you know the Visual Analytics operation shape. Otherwise,
create the closest conservative dashboard and explain the limitation in the chat
response.

### Dashboard Safety

- Do not invent table names or columns. Discover them first or ask the user.
- Do not create a dashboard against sensitive data unless the user has clearly
  requested that table and output.
- Do not use `replace` conflict behavior unless the user asked to overwrite.
- Keep generated specs deterministic and small enough for review.
- Prefer `-o json` for dashboard creation so the agent can capture the created
  report ID.

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
