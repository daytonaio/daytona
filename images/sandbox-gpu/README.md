# Daytona Sandbox GPU Image

[Dockerfile](./Dockerfile) contains the definition for [daytonaio/sandbox-gpu](https://hub.docker.com/r/daytonaio/sandbox-gpu) images (`:$VERSION`), intended for GPU-enabled sandboxes.

It is built on the official [`vllm/vllm-openai`](https://hub.docker.com/r/vllm/vllm-openai) image, which already ships a working CUDA toolchain and inference stack:

- CUDA 13.0 toolkit (`nvcc`)
- PyTorch (CUDA build)
- vLLM
- FlashAttention, FlashInfer, cuDNN

On top of that base it adds general GPU/ML development tooling.

## NOTE

This image is **amd64-only** — it targets x86 GPU hosts (e.g. H100). There is no arm64 variant.

## System tooling

- build-essential, cmake, ninja-build, pkg-config (for compiling CUDA extensions)
- git, git-lfs (model/dataset repos)
- curl, wget, aria2, rsync
- nvtop, htop (GPU/system monitoring)
- tmux, vim, less, jq, unzip, zip
- uv (Python package/environment manager)

## Python packages

GPU / ML / fine-tuning:

- transformers
- accelerate
- datasets
- safetensors
- peft
- bitsandbytes
- trl
- einops
- sentencepiece
- sentence-transformers
- huggingface-hub (with hf_xet)
- nvitop

Experiment tracking:

- wandb
- tensorboard

Interactive:

- jupyterlab
- ipython
- ipykernel
- ipywidgets

Data science / viz:

- numpy
- pandas
- scipy
- scikit-learn
- matplotlib
- seaborn
- pillow

LLM clients:

- openai
- anthropic
- ollama

Agent / app frameworks:

- langchain
- llama-index
- instructor
- pydantic-ai
- openai-agents
- claude-agent-sdk

Web / DB / misc:

- django
- flask
- sqlalchemy
- requests
- beautifulsoup4

Daytona:

- daytona
