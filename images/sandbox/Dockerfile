FROM mcr.microsoft.com/devcontainers/python

# Update package list and install required dependencies
RUN apt-get update && apt-get install -y \
    curl \
    sudo \
    python3-pip \
    python3-venv \
    ripgrep \
    chromium \
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