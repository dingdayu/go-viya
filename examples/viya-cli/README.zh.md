# viya-cli

[![English](https://img.shields.io/badge/docs-English-blue)](README.md)

`viya-cli` 是一个小巧的 CLI 示例，适用于需要在 SAS Viya 上执行 SAS 代码、发现和管理 CAS 数据资产、使用 Viya 文件、提交异步作业执行作业以及编排基于 Compute 的工作流计划的代理框架。

它通过命令行标志、环境变量和本地 SAS CLI 风格的文件读取配置：

- `~/.sas/config.json`
- `~/.sas/credentials.json`
- `~/.viya-cli/workflow.yaml` 或 `~/.viya-cli/workflow.json`（用户级工作流默认值）

解析器接受简单的顶级键和常见的配置文件容器，如 `profiles`、`contexts` 和 `credentials`。

项目工作流文件与用户工作流配置是分开的：

- 项目工作流文件描述一次运行、其 DAG 形状、Compute 上下文默认值、文件路径和输出产物。
- 用户工作流配置提供用户级的 Compute 上下文默认值、`autoexec`、`preCode`、`postCode` 和非机密变量。
- 项目工作流的 `contextId`/`contextName` 会覆盖该运行的用户工作流上下文默认值。

## 安装

从仓库根目录：

```bash
go install ./examples/viya-cli
```

或者不安装直接运行：

```bash
go run ./examples/viya-cli run --code "data _null_; put 'hello from viya-cli'; run;"
```

## 配置

环境变量会覆盖 `~/.sas` 文件：

```bash
export VIYA_BASE_URL="https://example.viya.sas.com"
export VIYA_CLIENT_ID="go-viya"
export VIYA_USERNAME="my-user"
export VIYA_PASSWORD="my-password"
export VIYA_COMPUTE_CONTEXT_NAME="SAS Job Execution compute context"
```

支持的认证输入：

- `VIYA_ACCESS_TOKEN` / `SAS_ACCESS_TOKEN`
- `VIYA_CLIENT_ID` 和 `VIYA_CLIENT_SECRET`
- `VIYA_USERNAME` 和 `VIYA_PASSWORD`，可选配 `VIYA_CLIENT_ID` 和 `VIYA_CLIENT_SECRET`

支持的配置输入：

- `VIYA_BASE_URL`、`SAS_VIYA_URL`、`SAS_SERVICES_ENDPOINT` 或 `SAS_BASE_URL`
- `VIYA_COMPUTE_CONTEXT_ID` 或 `VIYA_COMPUTE_CONTEXT_NAME`
- `VIYA_PROFILE` 或 `SAS_PROFILE`

等效的值也可以通过标志传递，如 `-base-url`、`-context-id`、`-context-name`、`-username`、`-password` 和 `-access-token`。

工作流命令标志接受相同的连接设置，外加 `--max-parallel`、`--keep-session`、`--include-output` 和 `--user-config`。

## 用法

执行内联 SAS 代码：

```bash
viya-cli run --code "data _null_; put 'hello'; run;"
```

执行本地 SAS 程序：

```bash
viya-cli run --file ./program.sas
```

从标准输入读取 SAS 代码：

```bash
cat ./program.sas | viya-cli run --file -
```

执行后保持 Compute 会话：

```bash
viya-cli run --file ./program.sas --keep-session
```

命令在失败时以非零状态退出。默认输出控制台友好的文本。当代理或脚本需要完整的机器可读结果时，使用 `-o json`，包括 Compute 上下文、会话、作业、最终状态、条件代码、日志文本和列表文本。

发现 CAS 服务器、casiibs、表、列和示例行：

```bash
viya-cli cas servers
viya-cli cas caslibs --server cas-shared-default
viya-cli cas tables --server cas-shared-default --caslib Public
viya-cli cas table-info --server cas-shared-default --caslib Public --table HMEQ
viya-cli cas columns --server cas-shared-default --caslib Public --table HMEQ
viya-cli cas rows --server cas-shared-default --caslib Public --table HMEQ --limit 25
```

CAS 发现命令默认输出表格文本。使用 `-o json` 在成功时输出 `{ "ok": true, "data": ... }`，在失败时输出 `{ "ok": false, "error": "..." }`。

上传和提升 CAS 数据：

```bash
viya-cli data upload-csv --server cas-shared-default --caslib Public --table HMEQ_UPLOAD --file ./hmeq.csv
cat ./hmeq.csv | viya-cli data upload-csv --server cas-shared-default --caslib Public --table HMEQ_UPLOAD --file -
viya-cli data promote --server cas-shared-default --caslib Public --table HMEQ_UPLOAD
```

使用 Viya 文件服务：

```bash
viya-cli files list --limit 50
viya-cli files list --filter-name report
viya-cli files upload --name report.txt --file ./report.txt --content-type text/plain
viya-cli files download --id file-id
```

提交和检查作业执行服务作业：

```bash
viya-cli jobs submit --code "proc options; run;" --name options-check
viya-cli jobs submit --file ./program.sas --context-name "SAS Job Execution compute context"
viya-cli jobs list --limit 20
viya-cli jobs status --id job-id
viya-cli jobs log --id job-id
viya-cli jobs cancel --id job-id
```

验证和运行工作流计划：

```bash
viya-cli workflow validate --file ./examples/workflow.yaml
viya-cli workflow run --file ./examples/workflow.yaml
viya-cli workflow run --file ./examples/workflow.json -o json
```

工作流文件使用简单的结构：

- 顶层的 `steps` 按顺序运行。
- `steps` 内的嵌套数组并行运行其子工作项；同一并行组中的每个工作项会使用独立的 Compute session，以降低代码域冲突风险。
- 每个工作项可以定义 `name`、`file`/`work`/`path`、`code`、`log`、`listing` 和 `variables`。
- `file`、`work` 和 `path` 相对于工作流文件目录解析。
- 基于文件的工作项还会接收 `_SASPROGRAMFILE` 和 `_SASPROGRAMDIR`，以便 SAS 代码可以像 SAS VS Code 扩展那样推导相对路径。
- `log` 和 `listing` 是相对于工作流文件目录的输出产物路径。

支持字段请参见 `workflow.schema.json` 和 `workflow-user.schema.json`。

## 代理集成

代理应将此 CLI 视为一个具有单一主同步执行操作的工具：

```bash
viya-cli run --file path/to/program.sas
```

对于生成的短程序使用 `-code`，当代理通过标准输入流式传输代码时使用 `-file -`。使用 `-o json` 进行机器解析，然后检查：

- `ok`
- `state`
- `jobConditionCode`
- `log`
- `listing`
- `error`

当用户明确希望异步作业执行服务行为（而非 Compute 会话运行）时，使用 `viya-cli jobs submit`。当用户希望使用一个 Compute 会话和多个 Compute 作业运行可重复的项目工作流时，使用 `viya-cli workflow run`。

CLI 有意不打印机密信息，设计为可从现代代理框架、Shell 工具或 MCP 风格的包装器中调用。

## 代理技能

此目录包含一个代理技能指南 `SKILLS.md`。请将其保留在此仓库中，作为 `viya-cli` 的文档化代理工作流。

剩余的 sas-mcp-server 替换工作跟踪在 `todo.md` 中。此示例的运行时执行仅使用 Go，通过 `viya-cli` 完成。
