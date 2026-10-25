# syntax=docker/dockerfile:1
#
# 运行镜像使用 CI 预先构建的二进制，避免在模拟环境中编译 Go 和前端。
# 对应架构的二进制放在构建上下文的 dist/<TARGETARCH>/argus。
#
# 在本地手动构建映像时，请先自行准备二进制文件。:
#   cd web && npm run build:static && cd ..
#   mkdir -p server/webui && cp -r web/out server/webui/dist
#   CGO_ENABLED=0 GOARCH=amd64 go build -tags embedui -o dist/amd64/argus ./cmd/argus
#   docker build -t argus:local .
FROM python:3.12-slim-bookworm
ARG TARGETARCH
ENV ARGUS_HOME=/data ARGUS_NO_BROWSER=1 HOME=/home/argus \
    PLAYWRIGHT_BROWSERS_PATH=/opt/argus-browsers
# 安装常用研究工具及 Playwright 所需的 Node.js。
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates ripgrep curl wget vim git jq unzip \
      dnsutils iputils-ping netcat-openbsd inetutils-telnet whois nmap \
    && curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*
# 构建时安装浏览器工具、Chromium 及其依赖，首次启动无需下载。
RUN npm install -g @playwright/mcp@latest @playwright/cli@latest playwright@latest \
    && playwright-cli --help \
    && playwright install --with-deps chromium \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
# 该架构的预编译二进制文件(dist/amd64/argus 或 dist/arm64/argus)
COPY dist/${TARGETARCH}/argus /app/argus
# 启动器转发停止信号，仅在程序明确请求更新重启时重新启动。
COPY start.sh /app/start.sh
RUN chmod +x /app/argus /app/start.sh
COPY skills/ /app/skills/
COPY scripts/bundle-database.py /tmp/bundle-database.py
RUN python /tmp/bundle-database.py --os linux --arch "$TARGETARCH" --dest /app/database-cache \
    && groupadd --gid 1000 argus \
    && useradd --uid 1000 --gid argus --create-home --shell /bin/sh argus \
    && mkdir -p /data \
    && chown -R argus:argus /app /data /home/argus /opt/argus-browsers \
    && rm /tmp/bundle-database.py
USER argus
# 内置 PostgreSQL、设置和密钥保存在持久化用户数据目录。
VOLUME ["/data"]
HEALTHCHECK --interval=5s --timeout=3s --start-period=90s \
  CMD python -c "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8787/api/health', timeout=2)"
EXPOSE 8787 8788
ENTRYPOINT ["/app/start.sh"]
CMD ["-addr", ":8787", "-proxy", ":8788"]
