#!/usr/bin/env bash
# Argus 本地开发模式：内置数据库 + Go Agent 运行时 + Next.js 前端，无需 Docker。
# 浏览器访问 http://127.0.0.1:8788；Ctrl-C 会停止后端和前端，数据库保留。
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi
export ARGUS_NO_BROWSER=1
export ARGUS_CORS_ORIGINS="${ARGUS_CORS_ORIGINS:-http://127.0.0.1:8788,http://localhost:8788}"
export ARGUS_LANG="${ARGUS_LANG:-zh}"
unset NEXT_PUBLIC_MOCK

backend_pid=""
frontend_pid=""
cleanup() {
  [[ -n "$frontend_pid" ]] && kill "$frontend_pid" 2>/dev/null || true
  [[ -n "$backend_pid" ]] && kill "$backend_pid" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

echo "[Argus] 正在启动真实 Agent 运行时…"
go run ./cmd/argus -addr 127.0.0.1:8787 -proxy 127.0.0.1:8789 &
backend_pid=$!

for _ in {1..60}; do
  if curl -fsS http://127.0.0.1:8787/api/health >/dev/null 2>&1; then
    break
  fi
  if ! kill -0 "$backend_pid" 2>/dev/null; then
    echo "[Argus] Agent 运行时启动失败。" >&2
    exit 1
  fi
  sleep 1
done

if [[ ! -d web/node_modules ]]; then
  echo "[Argus] 首次运行，正在安装前端依赖…"
  (cd web && npm install)
fi

# Next.js 会把 NEXT_PUBLIC_* 编译进缓存。删除旧缓存可避免之前的演示模式残留。
rm -rf "$ROOT/web/.next"
echo "[Argus] 正在启动本地界面…"
(
  cd web
  npm run dev -- --hostname 127.0.0.1 --port 8788
) &
frontend_pid=$!

echo "[Argus] 已启动：http://127.0.0.1:8788"
echo "[Argus] 当前为真实运行模式，对话、执行记录和用量均来自本机数据库。"
wait "$backend_pid" "$frontend_pid"
