# viya-cli TODO

Goals and constraints for remaining sas-mcp-server replacement work.

## Goals

- Keep the runtime Go-only: all new capabilities should be implemented through the root `viya` package and exposed through `examples/viya-cli`.
- Preserve the CLI as the agent-facing contract: commands should default to useful console text, support `-o json` for machine parsing, and avoid printing secrets.
- Expand coverage incrementally beyond the implemented SAS execution and CAS discovery commands.
- Keep `SKILLS.md` aligned with every new command so agents know when and how to call it.

## Remaining Capabilities

- Data operations and files:
  - upload CSV data into a CAS table
  - promote a CAS table to global scope
  - list, upload, and download Viya Files service files
- Reports and visualization:
  - list Visual Analytics reports
  - get report metadata and definition
  - request report section image rendering
- Job Execution service:
  - submit SAS code as an asynchronous job
  - list jobs, get job status, cancel jobs, and retrieve job logs
- Model management and scoring:
  - list AutoML projects
  - create and run AutoML projects
  - list model repository models
  - list MAS modules
  - score input data through a MAS module step
- Prompt workflows:
  - encode log debugging, dataset exploration, data quality, statistical analysis, code optimization, macro building, and report generation as skill guidance or CLI-friendly templates without reintroducing Python runtime dependencies

## Constraints

- Public root-package additions must accept `context.Context`, use typed request/response structs where response shapes are stable, and return operation-specific errors with HTTP status when possible.
- Use `map[string]any` only for dynamic payloads such as scoring inputs, configuration-like bodies, or flexible service responses.
- Preserve Resty token middleware behavior and existing Compute execution output.
- Add focused tests for every new client method and CLI command, including method, path, query params, authentication propagation, response decoding, text output, `-o json` output, failure output, and URL escaping.
- Keep generated or user-provided SAS code free of credentials, tokens, tenant secrets, and customer data.
- Update `README.md` and `SKILLS.md` whenever new commands become available.

## Done

- Execute SAS code through `viya-cli run`.
- Discover CAS servers, caslibs, tables, table metadata, columns, and sample rows through `viya-cli cas ...`.
- Provide an agent skill guide that uses `viya-cli` instead of an MCP service.
