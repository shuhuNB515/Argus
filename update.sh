#!/usr/bin/env bash
# Argus 更新脚本：支持 Docker 部署更新和本地源码重新编译。
# install.sh 用于首次安装，update.sh 用于升级。数据库迁移会在启动时自动执行，
# 已有的数据库卷、data 目录和 skills 目录不会被删除。
set -euo pipefail
cd "$(cd "$(dirname "$0")" && pwd)"

info(){ printf '\033[36m[*]\033[0m %s\n' "$*"; }
ok(){   printf '\033[32m[+]\033[0m %s\n' "$*"; }
warn(){ printf '\033[33m[!]\033[0m %s\n' "$*"; }
die(){  printf '\033[31m[x]\033[0m %s\n' "$*" >&2; exit 1; }
ask(){  local p="$1" d="${2:-}" a; read -rp "$p${d:+ [$d]}: " a; echo "${a:-$d}"; }

# 可选：将当前仓库同步到最新代码。
sync_repo(){
  [ -d .git ] && command -v git >/dev/null 2>&1 || { warn "当前目录不是 Git 工作副本，跳过代码同步"; return; }
  [ "$(ask '是否获取最新代码（git pull --ff-only）？(y/n)' y)" = y ] || return
  if ! git pull --ff-only; then
    warn "无法快进更新，可能存在本地修改或分支已分叉；本次继续使用当前代码"
  fi
}

# ── ① Docker 更新 ───────────────────────────────
update_docker(){
  command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1 \
    || die "未找到 Docker 或 Docker Compose，请先运行 ./install.sh 完成部署"
  [ -f .env ] || die "未找到 .env，请先运行 ./install.sh 完成首次部署"

  # 可指定目标版本标签；留空时沿用 .env 中的 ARGUS_TAG，默认 latest。
  local tag; tag="$(ask '目标镜像标签（直接回车沿用 .env/latest）' '')"
  if [ -n "$tag" ]; then
    if grep -q '^ARGUS_TAG=' .env; then
      sed -i.bak "s|^ARGUS_TAG=.*|ARGUS_TAG=${tag}|" .env && rm -f .env.bak
    else
      printf '\nARGUS_TAG=%s\n' "$tag" >> .env
    fi
    ok "ARGUS_TAG 已设置为 ${tag}"
  fi

  # 只更新 Argus 服务；PostgreSQL 版本固定，避免无意义下载和主版本兼容风险。
  info "正在拉取 Argus 镜像…"
  if ! docker compose pull argus; then
    warn "无法拉取 ghcr.io/shuhunb515/argus，请检查版本标签和网络"
    die "镜像更新失败；也可以选择本地源码重新编译"
  fi
  info "正在重新创建并启动 Argus（启动时自动迁移数据库）…"
  docker compose up -d argus
  ok "更新完成 → http://localhost:8787"
  info "查看日志：docker compose logs -f argus"
  info "可选清理旧镜像：docker image prune -f"
}

# ── ② 本地编译更新 ──────────────────────────────
update_local(){
  command -v go >/dev/null 2>&1 || die "未找到 Go（需要 1.26 或更高版本）：https://go.dev/dl/"
  [ -f config.json ] || warn "未找到 config.json；首次部署请先运行 ./install.sh"
  ok "Go: $(go version)"

  if command -v npm >/dev/null 2>&1; then
    info "正在重建前端静态文件…"
    ( cd web && npm ci && npm run build:static )
    rm -rf server/webui/dist && cp -r web/out server/webui/dist
    info "正在编译包含前端的单文件程序…"
    CGO_ENABLED=0 go build -tags embedui -trimpath -o argus ./cmd/argus
  else
    warn "未找到 npm，仅编译后端；前端需要单独运行 npm run dev"
    CGO_ENABLED=0 go build -o argus ./cmd/argus
  fi
  ok "编译完成 → ./argus"
  warn "请重启 Argus 进程以应用更改（启动时会自动迁移数据库）"
}

echo "=============================="
echo "  ARGUS 更新"
echo "  1) 更新 Docker 部署"
echo "  2) 本地源码重新编译"
echo "=============================="
case "$(ask '请选择' 1)" in
  1) sync_repo; update_docker ;;
  2) sync_repo; update_local ;;
  *) die "无效的选择" ;;
esac
