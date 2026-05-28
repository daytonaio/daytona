ARG SANDBOX_IMAGE=daytonaio/sandbox:latest
FROM ${SANDBOX_IMAGE}
USER root

# CUDA 13 toolkit (nvcc) via runfile -- NVIDIA's apt repo is unusable on Debian trixie.
# --toolkit only; the host provides the driver. SHA-256 verified before executing as root.
ARG CUDA_RUN_SHA256=81a5d0d0870ba2022efb0a531dcc60adbdc2bbff7b3ef19d6fd6d8105406c775
RUN curl -fsSL -o /tmp/cuda.run https://developer.download.nvidia.com/compute/cuda/13.0.2/local_installers/cuda_13.0.2_580.95.05_linux.run \
    && echo "${CUDA_RUN_SHA256}  /tmp/cuda.run" | sha256sum -c - \
    && sh /tmp/cuda.run --silent --toolkit --override \
    && rm -f /tmp/cuda.run
ENV CUDA_HOME=/usr/local/cuda
ENV PATH=/usr/local/cuda/bin:${PATH}
# Put CUDA libs (libnvrtc.so.13) on the loader path, else vLLM's DeepGEMM warmup crashes.
RUN printf '/usr/local/cuda/lib64\n/usr/local/cuda/targets/x86_64-linux/lib\n' > /etc/ld.so.conf.d/cuda.conf \
    && ldconfig

RUN python3 -m pip install --no-cache-dir --break-system-packages \
      torch==2.11.0 torchvision==0.26.0 torchaudio==2.11.0 \
      --index-url https://download.pytorch.org/whl/cu130

RUN python3 -m pip install --no-cache-dir --break-system-packages vllm==0.21.0

# Pre-stage FlashInfer kernels, else vLLM JIT-compiles them on first serve (~60s cold start).
RUN FIV=$(python3 -c "import flashinfer; print(flashinfer.__version__)") \
    && python3 -m pip install --no-cache-dir --break-system-packages "flashinfer-jit-cache==${FIV}" \
       --extra-index-url https://flashinfer.ai/whl/cu130

# GPU/ML packages the base image lacks
RUN python3 -m pip install --no-cache-dir --break-system-packages \
      accelerate==1.13.0 \
      datasets==4.8.5 \
      safetensors==0.7.0 \
      peft==0.19.1 \
      bitsandbytes==0.49.2 \
      trl==1.5.0 \
      einops==0.8.2 \
      sentencepiece==0.2.1 \
      sentence-transformers==5.5.1 \
      nvitop==1.7.0 \
      wandb==0.27.0 \
      tensorboard==2.20.0 \
      jupyterlab==4.5.7 \
      ipykernel==7.2.0 \
      ipywidgets==8.1.8 \
      transformers==5.9.0

USER daytona
# Cache FlashInfer cubins as daytona so they land in a runtime-readable home.
RUN flashinfer download-cubin
RUN python3 -c "import torch, vllm, flashinfer"
ENTRYPOINT ["sleep", "infinity"]
