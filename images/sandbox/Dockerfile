FROM mcr.microsoft.com/devcontainers/python

# Update package list and install required dependencies
RUN apt-get update && apt-get install -y \
    curl sudo python3-pip python3-venv ripgrep && \
    rm -rf /var/lib/apt/lists/*

# Install pipx and uv
RUN python3 -m pip install pipx && pipx ensurepath && pipx install uv

# Install the Python Language Server
RUN python3 -m pip install python-lsp-server

# Install common pip packages
RUN python3 -m pip install \
    numpy pandas scikit-learn keras torch scipy seaborn matplotlib \
    django flask beautifulsoup4 requests opencv-python pillow sqlalchemy \
    daytona pydantic-ai langchain transformers openai anthropic llama-index instructor huggingface ollama

# Create the Daytona user and configure sudo access
RUN useradd -m daytona && echo "daytona ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/91-daytona

# Install latest Node.js using nvm
RUN bash -c "source /usr/local/share/nvm/nvm.sh && nvm install node && nvm use node" \
  && chown -R daytona:daytona /usr/local/share/nvm

RUN npm install -g ts-node typescript typescript-language-server

# Set locale
ENV LANG=en_US.UTF-8 \
    LC_ALL=en_US.UTF-8
# Vnc Setup

# Prevent interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive
ENV DISPLAY=:1
ENV VNC_PORT=5901
ENV NO_VNC_PORT=6901
ENV VNC_RESOLUTION=1280x720

# Install base packages
RUN apt-get update && apt-get install -y \
    wget \
    git \
    vim \
    xfce4 \
    xfce4-terminal \
    dbus-x11 \
    xfonts-base \
    xfonts-100dpi \
    xfonts-75dpi \
    xfonts-scalable \
    x11vnc \
    novnc \
    supervisor \
    net-tools \
    locales \
    xvfb \
    x11-utils \
    x11-xserver-utils \
    gnome-screenshot \
    scrot \
    imagemagick \
    xdotool \
    xautomation \
    wmctrl \
    build-essential \
    libx11-dev \
    libxext-dev \
    libxtst-dev \
    libxinerama-dev \
    libx11-xcb-dev \
    libxkbcommon-dev \
    libxkbcommon-x11-dev \
    libxcb-xkb-dev \
    libpng-dev \
    chromium \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Setup VNC
RUN mkdir -p /home/daytona/.vnc && \
    chown -R daytona:daytona /home/daytona/.vnc

# NoVNC setup
RUN ln -sf /usr/share/novnc/vnc.html /usr/share/novnc/index.html && \
    sed -i 's/websockify =/websockify = --heartbeat 30/' /usr/share/novnc/utils/launch.sh

# Switch to Daytona user
USER daytona

# Create .zshrc to suppress zsh-newuser-install prompt
RUN touch ~/.zshrc

# Keep the container running indefinitely
ENTRYPOINT [ "sleep", "infinity" ]