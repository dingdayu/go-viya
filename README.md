# go-viya

[![CI](https://github.com/dingdayu/go-viya/actions/workflows/ci.yml/badge.svg)](https://github.com/dingdayu/go-viya/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/dingdayu/go-viya.svg)](https://pkg.go.dev/github.com/dingdayu/go-viya)
[![Go Report Card](https://goreportcard.com/badge/github.com/dingdayu/go-viya)](https://goreportcard.com/report/github.com/dingdayu/go-viya)

`go-viya` is a Go client library for selected SAS Viya REST APIs. It provides token providers, a Resty-backed client, and helpers for identities, configuration, batch, and CAS table operations.

## Installation

```bash
go get github.com/dingdayu/go-viya
```

## Quick Start

```go
package main

import (
	"context"
	"log"

	"github.com/dingdayu/go-viya"
)

func main() {
	ctx := context.Background()
	baseURL := "https://viya.example.com"

	tokens, err := viya.NewClientCredentialsTokenProvider(
		baseURL,
		"client-id",
		"client-secret",
	)
	if err != nil {
		log.Fatal(err)
	}

	client := viya.NewClient(ctx, baseURL, viya.WithTokenProvider(tokens))

	users, err := client.GetIdentitiesUsers(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("users: %d", users.Count)
}
```

## Authentication

The client accepts any implementation of:

```go
type TokenProvider interface {
	Token(ctx context.Context) (string, error)
}
```

Built-in providers:

- `NewClientCredentialsTokenProvider(baseURL, clientID, clientSecret)`
- `NewPasswordTokenProvider(baseURL, username, password, opts...)`
- `NewAuthCodeTokenProvider(baseURL, code, opts...)`

Password and authorization-code flows can reuse OAuth client settings:

```go
provider, err := viya.NewPasswordTokenProvider(
	baseURL,
	"username",
	"password",
	viya.WithOAuthClient("client-id", "client-secret"),
)
```

## Supported Areas

- Identities: refresh identities cache and list users.
- Configuration: read configuration definitions.
- Batch: inspect contexts, file sets, files, jobs, and related resources.
- CAS: load and unload CAS library tables.
- OpenTelemetry: outbound token requests and client operations are instrumented with spans.

The API surface is intentionally small and grows around tested SAS Viya workflows.

## Development

```bash
go test ./...
go test -race ./...
go vet ./...
```

Before opening a pull request:

```bash
gofmt -w .
go mod tidy
go test ./...
```

## Releases

This project uses semantic versioning and Git tags in the form `vX.Y.Z`.

Create a release by pushing a tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow runs tests and then uses GoReleaser to publish a GitHub Release with generated release notes.

For local release validation:

```bash
goreleaser check
goreleaser release --snapshot --clean
```

For Go module compatibility checks before a public release:

```bash
go run golang.org/x/exp/cmd/gorelease@latest -base=latest
```

`gorelease` may fail before the first public version exists; after the first tag, use it to catch accidental API breakage.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).
