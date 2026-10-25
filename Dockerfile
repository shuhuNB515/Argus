# syntax=docker/dockerfile:1
#
# 仅运行图像(不要在图像内编译): 仅安装商业工具, **预编译的 Linux 单一二进制**放入.
# 二进制是 CI 的 binaries job 这个交叉编译(纯粹的 Go, QEMU 无), 按目标架构
# 在构建上下文中 dist/<TARGETARCH>/argus 把它放进去. 这允许多架构构建 arm64 是 apt 仅层次结构
# 模仿一下就可以了, Next/Go 由于不再需要模拟，编译速度要快得多。.
#
# 在本地手动构建映像时，请先自行准备二进制文件。:
#   cd web && npm run build:static && cd ..
#   mkdir -p server/webui && cp -r web/out server/webui/dist
#   CGO_ENABLED=0 GOARCH=amd64 go build -tags embedui -o dist/amd64/argus ./cmd/argus
#   docker build -t argus:local .
FROM python:3.12-slim-bookworm
ARG TARGETARCH
# 商业工具: ripgrep / curl / vim 除了 recon 常用工具包(根据需要添加·删除).
# Node 是 NodeSource 在 20.x 安装: bookworm 基本的 apt nodejs 是 18 并且, Playwright 是 >=20 要求.
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates ripgrep curl wget vim git jq unzip \
      dnsutils iputils-ping netcat-openbsd inetutils-telnet whois nmap \
    && curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*
# Playwright MCP 哇 CLI 提前全局安装(在运行时 npx 避免网络下载).
# @playwright/mcp: browser MCP 是 `npx @playwright/mcp` 立即执行(全局安装完成, -y/@latest 不必要的).
# @playwright/cli: playwright-cli 提供, 安装后 --help 检查它是否可执行.
# 下一个 playwright(提供浏览器管理)并安装, --with-deps 通过 chromium 并提前安装了其系统依赖项。.
# 这样，容器内 MCP/CLI 当您第一次启动它时，您可以立即使用它。, 不要通过网络下载浏览器.
RUN npm install -g @playwright/mcp@latest @playwright/cli@latest playwright@latest \
    && playwright-cli --help \
    && playwright install --with-deps chromium \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
# 该架构的预编译二进制文件(dist/amd64/argus 或 dist/arm64/argus)
COPY dist/${TARGETARCH}/argus /app/argus
# 监控启动脚本: 进程终止后，根据退出码决定是否重新启动。, 一键更新屏幕即可完成此脚本的替换。.
# 这个脚本同时 SIGTERM 到 argus 我还负责送货到: docker stop 银色信号 PID 1 因为我只发送给,
# 如果你不转发 argus 因为没有收到信号，所以无法优雅地结束。 10 几秒钟后 SIGKILL 被迫退出.
COPY start.sh /app/start.sh
RUN chmod +x /app/argus /app/start.sh
COPY skills/ /app/skills/
# data/(SQLite + jwt.key) 永久点
VOLUME ["/app/data"]
EXPOSE 8787 8788
ENTRYPOINT ["/app/start.sh"]
CMD ["-addr", ":8787", "-proxy", ":8788"]
