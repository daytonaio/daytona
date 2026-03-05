#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
LOG_DIR="$ROOT_DIR/logs"
PID_DIR="$ROOT_DIR/tmp"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
  printf "%b\n" "${BLUE}[INFO]${NC} $1"
}

log_success() {
  printf "%b\n" "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
  printf "%b\n" "${YELLOW}[WARN]${NC} $1"
}

mkdir -p "$LOG_DIR" "$PID_DIR"

printf "%b\n" "${GREEN}========================================${NC}"
printf "%b\n" "${GREEN} Daytona Lite 快速开发启动${NC}"
printf "%b\n" "${GREEN}========================================${NC}"

log_info "步骤 1/3: 启动开发基础设施..."
"$ROOT_DIR/scripts/dev.sh" start

echo
printf "%b" "${YELLOW}步骤 2/3: 是否后台启动 API? (y/N): ${NC}"
read -r start_api
if [ "$start_api" = "y" ] || [ "$start_api" = "Y" ]; then
  nohup bash -lc "cd '$ROOT_DIR' && ./scripts/dev.sh api" >"$LOG_DIR/dev-api.log" 2>&1 &
  echo $! >"$PID_DIR/dev-api.pid"
  log_success "API 已后台启动 (PID: $(cat "$PID_DIR/dev-api.pid"))"
  log_info "查看日志: tail -f $LOG_DIR/dev-api.log"
else
  log_warn "已跳过 API 启动"
  log_info "稍后运行: yarn dev:api"
fi

echo
printf "%b" "${YELLOW}步骤 3/3: 是否后台启动 Dashboard? (y/N): ${NC}"
read -r start_dashboard
if [ "$start_dashboard" = "y" ] || [ "$start_dashboard" = "Y" ]; then
  nohup bash -lc "cd '$ROOT_DIR' && ./scripts/dev.sh dashboard" >"$LOG_DIR/dev-dashboard.log" 2>&1 &
  echo $! >"$PID_DIR/dev-dashboard.pid"
  log_success "Dashboard 已后台启动 (PID: $(cat "$PID_DIR/dev-dashboard.pid"))"
  log_info "查看日志: tail -f $LOG_DIR/dev-dashboard.log"
else
  log_warn "已跳过 Dashboard 启动"
  log_info "稍后运行: yarn dev:dashboard"
fi

echo
printf "%b\n" "${GREEN}========================================${NC}"
printf "%b\n" "${GREEN} 启动完成${NC}"
printf "%b\n" "${GREEN}========================================${NC}"

echo "  - Dashboard: http://localhost:3000"
echo "  - API: http://localhost:3001"
echo "  - Runner: http://localhost:3003"
echo "  - MinIO Console: http://localhost:9001"

echo
log_info "管理命令:"
echo "  - 服务状态: yarn dev:status"
echo "  - 容器日志: yarn dev:logs"
echo "  - 停止基础设施: yarn dev:stop"
echo "  - 环境诊断: yarn dev:doctor"

echo
if [ -f "$PID_DIR/dev-api.pid" ]; then
  echo "  - 停止 API 后台进程: kill \$(cat $PID_DIR/dev-api.pid)"
fi
if [ -f "$PID_DIR/dev-dashboard.pid" ]; then
  echo "  - 停止 Dashboard 后台进程: kill \$(cat $PID_DIR/dev-dashboard.pid)"
fi
