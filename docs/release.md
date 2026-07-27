# Release Engineering

This project uses semantic versioning, release-please-compatible commit
messages, release-please configuration for version and changelog automation, and
GoReleaser for GitHub Release creation.

## Release Please Credentials

Release Please must use the repository secret `RELEASE_PLEASE_TOKEN` for its
release pull requests and subsequent updates to trigger CI automatically. Set
the secret to a fine-grained personal access token owned by a maintainer, scoped
to this repository, with read and write access to contents and pull requests.

The workflow falls back to the built-in `GITHUB_TOKEN` when the secret is not
configured so release automation does not fail during setup. GitHub treats
activity performed with that fallback as workflow-generated activity; CI for a
release pull request may therefore remain blocked pending manual approval or
may not be triggered. Do not remove the fallback until the repository secret
has been verified with a release pull request update.

Never store the token value in the repository, workflow, logs, or examples.

## Release Notes

Format user-facing changes for GitHub Releases using these categories:

- Added: new API or capability.
- Changed: behavior changes that users may notice.
- Fixed: bug fixes.
- Deprecated: API that remains available but should be avoided.
- Removed: breaking changes, only for major versions.

## Commit Message Rules

Use Conventional Commits when committing is explicitly requested. Commits must be
atomic: one coherent behavior, documentation, test, or tooling change per
commit. If a change cannot be described by one concise subject, split it before
committing.

```text
<type>(<scope>): <subject>

[optional body]

[optional footer(s)]
```

Supported types:

- `feat:` - new feature; minor version bump.
- `fix:` - bug fix; patch version bump.
- `refactor:` - code refactoring without behavior change; patch version.
- `docs:` - documentation only; patch version.
- `test:` - test changes only; no version bump.
- `chore:` - tooling or maintenance; no version bump.

Breaking changes use `!` after the type or a `BREAKING CHANGE:` footer.

Subject guidance:

- Keep the subject to one short sentence, preferably under 72 characters.
- Describe the user-visible or maintenance outcome, not an implementation list.
- Use the imperative mood when natural: `fix token refresh race`, not
  `fixed token refresh race`.
- Do not enumerate files, tests, or internal steps in the subject.
- Avoid multi-line bodies unless the commit needs rationale, migration notes, or
  a breaking-change footer.
- Match the release-please type to the release impact: `feat` for new capability,
  `fix` for bug fixes, `docs` for documentation-only changes, and `chore` for
  maintenance that should not affect users.

Examples:

```text
fix: add missing input validation for URLs and IDs
feat: add CAS table upload support
docs: split agent guidance into focused docs
feat!: rename PatchIdentitiesLDAPGroup to PatchIdentitiesLDAPUser

BREAKING CHANGE: Rename PatchIdentitiesLDAPGroup to PatchIdentitiesLDAPUser
```

Avoid messages like these:

```text
fix: update client.go, token.go, tests, docs, and README
docs: add lots of documentation and improve things
chore: miscellaneous cleanup
```

## Release Validation

For release preparation, validate release configuration when GoReleaser is
available:

```bash
goreleaser check
```

After the first public tag exists, use Go module compatibility checks before
tagging a release:

```bash
go run golang.org/x/exp/cmd/gorelease@latest -base=latest
```

`gorelease` may fail before the first public version exists; do not treat that
as a project failure before an initial release.
