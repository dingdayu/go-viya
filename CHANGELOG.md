# Changelog

All notable changes to this project will be documented in this file.

This project follows [Semantic Versioning](https://semver.org/).

## Unreleased

### ⚠ BREAKING CHANGES

* `RefreshIdentitiesCache`, `PatchIdentitiesLDAPUser`, and `UpdateIdentitiesLDAPObjectFilter` no longer return a redundant boolean — they now return `error` directly. Callers that check the boolean should switch to checking `err` only.

### Features

* add Accept header constants to eliminate string duplication across endpoint helpers
* add concurrent-access test for token providers
* add test helpers for httptest server setup

### Bug Fixes

* add nil check for baseURL in NewClient to prevent panic
* add missing Accept header to CAS table state operations
* add missing Accept header to DeleteBatchJob
* remove redundant parameter validation in uploadBatchFileFromReader
* tighten golangci-lint exclusion rules

### Code Refactoring

* simplify identities API by removing redundant boolean return values
* unify parameter naming: jobId→jobID, serverId→serverID, contextId→contextID
* remove otelhttp transport wrapper from token HTTP client (rely on application-level spans)
* remove unnecessary ErrInvalidParameter.Unwrap method

### Documentation

* add polling interval guidance to WaitBatchJobCompleted godoc
* translate Chinese comments to English in identities.go

### Dependencies

* update golang.org/x/net from v0.43.0 to v0.50.0
* add github.com/stretchr/testify for test assertions

### Test Improvements

* add table-driven tests for ParseURL, TokenURL, ErrorResponse
* add testify assertions across new test files
* add parallel test execution to most httptest-based tests
* add tests for UploadFile, GetComputeJobsList, GetIdentitiesUsers, GetConfiguration
* add t.Helper() to test helpers

## [0.7.0](https://github.com/dingdayu/go-viya/compare/v0.6.0...v0.7.0) (2026-05-14)


### ⚠ BREAKING CHANGES

* method parameters renamed for consistency
* RefreshIdentitiesCache, PatchIdentitiesLDAPUser, and UpdateIdentitiesLDAPObjectFilter now return error instead of (bool, error)

### Features

* add Accept header constants for SAS Viya REST APIs ([2975076](https://github.com/dingdayu/go-viya/commit/2975076e1bc5d6dfc17b77a168a8c6d15e6ff059))


### Bug Fixes

* add nil check for baseURL and Accept header ([53c3a92](https://github.com/dingdayu/go-viya/commit/53c3a92a66eb10906054e76c2c3f62a4729ff49b))
* remove omitempty from JobConditionCode fields ([613c2ed](https://github.com/dingdayu/go-viya/commit/613c2ed449595c63ef832414b285da3ed6d8b8d2))
* remove redundant parameter validation in uploadBatchFileFromReader ([23d0ad4](https://github.com/dingdayu/go-viya/commit/23d0ad4a9c740267814cb06e30f7c94c0bf02763))


### Code Refactoring

* rename jobId to jobID in batch jobs ([50d329e](https://github.com/dingdayu/go-viya/commit/50d329e79cb75c30ca7d23ea740fa3b5fda2ff86))
* simplify identities API to return error only ([2dc9dee](https://github.com/dingdayu/go-viya/commit/2dc9deee39aa98581b766e5a27d1a16354465473))

## [0.6.0](https://github.com/dingdayu/go-viya/compare/v0.5.0...v0.6.0) (2026-05-11)(https://github.com/dingdayu/go-viya/compare/v0.5.0...v0.6.0) (2026-05-11)


### Features

* add workflow orchestration ([#13](https://github.com/dingdayu/go-viya/issues/13)) ([f9c9233](https://github.com/dingdayu/go-viya/commit/f9c923370ae80313996582367c65d2b9393b6cd1))

## [0.5.0](https://github.com/dingdayu/go-viya/compare/v0.4.0...v0.5.0) (2026-05-09)


### ⚠ BREAKING CHANGES

* NewClientCredentialsTokenProvider, NewPasswordTokenProvider, NewAuthCodeTokenProvider now accept *url.URL instead of string baseURL.

### Features

* use *url.URL for baseURL in token providers ([d30aa69](https://github.com/dingdayu/go-viya/commit/d30aa69546f26920aea2402ad0e513ad11a6dca2))

## [0.4.0](https://github.com/dingdayu/go-viya/compare/v0.3.0...v0.4.0) (2026-05-07)


### ⚠ BREAKING CHANGES

* The auth middleware is now injected via Option. Direct modifications to NewClient's internal middleware logic should be replaced with WithAuthMiddleware.
* Rename PatchIdentitiesLDAPGroup to PatchIdentitiesLDAPUser to match its actual functionality (updating LDAP user provider configuration).

### Features

* rename PatchIdentitiesLDAPGroup to PatchIdentitiesLDAPUser ([7a13e53](https://github.com/dingdayu/go-viya/commit/7a13e5349dc94db04c896de6a7d671ce0e9931c7))


### Bug Fixes

* add input validation ([b9424f6](https://github.com/dingdayu/go-viya/commit/b9424f6aa7232d1ba021240915df149c5c2ae51a))
* improve error handling ([db7b75f](https://github.com/dingdayu/go-viya/commit/db7b75fa01cfcf8b4592b15ac5ab3bc196db074a))
* make defaultClient thread-safe ([65b8c8d](https://github.com/dingdayu/go-viya/commit/65b8c8d3e2a866aa5ad156dad0599ed70e739185))
* migrate golangci-lint config from v1 to v2 ([10420a4](https://github.com/dingdayu/go-viya/commit/10420a4a5262531678a52fb5f615489b169f2e3d))
* remove initial release version from release-please config ([07f15ea](https://github.com/dingdayu/go-viya/commit/07f15ea9f7fe407c7e9756156d786886af846dd6))
* revert release version to 0.3.0 in release-please manifest ([b82935c](https://github.com/dingdayu/go-viya/commit/b82935cd385bb03307930f5958eefea93a4e90bf))
* wrap token fetch errors with ErrViyaAuthFailed for errors.Is ([f079568](https://github.com/dingdayu/go-viya/commit/f079568eebbab8bd937e665c4078fda2910cc868))


### Code Refactoring

* decouple auth middleware from NewClient ([2192f48](https://github.com/dingdayu/go-viya/commit/2192f48f6ca5a09a7d29d8dfab8674874fe0ab67))

## [0.4.0](https://github.com/dingdayu/go-viya/compare/github.com/dingdayu/go-viya-v0.3.0...github.com/dingdayu/go-viya-v0.4.0) (2026-05-07)


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
* remove initial release version from release-please config ([07f15ea](https://github.com/dingdayu/go-viya/commit/07f15ea9f7fe407c7e9756156d786886af846dd6))
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
