# Argus

Argus 是面向授权安全研究和 SRC 漏洞验证的本地 Agent 工作台。它把目标范围、资产、工具调用、浏览器观察、证据、发现和报告放在同一条可追踪链路中。

## 安装使用

1. 下载与你电脑匹配的安装包并解压或安装。
2. macOS 双击 DMG，将 Argus 拖入“应用程序”；Windows 运行 `argus-*-windows-amd64-setup.exe`；Linux 解压后运行 `start.sh`。
3. 应用启动后打开 <http://127.0.0.1:8787>。

macOS 应用会把数据保存到当前用户的 `~/Library/Application Support/Argus`，不需要给应用目录写权限，也不需要 Docker。

Argus 自带本地数据库运行时。用户不需要安装 Docker、配置 PostgreSQL、创建数据库或填写数据库密码。第一次启动会自动准备对应系统的数据库运行文件，之后会保存在 Argus 数据目录中并持续复用。

Windows 安装包和便携 ZIP 已包含数据库运行文件，首次启动无需联网下载数据库。其他平台首次启动如需下载数据库运行文件，请保持网络连接。

## 支持的平台

- Windows x86-64：`argus-*-windows-amd64-setup.exe`（也提供便携 ZIP）
- macOS Intel：`argus-*-darwin-amd64.zip`
- macOS Apple Silicon：`argus-*-darwin-arm64.zip`
- Linux x86-64：`argus-*-linux-amd64.zip`
- Linux ARM64：`argus-*-linux-arm64.zip`

`amd64` 就是常见的 x86-64；普通 Intel/AMD Windows 电脑下载 Windows amd64 包。Apple Silicon Mac（M1/M2/M3/M4/M5）下载 `darwin-arm64`，Intel Mac 下载 `darwin-amd64`。下载后建议使用发布页的 `SHA256SUMS` 校验文件完整性。

## 工作流程

创建 SRC 会话后，先填写 Target 和 Test Range，再依次执行 Recon、Fingerprint、Asset Analysis、Discovery、非破坏性 Validation、Impact、Browser Evidence、Evidence Quality、Finding Correlation 和 Report Quality Gate。未经过证据和质量门禁的结果不会被标记为 confirmed vulnerability。

## 模型配置

首次打开后，在 Argus 的模型设置中填写你自己的 OpenAI-compatible API 地址和 Key。API Key 只保存在本机数据目录，不会写入安装包或 Git 仓库。

## 安全边界

只测试你拥有或明确获得授权的目标。不要对第三方系统进行未授权扫描、破坏性测试、拒绝服务、支付操作或账号权限变更。

## English

Argus is a local Agent workspace for authorized security research and SRC validation.

1. Download and extract the package for your platform.
2. On macOS, double-click `Argus.app`; if you downloaded a DMG, drag Argus to Applications and launch it there. Run the Windows `argus-*-windows-amd64-setup.exe` installer or use `start.sh` on Linux.
3. Open <http://127.0.0.1:8787> after the app starts.

The macOS app stores data in `~/Library/Application Support/Argus`, so it can run from Applications without write access to the app bundle and without Docker.

Argus starts its own local database runtime. Docker, PostgreSQL installation, database creation, and database passwords are not required. The first launch prepares the platform database runtime and caches it in the Argus data directory for later launches.

Windows installers and portable ZIPs include the database runtime for offline startup. On other platforms, keep the machine online if the first launch needs to download the runtime.

Configure your own OpenAI-compatible model endpoint and API key in Argus settings. Keys stay in the local data directory and are not included in packages or source control.

Use Argus only against systems you own or are explicitly authorized to test. Do not perform unauthorized scanning, destructive testing, denial of service, payment actions, or account/permission changes.
