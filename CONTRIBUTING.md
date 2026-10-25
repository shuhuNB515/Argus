# 贡献指南

感谢你对 Argus 的关注。提交代码、文档或问题前，请先阅读 [README.md](README.md) 和 [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)。

## 提交问题

- 使用 GitHub Issue 模板描述问题、复现步骤和运行环境。
- 不要在公开 Issue 中发布 API 密钥、密码、个人数据或未修复的安全漏洞。
- 安全漏洞请按照 [SECURITY.md](SECURITY.md) 的私密流程报告。

## 提交代码

1. 保持改动范围清晰，并补充必要的测试。
2. 用户可见文案只新增中文或英文，并同步更新两种语言资源。
3. 不要提交 `.env`、密钥、数据库文件、构建产物或个人数据。
4. 提交前运行相关 Go 测试、前端类型检查和生产构建。

## 安全边界

Argus 只能用于自有资产、明确授权的测试范围或本地隔离靶场。禁止对第三方系统进行未授权扫描、破坏性测试、拒绝服务或凭据攻击。

English version: [CONTRIBUTING.en.md](CONTRIBUTING.en.md)
