# viya-cli

`viya-cli` is a small CLI example for agent frameworks that need to execute SAS
code on SAS Viya through the Compute API and discover CAS data assets.

It reads configuration from flags, environment variables, and local SAS CLI-style
files:

- `~/.sas/config.json`
- `~/.sas/credentials.json`

The parser accepts both simple top-level keys and common profile containers such
as `profiles`, `contexts`, and `credentials`.

## Install

From the repository root:

```bash
go install ./examples/viya-cli
```

Or run without installing:

```bash
go run ./examples/viya-cli run --code "data _null_; put 'hello from viya-cli'; run;"
```

## Configuration

Environment variables override `~/.sas` files:

```bash
export VIYA_BASE_URL="https://example.viya.sas.com"
export VIYA_CLIENT_ID="go-viya"
export VIYA_USERNAME="my-user"
export VIYA_PASSWORD="my-password"
export VIYA_COMPUTE_CONTEXT_NAME="SAS Job Execution compute context"
```

Supported authentication inputs:

- `VIYA_ACCESS_TOKEN` / `SAS_ACCESS_TOKEN`
- `VIYA_CLIENT_ID` and `VIYA_CLIENT_SECRET`
- `VIYA_USERNAME` and `VIYA_PASSWORD`, optionally with `VIYA_CLIENT_ID` and `VIYA_CLIENT_SECRET`

Supported configuration inputs:

- `VIYA_BASE_URL`, `SAS_VIYA_URL`, `SAS_SERVICES_ENDPOINT`, or `SAS_BASE_URL`
- `VIYA_COMPUTE_CONTEXT_ID` or `VIYA_COMPUTE_CONTEXT_NAME`
- `VIYA_PROFILE` or `SAS_PROFILE`

The equivalent values can also be passed with flags such as `-base-url`,
`-context-id`, `-context-name`, `-username`, `-password`, and `-access-token`.

## Usage

Execute inline SAS code:

```bash
viya-cli run --code "data _null_; put 'hello'; run;"
```

Execute a local SAS program:

```bash
viya-cli run --file ./program.sas
```

Read SAS code from stdin:

```bash
cat ./program.sas | viya-cli run --file -
```

Keep the Compute session after execution:

```bash
viya-cli run --file ./program.sas --keep-session
```

The command exits non-zero on failure. By default it writes console-friendly
text. Use `-o json` when an agent or script needs the full machine-readable
result, including Compute context, session, job, final state, condition code,
log text, and listing text.

Discover CAS servers, caslibs, tables, columns, and sample rows:

```bash
viya-cli cas servers
viya-cli cas caslibs --server cas-shared-default
viya-cli cas tables --server cas-shared-default --caslib Public
viya-cli cas table-info --server cas-shared-default --caslib Public --table HMEQ
viya-cli cas columns --server cas-shared-default --caslib Public --table HMEQ
viya-cli cas rows --server cas-shared-default --caslib Public --table HMEQ --limit 25
```

CAS discovery commands default to table-like text. Use `-o json` to write
`{ "ok": true, "data": ... }` on success and `{ "ok": false, "error": "..." }`
on failure.

## Agent Integration

Agents should treat this CLI as a tool with one primary operation:

```bash
viya-cli run --file path/to/program.sas
```

Use `-code` for generated short programs and `-file -` when the agent streams
code through stdin. Use `-o json` for machine parsing, then inspect:

- `ok`
- `state`
- `jobConditionCode`
- `log`
- `listing`
- `error`

The CLI intentionally does not print secrets and is designed to be called from
modern agent frameworks, shell tools, or MCP-style wrappers.

## Agent Skill

This directory includes an agent skill guide at `SKILLS.md`. Keep it in this
repository as the documented agent workflow for `viya-cli`.

Remaining sas-mcp-server replacement work is tracked in `todo.md`. Runtime
execution for this example is Go-only and goes through `viya-cli`.
