#!/usr/bin/env bash
# =============================================================================
# Argus 管理员密码重置脚本
#
# 登录用户名固定为 Argus。密码以 bcrypt 哈希存储在 settings 表的
# auth.password_hash 键中。本脚本连接数据库，使用 pgcrypto 生成新的哈希并写回该键。
# 后端登录校验（golang.org/x/crypto/bcrypt）与此格式完全兼容。
#
# 两种部署模式：
#   local（默认）-- 从主机使用 psql 连接数据库。连接信息优先级为：
#                  命令行参数 > --dsn/$ARGUS_PG_DSN > config.json 的 database.*
#   docker       -- 通过 docker compose exec（或 docker exec）在 PostgreSQL 容器中运行 psql。
#
# 用法示例:
#   ./reset-password.sh                          # 本地的, config.json/自动读取您的环境并以交互方式输入新密码
#   ./reset-password.sh -p 'NewPass!'            # 本地的, 自己指定一个新密码
#   ./reset-password.sh --dsn postgres://u:p@h:5432/argus
#   ./reset-password.sh -H 127.0.0.1 -P 5433 -U autopentest -W pass -d argus
#   ./reset-password.sh -m docker                # docker 分布(.env 的 POSTGRES_* 读)
#   ./reset-password.sh -m docker -c postgres --exec docker
#
# 安全：新密码通过环境变量和 psql \getenv 传递，不会进入进程参数；SQL 值使用
# psql 变量安全转义。数据库密码同样通过 PGPASSWORD 传递，不会进入进程参数。
# =============================================================================
set -euo pipefail

PASS_KEY="auth.password_hash"
BCRYPT_COST=10

MODE=""            # local | docker(空时自动判断)
DSN=""
HOST="" PORT="" USER="" DBPASS="" DBNAME="" SSLMODE=""
CONFIG=""
CONTAINER=""       # docker 的模式 postgres 服务/集装箱名称(默认 postgres)
EXEC_KIND=""       # compose | docker(docker 在什么模式下 exec 我应该写吗. 空时自动)
NEWPASS=""
ASSUME_YES=0

die() { echo "错误: $*" >&2; exit 1; }
info() { echo "· $*" >&2; }

usage() { sed -n '2,25p' "$0" | sed 's/^# \{0,1\}//'; exit 0; }

# ---- 参数解析 -------------------------------------------------------------
while [[ $# -gt 0 ]]; do
  case "$1" in
    -m|--mode)        MODE="${2:-}"; shift 2 ;;
    --dsn)            DSN="${2:-}"; shift 2 ;;
    -H|--host)        HOST="${2:-}"; shift 2 ;;
    -P|--port)        PORT="${2:-}"; shift 2 ;;
    -U|--user)        USER="${2:-}"; shift 2 ;;
    -W|--db-password) DBPASS="${2:-}"; shift 2 ;;
    -d|--dbname)      DBNAME="${2:-}"; shift 2 ;;
    --sslmode)        SSLMODE="${2:-}"; shift 2 ;;
    --config)         CONFIG="${2:-}"; shift 2 ;;
    -c|--container)   CONTAINER="${2:-}"; shift 2 ;;
    --exec)           EXEC_KIND="${2:-}"; shift 2 ;;
    -p|--new-password) NEWPASS="${2:-}"; shift 2 ;;
    -y|--yes)         ASSUME_YES=1; shift ;;
    -h|--help)        usage ;;
    *) die "这是一个未知因素。: $1(-h 检查使用情况)" ;;
  esac
done

# ---- config.json 在 database.* 读(local 模式并且仅当未指定连接信息时。) -----
# python3 更喜欢解析(坚固耐用). python3 如果没有 grep 替换为(config.json 因为这是一个整齐的分离场).
read_config_json() {
  local path="$1"
  [[ -f "$path" ]] || return 1
  if command -v python3 >/dev/null 2>&1; then
    python3 - "$path" <<'PY'
import json, sys
try:
    d = json.load(open(sys.argv[1])).get("database", {})
except Exception:
    sys.exit(1)
# dsn 可以直接给它或者给它一个字段
if d.get("dsn"):
    print("DSN\t" + d["dsn"]); sys.exit(0)
for k in ("host","port","user","password","dbname","sslmode"):
    if d.get(k) is not None:
        print(k.upper() + "\t" + str(d[k]))
PY
  else
    # 非常简单的下属: 木田 grep(值是字符串或数字)
    local k
    for k in host port user password dbname sslmode; do
      local v
      v=$(grep -oE "\"$k\"[[:space:]]*:[[:space:]]*(\"[^\"]*\"|[0-9]+)" "$path" 2>/dev/null \
            | head -1 | sed -E "s/.*:[[:space:]]*//; s/^\"//; s/\"$//") || true
      [[ -n "$v" ]] && echo -e "${k^^}\t$v"
    done
  fi
}

apply_config_fields() {
  local line key val
  while IFS=$'\t' read -r key val; do
    [[ -z "$key" ]] && continue
    case "$key" in
      DSN)      [[ -z "$DSN" ]] && DSN="$val" ;;
      HOST)     [[ -z "$HOST" ]] && HOST="$val" ;;
      PORT)     [[ -z "$PORT" ]] && PORT="$val" ;;
      USER)     [[ -z "$USER" ]] && USER="$val" ;;
      PASSWORD) [[ -z "$DBPASS" ]] && DBPASS="$val" ;;
      DBNAME)   [[ -z "$DBNAME" ]] && DBNAME="$val" ;;
      SSLMODE)  [[ -z "$SSLMODE" ]] && SSLMODE="$val" ;;
    esac
  done
}

# ---- 模式自动判断 ---------------------------------------------------------
if [[ -z "$MODE" ]]; then
  if [[ -n "$DSN$HOST$USER$DBNAME" || -n "${ARGUS_PG_DSN:-}" || -f "${CONFIG:-config.json}" ]]; then
    MODE="local"
  elif command -v docker >/dev/null 2>&1 && [[ -f docker-compose.yml ]]; then
    MODE="docker"
  else
    MODE="local"
  fi
fi
info "部署方式: $MODE"

# ---- 收集新密码 -----------------------------------------------------------
if [[ -z "$NEWPASS" ]]; then
  read -r -s -p "输入新密码(您的用户名是 ARGUS 固定到):" NEWPASS; echo >&2
  [[ -n "$NEWPASS" ]] || die "密码不能为空"
  read -r -s -p "再次输入确认:" NEWPASS2; echo >&2
  [[ "$NEWPASS" == "$NEWPASS2" ]] || die "两次输入的值不匹配"
fi
[[ -n "$NEWPASS" ]] || die "密码不能为空"

# 密码作为环境变量 psql 交付至(\getenv 阅读与 argv/ps 不进入)
export ARGUS_RESET_NEWPASS="$NEWPASS"

# DB 里面 bcrypt 创造 upsert 我做. 密码是 :'newpw' 自动转义为. CREATE EXTENSION 是幂等的,
# 如果您的数据库角色没有创建扩展的权限，您将在此处收到错误。(说明如下 run 查看失败的分支).
SQL=$(cat <<SQL
\\set ON_ERROR_STOP on
\\getenv newpw ARGUS_RESET_NEWPASS
CREATE EXTENSION IF NOT EXISTS pgcrypto;
INSERT INTO settings(key, value)
VALUES ('$PASS_KEY', crypt(:'newpw', gen_salt('bf', $BCRYPT_COST)))
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now();
SQL
)

# ---- 跑 -----------------------------------------------------------------
if [[ "$MODE" == "local" ]]; then
  # 连接信息优先级: 命令行 > --dsn/$ARGUS_PG_DSN > config.json
  if [[ -z "$DSN" && -z "$HOST$USER$DBNAME" ]]; then
    [[ -n "${ARGUS_PG_DSN:-}" ]] && DSN="$ARGUS_PG_DSN"
  fi
  if [[ -z "$DSN" && -z "$HOST$USER$DBNAME" ]]; then
    cfg="${CONFIG:-config.json}"
    if [[ -f "$cfg" ]]; then
      info "$cfg 从以下位置读取数据库设置"
      apply_config_fields < <(read_config_json "$cfg")
    fi
  fi

  command -v psql >/dev/null 2>&1 || die "在这台计算机上 psql 未找到(postgresql-client 或安装 -m docker 请使用)"

  declare -a PSQL_ARGS=()
  if [[ -n "$DSN" ]]; then
    PSQL_ARGS=("$DSN")
    target="$DSN"
  else
    [[ -n "$USER"   ]] || die "数据库用户(-U)缺失或有效 config.json/DSN 没有"
    [[ -n "$DBNAME" ]] || die "数据库名称(-d)缺失或有效 config.json/DSN 没有"
    HOST="${HOST:-127.0.0.1}"; PORT="${PORT:-5432}"; SSLMODE="${SSLMODE:-disable}"
    PSQL_ARGS=(-h "$HOST" -p "$PORT" -U "$USER" -d "$DBNAME")
    [[ -n "$SSLMODE" ]] && export PGSSLMODE="$SSLMODE"
    [[ -n "$DBPASS" ]] && export PGPASSWORD="$DBPASS"
    target="$USER@$HOST:$PORT/$DBNAME"
  fi

  info "目标数据库: $target"
  if [[ "$ASSUME_YES" -ne 1 ]]; then
    read -r -p "在这个数据库中 ARGUS 我应该重设密码吗？? [y/N]" ans
    [[ "$ans" == "y" || "$ans" == "Y" ]] || die "我取消了"
  fi

  if ! printf '%s\n' "$SQL" | psql "${PSQL_ARGS[@]}" -v ON_ERROR_STOP=1 -q >/dev/null; then
    die "写入失败. pgcrypto 缺乏许可·如果出现遗漏错误，请使用具有创建扩展权限的角色或使用 CREATE EXTENSION pgcrypto 直接运行."
  fi

else
  # ---- docker ----
  command -v docker >/dev/null 2>&1 || die "docker 未找到"
  CONTAINER="${CONTAINER:-postgres}"

  # exec 选择方法: docker compose exec(服务名称)确定优先顺序, 或者 docker exec(集装箱名称)
  if [[ -z "$EXEC_KIND" ]]; then
    if docker compose version >/dev/null 2>&1 && [[ -f docker-compose.yml ]]; then
      EXEC_KIND="compose"
    else
      EXEC_KIND="docker"
    fi
  fi

  # 容器内 psql 凭证: 首先是命令行, 下一个 .env 的 POSTGRES_*, 最后 compose 默认(argus)
  if [[ -f .env ]]; then
    # shellcheck disable=SC1091
    set -a; . ./.env; set +a
  fi
  DUSER="${USER:-${POSTGRES_USER:-argus}}"
  DNAME="${DBNAME:-${POSTGRES_DB:-argus}}"
  [[ -n "$DBPASS" ]] && export PGPASSWORD="$DBPASS"
  [[ -z "${PGPASSWORD:-}" && -n "${POSTGRES_PASSWORD:-}" ]] && export PGPASSWORD="$POSTGRES_PASSWORD"

  info "目标: 容器 $CONTAINER 不 psql -U $DUSER -d $DNAME(exec=$EXEC_KIND)"
  if [[ "$ASSUME_YES" -ne 1 ]]; then
    read -r -p "在这个容器数据库中 ARGUS 我应该重设密码吗？? [y/N]" ans
    [[ "$ans" == "y" || "$ans" == "Y" ]] || die "我取消了"
  fi

  # -e 仅给出名称，不给出值。 → 由于是从当前环境继承的，所以密码是 docker 命令 argv 没有出现在.
  declare -a EXEC_CMD
  if [[ "$EXEC_KIND" == "compose" ]]; then
    EXEC_CMD=(docker compose exec -T -e ARGUS_RESET_NEWPASS -e PGPASSWORD "$CONTAINER"
              psql -U "$DUSER" -d "$DNAME" -v ON_ERROR_STOP=1 -q)
  else
    EXEC_CMD=(docker exec -i -e ARGUS_RESET_NEWPASS -e PGPASSWORD "$CONTAINER"
              psql -U "$DUSER" -d "$DNAME" -v ON_ERROR_STOP=1 -q)
  fi

  if ! printf '%s\n' "$SQL" | "${EXEC_CMD[@]}" >/dev/null; then
    die "写入失败. 集装箱名称(-c), 数据库帐户(.env 的 POSTGRES_*), 的作用 pgcrypto 检查您的权限."
  fi
fi

unset ARGUS_RESET_NEWPASS
echo "✓ ARGUS 我重置了管理员密码. 用户名 ARGUS + 请使用您的新密码登录(无需重启服务)."
echo "  先前发出的登录会话(代币)将在几秒内全部失效. 帐号被占用·用于泄漏后恢复的定制工具·MCP·LLM·另请检查您的通知设置."
