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

Use short, imperative commit messages:

```text
Add batch file set deletion errors
Fix token provider refresh race
Document CAS table loading
```

## Compatibility

This module follows semantic versioning. Do not break exported APIs in minor or patch releases. For release candidates, run:

```bash
go run golang.org/x/exp/cmd/gorelease@latest -base=latest
```
