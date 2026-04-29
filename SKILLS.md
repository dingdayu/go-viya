# SKILLS.md

This project benefits from contributors with the following skills.

## Go Library Design

- Design small exported APIs with clear names and stable behavior.
- Use contexts correctly for HTTP requests and token acquisition.
- Keep package-level state minimal.
- Maintain backward compatibility for exported symbols.

## SAS Viya REST APIs

- Understand SAS Viya authentication flows and token endpoints.
- Validate API behavior against real or representative Viya environments.
- Document service-specific assumptions such as CAS server names and media types.
- Keep CAS data operations, Files Service calls, reports, Visual Analytics report operations, report image rendering, and Job Execution service calls aligned with documented REST paths, media types, and status handling.

## Testing

- Add table-driven unit tests for request construction, error handling, and response parsing.
- Use `httptest.Server` for HTTP behavior.
- Cover CLI text and JSON output when adding agent-facing commands.
- Avoid tests that require live SAS Viya credentials unless they are explicitly marked as integration tests.

## Observability

- Preserve OpenTelemetry spans around networked operations.
- Avoid capturing sensitive request bodies or credentials.
- Include enough span status information to debug failed HTTP calls.

## Release Engineering

- Follow semantic versioning.
- Use `gorelease` to detect public API compatibility problems.
- Use GoReleaser for GitHub Release creation from version tags.
