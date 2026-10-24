# SRC Agent

一个面向授权安全研究和 SRC 漏洞验证的本地 Agent 工作台。

> 仅用于你拥有或明确获得授权的目标。不要用于未授权扫描、破坏性测试、拒绝服务、支付操作或账号权限变更。

## 项目功能

- SRC 会话：保存目标、Test Range 和端口范围，并在执行器中统一校验。
- 资产与侦察：资产登记、被动信息整理、指纹识别和范围分析。
- 安全验证：发现结果先进入 potential issue，经过验证、影响证据和质量门禁后才可确认。
- 浏览器取证：只读页面状态、截图、请求观察和证据关联。
- 证据与报告：证据质量、去重、冲突分流，以及 Markdown、HTML、PDF 报告输出。
- Agent 工具链：工具调用、失败恢复、执行事件、任务图和实时工作流。
- AOCI：运行时加载项目索引，帮助 Agent 快速理解当前架构。
- 模型配置：支持 OpenAI-compatible 模型配置、模型测试、备用配置和调用记录。
- 本地化：中文和英文用户界面，中文为默认 README 和默认使用说明。

## 下载

打开 [Releases](https://github.com/shuhuNB515/Argus/releases) 下载对应系统的压缩包。

- Windows Intel/AMD 电脑：`src-agent-*-windows-amd64.zip`
- macOS Intel：`src-agent-*-darwin-amd64.zip`
- macOS Apple Silicon：`src-agent-*-darwin-arm64.zip`
- Linux x86-64：`src-agent-*-linux-amd64.zip`
- Linux ARM64：`src-agent-*-linux-arm64.zip`

这里的 `amd64` 就是常说的 x86-64，Windows x86-64 电脑应下载 Windows amd64 包；ARM64 包只给 ARM 设备使用。

## 本地运行

当前版本使用 PostgreSQL 保存任务、会话、资产、证据和报告数据。最简单的方式是使用 Docker Desktop 启动 PostgreSQL，然后运行本地服务：

```bash
cp .env.example .env
# 编辑 .env，设置 POSTGRES_PASSWORD 和模型 API 配置
./dev.sh
```

浏览器打开 <http://127.0.0.1:8788>。

Docker Compose 只负责本地 PostgreSQL；下载的桌面压缩包仍需要按说明配置数据服务。真正的单机内置数据库模式尚未作为稳定功能发布。

## 从源码构建

```bash
cd web
npm ci
npm run build:static
cd ..
./build.sh --release --no-compress
```

构建脚本会生成 Windows、macOS 和 Linux 的 amd64/arm64 包，并生成 `SHA256SUMS`。

## 开发验证

```bash
go test ./...
cd web && npx tsc --noEmit && npm run check:i18n && npm run build:static
```

## English

See [README.en.md](README.en.md).

## 许可证

本项目沿用 AGPL-3.0。请同时阅读 [LICENSE](LICENSE) 和 [SECURITY.md](SECURITY.md)。
