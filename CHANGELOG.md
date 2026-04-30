# Changelog

All notable changes to this project will be documented in this file.

This project follows [Semantic Versioning](https://semver.org/).

## Unreleased

## v0.3.0 - 2026-04-30

### Added

- Added Compute API helpers for contexts, sessions, jobs, job state, logs, and listings.
- Added broader Batch API coverage for servers, contexts, jobs, file sets, file upload, input, output, and state management.
- Added CAS discovery helpers for servers, caslibs, tables, columns, and sample rows.
- Added CAS data operations for CSV upload, table promotion, and explicit server-aware table load and unload workflows.
- Added Files Service helpers for listing, uploading, downloading, and reader-based uploads.
- Added Job Execution helpers for listing jobs, submitting code, retrieving job details, and reading logs.
- Added runnable examples for authentication flows, batch jobs, configuration, CAS table state, and the `viya-cli` example.

### Changed

- Reworked batch support into focused files and typed request/response structures.
- Made CAS table state operations require an explicit CAS server identifier.
- Improved request path escaping for user-controlled path parameters.

### Fixed

- Refreshed Viya tokens with the current request context.
- Changed token provider test failures that run in background goroutines from fatal assertions to non-fatal error reporting.

### Removed

- Removed `CAS_SERVER_NAME`; callers now pass the CAS server ID explicitly.
- Removed `(*Client).LoadCasLibTableToMemory`; use `(*Client).LoadCASTableToMemory`.
- Removed `(*Client).UnLoadCasLibTableInMemory`; use `(*Client).UnloadCASTableFromMemory`.
