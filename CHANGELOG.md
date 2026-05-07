# Changelog

All notable changes to this project will be documented in this file.

This project follows [Semantic Versioning](https://semver.org/).

## Unreleased

## [1.0.0](https://github.com/dingdayu/go-viya/compare/github.com/dingdayu/go-viya-v0.3.0...github.com/dingdayu/go-viya-v1.0.0) (2026-05-07)


### ⚠ BREAKING CHANGES

* The auth middleware is now injected via Option. Direct modifications to NewClient's internal middleware logic should be replaced with WithAuthMiddleware.
* Rename PatchIdentitiesLDAPGroup to PatchIdentitiesLDAPUser to match its actual functionality (updating LDAP user provider configuration).
* GetDefaultClient now returns (*Client, error) and reports ErrDefaultClientNotSet when no default client is configured.

### Features

* add batch file reader upload ([b426a0a](https://github.com/dingdayu/go-viya/commit/b426a0acd8ff8aec1cf7638168587ebe30d6bdec))
* add compute api client ([1c1d2fc](https://github.com/dingdayu/go-viya/commit/1c1d2fccc527d301434743b16b87aa05fef012af))
* add data files and job execution APIs ([23f338d](https://github.com/dingdayu/go-viya/commit/23f338dfa4954c89b5bd18e5d2188bfeea02ad43))
* add viya CLI CAS discovery ([30156f6](https://github.com/dingdayu/go-viya/commit/30156f633fdceecbbf93d4a7a289b62dcd31cc30))
* expand batch api coverage ([a1ef1c4](https://github.com/dingdayu/go-viya/commit/a1ef1c4092808119ad406a0b481d2b0838c0d60a))
* rename PatchIdentitiesLDAPGroup to PatchIdentitiesLDAPUser ([7a13e53](https://github.com/dingdayu/go-viya/commit/7a13e5349dc94db04c896de6a7d671ce0e9931c7))
* report missing default client ([b37bd39](https://github.com/dingdayu/go-viya/commit/b37bd39db88091c3a81806fa5c857ba428b93d4c))
* return batch job from wait helper ([629f8c5](https://github.com/dingdayu/go-viya/commit/629f8c5b79adb0afc0def533a1b4f6e82053c46b))


### Bug Fixes

* add input validation ([b9424f6](https://github.com/dingdayu/go-viya/commit/b9424f6aa7232d1ba021240915df149c5c2ae51a))
* change fatal errors to non-fatal in token provider tests ([6efed61](https://github.com/dingdayu/go-viya/commit/6efed61a2558caaf1423ee2059d5cbf7d142b1c7))
* improve error handling ([db7b75f](https://github.com/dingdayu/go-viya/commit/db7b75fa01cfcf8b4592b15ac5ab3bc196db074a))
* make defaultClient thread-safe ([65b8c8d](https://github.com/dingdayu/go-viya/commit/65b8c8d3e2a866aa5ad156dad0599ed70e739185))
* migrate golangci-lint config from v1 to v2 ([10420a4](https://github.com/dingdayu/go-viya/commit/10420a4a5262531678a52fb5f615489b169f2e3d))
* refresh viya tokens with current context ([6e831d0](https://github.com/dingdayu/go-viya/commit/6e831d0fc15c2492ba3cb41c3035850c0ac426f3))
* refresh viya tokens with current context ([96d1999](https://github.com/dingdayu/go-viya/commit/96d1999c385fd3225cbf1f0cac21ef7e7ddd5274))
* wrap token fetch errors with ErrViyaAuthFailed for errors.Is ([f079568](https://github.com/dingdayu/go-viya/commit/f079568eebbab8bd937e665c4078fda2910cc868))


### Code Refactoring

* decouple auth middleware from NewClient ([2192f48](https://github.com/dingdayu/go-viya/commit/2192f48f6ca5a09a7d29d8dfab8674874fe0ab67))

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
