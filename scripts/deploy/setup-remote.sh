#!/bin/bash
set -e

# =============================================================================
# Daytona Remote Server Setup Script
# Sets up Docker, nginx, and SSL on a fresh Ubuntu 24 server
# =============================================================================

# Configuration
DOMAIN="${DOMAIN:-win.trydaytona.com}"
PROXY_DOMAIN="${PROXY_DOMAIN:-preview.win.trydaytona.com}"
CERTBOT_EMAIL="${CERTBOT_EMAIL:-admin@daytona.io}"
DEPLOY_PATH="${DEPLOY_PATH:-/opt/daytona/scripts/deploy}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# =============================================================================
# System Updates
# =============================================================================
install_system_packages() {
    log_info "Updating system packages..."
    apt-get update
    apt-get upgrade -y
    
    log_info "Installing required packages..."
    apt-get install -y \
        apt-transport-https \
        ca-certificates \
        curl \
        gnupg \
        lsb-release \
        software-properties-common \
        nginx \
        certbot \
        python3-certbot-nginx \
        rsync \
        git \
        htop \
        vim
    
    log_success "System packages installed!"
}

# =============================================================================
# Docker Installation
# =============================================================================
install_docker() {
    if command -v docker &> /dev/null; then
        log_info "Docker already installed: $(docker --version)"
        return
    fi
    
    log_info "Installing Docker..."
    
    # Add Docker's official GPG key
    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
    chmod a+r /etc/apt/keyrings/docker.asc
    
    # Add the repository to Apt sources
    echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
        $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
        tee /etc/apt/sources.list.d/docker.list > /dev/null
    
    apt-get update
    apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    
    # Start and enable Docker
    systemctl start docker
    systemctl enable docker
    
    log_success "Docker installed: $(docker --version)"
}

# =============================================================================
# Nginx Configuration
# =============================================================================
configure_nginx() {
    log_info "Configuring nginx..."
    
    # Create nginx config directory for our configs
    mkdir -p /etc/nginx/sites-available
    mkdir -p /etc/nginx/sites-enabled
    
    # Create directory for certbot challenges
    mkdir -p /var/www/certbot
    
    # Copy nginx configs from deploy folder
    if [ -f "$DEPLOY_PATH/nginx/daytona.conf" ]; then
        cp "$DEPLOY_PATH/nginx/daytona.conf" /etc/nginx/sites-available/daytona.conf
    fi
    
    if [ -f "$DEPLOY_PATH/nginx/proxy.conf" ]; then
        cp "$DEPLOY_PATH/nginx/proxy.conf" /etc/nginx/sites-available/proxy.conf
    fi
    
    # Create temporary configs for initial cert acquisition (HTTP only)
    cat > /etc/nginx/sites-available/daytona-temp.conf << EOF
server {
    listen 80;
    listen [::]:80;
    server_name $DOMAIN;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 200 'Waiting for SSL setup...';
        add_header Content-Type text/plain;
    }
}
EOF

    cat > /etc/nginx/sites-available/proxy-temp.conf << EOF
server {
    listen 80;
    listen [::]:80;
    server_name *.$PROXY_DOMAIN $PROXY_DOMAIN;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 200 'Waiting for SSL setup...';
        add_header Content-Type text/plain;
    }
}
EOF
    
    # Enable temporary configs
    rm -f /etc/nginx/sites-enabled/default
    ln -sf /etc/nginx/sites-available/daytona-temp.conf /etc/nginx/sites-enabled/
    ln -sf /etc/nginx/sites-available/proxy-temp.conf /etc/nginx/sites-enabled/
    
    # Test and reload nginx
    nginx -t
    systemctl reload nginx
    
    log_success "Nginx configured!"
}

# =============================================================================
# SSL Certificate Setup
# =============================================================================
setup_ssl() {
    log_info "Setting up SSL certificates with Let's Encrypt..."
    
    # Check if certificates already exist
    if [ -d "/etc/letsencrypt/live/$DOMAIN" ]; then
        log_info "SSL certificate for $DOMAIN already exists"
    else
        log_info "Obtaining SSL certificate for $DOMAIN..."
        certbot certonly --webroot \
            -w /var/www/certbot \
            -d "$DOMAIN" \
            --email "$CERTBOT_EMAIL" \
            --agree-tos \
            --non-interactive
    fi
    
    # For wildcard cert, we need DNS challenge
    if [ -d "/etc/letsencrypt/live/$PROXY_DOMAIN" ]; then
        log_info "SSL certificate for $PROXY_DOMAIN already exists"
    else
        log_warning "Wildcard certificate for *.$PROXY_DOMAIN requires DNS challenge."
        log_warning "You'll need to manually run:"
        echo ""
        echo "  certbot certonly --manual --preferred-challenges dns \\"
        echo "    -d '$PROXY_DOMAIN' -d '*.$PROXY_DOMAIN' \\"
        echo "    --email '$CERTBOT_EMAIL' --agree-tos"
        echo ""
        log_warning "After DNS validation, run this script again or manually enable nginx configs."
        
        # Create a placeholder directory so nginx doesn't fail
        mkdir -p /etc/letsencrypt/live/$PROXY_DOMAIN
        
        # For now, use the main domain cert as a fallback
        if [ -d "/etc/letsencrypt/live/$DOMAIN" ]; then
            ln -sf /etc/letsencrypt/live/$DOMAIN/fullchain.pem /etc/letsencrypt/live/$PROXY_DOMAIN/fullchain.pem
            ln -sf /etc/letsencrypt/live/$DOMAIN/privkey.pem /etc/letsencrypt/live/$PROXY_DOMAIN/privkey.pem
        fi
    fi
    
    # Create ssl-dhparams if not exists
    if [ ! -f /etc/letsencrypt/ssl-dhparams.pem ]; then
        log_info "Generating DH parameters (this may take a while)..."
        openssl dhparam -out /etc/letsencrypt/ssl-dhparams.pem 2048
    fi
    
    # Create options-ssl-nginx.conf if not exists
    if [ ! -f /etc/letsencrypt/options-ssl-nginx.conf ]; then
        cat > /etc/letsencrypt/options-ssl-nginx.conf << 'EOF'
ssl_session_cache shared:le_nginx_SSL:10m;
ssl_session_timeout 1440m;
ssl_session_tickets off;
ssl_protocols TLSv1.2 TLSv1.3;
ssl_prefer_server_ciphers off;
ssl_ciphers "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384";
EOF
    fi
    
    log_success "SSL setup completed!"
}

# =============================================================================
# Enable Full Nginx Configuration
# =============================================================================
enable_nginx_ssl() {
    log_info "Enabling SSL nginx configurations..."
    
    # Remove temporary configs
    rm -f /etc/nginx/sites-enabled/daytona-temp.conf
    rm -f /etc/nginx/sites-enabled/proxy-temp.conf
    
    # Enable full configs
    if [ -f /etc/nginx/sites-available/daytona.conf ]; then
        ln -sf /etc/nginx/sites-available/daytona.conf /etc/nginx/sites-enabled/
    fi
    
    if [ -f /etc/nginx/sites-available/proxy.conf ]; then
        ln -sf /etc/nginx/sites-available/proxy.conf /etc/nginx/sites-enabled/
    fi
    
    # Test and reload
    if nginx -t; then
        systemctl reload nginx
        log_success "Nginx SSL configuration enabled!"
    else
        log_error "Nginx configuration test failed. Check the configs."
        exit 1
    fi
}

# =============================================================================
# Firewall Setup
# =============================================================================
setup_firewall() {
    log_info "Configuring firewall..."
    
    # Install ufw if not present
    apt-get install -y ufw
    
    # Allow SSH
    ufw allow 22/tcp
    
    # Allow HTTP and HTTPS
    ufw allow 80/tcp
    ufw allow 443/tcp
    
    # Allow SSH Gateway port
    ufw allow 2222/tcp
    
    # Enable firewall
    ufw --force enable
    
    log_success "Firewall configured!"
}

# =============================================================================
# Setup Certbot Auto-renewal
# =============================================================================
setup_certbot_renewal() {
    log_info "Setting up certbot auto-renewal..."
    
    # Create renewal hook to reload nginx
    mkdir -p /etc/letsencrypt/renewal-hooks/deploy
    cat > /etc/letsencrypt/renewal-hooks/deploy/reload-nginx.sh << 'EOF'
#!/bin/bash
systemctl reload nginx
EOF
    chmod +x /etc/letsencrypt/renewal-hooks/deploy/reload-nginx.sh
    
    # Test renewal
    certbot renew --dry-run || true
    
    log_success "Certbot auto-renewal configured!"
}

# =============================================================================
# Main Setup Flow
# =============================================================================
main() {
    log_info "Starting Daytona server setup..."
    log_info "Domain: $DOMAIN"
    log_info "Proxy Domain: $PROXY_DOMAIN"
    echo ""
    
    # Check if running as root
    if [ "$EUID" -ne 0 ]; then
        log_error "Please run as root"
        exit 1
    fi
    
    install_system_packages
    install_docker
    configure_nginx
    setup_ssl
    enable_nginx_ssl
    setup_firewall
    setup_certbot_renewal
    
    echo ""
    log_success "========================================="
    log_success "Server setup completed!"
    log_success "========================================="
    echo ""
    echo "Next steps:"
    echo "1. If you need a wildcard cert for the proxy domain, run:"
    echo "   certbot certonly --manual --preferred-challenges dns \\"
    echo "     -d '$PROXY_DOMAIN' -d '*.$PROXY_DOMAIN'"
    echo ""
    echo "2. Configure your DNS records:"
    echo "   A record: $DOMAIN -> <server-ip>"
    echo "   A record: *.$PROXY_DOMAIN -> <server-ip>"
    echo ""
    echo "3. Copy env.example to .env and configure your settings:"
    echo "   cd $DEPLOY_PATH && cp env.example .env && vim .env"
    echo ""
    echo "4. Build and start the services:"
    echo "   cd $DEPLOY_PATH && docker compose -f docker-compose.production.yaml up -d"
    echo ""
}

# Run main
main "$@"


