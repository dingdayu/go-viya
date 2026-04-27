# go-viya

[![CI](https://github.com/dingdayu/go-viya/actions/workflows/ci.yml/badge.svg)](https://github.com/dingdayu/go-viya/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/dingdayu/go-viya.svg)](https://pkg.go.dev/github.com/dingdayu/go-viya)
[![Go Report Card](https://goreportcard.com/badge/github.com/dingdayu/go-viya)](https://goreportcard.com/report/github.com/dingdayu/go-viya)

`go-viya` is a Go client library for selected SAS Viya REST APIs. It follows the REST protocols and media types documented at <https://developer.sas.com/rest-apis>, and provides token providers, a Resty-backed client, and helpers for identities, configuration, batch, and CAS table operations.

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

### Distributed services

The built-in providers cache and refresh tokens in the current Go process. This
is suitable for command-line tools, tests, and simple services, but it is not a
distributed token cache.

For multi-instance deployments, implement `TokenProvider` in your application
and keep refresh-token handling behind your own operational boundary. A typical
implementation reads a valid access token from a shared cache or internal
authentication service, refreshes it with a distributed lock before expiry, and
stores refresh tokens in a secret manager such as Vault, KMS-backed storage, or
your platform's secret store.

`go-viya` intentionally asks only for a bearer access token. It does not expose
refresh tokens, because refresh-token storage, rotation, revocation, encryption,
auditing, tenant isolation, and cross-instance locking are deployment-specific
security concerns.

```go
type DistributedTokenProvider struct {
	cache SharedTokenCache
}

func (p DistributedTokenProvider) Token(ctx context.Context) (string, error) {
	token, err := p.cache.AccessToken(ctx)
	if err != nil {
		return "", fmt.Errorf("viya access token: %w", err)
	}
	if token == "" {
		return "", viya.ErrViyaAuthFailed
	}
	return token, nil
}

provider := DistributedTokenProvider{cache: cache}
client := viya.NewClient(ctx, baseURL, viya.WithTokenProvider(provider))
```

See `examples/` for complete custom provider and workflow examples.

## Examples

- `examples/client-credentials`: create a client with OAuth2 client credentials and list identity users.
- `examples/password-flow`: use the OAuth2 password grant when SAS Logon allows it.
- `examples/distributed-token-provider`: connect `go-viya` to an application-managed shared token cache.
- `examples/configuration`: read a dynamic SAS Viya configuration definition.
- `examples/default-client`: configure and retrieve the process-wide default client.
- `examples/batch-job`: create a file set, upload a SAS program, submit a batch job, and wait for completion.
- `examples/cas-table-state`: load and optionally unload a CAS table.

## API Basis

This package is implemented against the public SAS Viya REST API documentation:

- SAS Viya REST APIs: <https://developer.sas.com/rest-apis>
- SAS Logon API: <https://developer.sas.com/rest-apis/SASLogon>
- Batch API: <https://developer.sas.com/rest-apis/batch>
- Compute API: <https://developer.sas.com/rest-apis/compute>

The API surface is intentionally small and grows around tested SAS Viya workflows. It is not a generated client for every SAS Viya endpoint.

## Supported Features

Current implemented areas include:

- Authentication:
  - OAuth2 client credentials token provider.
  - OAuth2 password token provider.
  - OAuth2 authorization-code token provider.
  - Custom `TokenProvider` support.
- Default client wiring:
  - Set, get, and must-get helpers for a process-wide default client.
- Identities:
  - Refresh the identities cache.
  - List identity users.
  - Read LDAP user configuration.
  - Patch LDAP group configuration.
  - Update LDAP object filters from usernames.
- Configuration:
  - Read configuration definitions.
- Batch:
  - List batch contexts and inspect contexts by name.
  - List, create, inspect, and delete batch file sets.
  - List, inspect, download, and upload files in batch file sets.
  - List, create, inspect, delete, cancel, wait for, and retrieve state/output for batch jobs.
  - Send STDIN to running batch jobs.
  - List, inspect, and delete reusable batch servers.
- Compute:
  - List and inspect compute contexts.
  - List, create, inspect, cancel, and delete compute sessions.
  - List, create, inspect, cancel, delete, and retrieve state for compute jobs.
  - Retrieve compute job log and listing output as collections or plain text.
- CAS:
  - Load CAS library tables to memory.
  - Unload CAS library tables from memory.
- Observability:
  - OpenTelemetry spans for outbound token requests and client operations.

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
