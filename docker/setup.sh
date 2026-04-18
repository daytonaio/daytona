#!/usr/bin/env bash
# =============================================================================
# Daytona Self-Hosted Setup Wizard
# =============================================================================
# Interactive setup for deploying Daytona with an external domain, HTTPS,
# and optional Cloudflare Tunnel / Nginx Proxy Manager configuration.
#
# Usage:
#   cd docker
#   chmod +x setup.sh
#   ./setup.sh
#   ./setup.sh --skip-preflight    # skip Docker/tool checks
#
# What this script does:
#   1. Prompts for your domain configuration
#   2. Auto-generates all secrets (encryption keys, API keys, passwords)
#   3. Creates the .env file
#   4. Creates the dex/config.yaml (OIDC provider config)
#   5. Shows a summary with next steps
#
# =============================================================================

set -euo pipefail

# ── Flags ────────────────────────────────────────────────────────────────────

SKIP_PREFLIGHT=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --skip-preflight|-s)
            SKIP_PREFLIGHT=true
            shift
            ;;
        --help|-h)
            echo "Usage: ./setup.sh [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --skip-preflight, -s   Skip Docker and tool checks"
            echo "  --help, -h             Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1 (use --help for usage)"
            exit 1
            ;;
    esac
done

# ── Colors & formatting ─────────────────────────────────────────────────────

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m' # No Color

# ── Helper functions ─────────────────────────────────────────────────────────

banner() {
    echo ""
    echo -e "${BLUE}${BOLD}"
    echo "  ╔══════════════════════════════════════════════════════════════╗"
    echo "  ║                                                              ║"
    echo "  ║              Daytona Self-Hosted Setup Wizard                ║"
    echo "  ║                                                              ║"
    echo "  ╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

section() {
    echo ""
    echo -e "${CYAN}${BOLD}━━━ $1 ━━━${NC}"
    echo ""
}

info() {
    echo -e "  ${BLUE}ℹ${NC}  $1"
}

success() {
    echo -e "  ${GREEN}✓${NC}  $1"
}

warn() {
    echo -e "  ${YELLOW}!${NC}  $1"
}

error() {
    echo -e "  ${RED}✗${NC}  $1"
}

prompt() {
    local varname="$1"
    local message="$2"
    local default="${3:-}"
    local value

    if [[ -n "$default" ]]; then
        echo -en "  ${BOLD}$message${NC} ${DIM}[$default]${NC}: "
        read -r value
        value="${value:-$default}"
    else
        echo -en "  ${BOLD}$message${NC}: "
        read -r value
    fi

    eval "$varname=\$value"
}

prompt_password() {
    local varname="$1"
    local message="$2"
    local value confirm

    while true; do
        echo -en "  ${BOLD}$message${NC}: "
        read -rs value
        echo ""
        echo -en "  ${BOLD}Confirm${NC}: "
        read -rs confirm
        echo ""

        if [[ "$value" == "$confirm" ]]; then
            if [[ -z "$value" ]]; then
                warn "Password cannot be empty. Try again."
            else
                break
            fi
        else
            warn "Passwords don't match. Try again."
        fi
    done

    eval "$varname=\$value"
}

prompt_yn() {
    local varname="$1"
    local message="$2"
    local default="${3:-y}"
    local value

    if [[ "$default" == "y" ]]; then
        echo -en "  ${BOLD}$message${NC} ${DIM}[Y/n]${NC}: "
    else
        echo -en "  ${BOLD}$message${NC} ${DIM}[y/N]${NC}: "
    fi
    read -r value
    value="${value:-$default}"

    case "$value" in
        [yY]|[yY][eE][sS]) eval "$varname=true" ;;
        *) eval "$varname=false" ;;
    esac
}

generate_secret() {
    local length="${1:-32}"
    openssl rand -hex "$length" 2>/dev/null || LC_ALL=C tr -dc 'a-f0-9' < /dev/urandom | head -c "$((length * 2))"
}

generate_password() {
    local length="${1:-24}"
    openssl rand -base64 "$length" 2>/dev/null | tr -d '=/+' | head -c "$length" || LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c "$length"
}

check_command() {
    command -v "$1" &>/dev/null
}

# ── Preflight checks ────────────────────────────────────────────────────────

preflight() {
    section "Preflight Checks"

    local ok=true

    if check_command docker; then
        success "Docker found: $(docker --version | head -1)"
    else
        error "Docker not found. Install Docker first: https://docs.docker.com/get-docker/"
        ok=false
    fi

    if check_command docker && docker compose version &>/dev/null; then
        success "Docker Compose found: $(docker compose version --short 2>/dev/null || echo 'available')"
    elif check_command docker-compose; then
        success "Docker Compose (standalone) found"
    else
        error "Docker Compose not found."
        ok=false
    fi

    if check_command openssl; then
        success "OpenSSL found (used for generating secrets)"
    else
        warn "OpenSSL not found. Will use /dev/urandom fallback for secret generation."
    fi

    if check_command htpasswd; then
        success "htpasswd found (used for Dex password hashing)"
        HAS_HTPASSWD=true
    else
        HAS_HTPASSWD=false
        if check_command python3; then
            success "python3 found (will use bcrypt for Dex password hashing)"
            HAS_PYTHON3=true
        else
            HAS_PYTHON3=false
            warn "Neither htpasswd nor python3 found. You'll need to set the Dex password hash manually."
        fi
    fi

    # Check we're in the docker/ directory
    if [[ ! -f "docker-compose.yaml" ]]; then
        if [[ -f "docker/docker-compose.yaml" ]]; then
            info "Changing to docker/ directory..."
            cd docker
        else
            error "Cannot find docker-compose.yaml. Run this script from the docker/ directory."
            ok=false
        fi
    fi

    if [[ "$ok" == "false" ]]; then
        echo ""
        error "Preflight checks failed. Fix the issues above and try again."
        exit 1
    fi

    echo ""
    success "All checks passed!"
}

# ── Existing config detection ────────────────────────────────────────────────

detect_existing() {
    if [[ -f ".env" ]]; then
        echo ""
        warn "An existing .env file was found."
        prompt_yn OVERWRITE_ENV "Overwrite it?" "n"
        if [[ "$OVERWRITE_ENV" != "true" ]]; then
            info "Keeping existing .env. Only dex/config.yaml will be regenerated if needed."
            SKIP_ENV=true
        else
            SKIP_ENV=false
            # Backup
            cp .env ".env.backup.$(date +%Y%m%d%H%M%S)"
            success "Backed up existing .env"
        fi
    else
        SKIP_ENV=false
    fi
}

# ── Domain configuration ────────────────────────────────────────────────────

configure_domains() {
    section "Domain Configuration"

    info "Daytona needs 3 hostnames for HTTP(S) traffic and optionally 1 for SSH."
    info "All HTTP(S) traffic goes through your reverse proxy (e.g., Nginx Proxy Manager)."
    echo ""
    echo -e "  ${DIM}Example domain layout:${NC}"
    echo -e "  ${DIM}  Dashboard + API:    daytona.example.com     -> api:3000${NC}"
    echo -e "  ${DIM}  OIDC Provider:      dex.example.com         -> dex:5556${NC}"
    echo -e "  ${DIM}  Sandbox Proxy:      proxy.daytona.example.com -> proxy:4000${NC}"
    echo -e "  ${DIM}  Sandbox Wildcard:   *.proxy.daytona.example.com -> proxy:4000${NC}"
    echo ""

    prompt MAIN_DOMAIN "Main domain (dashboard + API)" ""
    while [[ -z "$MAIN_DOMAIN" ]]; do
        warn "Domain is required."
        prompt MAIN_DOMAIN "Main domain (dashboard + API)" ""
    done

    # Strip protocol if accidentally included
    MAIN_DOMAIN="${MAIN_DOMAIN#https://}"
    MAIN_DOMAIN="${MAIN_DOMAIN#http://}"
    MAIN_DOMAIN="${MAIN_DOMAIN%/}"

    # Suggest defaults based on main domain
    # Extract base domain (e.g., "example.com" from "daytona.example.com")
    local base_domain
    if [[ "$MAIN_DOMAIN" == *.*.* ]]; then
        # Has subdomain: daytona.example.com -> example.com
        base_domain="${MAIN_DOMAIN#*.}"
    else
        # Is root domain: example.com
        base_domain="$MAIN_DOMAIN"
    fi

    local dex_default="dex.${base_domain}"
    local proxy_default="proxy.${MAIN_DOMAIN}"

    echo ""
    prompt DEX_DOMAIN "Dex (OIDC) domain" "$dex_default"
    DEX_DOMAIN="${DEX_DOMAIN#https://}"
    DEX_DOMAIN="${DEX_DOMAIN#http://}"

    echo ""
    prompt PROXY_DOMAIN "Sandbox proxy base domain" "$proxy_default"
    PROXY_DOMAIN="${PROXY_DOMAIN#https://}"
    PROXY_DOMAIN="${PROXY_DOMAIN#http://}"

    echo ""
    info "Sandbox URLs will look like: https://22222-{sandboxId}.${PROXY_DOMAIN}"

    # Protocol
    echo ""
    prompt_yn USE_HTTPS "Use HTTPS? (requires SSL termination at your reverse proxy or CDN)" "y"
    if [[ "$USE_HTTPS" == "true" ]]; then
        PROTOCOL="https"
    else
        PROTOCOL="http"
    fi
}

# ── SSH Gateway ──────────────────────────────────────────────────────────────

configure_ssh() {
    section "SSH Gateway (Optional)"

    info "The SSH gateway allows direct SSH access to sandboxes."
    info "It uses raw TCP — it cannot go through Cloudflare or HTTP reverse proxies."
    echo ""

    prompt_yn ENABLE_SSH "Enable SSH gateway?" "y"

    if [[ "$ENABLE_SSH" == "true" ]]; then
        echo ""
        info "Enter the public hostname/IP and port where SSH will be reachable."
        info "This must be directly accessible from the internet (port forwarded if behind NAT)."
        echo ""

        prompt SSH_HOST "SSH gateway public hostname or IP" "$MAIN_DOMAIN"
        prompt SSH_PORT "SSH gateway public port" "2222"

        SSH_URL="${SSH_HOST}:${SSH_PORT}"
        SSH_COMMAND="ssh -p ${SSH_PORT} {{TOKEN}}@${SSH_HOST}"
    else
        SSH_HOST=""
        SSH_PORT=""
        SSH_URL=""
        SSH_COMMAND=""
    fi
}

# ── Dex admin user ───────────────────────────────────────────────────────────

configure_dex_admin() {
    section "Admin Account (Dex)"

    info "Set up the initial admin account for logging into Daytona."
    info "This creates a static user in Dex (the OIDC provider)."
    echo ""

    prompt DEX_ADMIN_EMAIL "Admin email" "admin@${MAIN_DOMAIN}"
    prompt DEX_ADMIN_USERNAME "Admin username" "admin"

    echo ""
    prompt_yn USE_DEFAULT_PASSWORD "Use default password 'password'? (change later for production)" "n"

    if [[ "$USE_DEFAULT_PASSWORD" == "true" ]]; then
        DEX_ADMIN_PASSWORD="password"
        # Default bcrypt hash for 'password'
        DEX_PASSWORD_HASH='$2a$10$2b2cU8CPhOTaGrs1HRQuAueS7JTT5ZHsHSzYiFPm1leZck7Mc8T4W'
    else
        prompt_password DEX_ADMIN_PASSWORD "Admin password"
        echo ""
        info "Generating bcrypt hash..."

        if [[ "${HAS_HTPASSWD:-false}" == "true" ]]; then
            DEX_PASSWORD_HASH=$(htpasswd -nbBC 10 "" "$DEX_ADMIN_PASSWORD" | tr -d ':\n' | sed 's/$2y/$2a/')
        elif [[ "${HAS_PYTHON3:-false}" == "true" ]]; then
            DEX_PASSWORD_HASH=$(python3 -c "
import hashlib, os, base64
try:
    import bcrypt
    print(bcrypt.hashpw(b'''${DEX_ADMIN_PASSWORD}''', bcrypt.gensalt(10)).decode())
except ImportError:
    print('NEED_MANUAL')
" 2>/dev/null)
            if [[ "$DEX_PASSWORD_HASH" == "NEED_MANUAL" ]]; then
                warn "Python bcrypt module not installed."
                # Try using docker to generate the hash
                info "Trying Docker to generate hash..."
                DEX_PASSWORD_HASH=$(docker run --rm alpine sh -c "apk add --no-cache apache2-utils >/dev/null 2>&1 && htpasswd -nbBC 10 '' '$DEX_ADMIN_PASSWORD' | tr -d ':\n' | sed 's/\$2y/\$2a/'" 2>/dev/null) || true
                if [[ -z "$DEX_PASSWORD_HASH" ]]; then
                    warn "Could not generate bcrypt hash. Using default password hash."
                    warn "You'll need to update dex/config.yaml manually."
                    DEX_PASSWORD_HASH='$2a$10$2b2cU8CPhOTaGrs1HRQuAueS7JTT5ZHsHSzYiFPm1leZck7Mc8T4W'
                    DEX_ADMIN_PASSWORD="password"
                fi
            fi
        else
            info "Trying Docker to generate hash..."
            DEX_PASSWORD_HASH=$(docker run --rm alpine sh -c "apk add --no-cache apache2-utils >/dev/null 2>&1 && htpasswd -nbBC 10 '' '$DEX_ADMIN_PASSWORD' | tr -d ':\n' | sed 's/\$2y/\$2a/'" 2>/dev/null) || true
            if [[ -z "$DEX_PASSWORD_HASH" ]]; then
                warn "Could not generate bcrypt hash. Using default password hash."
                warn "You'll need to update dex/config.yaml manually."
                DEX_PASSWORD_HASH='$2a$10$2b2cU8CPhOTaGrs1HRQuAueS7JTT5ZHsHSzYiFPm1leZck7Mc8T4W'
                DEX_ADMIN_PASSWORD="password"
            fi
        fi
        success "Password hash generated."
    fi
}

# ── SMTP ─────────────────────────────────────────────────────────────────────

configure_smtp() {
    section "Email / SMTP (Optional)"

    info "Daytona sends emails for team invitations. The default uses MailDev (test only)."
    echo ""

    prompt_yn CONFIGURE_SMTP "Configure a real SMTP server?" "n"

    if [[ "$CONFIGURE_SMTP" == "true" ]]; then
        prompt SMTP_HOST "SMTP host" "smtp.gmail.com"
        prompt SMTP_PORT "SMTP port" "587"
        prompt SMTP_USER "SMTP username" ""
        prompt SMTP_PASSWORD_VAL "SMTP password" ""
        prompt SMTP_FROM "From address" "no-reply@${MAIN_DOMAIN}"
        prompt_yn SMTP_SECURE_VAL "Use TLS?" "y"
        if [[ "$SMTP_SECURE_VAL" == "true" ]]; then
            SMTP_SECURE="true"
        else
            SMTP_SECURE=""
        fi
    else
        SMTP_HOST="maildev"
        SMTP_PORT="1025"
        SMTP_USER=""
        SMTP_PASSWORD_VAL=""
        SMTP_FROM="Daytona Team <no-reply@daytona.io>"
        SMTP_SECURE=""
        info "Using MailDev. Emails will be visible at http://your-server:1080"
    fi
}

# ── Generate secrets ─────────────────────────────────────────────────────────

generate_secrets() {
    section "Generating Secrets"

    info "Auto-generating cryptographic secrets for all services..."
    echo ""

    SECRET_ENCRYPTION_KEY=$(generate_secret 32)
    success "Encryption key"

    SECRET_ENCRYPTION_SALT=$(generate_secret 16)
    success "Encryption salt"

    SECRET_PROXY_API_KEY=$(generate_secret 32)
    success "Proxy API key"

    SECRET_RUNNER_API_KEY=$(generate_secret 32)
    success "Runner API key"

    SECRET_SSH_GATEWAY_API_KEY=$(generate_secret 32)
    success "SSH gateway API key"

    SECRET_OTEL_API_KEY=$(generate_secret 32)
    success "OTel collector API key"

    SECRET_HEALTH_CHECK_KEY=$(generate_secret 32)
    success "Health check API key"

    SECRET_DB_PASSWORD=$(generate_password 24)
    success "Database password"

    SECRET_S3_ACCESS_KEY=$(generate_password 20)
    success "S3 access key"

    SECRET_S3_SECRET_KEY=$(generate_password 40)
    success "S3 secret key"

    SECRET_REGISTRY_PASSWORD=$(generate_password 24)
    success "Registry password"

    echo ""
    success "All secrets generated."
}

# ── Write .env file ──────────────────────────────────────────────────────────

write_env_file() {
    section "Writing .env File"

    cat > .env << ENVEOF
# =============================================================================
# Daytona Self-Hosted Configuration
# =============================================================================
# Generated by setup.sh on $(date -u '+%Y-%m-%d %H:%M:%S UTC')
#
# Architecture:
#   Browser -> Reverse Proxy (SSL termination) -> Docker services
#
# Required DNS entries:
#   - ${MAIN_DOMAIN}              -> reverse proxy -> api:3000
#   - ${DEX_DOMAIN}               -> reverse proxy -> dex:5556
#   - ${PROXY_DOMAIN}             -> reverse proxy -> proxy:4000
#   - *.${PROXY_DOMAIN}           -> reverse proxy -> proxy:4000 (wildcard!)
#
# =============================================================================

# ── Main domain ──
DAYTONA_DOMAIN=${MAIN_DOMAIN}
DAYTONA_URL=${PROTOCOL}://${MAIN_DOMAIN}

# ── Dashboard ──
DASHBOARD_URL=${PROTOCOL}://${MAIN_DOMAIN}/dashboard
DASHBOARD_BASE_API_URL=${PROTOCOL}://${MAIN_DOMAIN}

# ── OIDC / Dex ──
# Public URL: what browsers use to reach Dex
DEX_PUBLIC_URL=${PROTOCOL}://${DEX_DOMAIN}/dex
# Internal URL: how API/proxy reach Dex inside Docker network (do not change)
DEX_INTERNAL_URL=http://dex:5556/dex
# OIDC client settings
OIDC_CLIENT_ID=daytona
OIDC_AUDIENCE=daytona

# ── Proxy ──
# Sandbox preview URLs: ${PROTOCOL}://{port}-{sandboxId}.${PROXY_DOMAIN}
PROXY_DOMAIN=${PROXY_DOMAIN}
PROXY_PROTOCOL=${PROTOCOL}
PROXY_TEMPLATE_URL=${PROTOCOL}://{{PORT}}-{{sandboxId}}.${PROXY_DOMAIN}
PROXY_TOOLBOX_BASE_URL=${PROTOCOL}://${PROXY_DOMAIN}
PROXY_API_KEY=${SECRET_PROXY_API_KEY}

# ── SSH Gateway ──
SSH_GATEWAY_HOST=${SSH_HOST}
SSH_GATEWAY_PORT=${SSH_PORT}
SSH_GATEWAY_URL=${SSH_URL}
SSH_GATEWAY_COMMAND=${SSH_COMMAND}
SSH_GATEWAY_API_KEY=${SECRET_SSH_GATEWAY_API_KEY}

# ── Database ──
DB_HOST=db
DB_PORT=5432
DB_USERNAME=daytona
DB_PASSWORD=${SECRET_DB_PASSWORD}
DB_DATABASE=daytona

# ── Redis ──
REDIS_HOST=redis
REDIS_PORT=6379

# ── Encryption ──
ENCRYPTION_KEY=${SECRET_ENCRYPTION_KEY}
ENCRYPTION_SALT=${SECRET_ENCRYPTION_SALT}

# ── Object Storage (MinIO) ──
S3_ENDPOINT=http://minio:9000
S3_STS_ENDPOINT=http://minio:9000/minio/v1/assume-role
S3_REGION=us-east-1
S3_ACCESS_KEY=${SECRET_S3_ACCESS_KEY}
S3_SECRET_KEY=${SECRET_S3_SECRET_KEY}
S3_DEFAULT_BUCKET=daytona

# ── Container Registry ──
REGISTRY_URL=http://registry:6000
REGISTRY_ADMIN=admin
REGISTRY_PASSWORD=${SECRET_REGISTRY_PASSWORD}
REGISTRY_PROJECT_ID=daytona

# ── SMTP ──
SMTP_HOST=${SMTP_HOST}
SMTP_PORT=${SMTP_PORT}
SMTP_USER=${SMTP_USER}
SMTP_PASSWORD=${SMTP_PASSWORD_VAL}
SMTP_SECURE=${SMTP_SECURE}
SMTP_EMAIL_FROM=${SMTP_FROM}

# ── Runner ──
DEFAULT_RUNNER_API_KEY=${SECRET_RUNNER_API_KEY}

# ── Telemetry ──
OTEL_ENABLED=true
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
OTEL_COLLECTOR_API_KEY=${SECRET_OTEL_API_KEY}

# ── Health Check ──
HEALTH_CHECK_API_KEY=${SECRET_HEALTH_CHECK_KEY}
ENVEOF

    success "Written: .env"
}

# ── Write dex/config.yaml ───────────────────────────────────────────────────

write_dex_config() {
    section "Writing dex/config.yaml"

    mkdir -p dex

    # Escape special characters in the hash for YAML
    local escaped_hash
    escaped_hash=$(echo "$DEX_PASSWORD_HASH" | sed 's/\$/\\$/g')

    cat > dex/config.yaml << DEXEOF
# Dex OIDC Provider Configuration
# Generated by setup.sh on $(date -u '+%Y-%m-%d %H:%M:%S UTC')
#
# IMPORTANT: The issuer URL must match DEX_PUBLIC_URL in .env exactly.
issuer: ${PROTOCOL}://${DEX_DOMAIN}/dex
storage:
  type: sqlite3
  config:
    file: /var/dex/dex.db

web:
  http: 0.0.0.0:5556
  allowedOrigins: ['*']
  allowedHeaders: ['x-requested-with']
staticClients:
  - id: daytona
    redirectURIs:
      # External domain URLs
      - '${PROTOCOL}://${MAIN_DOMAIN}'
      - '${PROTOCOL}://${MAIN_DOMAIN}/api/oauth2-redirect.html'
      - '${PROTOCOL}://${PROXY_DOMAIN}/callback'
      # Localhost fallback (for local development / debugging)
      - 'http://localhost:3000'
      - 'http://localhost:3000/api/oauth2-redirect.html'
      - 'http://proxy.localhost:4000/callback'
    name: 'Daytona'
    public: true
enablePasswordDB: true
staticPasswords:
  - email: '${DEX_ADMIN_EMAIL}'
    hash: '${DEX_PASSWORD_HASH}'
    username: '${DEX_ADMIN_USERNAME}'
    userID: '1234'
DEXEOF

    success "Written: dex/config.yaml"
}

# ── Summary ──────────────────────────────────────────────────────────────────

show_summary() {
    section "Configuration Summary"

    echo -e "  ${BOLD}Domains:${NC}"
    echo -e "    Dashboard + API:  ${GREEN}${PROTOCOL}://${MAIN_DOMAIN}${NC}"
    echo -e "    OIDC (Dex):       ${GREEN}${PROTOCOL}://${DEX_DOMAIN}/dex${NC}"
    echo -e "    Sandbox Proxy:    ${GREEN}${PROTOCOL}://${PROXY_DOMAIN}${NC}"
    echo -e "    Sandbox Wildcard: ${GREEN}${PROTOCOL}://*.${PROXY_DOMAIN}${NC}"
    if [[ -n "$SSH_HOST" ]]; then
        echo -e "    SSH Gateway:      ${GREEN}${SSH_HOST}:${SSH_PORT}${NC} (direct TCP)"
    else
        echo -e "    SSH Gateway:      ${DIM}disabled${NC}"
    fi

    echo ""
    echo -e "  ${BOLD}Admin Account:${NC}"
    echo -e "    Email:    ${GREEN}${DEX_ADMIN_EMAIL}${NC}"
    echo -e "    Username: ${GREEN}${DEX_ADMIN_USERNAME}${NC}"
    echo -e "    Password: ${DIM}(as configured)${NC}"

    echo ""
    echo -e "  ${BOLD}Generated Files:${NC}"
    if [[ "${SKIP_ENV:-false}" != "true" ]]; then
        echo -e "    ${GREEN}✓${NC} .env"
    fi
    echo -e "    ${GREEN}✓${NC} dex/config.yaml"

    echo ""
    echo -e "${CYAN}${BOLD}━━━ Next Steps ━━━${NC}"
    echo ""
    echo -e "  ${BOLD}1. Set up DNS / Reverse Proxy${NC}"
    echo ""
    echo -e "     You need to route these hostnames to the Docker services:"
    echo ""
    echo -e "     ${BOLD}Hostname${NC}                              ${BOLD}Target${NC}           ${BOLD}WebSocket${NC}"
    echo -e "     ─────────────────────────────────────────────────────────────────"
    printf "     %-37s %-16s %s\n" "${MAIN_DOMAIN}" "api:3000" "Yes"
    printf "     %-37s %-16s %s\n" "${DEX_DOMAIN}" "dex:5556" "No"
    printf "     %-37s %-16s %s\n" "${PROXY_DOMAIN}" "proxy:4000" "Yes"
    printf "     %-37s %-16s %s\n" "*.${PROXY_DOMAIN}" "proxy:4000" "Yes"
    echo ""

    echo -e "     If using ${BOLD}Nginx Proxy Manager${NC}:"
    echo -e "       - Create 3 proxy hosts (one for each non-wildcard hostname)"
    echo -e "       - For the proxy host, add both ${CYAN}${PROXY_DOMAIN}${NC}"
    echo -e "         and ${CYAN}*.${PROXY_DOMAIN}${NC} as domain names"
    echo -e "       - Enable 'Websockets Support' for the API and Proxy hosts"
    echo ""

    echo -e "     If using ${BOLD}Cloudflare Tunnels${NC}:"
    echo -e "       - Create tunnel routes for all 4 hostnames"
    echo -e "       - Point them to your NPM or directly to the Docker services"
    echo -e "       - Add a wildcard CNAME: ${CYAN}*.${PROXY_DOMAIN}${NC} -> tunnel"
    echo ""

    echo -e "  ${BOLD}2. Start Daytona${NC}"
    echo ""
    echo -e "     ${DIM}\$ docker compose up -d${NC}"
    echo ""

    echo -e "  ${BOLD}3. Verify${NC}"
    echo ""
    echo -e "     ${DIM}\$ curl -s ${PROTOCOL}://${MAIN_DOMAIN}/api/config | jq .${NC}"
    echo ""
    echo -e "     Check that all URLs point to your external domain, not localhost."
    echo ""

    if [[ -n "$SSH_HOST" ]]; then
        echo -e "  ${BOLD}4. SSH Gateway${NC}"
        echo ""
        echo -e "     Make sure port ${CYAN}${SSH_PORT}${NC} on ${CYAN}${SSH_HOST}${NC} is directly"
        echo -e "     reachable from the internet (not through Cloudflare)."
        echo -e "     The docker-compose.yaml maps host port 2222 to the container."
        if [[ "$SSH_PORT" != "2222" ]]; then
            echo ""
            warn "Your external port (${SSH_PORT}) differs from the container port (2222)."
            echo -e "     Update the port mapping in docker-compose.yaml:"
            echo -e "     ${DIM}  ssh-gateway:"
            echo -e "       ports:"
            echo -e "         - \"${SSH_PORT}:2222\"${NC}"
        fi
        echo ""
    fi

    echo -e "  ${BOLD}Documentation:${NC} See DEPLOYMENT-GUIDE.md for detailed instructions."
    echo ""
}

# ── Launch offer ─────────────────────────────────────────────────────────────

offer_launch() {
    prompt_yn DO_LAUNCH "Start Daytona now with 'docker compose up -d'?" "n"

    if [[ "$DO_LAUNCH" == "true" ]]; then
        echo ""
        info "Starting Daytona services..."
        echo ""

        if docker compose up -d; then
            echo ""
            success "Daytona is starting!"
            echo ""
            info "Watch logs: docker compose logs -f api proxy dex"
            info "Dashboard:  ${PROTOCOL}://${MAIN_DOMAIN}"
        else
            echo ""
            error "Failed to start. Check the errors above."
            info "You can try manually: docker compose up -d"
        fi
    else
        echo ""
        info "When ready, run: ${BOLD}docker compose up -d${NC}"
    fi
}

# ── Main ─────────────────────────────────────────────────────────────────────

main() {
    banner

    if [[ "$SKIP_PREFLIGHT" == "true" ]]; then
        warn "Skipping preflight checks (--skip-preflight)"
        # Still need to check we're in the right directory
        if [[ ! -f "docker-compose.yaml" ]]; then
            if [[ -f "docker/docker-compose.yaml" ]]; then
                info "Changing to docker/ directory..."
                cd docker
            else
                error "Cannot find docker-compose.yaml. Run this script from the docker/ directory."
                exit 1
            fi
        fi
        # Default tool availability flags
        HAS_HTPASSWD=false
        HAS_PYTHON3=false
        check_command htpasswd && HAS_HTPASSWD=true
        check_command python3 && HAS_PYTHON3=true
    else
        preflight
    fi
    detect_existing

    configure_domains
    configure_ssh
    configure_dex_admin
    configure_smtp

    if [[ "${SKIP_ENV:-false}" != "true" ]]; then
        generate_secrets
        write_env_file
    fi

    write_dex_config
    show_summary
    offer_launch

    echo ""
    success "Setup complete!"
    echo ""
}

main "$@"
