# Daytona Sandbox GPU Image

[Dockerfile](./Dockerfile) defines [daytonaio/sandbox-gpu](https://hub.docker.com/r/daytonaio/sandbox-gpu), a GPU-enabled sandbox image for x86 GPU hosts (e.g. H100).

It is a **superset of the default [`daytonaio/sandbox`](../sandbox) image** — it builds `FROM daytonaio/sandbox` and layers the GPU stack on top, so everything in the standard sandbox (Python, Node, language servers, the computer-use/VNC tooling, the default Python/agent packages, the `daytona` user) is present, plus GPU support.

## NOTE

This image is **amd64-only** — there is no arm64 variant.

## Added on top of the base image

CUDA / GPU runtime:

- CUDA 13 toolkit (`nvcc`), installed via the NVIDIA runfile
- PyTorch (CUDA 13 build) — `torch`, `torchvision`, `torchaudio`
- vLLM, with FlashInfer kernels pre-staged so first-serve cold-start stays fast
- FlashAttention (bundled with vLLM), cuDNN

GPU / ML / fine-tuning:

- accelerate
- datasets
- safetensors
- peft
- bitsandbytes
- trl
- einops
- sentencepiece
- sentence-transformers
- nvitop

Experiment tracking:

- wandb
- tensorboard

Interactive:

- jupyterlab
- ipykernel
- ipywidgets
