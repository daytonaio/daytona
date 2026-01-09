FROM python:3.11.14-slim

# Update package list and install required dependencies
RUN apt-get update && apt-get install -y \
    curl \
    sudo \
    bash \
    python3-pip \
    python3-venv \
    ripgrep \
    && rm -rf /var/lib/apt/lists/*

# Change default shell to bash
RUN chsh -s /bin/bash

SHELL ["/bin/bash", "-c"]

# Install pipx and uv
RUN python3 -m pip install --no-cache-dir pipx && pipx ensurepath && pipx install uv

# Install Python LSP, daytona and essential pip packages
RUN python3 -m pip install --no-cache-dir \
    python-lsp-server==1.14.0 \
    daytona==0.22.0 \
    matplotlib==3.10.8 \
    pandas==2.2.3 \
    numpy==2.2.6

# Create the Daytona user and configure sudo access
RUN useradd -m daytona && echo "daytona ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/91-daytona

ENV NVM_DIR=/usr/local/nvm
ENV NODE_VERSION=22.14.0
RUN mkdir -p $NVM_DIR

# Install nvm with node and npm
RUN curl https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash

# install node and npm
RUN source $NVM_DIR/nvm.sh \
    && nvm install $NODE_VERSION \
    && nvm alias default $NODE_VERSION \
    && nvm use default

# add node and npm to path so the commands are available
ENV NODE_PATH=$NVM_DIR/v$NODE_VERSION/lib/node_modules
ENV PATH=$NVM_DIR/versions/node/v$NODE_VERSION/bin:$PATH

RUN npm install -g ts-node@10.9.2 typescript@5.9.3 typescript-language-server@5.1.3

# Switch to Daytona user
USER daytona

# Create directory for computer use plugin
RUN mkdir -p /usr/local/lib && sudo chown daytona:daytona /usr/local/lib

# Keep the container running indefinitely
ENTRYPOINT [ "sleep", "infinity" ]
