# Contributing

Thanks for helping improve `go-viya`.

## Workflow

1. Open an issue for larger changes before implementation.
2. Keep pull requests focused on one behavior change.
3. Add or update tests for user-visible behavior.
4. Run the required checks locally.

## Required Checks

```bash
gofmt -w .
go vet ./...
go test ./...
```

For shared client, authentication, or concurrency changes:

```bash
go test -race ./...
```

## Commit Style

Use Conventional Commits so Release Please can choose the next version and
write release notes:

```text
feat: add batch file set deletion errors
fix: refresh tokens with request context
docs: document CAS table loading
```

Use `feat:` for minor releases, `fix:` for patch releases, and `!` or a
`BREAKING CHANGE:` footer for breaking changes.

## Releases

Release Please manages release pull requests, `CHANGELOG.md`, tags, and GitHub
Releases from commits merged to `main`.

1. Merge changes with Conventional Commit messages.
2. Review and merge the Release Please pull request.
3. The Release Please workflow creates the tag and GitHub Release.
4. The same workflow runs GoReleaser when a release is created.

The manual GoReleaser workflow remains available for explicitly pushed tags.

## Compatibility

This module follows semantic versioning. Do not break exported APIs in minor or patch releases. For release candidates, run:

```bash
go run golang.org/x/exp/cmd/gorelease@latest -base=latest
```
