# 维护说明

本文档记录 ARTEX 的日常维护原则。公开用户文档只提供中文和英文两种版本；新增用户可见文案必须同时更新两种语言。

## 发布前检查

- `go test ./...`
- 前端 TypeScript 检查、国际化键检查和生产构建
- `git diff --check`
- 检查发布包不包含 `.env`、密钥、数据库和本地数据
- 分别验证 Windows、macOS、Linux 的 amd64（x86-64）包；ARM64 作为可选包

## 安全与数据

不得把 API 密钥、数据库文件、用户会话、测试目标数据或运行日志提交到仓库。安全测试功能必须遵守 Target/Test Range 授权边界。

English version: [MAINTAINING.en.md](MAINTAINING.en.md)
