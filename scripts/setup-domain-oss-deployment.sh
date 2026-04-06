#!/usr/bin/env bash
# Copyright 2026 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

# Daytona Domain Setup
# Automated deployment of Daytona OSS behind a custom domain with Caddy + TLS.
# Supports: Ubuntu, Debian, Fedora, CentOS, RHEL, Rocky, AlmaLinux, macOS
# Usage: ./setup.sh
set -euo pipefail

# ── Colors ──────────────────────────────────────────────────
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# ── Globals (filled during input/detection) ─────────────────
OS="" PKG_MANAGER="" FIREWALL="" ARCH="" CADDY_OS="" CADDY_BIN="" CADDY_CONF_DIR=""
DOMAIN="" EMAIL="" ADMIN_EMAIL="" ADMIN_PASSWORD="" ADMIN_PASSWORD_HASH=""
DNS_PROVIDER_NAME="" DNS_CADDY_MODULE=""
DNS_TLS_BLOCK="" DNS_ENV_NAME="" DNS_ENV_NAME_EXTRA=""
DNS_TOKEN="" DNS_TOKEN_EXTRA=""
ENCRYPTION_KEY="" ENCRYPTION_SALT=""
PROXY_API_KEY="" RUNNER_API_KEY="" SSH_GATEWAY_API_KEY=""
PGADMIN_EMAIL="" PGADMIN_PASSWORD=""
MINIO_USER="" MINIO_PASSWORD=""
DB_USER="" DB_PASS=""
REGISTRY_USER="" REGISTRY_PASSWORD=""
HEALTH_CHECK_KEY="" OTEL_COLLECTOR_KEY=""
CLICKHOUSE_ENABLED="" CLICKHOUSE_HOST_VAL="" CLICKHOUSE_PORT_VAL=""
CLICKHOUSE_USER="" CLICKHOUSE_PASS="" CLICKHOUSE_DB_VAL="" CLICKHOUSE_PROTO=""
REPO_DIR="$HOME/daytona"

# ── Helpers ─────────────────────────────────────────────────
info()    { printf "  ${CYAN}▸${NC} %s\n" "$*"; }
ok()      { printf "  ${GREEN}✓${NC} %s\n" "$*"; }
warn()    { printf "  ${YELLOW}!${NC} %s\n" "$*"; }
fail()    { printf "  ${RED}✗${NC} %s\n" "$*"; }
die()     { fail "$*"; exit 1; }

# Portable sed in-place: BSD sed (macOS) requires -i '', GNU sed requires -i
sedi() {
    if [ "$OS" = "macos" ]; then
        sed -i '' "$@"
    else
        sed -i "$@"
    fi
}

# Replace a literal placeholder in a file with a value.
# Uses awk + ENVIRON so the value never appears in the process command line.
_file_replace() {
    local file="$1" placeholder="$2" val="$3"
    _DAYTONA_VAL="$val" awk -v ph="$placeholder" '
    {
        while ((i = index($0, ph)) > 0)
            $0 = substr($0, 1, i-1) ENVIRON["_DAYTONA_VAL"] substr($0, i + length(ph))
        print
    }' "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"
}

run_step() {
    local name="$1"; shift
    printf "  ${CYAN}▸${NC} %s" "$name"
    local log="/tmp/daytona-step-$$.log"
    install -m 600 /dev/null "$log"
    if "$@" > "$log" 2>&1; then
        printf "\r  ${GREEN}✓${NC} %s\n" "$name"
    else
        printf "\r  ${RED}✗${NC} %s\n" "$name"
        sed 's/^/    /' "$log"
        rm -f "$log"
        exit 1
    fi
    rm -f "$log"
}

# ── Platform Detection ──────────────────────────────────────
detect_platform() {
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64)  ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        arm64)   ARCH="arm64" ;;
        *) die "Unsupported architecture: $ARCH" ;;
    esac

    case "$(uname -s)" in
        Darwin)
            OS="macos"
            PKG_MANAGER="brew"
            FIREWALL="none"
            CADDY_OS="darwin"
            CADDY_BIN="/usr/local/bin/caddy"
            CADDY_CONF_DIR="/usr/local/etc/caddy"
            info "Detected macOS ($ARCH)"
            return
            ;;
        Linux)
            CADDY_OS="linux"
            ;;
        *) die "Unsupported operating system: $(uname -s)" ;;
    esac

    # Linux: detect distro from /etc/os-release
    [ -f /etc/os-release ] || die "Cannot detect OS — /etc/os-release not found"
    # shellcheck disable=SC1091
    . /etc/os-release

    CADDY_BIN="/usr/bin/caddy"
    CADDY_CONF_DIR="/etc/caddy"

    case "${ID:-}" in
        ubuntu|debian)
            OS="$ID"; PKG_MANAGER="apt"; FIREWALL="ufw" ;;
        fedora)
            OS="fedora"; PKG_MANAGER="dnf"; FIREWALL="firewalld" ;;
        centos|rhel|rocky|almalinux)
            OS="$ID"; PKG_MANAGER="dnf"; FIREWALL="firewalld" ;;
        *)
            case "${ID_LIKE:-}" in
                *debian*|*ubuntu*) OS="debian"; PKG_MANAGER="apt"; FIREWALL="ufw" ;;
                *fedora*|*rhel*)   OS="fedora"; PKG_MANAGER="dnf"; FIREWALL="firewalld" ;;
                *) die "Unsupported OS: ${PRETTY_NAME:-$ID}" ;;
            esac ;;
    esac

    info "Detected ${PRETTY_NAME:-$ID} ($ARCH)"
}

# ── Input Collection ────────────────────────────────────────
collect_input() {
    printf "\n${BOLD}  Daytona Domain Setup${NC}\n"
    printf "  ═══════════════════════\n\n"

    # Domain
    while true; do
        printf "  Domain (e.g. daytona.example.com): "; read -r DOMAIN
        DOMAIN="${DOMAIN#https://}"; DOMAIN="${DOMAIN#http://}"
        DOMAIN="$(echo "$DOMAIN" | tr -d '[:space:]')"
        [ -z "$DOMAIN" ] && { fail "Required"; continue; }
        echo "$DOMAIN" | grep -q '\.' || { fail "Must contain a dot"; continue; }
        break
    done

    # DNS provider
    printf "\n  DNS Provider:\n"
    printf "    1) Cloudflare\n"
    printf "    2) DigitalOcean\n"
    printf "    3) AWS Route 53\n"
    printf "    4) Google Cloud DNS\n"
    printf "    5) Hetzner\n"
    printf "    6) Namecheap\n"
    while true; do
        printf "  Select [1-6]: "; read -r choice
        case "$choice" in
            1) DNS_PROVIDER_NAME="Cloudflare"
               DNS_CADDY_MODULE="github.com/caddy-dns/cloudflare"
               DNS_TLS_BLOCK="dns cloudflare {env.CLOUDFLARE_API_TOKEN}"
               DNS_ENV_NAME="CLOUDFLARE_API_TOKEN"; break ;;
            2) DNS_PROVIDER_NAME="DigitalOcean"
               DNS_CADDY_MODULE="github.com/caddy-dns/digitalocean"
               DNS_TLS_BLOCK="dns digitalocean {env.DO_AUTH_TOKEN}"
               DNS_ENV_NAME="DO_AUTH_TOKEN"; break ;;
            3) DNS_PROVIDER_NAME="AWS Route 53"
               DNS_CADDY_MODULE="github.com/caddy-dns/route53"
               DNS_TLS_BLOCK="dns route53 {
            access_key_id {env.AWS_ACCESS_KEY_ID}
            secret_access_key {env.AWS_SECRET_ACCESS_KEY}
        }"
               DNS_ENV_NAME="AWS_ACCESS_KEY_ID"
               DNS_ENV_NAME_EXTRA="AWS_SECRET_ACCESS_KEY"; break ;;
            4) DNS_PROVIDER_NAME="Google Cloud DNS"
               DNS_CADDY_MODULE="github.com/caddy-dns/googleclouddns"
               DNS_TLS_BLOCK="dns googleclouddns {
            gcp_project {env.GCP_PROJECT}
            service_account_json {env.GCP_SERVICE_ACCOUNT_JSON}
        }"
               DNS_ENV_NAME="GCP_PROJECT"
               DNS_ENV_NAME_EXTRA="GCP_SERVICE_ACCOUNT_JSON"; break ;;
            5) DNS_PROVIDER_NAME="Hetzner"
               DNS_CADDY_MODULE="github.com/caddy-dns/hetzner"
               DNS_TLS_BLOCK="dns hetzner {env.HETZNER_API_TOKEN}"
               DNS_ENV_NAME="HETZNER_API_TOKEN"; break ;;
            6) DNS_PROVIDER_NAME="Namecheap"
               DNS_CADDY_MODULE="github.com/caddy-dns/namecheap"
               DNS_TLS_BLOCK="dns namecheap {
            api_key {env.NAMECHEAP_API_KEY}
            user {env.NAMECHEAP_API_USER}
        }"
               DNS_ENV_NAME="NAMECHEAP_API_KEY"
               DNS_ENV_NAME_EXTRA="NAMECHEAP_API_USER"; break ;;
            *) fail "Invalid choice" ;;
        esac
    done

    # DNS credentials
    printf "\n  %s %s: " "$DNS_PROVIDER_NAME" "$DNS_ENV_NAME"
    read -rs DNS_TOKEN; echo
    [ -z "$DNS_TOKEN" ] && die "Token is required"

    if [ -n "${DNS_ENV_NAME_EXTRA:-}" ]; then
        printf "  %s %s: " "$DNS_PROVIDER_NAME" "$DNS_ENV_NAME_EXTRA"
        read -rs DNS_TOKEN_EXTRA; echo
        [ -z "$DNS_TOKEN_EXTRA" ] && die "Required"
    fi

    # Email
    while true; do
        printf "\n  Email (for Let's Encrypt certificates): "; read -r EMAIL
        EMAIL="$(echo "$EMAIL" | tr -d '[:space:]')"
        echo "$EMAIL" | grep -q '@.*\.' && break
        fail "Must be a valid email"
    done

    # Admin email
    while true; do
        printf "  Admin login email: "; read -r ADMIN_EMAIL
        ADMIN_EMAIL="$(echo "$ADMIN_EMAIL" | tr -d '[:space:]')"
        echo "$ADMIN_EMAIL" | grep -q '@' && break
        fail "Must be a valid email"
    done

    # Admin password
    while true; do
        printf "  Admin password (min 8 chars): "; read -rs ADMIN_PASSWORD; echo
        [ ${#ADMIN_PASSWORD} -lt 8 ] && { fail "Too short"; continue; }
        printf "  Confirm password: "; read -rs pw_confirm; echo
        [ "$ADMIN_PASSWORD" = "$pw_confirm" ] && break
        fail "Passwords do not match"
    done

    # Service credentials
    printf "\n${BOLD}  Service Credentials${NC}\n"
    printf "  These replace default placeholder credentials and secure services behind HTTPS.\n\n"

    # PostgreSQL
    while true; do
        printf "  PostgreSQL username: "; read -r DB_USER
        DB_USER="$(echo "$DB_USER" | tr -d '[:space:]')"
        [ -n "$DB_USER" ] && break
        fail "Required"
    done
    while true; do
        printf "  PostgreSQL password (min 8 chars): "; read -rs DB_PASS; echo
        [ ${#DB_PASS} -lt 8 ] && { fail "Too short"; continue; }
        break
    done

    # PgAdmin
    while true; do
        printf "  PgAdmin email: "; read -r PGADMIN_EMAIL
        PGADMIN_EMAIL="$(echo "$PGADMIN_EMAIL" | tr -d '[:space:]')"
        echo "$PGADMIN_EMAIL" | grep -q '@' && break
        fail "Must be a valid email"
    done
    while true; do
        printf "  PgAdmin password (min 8 chars): "; read -rs PGADMIN_PASSWORD; echo
        [ ${#PGADMIN_PASSWORD} -lt 8 ] && { fail "Too short"; continue; }
        break
    done

    # MinIO
    while true; do
        printf "  MinIO admin username: "; read -r MINIO_USER
        MINIO_USER="$(echo "$MINIO_USER" | tr -d '[:space:]')"
        [ -n "$MINIO_USER" ] && break
        fail "Required"
    done
    while true; do
        printf "  MinIO admin password (min 8 chars): "; read -rs MINIO_PASSWORD; echo
        [ ${#MINIO_PASSWORD} -lt 8 ] && { fail "Too short"; continue; }
        break
    done

    # Registry
    while true; do
        printf "  Registry admin username: "; read -r REGISTRY_USER
        REGISTRY_USER="$(echo "$REGISTRY_USER" | tr -d '[:space:]')"
        [ -n "$REGISTRY_USER" ] && break
        fail "Required"
    done
    while true; do
        printf "  Registry admin password (min 8 chars): "; read -rs REGISTRY_PASSWORD; echo
        [ ${#REGISTRY_PASSWORD} -lt 8 ] && { fail "Too short"; continue; }
        break
    done

    # ClickHouse (optional)
    printf "\n  Configure ClickHouse for sandbox telemetry? [y/N] "; read -r ch_yn
    case "${ch_yn:-n}" in
        [yY]*)
            CLICKHOUSE_ENABLED="true"
            printf "  ClickHouse host: "; read -r CLICKHOUSE_HOST_VAL
            CLICKHOUSE_HOST_VAL="$(echo "$CLICKHOUSE_HOST_VAL" | tr -d '[:space:]')"
            [ -z "$CLICKHOUSE_HOST_VAL" ] && die "Host is required"

            printf "  ClickHouse port [8123]: "; read -r CLICKHOUSE_PORT_VAL
            CLICKHOUSE_PORT_VAL="${CLICKHOUSE_PORT_VAL:-8123}"

            printf "  ClickHouse database [otel]: "; read -r CLICKHOUSE_DB_VAL
            CLICKHOUSE_DB_VAL="${CLICKHOUSE_DB_VAL:-otel}"

            printf "  ClickHouse protocol (http/https) [https]: "; read -r CLICKHOUSE_PROTO
            CLICKHOUSE_PROTO="${CLICKHOUSE_PROTO:-https}"

            printf "  ClickHouse username: "; read -r CLICKHOUSE_USER
            CLICKHOUSE_USER="$(echo "$CLICKHOUSE_USER" | tr -d '[:space:]')"

            printf "  ClickHouse password: "; read -rs CLICKHOUSE_PASS; echo
            ;;
    esac

    # Confirmation
    printf "\n${BOLD}  Configuration${NC}\n"
    printf "  ─────────────\n"
    printf "  Domain:       %s\n" "$DOMAIN"
    printf "  DNS Provider: %s\n" "$DNS_PROVIDER_NAME"
    printf "  Email:        %s\n" "$EMAIL"
    printf "  Admin:        %s\n" "$ADMIN_EMAIL"
    printf "  DB user:      %s\n" "$DB_USER"
    printf "  PgAdmin:      %s\n" "$PGADMIN_EMAIL"
    printf "  MinIO user:   %s\n" "$MINIO_USER"
    printf "  Registry:     %s\n" "$REGISTRY_USER"
    [ "$CLICKHOUSE_ENABLED" = "true" ] && printf "  ClickHouse:   %s:%s\n" "$CLICKHOUSE_HOST_VAL" "$CLICKHOUSE_PORT_VAL"
    printf "\n  Proceed? [Y/n] "; read -r yn
    case "${yn:-y}" in [nN]*) echo "  Aborted."; exit 0 ;; esac
}

# ── Steps ───────────────────────────────────────────────────

step_clean() {
    local cf="$REPO_DIR/docker/docker-compose.yaml"
    [ -f "$cf" ] && docker compose -f "$cf" down -v --remove-orphans 2>/dev/null || true
    # Kill any orphaned daytona containers from a previous failed run
    docker ps -aq --filter "name=daytona-" 2>/dev/null | xargs -r docker rm -f 2>/dev/null || true
    # Clean transient files only — preserve $REPO_DIR so re-runs are non-destructive
    # (step_clone is now idempotent and skips when the repo is already present)
    rm -rf /tmp/dashboard-extract /opt/daytona-dashboard-assets "$HOME/.daytona-dashboard-assets"
}

step_packages() {
    command -v docker >/dev/null 2>&1 || die "Docker is not installed. See https://docs.docker.com/engine/install/"
    docker compose version >/dev/null 2>&1 || die "docker compose v2 plugin not found"

    case "$PKG_MANAGER" in
        apt) apt-get update -y && apt-get install -y git curl openssl apache2-utils ufw ;;
        dnf) dnf install -y git curl openssl httpd-tools firewalld ;;
        brew)
            command -v brew >/dev/null 2>&1 || die "Homebrew is not installed. See https://brew.sh"
            command -v htpasswd >/dev/null 2>&1 || brew install httpd
            ;;
    esac
}

step_clone() {
    # Idempotent: clone only if the repo isn't already present.
    # This makes re-runs non-destructive and lets users pre-stage the repo
    # (e.g. for testing local changes) without needing to edit this function.
    [ -d "$REPO_DIR/.git" ] && return 0
    git clone https://github.com/daytonaio/daytona.git "$REPO_DIR"
}

step_secrets() {
    ENCRYPTION_KEY=$(openssl rand -hex 16)
    ENCRYPTION_SALT=$(openssl rand -hex 16)
    PROXY_API_KEY=$(openssl rand -hex 16)
    RUNNER_API_KEY=$(openssl rand -hex 16)
    SSH_GATEWAY_API_KEY=$(openssl rand -hex 16)
    # Hash passwords via stdin (-i) so they never appear in process arguments
    ADMIN_PASSWORD_HASH=$(printf '%s' "$ADMIN_PASSWORD" | htpasswd -niBC 10 "" | cut -d: -f2)
    HEALTH_CHECK_KEY=$(openssl rand -hex 16)
    OTEL_COLLECTOR_KEY=$(openssl rand -hex 16)
}

step_dex() {
    local dir="$REPO_DIR/docker/dex"
    mkdir -p "$dir"

    # Use a quoted heredoc so $2b in the bcrypt hash isn't expanded
    cat > "$dir/config.yaml" <<'DEXEOF'
issuer: https://DOMAIN_PH/dex
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
      - 'https://DOMAIN_PH'
      - 'https://DOMAIN_PH/api/oauth2-redirect.html'
      - 'https://DOMAIN_PH/callback'
      - 'https://proxy.DOMAIN_PH/callback'
    name: 'Daytona'
    public: true
enablePasswordDB: true
staticPasswords:
  - email: 'ADMIN_EMAIL_PH'
    hash: 'ADMIN_HASH_PH'
    username: 'admin'
    userID: '1234'
DEXEOF

    _file_replace "$dir/config.yaml" "DOMAIN_PH"     "$DOMAIN"
    _file_replace "$dir/config.yaml" "ADMIN_EMAIL_PH" "$ADMIN_EMAIL"
    _file_replace "$dir/config.yaml" "ADMIN_HASH_PH"  "$ADMIN_PASSWORD_HASH"
}

step_compose() {
    local cf="$REPO_DIR/docker/docker-compose.yaml"

    # Helper: replace KEY: value and - KEY=value lines in docker-compose.
    # Uses awk + ENVIRON so credential values never appear in process arguments.
    _set() {
        local key="$1" val="$2"
        _DAYTONA_VAL="$val" awk -v key="$key" '
        {
            if (match($0, "^[[:space:]]*" key ":")) {
                n = index($0, key ":")
                print substr($0, 1, n-1) key ": " ENVIRON["_DAYTONA_VAL"]
            } else if (match($0, "^[[:space:]]*-[[:space:]]*" key "=")) {
                n = index($0, key "=")
                print substr($0, 1, n-1) key "=" ENVIRON["_DAYTONA_VAL"]
            } else {
                print
            }
        }' "$cf" > "${cf}.tmp" && mv "${cf}.tmp" "$cf"
    }

    _set ENCRYPTION_KEY        "$ENCRYPTION_KEY"
    _set ENCRYPTION_SALT       "$ENCRYPTION_SALT"
    _set PROXY_DOMAIN          "proxy.$DOMAIN"
    _set PROXY_TEMPLATE_URL    "https://{{PORT}}-{{sandboxId}}.proxy.$DOMAIN"
    _set PROXY_API_KEY         "$PROXY_API_KEY"
    _set PROXY_PROTOCOL        "https"
    _set DASHBOARD_URL         "https://$DOMAIN/dashboard"
    _set DASHBOARD_BASE_API_URL "https://$DOMAIN"
    _set PUBLIC_OIDC_DOMAIN    "https://$DOMAIN/dex"
    _set SSH_GATEWAY_URL       "$DOMAIN:2222"
    _set SSH_GATEWAY_COMMAND   "ssh -p 2222 {{TOKEN}}@$DOMAIN"
    _set SSH_GATEWAY_API_KEY   "$SSH_GATEWAY_API_KEY"
    _set DEFAULT_RUNNER_API_KEY "$RUNNER_API_KEY"
    _set COOKIE_DOMAIN         "proxy.$DOMAIN"
    _set OIDC_PUBLIC_DOMAIN    "https://$DOMAIN/dex"
    _set DAYTONA_RUNNER_TOKEN  "$RUNNER_API_KEY"

    # API_KEY inside ssh-gateway section only
    # Uses ENVIRON so the key value never appears in process arguments
    _DAYTONA_VAL="$SSH_GATEWAY_API_KEY" awk '
        /^  ssh-gateway:/ { in_svc=1 }
        in_svc && /^  [a-z]/ && !/ssh-gateway:/ { in_svc=0 }
        in_svc && /API_KEY:/ {
            n = index($0, "API_KEY:")
            $0 = substr($0, 1, n-1) "API_KEY: " ENVIRON["_DAYTONA_VAL"]
        }
        in_svc && /- API_KEY=/ {
            n = index($0, "API_KEY=")
            $0 = substr($0, 1, n-1) "API_KEY=" ENVIRON["_DAYTONA_VAL"]
        }
        { print }
    ' "$cf" > "${cf}.tmp" && mv "${cf}.tmp" "$cf"

    # Dex ports
    if ! grep -q '5556:5556' "$cf"; then
        awk '
            /^  dex:/ { in_dex=1; print; next }
            in_dex && /^  [a-z]/ && !/dex:/ {
                print "    ports:"; print "      - \"5556:5556\""
                in_dex=0
            }
            { print }
        ' "$cf" > "${cf}.tmp" && mv "${cf}.tmp" "$cf"
    fi

    # Remap auxiliary service ports to localhost-only offset ports so they don't
    # collide with Caddy, which listens on the original ports for HTTPS.
    # Actual compose file uses: 5050:80 (pgadmin), 5100:80 (registry-ui), 9001:9001 (minio)
    _remap_port() {
        local host_port="$1" new_host_port="$2"
        awk -v hp="$host_port" -v nhp="$new_host_port" '{
            # Match "- HP:CP" or "- \"HP:CP\"" where HP is the host port
            if (match($0, hp ":[0-9]+")) {
                sub(hp ":", "127.0.0.1:" nhp ":")
            }
            print
        }' "$cf" > "${cf}.tmp" && mv "${cf}.tmp" "$cf"
    }
    _remap_port 5050 15050
    _remap_port 5100 15100
    _remap_port 9001 19001

    # Update PostgreSQL credentials (db service + API service connection)
    _set POSTGRES_USER         "$DB_USER"
    _set POSTGRES_PASSWORD     "$DB_PASS"
    _set DB_USERNAME           "$DB_USER"
    _set DB_PASSWORD           "$DB_PASS"

    # Update PgAdmin credentials and enable server mode (default is desktop mode with no login)
    _set PGADMIN_DEFAULT_EMAIL    "$PGADMIN_EMAIL"
    _set PGADMIN_DEFAULT_PASSWORD "$PGADMIN_PASSWORD"
    _set PGADMIN_CONFIG_SERVER_MODE       "'True'"
    _set PGADMIN_CONFIG_MASTER_PASSWORD_REQUIRED "'True'"

    # Update MinIO credentials (also used by API and Runner for S3 access)
    _set MINIO_ROOT_USER       "$MINIO_USER"
    _set MINIO_ROOT_PASSWORD   "$MINIO_PASSWORD"
    _set S3_ACCESS_KEY         "$MINIO_USER"
    _set S3_SECRET_KEY         "$MINIO_PASSWORD"
    _set AWS_ACCESS_KEY_ID     "$MINIO_USER"
    _set AWS_SECRET_ACCESS_KEY "$MINIO_PASSWORD"

    # Update Registry credentials (transient + internal point to same registry)
    _set TRANSIENT_REGISTRY_ADMIN    "$REGISTRY_USER"
    _set TRANSIENT_REGISTRY_PASSWORD "$REGISTRY_PASSWORD"
    _set INTERNAL_REGISTRY_ADMIN     "$REGISTRY_USER"
    _set INTERNAL_REGISTRY_PASSWORD  "$REGISTRY_PASSWORD"

    # Auto-generated API keys
    _set HEALTH_CHECK_API_KEY    "$HEALTH_CHECK_KEY"
    _set OTEL_COLLECTOR_API_KEY  "$OTEL_COLLECTOR_KEY"

    # ClickHouse (optional)
    if [ "$CLICKHOUSE_ENABLED" = "true" ]; then
        _set CLICKHOUSE_HOST     "$CLICKHOUSE_HOST_VAL"
        _set CLICKHOUSE_PORT     "$CLICKHOUSE_PORT_VAL"
        _set CLICKHOUSE_DATABASE "$CLICKHOUSE_DB_VAL"
        _set CLICKHOUSE_USERNAME "$CLICKHOUSE_USER"
        _set CLICKHOUSE_PASSWORD "$CLICKHOUSE_PASS"
        _set CLICKHOUSE_PROTOCOL "$CLICKHOUSE_PROTO"
    fi

    # SELinux :z labels on bind-mount volumes — Linux only
    if [ "$OS" != "macos" ]; then
        sedi -E '/^\s*-\s*(\.\/|\/)[^:]+:[^:]+$/s/$/:z/' "$cf"
    fi

    # Restrict permissions — compose file now contains plaintext credentials
    chmod 600 "$cf"
}

step_firewall() {
    case "$FIREWALL" in
        ufw)
            ufw allow 22/tcp && ufw allow 80/tcp && ufw allow 443/tcp && ufw allow 2222/tcp \
                && ufw allow 5050/tcp && ufw allow 5100/tcp && ufw allow 9001/tcp
            ufw --force enable
            ;;
        firewalld)
            systemctl enable --now firewalld
            for p in 22/tcp 80/tcp 443/tcp 2222/tcp 5050/tcp 5100/tcp 9001/tcp; do
                firewall-cmd --permanent --add-port="$p"
            done
            firewall-cmd --permanent --add-masquerade
            firewall-cmd --reload
            # SELinux: Caddy's binary at /usr/bin/caddy is auto-labeled httpd_exec_t on
            # Fedora and runs in the httpd_t domain under systemd. By default httpd_t cannot
            # bind the auxiliary ports (5050/5100/9001), so we relabel them. Caddy's storage
            # is moved to /var/lib/caddy (the conventional location) via XDG_DATA_HOME in the
            # systemd unit, which avoids the /root home-directory traversal entirely.
            # No-op if SELinux is disabled.
            if command -v getenforce >/dev/null 2>&1 && [ "$(getenforce 2>/dev/null)" != "Disabled" ]; then
                command -v semanage >/dev/null 2>&1 || dnf install -y policycoreutils-python-utils >/dev/null 2>&1 || true
                if command -v semanage >/dev/null 2>&1; then
                    # Allow httpd_t to bind the aux ports (TCP for h1/h2, UDP for h3 QUIC)
                    for proto in tcp udp; do
                        for p in 5050 5100 9001; do
                            semanage port -m -t http_port_t -p "$proto" "$p" 2>/dev/null || \
                                semanage port -a -t http_port_t -p "$proto" "$p" 2>/dev/null || true
                        done
                    done
                    # Pre-create Caddy storage at the conventional /var/lib/caddy location
                    # and label it so httpd_t can read/write its certs and ACME state.
                    mkdir -p /var/lib/caddy
                    chmod 700 /var/lib/caddy
                    semanage fcontext -a -t httpd_var_lib_t '/var/lib/caddy(/.*)?' 2>/dev/null || true
                    restorecon -R /var/lib/caddy 2>/dev/null || true
                fi
            fi
            ;;
        none) true ;;
    esac
}

step_caddy_install() {
    if [ "$OS" = "macos" ]; then
        launchctl bootout "gui/$(id -u)/com.caddyserver.caddy" 2>/dev/null || true
    else
        systemctl stop caddy 2>/dev/null || true
    fi
    local url="https://caddyserver.com/api/download?os=${CADDY_OS}&arch=${ARCH}&p=${DNS_CADDY_MODULE}"
    if [ "$OS" = "macos" ]; then
        sudo mkdir -p "$(dirname "$CADDY_BIN")"
        sudo curl -fsSL "$url" -o "$CADDY_BIN"
        sudo chmod +x "$CADDY_BIN"
    else
        curl -fsSL "$url" -o "$CADDY_BIN"
        chmod +x "$CADDY_BIN"
    fi
}

step_caddy_configure() {
    if [ "$OS" = "macos" ]; then
        sudo mkdir -p "$CADDY_CONF_DIR"
        sudo chown "$(whoami)" "$CADDY_CONF_DIR"
    else
        mkdir -p "$CADDY_CONF_DIR"
    fi

    # Caddyfile
    cat > "$CADDY_CONF_DIR/Caddyfile" <<CADDYEOF
{
    admin off
    email $EMAIL
}

$DOMAIN {
    tls {
        $DNS_TLS_BLOCK
    }

    handle /dex/* {
        reverse_proxy localhost:5556
    }

    handle {
        reverse_proxy localhost:3000
    }
}

proxy.$DOMAIN, *.proxy.$DOMAIN {
    tls {
        $DNS_TLS_BLOCK
    }

    reverse_proxy localhost:4000
}

$DOMAIN:5050 {
    tls {
        $DNS_TLS_BLOCK
    }
    basic_auth {
        PGADMIN_AUTH_PH
    }
    reverse_proxy localhost:15050
}

$DOMAIN:9001 {
    tls {
        $DNS_TLS_BLOCK
    }
    reverse_proxy localhost:19001
}

$DOMAIN:5100 {
    tls {
        $DNS_TLS_BLOCK
    }
    basic_auth {
        REGISTRY_AUTH_PH
    }
    reverse_proxy localhost:15100
}
CADDYEOF

    # Generate basicauth hashes using Caddy's own hasher — guaranteed format compatibility
    # Hash passwords via stdin so they never appear in process arguments.
    # Note: Caddy 2.11+ requires a newline-terminated password on stdin (the trailing \n
    # is significant — without it the reader returns EOF before getting any input).
    local pgadmin_hash registry_hash
    pgadmin_hash=$(printf '%s\n' "$PGADMIN_PASSWORD" | "$CADDY_BIN" hash-password)
    registry_hash=$(printf '%s\n' "$ADMIN_PASSWORD" | "$CADDY_BIN" hash-password)
    _file_replace "$CADDY_CONF_DIR/Caddyfile" "PGADMIN_AUTH_PH" "$PGADMIN_EMAIL $pgadmin_hash"
    _file_replace "$CADDY_CONF_DIR/Caddyfile" "REGISTRY_AUTH_PH" "admin $registry_hash"

    # Environment file
    printf '%s=%s\n' "$DNS_ENV_NAME" "$DNS_TOKEN" > "$CADDY_CONF_DIR/environment"
    [ -n "${DNS_ENV_NAME_EXTRA:-}" ] && [ -n "${DNS_TOKEN_EXTRA:-}" ] && \
        printf '%s=%s\n' "$DNS_ENV_NAME_EXTRA" "$DNS_TOKEN_EXTRA" >> "$CADDY_CONF_DIR/environment"
    chmod 600 "$CADDY_CONF_DIR/environment"

    if [ "$OS" = "macos" ]; then
        # Wrapper script to load env vars and run Caddy
        sudo tee /usr/local/bin/caddy-daytona > /dev/null <<WRAPEOF
#!/bin/bash
set -a
source "$CADDY_CONF_DIR/environment"
set +a
exec "$CADDY_BIN" run --config "$CADDY_CONF_DIR/Caddyfile" --adapter caddyfile
WRAPEOF
        sudo chmod +x /usr/local/bin/caddy-daytona

        # LaunchAgent plist
        mkdir -p "$HOME/Library/LaunchAgents"
        cat > "$HOME/Library/LaunchAgents/com.caddyserver.caddy.plist" <<'PLISTEOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.caddyserver.caddy</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/caddy-daytona</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/usr/local/var/log/caddy/access.log</string>
    <key>StandardErrorPath</key>
    <string>/usr/local/var/log/caddy/error.log</string>
</dict>
</plist>
PLISTEOF
        sudo mkdir -p /usr/local/var/log/caddy
        sudo chown "$(whoami)" /usr/local/var/log/caddy
        launchctl bootstrap "gui/$(id -u)" "$HOME/Library/LaunchAgents/com.caddyserver.caddy.plist"
    else
        # Systemd unit
        cat > /etc/systemd/system/caddy.service <<'SVCEOF'
[Unit]
Description=Caddy web server
After=network.target network-online.target
Requires=network-online.target

[Service]
# Caddy stores certs/state at $XDG_DATA_HOME/caddy and autosave at $XDG_CONFIG_HOME/caddy.
# Pointing both at /var/lib puts everything in /var/lib/caddy — the conventional location,
# matches Fedora's packaged caddy, and avoids any /root home-directory traversal under SELinux.
Environment=XDG_DATA_HOME=/var/lib
Environment=XDG_CONFIG_HOME=/var/lib
EnvironmentFile=/etc/caddy/environment
ExecStart=/usr/bin/caddy run --config /etc/caddy/Caddyfile --adapter caddyfile
ExecReload=/usr/bin/caddy reload --config /etc/caddy/Caddyfile --adapter caddyfile
TimeoutStopSec=5s
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
SVCEOF

        systemctl daemon-reload
        systemctl enable --now caddy

        # If Caddy failed to start, run it directly to capture the actual error
        if ! systemctl is-active --quiet caddy; then
            echo "Caddy failed to start via systemd. Capturing error:"
            (set -a; source "$CADDY_CONF_DIR/environment"; set +a
             timeout 5 "$CADDY_BIN" run --config "$CADDY_CONF_DIR/Caddyfile" --adapter caddyfile 2>&1) || true
            return 1
        fi
    fi
}

step_docker_start() {
    local cf="$REPO_DIR/docker/docker-compose.yaml"

    docker compose -f "$cf" up -d

    # Wait for services (90s timeout)
    local deadline=$(( $(date +%s) + 90 ))
    for svc in api proxy dex ssh-gateway; do
        while true; do
            [ "$(date +%s)" -gt "$deadline" ] && {
                echo "Timeout waiting for $svc"
                docker compose -f "$cf" logs "$svc" --tail 10
                return 1
            }
            local state
            state=$(docker compose -f "$cf" ps --format '{{.State}}' "$svc" 2>/dev/null || true)
            case "$state" in
                running|healthy) break ;;
                exited|dead)
                    echo "Service $svc failed:"
                    docker compose -f "$cf" logs "$svc" --tail 10
                    return 1 ;;
            esac
            sleep 3
        done
    done

    # Let API finish initializing
    sleep 10
}

step_verify() {
    local cf="$REPO_DIR/docker/docker-compose.yaml"
    local passed=0 failed=0

    printf "\n${BOLD}  Verification Results${NC}\n\n"

    # Container health
    local all_ok=true
    for svc in api proxy dex ssh-gateway; do
        local st
        st=$(docker compose -f "$cf" ps --format '{{.State}}' "$svc" 2>/dev/null || true)
        [ "$st" = "running" ] || [ "$st" = "healthy" ] || all_ok=false
    done
    if $all_ok; then ok "Docker container health"; passed=$((passed+1))
    else fail "Docker container health"; failed=$((failed+1)); fi

    # Dex OIDC
    if curl -sf --max-time 5 "http://localhost:5556/dex/.well-known/openid-configuration" | grep -q "https://$DOMAIN/dex"; then
        ok "Dex OIDC issuer"; passed=$((passed+1))
    else fail "Dex OIDC issuer"; failed=$((failed+1)); fi

    # DNS — use host(1) on macOS, getent on Linux
    if [ "$OS" = "macos" ]; then
        if host "$DOMAIN" >/dev/null 2>&1; then
            ok "Base domain DNS"; passed=$((passed+1))
        else fail "Base domain DNS"; failed=$((failed+1)); fi

        if host "proxy.$DOMAIN" >/dev/null 2>&1; then
            ok "Proxy wildcard DNS"; passed=$((passed+1))
        else fail "Proxy wildcard DNS"; failed=$((failed+1)); fi
    else
        if getent hosts "$DOMAIN" >/dev/null 2>&1; then
            ok "Base domain DNS"; passed=$((passed+1))
        else fail "Base domain DNS"; failed=$((failed+1)); fi

        if getent hosts "proxy.$DOMAIN" >/dev/null 2>&1; then
            ok "Proxy wildcard DNS"; passed=$((passed+1))
        else fail "Proxy wildcard DNS"; failed=$((failed+1)); fi
    fi

    # HTTPS — wait up to 120s for TLS certificate
    info "Waiting for TLS certificate (up to 120s)..."
    local https_ok=false code=""
    for _ in $(seq 1 24); do
        code=$(curl -sk -o /dev/null -w '%{http_code}' --max-time 5 "https://$DOMAIN" 2>/dev/null || echo "000")
        case "$code" in 200|301|302) https_ok=true; break ;; esac
        sleep 5
    done
    if $https_ok; then ok "HTTPS dashboard (HTTP $code)"; passed=$((passed+1))
    else
        fail "HTTPS dashboard (HTTP $code)"; failed=$((failed+1))
        # Check if Caddy is running and diagnose why TLS failed
        if [ "$OS" != "macos" ]; then
            if ! systemctl is-active --quiet caddy 2>/dev/null; then
                warn "Caddy is not running — check: systemctl status caddy.service"
            else
                # Caddy is running but TLS failed — check for rate limiting
                local caddy_err
                caddy_err=$(journalctl -u caddy.service --since "5 min ago" --no-pager 2>/dev/null || true)
                if echo "$caddy_err" | grep -qi "rateLimited" 2>/dev/null; then
                    local retry_time
                    retry_time=$(echo "$caddy_err" | grep -o 'retry after [0-9A-Z: -]*' | tail -1)
                    warn "Let's Encrypt rate limit hit — $retry_time"
                    warn "Caddy will retry automatically. Leave it running."
                elif echo "$caddy_err" | grep -qi "error" 2>/dev/null; then
                    warn "Caddy reported errors — check: journalctl -u caddy.service --no-pager | tail -20"
                else
                    warn "TLS certificates may still be issuing. Caddy is running and will retry."
                fi
            fi
        fi
    fi

    # Wildcard TLS
    local proxy_ip
    if [ "$OS" = "macos" ]; then
        proxy_ip=$(dig +short "proxy.$DOMAIN" A 2>/dev/null | head -1)
    else
        proxy_ip=$(getent hosts "proxy.$DOMAIN" 2>/dev/null | awk '{print $1; exit}' || true)
    fi
    if [ -n "$proxy_ip" ]; then
        local san
        san=$(echo | openssl s_client -servername "test.proxy.$DOMAIN" -connect "$proxy_ip:443" 2>/dev/null \
            | openssl x509 -noout -text 2>/dev/null | grep -A1 "Subject Alternative Name" || true)
        if echo "$san" | grep -q "\\*.proxy.$DOMAIN"; then
            ok "Wildcard TLS certificate"; passed=$((passed+1))
        else fail "Wildcard TLS certificate (may still be issuing)"; failed=$((failed+1)); fi
    else fail "Wildcard TLS certificate (proxy DNS not resolving)"; failed=$((failed+1)); fi

    # SSH Gateway — nc for macOS (no timeout command), bash /dev/tcp for Linux
    if [ "$OS" = "macos" ]; then
        if nc -z -w 5 "$DOMAIN" 2222 2>/dev/null; then
            ok "SSH Gateway port 2222"; passed=$((passed+1))
        else fail "SSH Gateway port 2222"; failed=$((failed+1)); fi
    else
        if timeout 5 bash -c "echo >/dev/tcp/$DOMAIN/2222" 2>/dev/null; then
            ok "SSH Gateway port 2222"; passed=$((passed+1))
        else fail "SSH Gateway port 2222"; failed=$((failed+1)); fi
    fi

    printf "\n  %d passed, %d failed\n" "$passed" "$failed"

    if [ "$failed" -eq 0 ]; then
        printf "\n${GREEN}${BOLD}  Setup complete!${NC}\n"
    else
        printf "\n${YELLOW}  Some checks failed. TLS certificates may still be issuing — retry in a few minutes.${NC}\n"
    fi

    printf "\n${BOLD}  Endpoints${NC}\n"
    printf "  ─────────\n"
    printf "  Dashboard:   https://%s\n" "$DOMAIN"
    printf "               Login: %s\n" "$ADMIN_EMAIL"
    printf "  PgAdmin:     https://%s:5050\n" "$DOMAIN"
    printf "               Basic Auth: %s / (password set during setup)\n" "$PGADMIN_EMAIL"
    printf "               PgAdmin Login: same email and password\n"
    printf "  MinIO:       https://%s:9001\n" "$DOMAIN"
    printf "               Login: %s / (password set during setup)\n" "$MINIO_USER"
    printf "  Registry UI: https://%s:5100\n" "$DOMAIN"
    printf "               Basic Auth: admin / (admin password set during setup)\n"
    printf "\n"
}

# ── Main ────────────────────────────────────────────────────
main() {
    # Root required on Linux, not on macOS (uses sudo where needed)
    if [ "$(uname -s)" != "Darwin" ] && [ "$(id -u)" -ne 0 ]; then
        die "This script must be run as root"
    fi

    detect_platform
    collect_input

    printf "\n${BOLD}  Setting up Daytona...${NC}\n\n"

    run_step "Cleaning previous installation"    step_clean
    run_step "Installing system packages"        step_packages
    run_step "Cloning Daytona repository"        step_clone
    run_step "Generating security credentials"   step_secrets
    run_step "Configuring Dex OIDC provider"     step_dex
    run_step "Configuring Docker Compose"        step_compose
    run_step "Configuring firewall rules"        step_firewall
    run_step "Installing Caddy with DNS module"  step_caddy_install
    run_step "Configuring Caddy reverse proxy"   step_caddy_configure
    run_step "Starting Docker services"          step_docker_start

    # Pull sandbox image with visible progress (blocking — sandboxes won't work without it)
    printf "  ${CYAN}▸${NC} Pulling sandbox image (this may take a few minutes)\n"
    if docker pull daytonaio/sandbox:0.5.0-slim; then
        printf "  ${GREEN}✓${NC} Sandbox image ready\n"
    else
        printf "  ${RED}✗${NC} Failed to pull sandbox image — sandbox creation will download it on first use\n"
    fi

    step_verify
}

main "$@"
