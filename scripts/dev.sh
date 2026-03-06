#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_FILE="$ROOT_DIR/docker/docker-compose.dev.yml"
API_ENV_EXAMPLE="$ROOT_DIR/apps/api/.env.example"
API_ENV_FILE="$ROOT_DIR/apps/api/.env"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

DOCKER_COMPOSE_BIN=""
DOCKER_COMPOSE_SUBCMD=""

log_info() {
  printf "%b\n" "${BLUE}[INFO]${NC} $1"
}

log_success() {
  printf "%b\n" "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
  printf "%b\n" "${YELLOW}[WARN]${NC} $1"
}

log_error() {
  printf "%b\n" "${RED}[ERROR]${NC} $1"
}

detect_compose_cmd() {
  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    DOCKER_COMPOSE_BIN="docker"
    DOCKER_COMPOSE_SUBCMD="compose"
    return 0
  fi

  if command -v docker-compose >/dev/null 2>&1 && docker-compose version >/dev/null 2>&1; then
    DOCKER_COMPOSE_BIN="docker-compose"
    DOCKER_COMPOSE_SUBCMD=""
    return 0
  fi

  return 1
}

compose_cmd() {
  if [ "$DOCKER_COMPOSE_BIN" = "docker" ]; then
    docker compose -f "$COMPOSE_FILE" "$@"
  else
    docker-compose -f "$COMPOSE_FILE" "$@"
  fi
}

ensure_runtime() {
  if ! command -v docker >/dev/null 2>&1; then
    log_error "Docker 未安装，请先安装 Docker Desktop"
    exit 1
  fi

  if ! detect_compose_cmd; then
    log_error "未检测到 docker compose 或 docker-compose"
    exit 1
  fi

  if ! docker info >/dev/null 2>&1; then
    log_error "Docker 未启动，请先启动 Docker Desktop"
    exit 1
  fi
}

ensure_api_env() {
  if [ -f "$API_ENV_FILE" ]; then
    return 0
  fi

  if [ ! -f "$API_ENV_EXAMPLE" ]; then
    log_error "缺少模板文件: $API_ENV_EXAMPLE"
    exit 1
  fi

  cp "$API_ENV_EXAMPLE" "$API_ENV_FILE"
  log_warn "检测到 apps/api/.env 不存在，已从 .env.example 自动生成。"
  log_warn "请按需修改 ADMIN_PASSWORD、ENCRYPTION_KEY、ENCRYPTION_SALT 等配置。"
}

get_api_port() {
  if [ ! -f "$API_ENV_FILE" ]; then
    echo "3001"
    return 0
  fi

  local port
  port="$(grep -E '^PORT=' "$API_ENV_FILE" | tail -n 1 | cut -d '=' -f2- || true)"
  if [ -z "$port" ]; then
    echo "3001"
  else
    echo "$port"
  fi
}

get_api_env_value() {
  local key="$1"
  local default_value="${2:-}"

  if [ ! -f "$API_ENV_FILE" ]; then
    echo "$default_value"
    return 0
  fi

  local value
  value="$(grep -E "^${key}=" "$API_ENV_FILE" | tail -n 1 | cut -d '=' -f2- || true)"
  if [ -z "$value" ]; then
    echo "$default_value"
  else
    echo "$value"
  fi
}

start_services() {
  ensure_runtime
  ensure_api_env

  local profiles=()
  while [ $# -gt 0 ]; do
    case "$1" in
      --tools)
        profiles+=(--profile tools)
        ;;
      --observability)
        profiles+=(--profile observability)
        ;;
      --full)
        profiles+=(--profile tools --profile observability)
        ;;
      *)
        log_warn "忽略未知参数: $1"
        ;;
    esac
    shift
  done

  log_info "启动开发基础设施（docker-compose.dev.yml）..."
  local runner_api_key
  runner_api_key="$(get_api_env_value "DEFAULT_RUNNER_API_KEY" "local_runner_key")"

  if [ ${#profiles[@]} -gt 0 ]; then
    DEFAULT_RUNNER_API_KEY="$runner_api_key" compose_cmd up -d "${profiles[@]}"
  else
    DEFAULT_RUNNER_API_KEY="$runner_api_key" compose_cmd up -d
  fi
  log_success "基础设施已启动"

  local api_port
  api_port="$(get_api_port)"

  echo
  log_info "常用地址:"
  echo "  - API（本机启动后）: http://localhost:${api_port}"
  echo "  - Dashboard（本机启动后）: http://localhost:3000"
  echo "  - Runner API: http://localhost:3003"
  echo "  - PostgreSQL: localhost:5432"
  echo "  - Redis: localhost:6379"
  echo "  - MinIO API: http://localhost:9000"
  echo "  - MinIO Console: http://localhost:9001"
  echo "  - Registry: http://localhost:6000"
  echo
  log_info "下一步:"
  echo "  - 启动 API: yarn dev:api"
  echo "  - 启动 Dashboard: yarn dev:dashboard"
  echo "  - 一键启动本机应用: yarn dev:full"
}

stop_services() {
  ensure_runtime

  log_info "停止开发基础设施..."
  compose_cmd down
  log_success "开发基础设施已停止"
}

restart_services() {
  stop_services
  start_services "$@"
}

show_status() {
  ensure_runtime
  compose_cmd ps
}

show_logs() {
  ensure_runtime
  compose_cmd logs -f "$@"
}

run_doctor() {
  local has_issue=false

  echo "[doctor] 检查基础环境"

  if command -v docker >/dev/null 2>&1; then
    echo "  - docker: $(docker --version)"
  else
    echo "  - docker: missing"
    has_issue=true
  fi

  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    echo "  - compose: $(docker compose version | head -n 1)"
  elif command -v docker-compose >/dev/null 2>&1; then
    echo "  - compose: $(docker-compose --version)"
  else
    echo "  - compose: missing"
    has_issue=true
  fi

  if command -v node >/dev/null 2>&1; then
    echo "  - node: $(node --version)"
  else
    echo "  - node: missing"
    has_issue=true
  fi

  if command -v yarn >/dev/null 2>&1; then
    echo "  - yarn: $(yarn --version)"
  else
    echo "  - yarn: missing"
    has_issue=true
  fi

  if [ -f "$API_ENV_FILE" ]; then
    echo "  - apps/api/.env: present"
  else
    echo "  - apps/api/.env: missing (将使用 .env.example 自动生成)"
  fi

  if [ -f "$SCRIPT_DIR/setup-proxy-dns.sh" ]; then
    echo "  - proxy dns script: scripts/setup-proxy-dns.sh"
  else
    echo "  - proxy dns script: missing"
  fi

  if [ "$has_issue" = true ]; then
    log_warn "检测到缺失项，建议先修复后再运行开发命令。"
    return 1
  fi

  log_success "基础环境检查完成"
}

ensure_api_database() {
  local db_name="${DB_DATABASE:-daytona}"
  local db_user="${DB_USERNAME:-user}"

  if ! detect_compose_cmd; then
    return 0
  fi

  if ! docker info >/dev/null 2>&1; then
    return 0
  fi

  if ! compose_cmd ps --services --status running | grep -qx "db"; then
    log_warn "PostgreSQL 容器未运行，跳过数据库自动检查。"
    return 0
  fi

  local exists
  exists="$(
    compose_cmd exec -T db psql -U "$db_user" -d postgres \
      -tAc "SELECT 1 FROM pg_database WHERE datname='${db_name}';" 2>/dev/null \
      | tr -d '[:space:]' || true
  )"

  if [ "$exists" = "1" ]; then
    return 0
  fi

  log_warn "检测到数据库 '$db_name' 不存在，正在自动创建..."
  if compose_cmd exec -T db psql -U "$db_user" -d postgres -c "CREATE DATABASE \"$db_name\";" >/dev/null 2>&1; then
    log_success "数据库 '$db_name' 创建完成"
  else
    log_warn "自动创建数据库 '$db_name' 失败，请手动检查 PostgreSQL 权限或数据库名。"
  fi
}

run_api() {
  ensure_api_env

  # Export vars from apps/api/.env so Nx serve uses local dev settings.
  set -a
  # shellcheck disable=SC1090
  source "$API_ENV_FILE"
  set +a

  # Normalize container-oriented defaults for host-run local development.
  if [ "${DB_HOST:-}" = "db" ] || [ -z "${DB_HOST:-}" ]; then
    export DB_HOST="localhost"
  fi
  if [ "${REDIS_HOST:-}" = "redis" ] || [ -z "${REDIS_HOST:-}" ]; then
    export REDIS_HOST="localhost"
  fi
  if [[ "${S3_ENDPOINT:-}" == *"minio:9000"* ]] || [ -z "${S3_ENDPOINT:-}" ]; then
    export S3_ENDPOINT="http://localhost:9000"
  fi
  if [[ "${S3_STS_ENDPOINT:-}" == *"minio:9000"* ]] || [ -z "${S3_STS_ENDPOINT:-}" ]; then
    export S3_STS_ENDPOINT="http://localhost:9000/minio/v1/assume-role"
  fi
  if [[ "${TRANSIENT_REGISTRY_URL:-}" == *"registry:5000"* ]] || \
     [[ "${TRANSIENT_REGISTRY_URL:-}" == *"registry:6000"* ]] || \
     [[ "${TRANSIENT_REGISTRY_URL:-}" == *"localhost:6000"* ]] || \
     [ -z "${TRANSIENT_REGISTRY_URL:-}" ]; then
    export TRANSIENT_REGISTRY_URL="http://host.docker.internal:6000"
  fi
  if [[ "${INTERNAL_REGISTRY_URL:-}" == *"registry:5000"* ]] || \
     [[ "${INTERNAL_REGISTRY_URL:-}" == *"registry:6000"* ]] || \
     [[ "${INTERNAL_REGISTRY_URL:-}" == *"localhost:6000"* ]] || \
     [ -z "${INTERNAL_REGISTRY_URL:-}" ]; then
    export INTERNAL_REGISTRY_URL="http://host.docker.internal:6000"
  fi
  if [[ "${DEFAULT_RUNNER_DOMAIN:-}" == *"runner"* ]] || [ -z "${DEFAULT_RUNNER_DOMAIN:-}" ]; then
    export DEFAULT_RUNNER_DOMAIN="localhost:3003"
  fi
  if [[ "${DEFAULT_RUNNER_API_URL:-}" == *"runner"* ]] || [ -z "${DEFAULT_RUNNER_API_URL:-}" ]; then
    export DEFAULT_RUNNER_API_URL="http://localhost:3003"
  fi
  if [[ "${DEFAULT_RUNNER_PROXY_URL:-}" == *"runner"* ]] || [ -z "${DEFAULT_RUNNER_PROXY_URL:-}" ]; then
    export DEFAULT_RUNNER_PROXY_URL="http://localhost:3003"
  fi
  if [ "${DASHBOARD_BASE_API_URL:-}" = "http://localhost:3000" ]; then
    export DASHBOARD_BASE_API_URL="http://localhost:${PORT:-3001}"
  fi
  export NX_TUI="false"
  export NX_DAEMON="false"

  ensure_api_database

  cd "$ROOT_DIR"
  exec yarn nx serve api --configuration=development --output-style=stream
}

run_dashboard() {
  cd "$ROOT_DIR"
  export NX_TUI="false"
  export NX_DAEMON="false"
  exec yarn nx serve dashboard --output-style=stream
}

run_full() {
  start_services

  log_info "启动本机 API 与 Dashboard（Ctrl+C 结束）..."

  (
    cd "$ROOT_DIR"
    ./scripts/dev.sh api
  ) &
  local api_pid=$!

  (
    cd "$ROOT_DIR"
    ./scripts/dev.sh dashboard
  ) &
  local dashboard_pid=$!

  cleanup() {
    kill "$api_pid" "$dashboard_pid" >/dev/null 2>&1 || true
  }

  trap cleanup INT TERM EXIT
  wait "$api_pid" "$dashboard_pid"
}

show_help() {
  cat <<USAGE
Daytona Lite 本地开发脚本

用法:
  ./scripts/dev.sh <command> [options]

命令:
  start [--tools|--observability|--full]  启动开发基础设施
  stop                                    停止开发基础设施
  restart [--tools|--observability|--full] 重启开发基础设施
  status                                  查看容器状态
  logs [service]                          查看日志
  doctor                                  检查本地开发环境
  api                                     本机启动 API（热重载）
  dashboard                               本机启动 Dashboard（Vite）
  full                                    启动基础设施 + API + Dashboard
  help                                    显示帮助

常用:
  yarn dev:start
  yarn dev:api
  yarn dev:dashboard
  yarn dev:full
USAGE
}

CMD="${1:-help}"
shift || true

case "$CMD" in
  start)
    start_services "$@"
    ;;
  stop)
    stop_services
    ;;
  restart)
    restart_services "$@"
    ;;
  status)
    show_status
    ;;
  logs)
    show_logs "$@"
    ;;
  doctor)
    run_doctor
    ;;
  api)
    run_api
    ;;
  dashboard)
    run_dashboard
    ;;
  full)
    run_full
    ;;
  help|--help|-h)
    show_help
    ;;
  *)
    log_error "未知命令: $CMD"
    show_help
    exit 1
    ;;
esac
