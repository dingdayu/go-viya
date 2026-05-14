# go-viya

[![CI](https://github.com/dingdayu/go-viya/actions/workflows/ci.yml/badge.svg)](https://github.com/dingdayu/go-viya/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/dingdayu/go-viya.svg)](https://pkg.go.dev/github.com/dingdayu/go-viya)
[![Go Report Card](https://goreportcard.com/badge/github.com/dingdayu/go-viya)](https://goreportcard.com/report/github.com/dingdayu/go-viya)
[![English](https://img.shields.io/badge/docs-English-blue)](README.md)

`go-viya` 是一个精选的 SAS Viya REST API 的 Go 客户端库。它遵循 <https://developer.sas.com/rest-apis> 中记录的 REST 协议和媒体类型，提供 token 提供者、基于 Resty 的客户端，以及身份、配置、批处理、CAS 表操作、文件和作业执行等辅助功能。

对于 AI 代理和工具，可以复制以下提示直接发送给您的 AI 代理：

```text
您正在处理开源仓库 `go-viya`。

仓库：https://github.com/dingdayu/go-viya

请先阅读 `llms.txt` 中的仓库指南：
- GitHub 原始 URL：https://raw.githubusercontent.com/dingdayu/go-viya/main/llms.txt
- jsDelivr CDN URL：https://cdn.jsdelivr.net/gh/dingdayu/go-viya@main/llms.txt

使用该指南了解项目范围、公共 API、示例和开发约束。
然后按照现有代码风格实现所需的更改，保持 API 小巧且经过测试，并保持与当前导出行为的兼容性。

如果请求存在歧义，请先检查 README、示例和现有测试再进行更改。
```

## 安装

```bash
go get github.com/dingdayu/go-viya
```

## 快速开始

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

	log.Printf("用户数：%d", users.Count)
}
```

## 认证

客户端接受任何实现以下接口的对象：

```go
type TokenProvider interface {
	Token(ctx context.Context) (string, error)
}
```

内置提供者：

- `NewClientCredentialsTokenProvider(clientID, clientSecret string, baseURL *url.URL)`
- `NewPasswordTokenProvider(username, password string, baseURL *url.URL, opts ...TokenProviderOption)`
- `NewAuthCodeTokenProvider(code string, baseURL *url.URL, opts ...TokenProviderOption)`

密码和授权码流程可以复用 OAuth 客户端设置：

```go
baseURL, _ := url.Parse("https://viya.example.com")
provider, err := viya.NewPasswordTokenProvider(
	"username",
	"password",
	baseURL,
	viya.WithOAuthClient("client-id", "client-secret"),
)
```

### 分布式服务

内置提供者在当前 Go 进程中缓存和刷新令牌。这适用于命令行工具、测试和简单的服务，但不适用于分布式令牌缓存。

对于多实例部署，请在应用程序中实现 `TokenProvider`，并将刷新令牌处理放在您自己的操作边界内。典型的实现是从共享缓存或内部认证服务读取有效的访问令牌，在过期前使用分布式锁刷新令牌，并将刷新令牌存储在 Vault、KMS 支持的存储或您平台的密钥存储等机密管理器中。

`go-viya` 仅请求持有者访问令牌。它不暴露刷新令牌，因为刷新令牌的存储、轮换、撤销、加密、审计、租户隔离和跨实例锁定是与部署相关的安全关注点。

```go
type DistributedTokenProvider struct {
	cache SharedTokenCache
}

func (p DistributedTokenProvider) Token(ctx context.Context) (string, error) {
	token, err := p.cache.AccessToken(ctx)
	if err != nil {
		return "", fmt.Errorf("viya 访问令牌：%w", err)
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

参见 `examples/` 获取完整的自定义提供者和工作流示例。

## 示例

- `examples/client-credentials`：使用 OAuth2 客户端凭据创建客户端并列出身份用户。
- `examples/password-flow`：在 SAS Logon 允许时使用 OAuth2 密码授权。
- `examples/distributed-token-provider`：将 `go-viya` 连接到应用管理的共享令牌缓存。
- `examples/configuration`：读取动态 SAS Viya 配置定义。
- `examples/default-client`：配置和获取进程级默认客户端。
- `examples/batch-job`：创建文件集、上传 SAS 程序、提交批处理作业并等待完成。
- `examples/cas-table-state`：加载并可选地卸载 CAS 表。
- [`examples/viya-cli`](#viya-cli)：用于代理执行 SAS 代码、发现和管理 CAS 数据、使用 Viya 文件以及提交作业执行作业的 CLI。

### viya-cli

`viya-cli` 是一个 CLI 工具，用于在 SAS Viya 上执行 SAS 代码、发现和管理 CAS 数据资产、使用 Viya 文件服务、提交作业执行作业以及编排基于 Compute 的工作流计划。它专为代理框架、CI/CD 流水线和交互式使用而设计。

```bash
# 从仓库根目录安装
go install ./examples/viya-cli

# 内联运行 SAS 代码
viya-cli run --code "data _null_; put 'hello from viya-cli'; run;"

# 运行本地 SAS 程序
viya-cli run --file ./program.sas

# 执行后保持 Compute 会话
viya-cli run --file ./program.sas --keep-session

# 发现 CAS 服务器
viya-cli cas servers

# 列出 Viya 文件
viya-cli files list --limit 50

# 提交作业执行作业
viya-cli jobs submit --code "proc options; run;" --name options-check

# 验证和运行工作流计划
viya-cli workflow validate --file ./examples/workflow.yaml
viya-cli workflow run --file ./examples/workflow.yaml -o json
```

使用 `-o json` 获取机器可读输出。请参阅
[examples/viya-cli/README.md](examples/viya-cli/README.md) 获取完整文档、配置和代理集成指南。

## API 基础

此包基于公开的 SAS Viya REST API 文档实现：

- SAS Viya REST API：<https://developer.sas.com/rest-apis>
- SAS Logon API：<https://developer.sas.com/rest-apis/SASLogon>
- 批处理 API：<https://developer.sas.com/rest-apis/batch>
- Compute API：<https://developer.sas.com/rest-apis/compute>

API 面设计得较小，并围绕经过测试的 SAS Viya 工作流逐步扩展。它不是为每个 SAS Viya 端点生成的客户端。

## 支持的功能

当前实现的功能包括：

- 认证：
  - OAuth2 客户端凭据令牌提供者。
  - OAuth2 密码令牌提供者。
  - OAuth2 授权码令牌提供者。
  - 自定义 `TokenProvider` 支持。
- 默认客户端配置：
  - 设置、获取和必须获取进程级默认客户端的辅助函数。
- 身份：
  - 刷新身份缓存。
  - 列出身份用户。
  - 读取 LDAP 用户配置。
  - 修补 LDAP 组配置。
  - 根据用户名更新 LDAP 对象过滤器。
- 配置：
  - 读取配置定义。
- 批处理：
  - 列出批处理上下文并按名称检查上下文。
  - 列出、创建、检查和删除批处理文件集。
  - 列出、检查、下载和上传批处理文件集中的文件。
  - 列出、创建、检查、删除、取消、等待并获取批处理作业的状态/输出。
  - 向正在运行的批处理作业发送 STDIN。
  - 列出、检查和删除可复用的批处理服务器。
- Compute：
  - 列出和检查 Compute 上下文。
  - 列出、创建、检查、取消和删除 Compute 会话。
  - 列出、创建、检查、取消、删除并获取 Compute 作业的状态。
  - 以集合或纯文本形式获取 Compute 作业日志和列表输出。
- CAS：
  - 将 CAS 库表加载到内存。
  - 从内存卸载 CAS 库表。
  - 发现 CAS 服务器、casiibs、表、列和示例行。
  - 将 CSV 数据上传到 CAS 表。
  - 将 CAS 表提升到全局范围。
- 文件：
  - 列出、上传和下载 SAS Viya 文件服务文件。
- 作业执行：
  - 将 SAS 代码作为异步作业执行作业提交。
  - 列出、检查、取消和获取作业执行作业的日志。
- 可观测性：
  - 对外发令牌请求和客户端操作的 OpenTelemetry 链路。

## 开发

```bash
go test ./...
go test -race ./...
go vet ./...
```

在发起拉取请求之前：

```bash
gofmt -w .
go mod tidy
go test ./...
```

有关发布说明和提交约定，请参见 [docs/release.md](docs/release.md)。

## 许可证

本项目采用 MIT 许可证。请参见 [LICENSE](LICENSE)。
