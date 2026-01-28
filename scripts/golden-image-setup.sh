#!/bin/bash
# Golden Image Setup Script
# This script installs all dependencies to match the container-based sandbox
# Run this inside the Cloud Hypervisor VM

set -e

echo "=== Starting Golden Image Setup ==="
echo "Date: $(date)"

# Update package lists
echo "=== Updating package lists ==="
apt-get update

# Install system packages (matching Dockerfile)
echo "=== Installing system packages ==="
DEBIAN_FRONTEND=noninteractive apt-get install -y \
    curl \
    sudo \
    python3-pip \
    python3-venv \
    ripgrep \
    chromium-browser \
    iputils-ping \
    bind9-dnsutils \
    # X11 libraries required for computer use plugin
    libx11-6 \
    libxrandr2 \
    libxext6 \
    libxrender1 \
    libxfixes3 \
    libxss1 \
    libxtst6 \
    libxi6 \
    # VNC and desktop environment for computer use
    xvfb \
    x11vnc \
    novnc \
    xfce4 \
    xfce4-terminal \
    dbus-x11 \
    # Additional useful tools
    git \
    wget \
    unzip \
    build-essential \
    locales

# Generate locales
echo "=== Setting up locales ==="
locale-gen en_US.UTF-8
update-locale LANG=en_US.UTF-8 LC_ALL=en_US.UTF-8

# Create daytona user if it doesn't exist
echo "=== Setting up daytona user ==="
if ! id -u daytona &>/dev/null; then
    useradd -m -s /bin/bash daytona
fi
echo "daytona ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/91-daytona
chmod 0440 /etc/sudoers.d/91-daytona

# Ensure daytona user has a home directory
mkdir -p /home/daytona
chown -R daytona:daytona /home/daytona

# Install pipx and uv globally
echo "=== Installing pipx and uv ==="
python3 -m pip install --break-system-packages pipx
pipx ensurepath
pipx install uv

# Install Python Language Server
echo "=== Installing Python Language Server ==="
python3 -m pip install --break-system-packages python-lsp-server

# Install common pip packages (split into batches to avoid memory issues)
echo "=== Installing Python packages (batch 1: data science) ==="
python3 -m pip install --break-system-packages \
    numpy pandas scipy matplotlib seaborn

echo "=== Installing Python packages (batch 2: ML frameworks) ==="
python3 -m pip install --break-system-packages \
    scikit-learn

echo "=== Installing Python packages (batch 3: PyTorch) ==="
python3 -m pip install --break-system-packages \
    torch --index-url https://download.pytorch.org/whl/cpu

echo "=== Installing Python packages (batch 4: Keras) ==="
python3 -m pip install --break-system-packages \
    keras

echo "=== Installing Python packages (batch 5: web frameworks) ==="
python3 -m pip install --break-system-packages \
    django flask beautifulsoup4 requests sqlalchemy pillow

echo "=== Installing Python packages (batch 6: opencv) ==="
python3 -m pip install --break-system-packages \
    opencv-python-headless

echo "=== Installing Python packages (batch 7: AI/LLM tools) ==="
python3 -m pip install --break-system-packages \
    daytona pydantic-ai langchain openai anthropic

echo "=== Installing Python packages (batch 8: more AI tools) ==="
python3 -m pip install --break-system-packages \
    transformers llama-index instructor huggingface-hub ollama

# Install Node.js via nvm
echo "=== Installing Node.js via nvm ==="
export NVM_DIR="/usr/local/share/nvm"
mkdir -p "$NVM_DIR"

# Download and install nvm
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash

# Source nvm
export NVM_DIR="/usr/local/share/nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

# Install latest Node.js
nvm install node
nvm use node
nvm alias default node

# Install global npm packages
echo "=== Installing npm packages ==="
npm install -g ts-node typescript typescript-language-server

# Set ownership for nvm
chown -R daytona:daytona /usr/local/share/nvm

# Create directory for computer use plugin
echo "=== Setting up directories ==="
mkdir -p /usr/local/lib
chown daytona:daytona /usr/local/lib

# Create .zshrc for daytona user to suppress zsh-newuser-install prompt
touch /home/daytona/.zshrc
chown daytona:daytona /home/daytona/.zshrc

# Create .bashrc with nvm initialization for daytona user
cat >> /home/daytona/.bashrc << 'EOF'
export NVM_DIR="/usr/local/share/nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"
EOF
chown daytona:daytona /home/daytona/.bashrc

# Same for root
cat >> /root/.bashrc << 'EOF'
export NVM_DIR="/usr/local/share/nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"
EOF

# Set environment variables
echo "=== Setting environment variables ==="
cat >> /etc/environment << 'EOF'
LANG=en_US.UTF-8
LC_ALL=en_US.UTF-8
EOF

# Clean up
echo "=== Cleaning up ==="
apt-get clean
rm -rf /var/lib/apt/lists/*

# Summary
echo ""
echo "=== Installation Complete ==="
echo "Installed packages:"
echo "  - System: curl, sudo, python3, ripgrep, chromium, X11 libs, xfce4, VNC"
echo "  - Python: numpy, pandas, scikit-learn, torch, keras, scipy, matplotlib, seaborn"
echo "  - Python: django, flask, requests, beautifulsoup4, sqlalchemy, pillow, opencv"
echo "  - Python: daytona, pydantic-ai, langchain, openai, anthropic, transformers, llama-index"
echo "  - Node.js: latest via nvm"
echo "  - NPM: ts-node, typescript, typescript-language-server"
echo ""
echo "Verify installations:"
echo "  python3 --version"
echo "  node --version"
echo "  npm --version"
echo ""
