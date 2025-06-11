FROM mcr.microsoft.com/devcontainers/python

# Update package list and install required dependencies
RUN apt-get update && apt-get install -y curl sudo python3-pip python3-venv && rm -rf /var/lib/apt/lists/*

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

# Switch to Daytona user
USER daytona

# Keep the container running indefinitely
ENTRYPOINT [ "sleep", "infinity" ]