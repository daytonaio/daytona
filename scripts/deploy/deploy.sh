#!/bin/bash
set -e

# =============================================================================
# Daytona Deployment Script
# Deploys the Daytona project to a remote Ubuntu server
# =============================================================================

# Configuration
REMOTE_HOST="${REMOTE_HOST:-h1004.blinkbox.dev}"
REMOTE_USER="${REMOTE_USER:-root}"
REMOTE_PATH="${REMOTE_PATH:-/opt/daytona}"
SSH_KEY="${SSH_KEY:-}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# SSH command helper
ssh_cmd() {
    local ssh_opts="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
    if [ -n "$SSH_KEY" ]; then
        ssh_opts="$ssh_opts -i $SSH_KEY"
    fi
    ssh $ssh_opts "$REMOTE_USER@$REMOTE_HOST" "$@"
}

# Rsync command helper
rsync_cmd() {
    local rsync_opts="-avz --progress --delete"
    rsync_opts="$rsync_opts --exclude='.git'"
    rsync_opts="$rsync_opts --exclude='node_modules'"
    rsync_opts="$rsync_opts --exclude='dist'"
    rsync_opts="$rsync_opts --exclude='.tmp'"
    rsync_opts="$rsync_opts --exclude='tmp'"
    rsync_opts="$rsync_opts --exclude='.yarn/cache'"
    rsync_opts="$rsync_opts --exclude='*.log'"
    rsync_opts="$rsync_opts --exclude='.env'"
    
    if [ -n "$SSH_KEY" ]; then
        rsync_opts="$rsync_opts -e 'ssh -i $SSH_KEY -o StrictHostKeyChecking=no'"
    else
        rsync_opts="$rsync_opts -e 'ssh -o StrictHostKeyChecking=no'"
    fi
    
    eval rsync $rsync_opts "$@"
}

show_usage() {
    echo "Usage: $0 [OPTIONS] COMMAND"
    echo ""
    echo "Commands:"
    echo "  setup       Run initial server setup (Docker, nginx, certbot)"
    echo "  sync        Sync project files to remote server"
    echo "  build       Build Docker images on remote server"
    echo "  start       Start all services"
    echo "  stop        Stop all services"
    echo "  restart     Restart all services"
    echo "  logs        View service logs"
    echo "  status      Show service status"
    echo "  deploy      Full deployment (sync + build + start)"
    echo ""
    echo "Options:"
    echo "  -h, --host HOST     Remote host (default: h1004.blinkbox.dev)"
    echo "  -u, --user USER     Remote user (default: root)"
    echo "  -k, --key KEY       SSH private key path"
    echo "  -p, --path PATH     Remote path (default: /opt/daytona)"
    echo "  --help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 deploy                    # Full deployment with defaults"
    echo "  $0 -h myserver.com deploy    # Deploy to custom host"
    echo "  $0 -k ~/.ssh/id_rsa sync     # Sync with specific SSH key"
    echo "  $0 logs api                  # View API logs"
}

cmd_setup() {
    log_info "Running initial server setup on $REMOTE_HOST..."
    
    # Copy setup script to remote
    local setup_script="$SCRIPT_DIR/setup-remote.sh"
    if [ ! -f "$setup_script" ]; then
        log_error "Setup script not found: $setup_script"
        exit 1
    fi
    
    # Create remote directory
    ssh_cmd "mkdir -p $REMOTE_PATH/scripts/deploy"
    
    # Copy setup script
    local scp_opts="-o StrictHostKeyChecking=no"
    if [ -n "$SSH_KEY" ]; then
        scp_opts="$scp_opts -i $SSH_KEY"
    fi
    scp $scp_opts "$setup_script" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/scripts/deploy/"
    
    # Run setup script
    ssh_cmd "chmod +x $REMOTE_PATH/scripts/deploy/setup-remote.sh && $REMOTE_PATH/scripts/deploy/setup-remote.sh"
    
    log_success "Server setup completed!"
}

cmd_sync() {
    log_info "Syncing project to $REMOTE_HOST:$REMOTE_PATH..."
    
    # Create remote directory
    ssh_cmd "mkdir -p $REMOTE_PATH"
    
    # Sync project files
    rsync_cmd "$PROJECT_ROOT/" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/"
    
    # Copy deployment configs
    log_info "Copying deployment configuration files..."
    ssh_cmd "mkdir -p $REMOTE_PATH/scripts/deploy/dex"
    ssh_cmd "mkdir -p $REMOTE_PATH/scripts/deploy/nginx"
    ssh_cmd "mkdir -p $REMOTE_PATH/scripts/deploy/otel"
    ssh_cmd "mkdir -p $REMOTE_PATH/scripts/deploy/pgadmin4"
    
    # Copy otel config from docker folder
    if [ -f "$PROJECT_ROOT/docker/otel/otel-collector-config.yaml" ]; then
        local scp_opts="-o StrictHostKeyChecking=no"
        if [ -n "$SSH_KEY" ]; then
            scp_opts="$scp_opts -i $SSH_KEY"
        fi
        scp $scp_opts "$PROJECT_ROOT/docker/otel/otel-collector-config.yaml" \
            "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/scripts/deploy/otel/"
    fi
    
    # Copy pgadmin config from docker folder
    if [ -d "$PROJECT_ROOT/docker/pgadmin4" ]; then
        local scp_opts="-o StrictHostKeyChecking=no -r"
        if [ -n "$SSH_KEY" ]; then
            scp_opts="$scp_opts -i $SSH_KEY"
        fi
        scp $scp_opts "$PROJECT_ROOT/docker/pgadmin4/"* \
            "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/scripts/deploy/pgadmin4/"
    fi
    
    # Create .env from env.example if it doesn't exist
    ssh_cmd "cd $REMOTE_PATH/scripts/deploy && [ ! -f .env ] && cp env.example .env || true"
    
    log_success "Project synced successfully!"
}

cmd_build() {
    log_info "Building Docker images on $REMOTE_HOST..."
    
    ssh_cmd "cd $REMOTE_PATH/scripts/deploy && docker compose -f docker-compose.production.yaml build"
    
    log_success "Docker images built successfully!"
}

cmd_start() {
    log_info "Starting services on $REMOTE_HOST..."
    
    ssh_cmd "cd $REMOTE_PATH/scripts/deploy && docker compose -f docker-compose.production.yaml up -d"
    
    log_success "Services started!"
    cmd_status
}

cmd_stop() {
    log_info "Stopping services on $REMOTE_HOST..."
    
    ssh_cmd "cd $REMOTE_PATH/scripts/deploy && docker compose -f docker-compose.production.yaml down"
    
    log_success "Services stopped!"
}

cmd_restart() {
    log_info "Restarting services on $REMOTE_HOST..."
    
    ssh_cmd "cd $REMOTE_PATH/scripts/deploy && docker compose -f docker-compose.production.yaml restart"
    
    log_success "Services restarted!"
}

cmd_logs() {
    local service="${1:-}"
    
    if [ -n "$service" ]; then
        ssh_cmd "cd $REMOTE_PATH/scripts/deploy && docker compose -f docker-compose.production.yaml logs -f $service"
    else
        ssh_cmd "cd $REMOTE_PATH/scripts/deploy && docker compose -f docker-compose.production.yaml logs -f"
    fi
}

cmd_status() {
    log_info "Service status on $REMOTE_HOST:"
    ssh_cmd "cd $REMOTE_PATH/scripts/deploy && docker compose -f docker-compose.production.yaml ps"
}

cmd_deploy() {
    log_info "Starting full deployment to $REMOTE_HOST..."
    
    cmd_sync
    cmd_build
    cmd_start
    
    log_success "Deployment completed!"
    echo ""
    echo "Access your Daytona instance at:"
    echo "  Dashboard: https://h1004.blinkbox.dev/"
    echo "  API:       https://h1004.blinkbox.dev/api/"
    echo "  SSH:       ssh -p 2222 <token>@h1004.blinkbox.dev"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--host)
            REMOTE_HOST="$2"
            shift 2
            ;;
        -u|--user)
            REMOTE_USER="$2"
            shift 2
            ;;
        -k|--key)
            SSH_KEY="$2"
            shift 2
            ;;
        -p|--path)
            REMOTE_PATH="$2"
            shift 2
            ;;
        --help)
            show_usage
            exit 0
            ;;
        setup|sync|build|start|stop|restart|logs|status|deploy)
            COMMAND="$1"
            shift
            break
            ;;
        *)
            log_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Execute command
if [ -z "${COMMAND:-}" ]; then
    show_usage
    exit 1
fi

case $COMMAND in
    setup)
        cmd_setup
        ;;
    sync)
        cmd_sync
        ;;
    build)
        cmd_build
        ;;
    start)
        cmd_start
        ;;
    stop)
        cmd_stop
        ;;
    restart)
        cmd_restart
        ;;
    logs)
        cmd_logs "$@"
        ;;
    status)
        cmd_status
        ;;
    deploy)
        cmd_deploy
        ;;
esac


