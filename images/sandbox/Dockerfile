FROM mcr.microsoft.com/devcontainers/python:1-3.14

# Update package list and install required dependencies
RUN apt-get update && apt-get install -y \
    curl \
    sudo \
    python3-pip \
    python3-venv \
    ripgrep \
    chromium \
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
    && rm -rf /var/lib/apt/lists/*

# Install pipx and uv
RUN python3 -m pip install --no-cache-dir pipx && pipx ensurepath && pipx install uv

# Install the Python Language Server
RUN python3 -m pip install --no-cache-dir python-lsp-server==1.14.0

# Install common pip packages
RUN python3 -m pip install --no-cache-dir \
    numpy==2.2.6 \
    pandas==2.2.3 \
    scikit-learn==1.8.0 \
    keras==3.13.0 \
    torch==2.9.1 \
    scipy==1.16.3 \
    seaborn==0.13.2 \
    matplotlib==3.10.8 \
    django==6.0 \
    flask==3.1.2 \
    beautifulsoup4==4.14.3 \
    requests==2.32.5 \
    opencv-python==4.12.0.88 \
    pillow==12.0.0 \
    sqlalchemy==2.0.45 \
    daytona==0.22.0 \
    pydantic-ai==0.2.20 \
    langchain==1.2.0 \
    transformers==4.57.3 \
    openai==2.14.0 \
    anthropic==0.75.0 \
    llama-index==0.14.10 \
    instructor==1.13.0 \
    huggingface-hub==0.36.0 \
    ollama==0.6.1 \
    claude-agent-sdk==0.1.18

# Create the Daytona user and configure sudo access
RUN useradd -m daytona && echo "daytona ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/91-daytona

# Install latest Node.js using nvm
RUN bash -c "source /usr/local/share/nvm/nvm.sh && nvm install node && nvm use node" \
  && chown -R daytona:daytona /usr/local/share/nvm

RUN npm install -g ts-node@10.9.2 typescript@5.9.3 typescript-language-server@5.1.3 bun@1.3.5

# Create directory for computer use plugin
RUN mkdir -p /usr/local/lib && chown daytona:daytona /usr/local/lib

ENV LANG=en_US.UTF-8 \
    LC_ALL=en_US.UTF-8

# Switch to Daytona user
USER daytona

# Create .zshrc to suppress zsh-newuser-install prompt
RUN touch ~/.zshrc

# Keep the container running indefinitely
ENTRYPOINT [ "sleep", "infinity" ]
