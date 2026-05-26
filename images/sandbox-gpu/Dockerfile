FROM vllm/vllm-openai:v0.21.0@sha256:a230095847e93bd4df9888b33dab956fa9504537b828a23657d2b26fed57b5c9

ENV CUDA_HOME=/usr/local/cuda \
    HF_HOME=/home/daytona/.cache/huggingface \
    HF_XET_HIGH_PERFORMANCE=1 \
    DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
      sudo \
      build-essential cmake ninja-build pkg-config \
      git git-lfs curl wget aria2 rsync ca-certificates openssh-client \
      nvtop htop tmux vim less jq unzip zip \
    && git lfs install --system \
    && rm -rf /var/lib/apt/lists/*

RUN python3 -m pip install --no-cache-dir uv==0.11.16

RUN uv pip install --system --no-cache \
      transformers==5.8.1 \
      accelerate==1.13.0 \
      datasets==4.8.5 \
      safetensors==0.7.0 \
      peft==0.19.1 \
      bitsandbytes==0.49.2 \
      trl==1.5.0 \
      einops==0.8.2 \
      sentencepiece==0.2.1 \
      sentence-transformers==5.5.1 \
      huggingface_hub[hf_xet]==1.14.0 \
      nvitop==1.7.0 \
      wandb==0.27.0 \
      tensorboard==2.20.0 \
      jupyterlab==4.5.7 \
      ipython==9.13.0 \
      ipykernel==7.2.0 \
      ipywidgets==8.1.8 \
      numpy==2.2.6 \
      pandas==3.0.3 \
      scipy==1.17.1 \
      scikit-learn==1.8.0 \
      matplotlib==3.10.9 \
      seaborn==0.13.2 \
      pillow==12.2.0 \
      openai==2.36.0 \
      anthropic==0.102.0 \
      ollama==0.6.2 \
      langchain==1.3.1 \
      llama-index==0.14.22 \
      instructor==1.15.1 \
      mistralai==1.12.4 \
      pydantic-ai==1.94.0 \
      openai-agents==0.17.4 \
      claude-agent-sdk==0.2.87 \
      django==6.0.5 \
      flask==3.1.3 \
      sqlalchemy==2.0.50 \
      requests==2.34.2 \
      beautifulsoup4==4.14.3 \
      daytona==0.181.0

RUN useradd -m -s /bin/bash daytona \
    && echo "daytona ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/91-daytona \
    && chmod 0440 /etc/sudoers.d/91-daytona

ENV HOME=/home/daytona
WORKDIR /home/daytona
USER daytona
ENTRYPOINT ["sleep", "infinity"]
