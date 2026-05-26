ARG SANDBOX_IMAGE=daytonaio/sandbox:latest
FROM ${SANDBOX_IMAGE}
USER root

# CUDA 13 toolkit (nvcc) via runfile -- NVIDIA's apt repo is unusable on Debian trixie.
# --toolkit only; the host provides the driver.
RUN curl -fsSL -o /tmp/cuda.run https://developer.download.nvidia.com/compute/cuda/13.0.2/local_installers/cuda_13.0.2_580.95.05_linux.run \
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
      accelerate datasets safetensors peft trl einops sentencepiece sentence-transformers \
      huggingface_hub[hf_xet] nvitop wandb tensorboard jupyterlab ipykernel ipywidgets

RUN python3 -m pip install --no-cache-dir --break-system-packages bitsandbytes

ENV HF_XET_HIGH_PERFORMANCE=1
USER daytona
# Cache FlashInfer cubins as daytona so they land in a runtime-readable home.
RUN flashinfer download-cubin || true
ENTRYPOINT ["sleep", "infinity"]
