# go-viya

[![CI](https://github.com/dingdayu/go-viya/actions/workflows/ci.yml/badge.svg)](https://github.com/dingdayu/go-viya/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/dingdayu/go-viya.svg)](https://pkg.go.dev/github.com/dingdayu/go-viya)
[![Go Report Card](https://goreportcard.com/badge/github.com/dingdayu/go-viya)](https://goreportcard.com/report/github.com/dingdayu/go-viya)
[![中文文档](https://img.shields.io/badge/docs-中文-blue)](README.zh.md)

`go-viya` is a Go client library for selected SAS Viya REST APIs. It follows the REST protocols and media types documented at <https://developer.sas.com/rest-apis>, and provides token providers, a Resty-backed client, and helpers for identities, configuration, batch, CAS table operations, files, and Job Execution jobs.

For AI agents and tooling, you can copy the prompt below and send it directly to your AI Agent:

```text
You are working on the open-source repository `go-viya`.

Repository: https://github.com/dingdayu/go-viya

Please read the repository guide in `llms.txt` first:
- GitHub raw URL: https://raw.githubusercontent.com/dingdayu/go-viya/main/llms.txt
- jsDelivr CDN URL: https://cdn.jsdelivr.net/gh/dingdayu/go-viya@main/llms.txt

Use the guide to understand the project scope, public API, examples, and development constraints.
Then implement the requested change by following the existing code style, keeping the API small and tested, and preserving compatibility with the current exported behavior.

If the request is ambiguous, inspect the README, examples, and existing tests before making changes.
```

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
	"net/url"

	"github.com/dingdayu/go-viya"
)

func main() {
	ctx := context.Background()
	baseURLStr := "https://viya.example.com"
	clientID := "client-id"
	clientSecret := "client-secret"

	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		log.Fatal(err)
	}

	tokens, err := viya.NewClientCredentialsTokenProvider(clientID, clientSecret, baseURL)
	if err != nil {
		log.Fatal(err)
	}

	client := viya.NewClient(ctx, viya.WithBaseURL(baseURL), viya.WithTokenProvider(tokens))

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

- `NewClientCredentialsTokenProvider(clientID, clientSecret string, baseURL *url.URL)`
- `NewPasswordTokenProvider(username, password string, baseURL *url.URL, opts ...TokenProviderOption)`
- `NewAuthCodeTokenProvider(code string, baseURL *url.URL, opts ...TokenProviderOption)`

Password and authorization-code flows can reuse OAuth client settings:

```go
baseURL, _ := url.Parse("https://viya.example.com")
provider, err := viya.NewPasswordTokenProvider(
	"username",
	"password",
	baseURL,
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
baseURLOpt, err := viya.ParseURL(baseURL)
if err != nil {
	panic(err)
}
client := viya.NewClient(ctx, baseURLOpt, viya.WithTokenProvider(provider))
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
- [`examples/viya-cli`](#viya-cli): CLI for agents to execute SAS code, discover and manage CAS data, use Viya files, and submit Job Execution jobs.

### viya-cli

`viya-cli` is a CLI tool for executing SAS code on SAS Viya, discovering and managing
CAS data assets, using the Viya Files Service, submitting Job Execution jobs, and
orchestrating Compute-based workflow plans. It is designed for agent frameworks,
CI/CD pipelines, and interactive use.

```bash
# Install from repository root
go install ./examples/viya-cli

# Run SAS code inline
viya-cli run --code "data _null_; put 'hello from viya-cli'; run;"

# Run a local SAS program
viya-cli run --file ./program.sas

# Keep the Compute session after execution
viya-cli run --file ./program.sas --keep-session

# Discover CAS servers
viya-cli cas servers

# List Viya files
viya-cli files list --limit 50

# Submit a Job Execution job
viya-cli jobs submit --code "proc options; run;" --name options-check

# Validate and run a workflow plan
viya-cli workflow validate --file ./examples/workflow.yaml
viya-cli workflow run --file ./examples/workflow.yaml -o json
```

Use `-o json` for machine-readable output. See
[examples/viya-cli/README.md](examples/viya-cli/README.md) for complete
documentation, configuration, and agent integration guidance.

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
  - Discover CAS servers, caslibs, tables, columns, and sample rows.
  - Upload CSV data into CAS tables.
  - Promote CAS tables to global scope.
- Files:
  - List, upload, and download SAS Viya Files Service files.
- Job Execution:
  - Submit SAS code as an asynchronous Job Execution job.
  - List, inspect, cancel, and retrieve logs for Job Execution jobs.
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

For release notes and commit conventions, see [docs/release.md](docs/release.md).

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).
