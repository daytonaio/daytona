# Serving vLLM on GPU Sandboxes (vLLM + Daytona)

## Overview

This guide demonstrates how to serve an open-weights model with [vLLM](https://github.com/vllm-project/vllm) on a Daytona GPU sandbox and query it from anywhere through a token-authenticated preview URL. The server speaks the OpenAI-compatible API, so any OpenAI client works against it unchanged.

`serve_vllm.py` creates the sandbox, starts `vllm serve`, streams the startup logs, and prints the endpoint once the server is healthy. Three query examples are included: raw `curl` (`query.sh`), the OpenAI SDK with chat, streaming, and tool calling (`query_openai.py`), and LiteLLM (`query_litellm.py`).

## Features

- **GPU sandbox from the stock vLLM image:** No custom image build, the official `vllm/vllm-openai` image runs as-is
- **GPU type preference:** `gpu_type` requests an H100 first, falling back to an RTX PRO 6000
- **OpenAI-compatible endpoint:** Works with `curl`, the OpenAI SDK, LiteLLM, or anything else that speaks the OpenAI API
- **Token-authenticated preview URL:** The endpoint is reachable from anywhere; requests authenticate with the `x-daytona-preview-token` header
- **Live boot logs and fail-fast startup:** Server logs stream to your terminal while the model loads; if `vllm serve` dies, the script exits immediately with the full log saved locally
- **Tool calling and reasoning enabled:** The server is started with vLLM's tool-call and reasoning parsers for the served model family

## Requirements

- **Python:** 3.10 or higher
- **Daytona:** GPU sandboxes are currently experimental, make sure your organization has access to GPUs

> [!TIP]
> No local GPU is needed; the model runs entirely inside the sandbox.

## Environment Variables

- `DAYTONA_API_KEY`: Required for Daytona sandbox access. Get it from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)
- `HF_TOKEN`: Optional, required only for gated Hugging Face models; Hugging Face recommends a token for faster, less throttled downloads in general

## Getting Started

1. Create and activate a virtual environment:

```bash
python3.10 -m venv venv
source venv/bin/activate
```

2. Install dependencies:

```bash
pip install -e .
```

3. Set your Daytona API key:

```bash
cp .env.example .env
# edit .env with your API key
```

4. Start the server (model download and loading can take several minutes):

```bash
python serve_vllm.py
```

5. When the server is healthy, the script prints paste-ready exports:

```bash
export ENDPOINT=https://8000-{sandboxId}.{daytonaProxyDomain}
export TOKEN={previewToken}
```

6. Paste them into your shell, then query the endpoint:

```bash
./query.sh                # raw curl
python query_openai.py    # OpenAI SDK: chat, streaming, tool calling
python query_litellm.py   # LiteLLM
```

The endpoint authenticates via the token header. Two alternatives: `sb.create_signed_preview_url(PORT, expires_in_seconds=3600)` returns a URL with the token embedded (for clients that can't set headers), and `public=True` at sandbox creation drops proxy auth entirely. Independently, vLLM's `--api-key` flag adds the server's own key check; combined with a public preview, the endpoint takes the standard OpenAI shape of base URL plus `api_key`.

### Querying from inside the sandbox

Code running inside the sandbox can skip the preview URL and token and talk to `http://localhost:8000` directly. The vLLM image ships the `openai` package, so the SDK works there as-is:

```python
from daytona import Daytona, DaytonaConfig

sb = Daytona(DaytonaConfig(target="us-east-1")).get("SANDBOX_ID")
print(sb.process.code_run("""
from openai import OpenAI

client = OpenAI(base_url="http://localhost:8000/v1", api_key="EMPTY")
resp = client.chat.completions.create(
    model="gemma-4-moe",
    messages=[{"role": "user", "content": "Write a haiku about code that never leaves its sandbox."}],
    max_tokens=64,
)
print(resp.choices[0].message.content)
""").result)
```

Useful for colocated workloads, like batch inference over data uploaded into the sandbox.

### Cleanup

The sandbox stays up after `serve_vllm.py` exits, so the endpoint keeps working on success and the downloaded weights aren't lost on failure. Delete it when you're done:

```bash
python -c "from daytona import Daytona; Daytona().get('SANDBOX_ID').delete()"
```

The sandbox ID is printed by `serve_vllm.py`.

## Configuration

Constants at the top of `serve_vllm.py`:

- `MODEL`: Hugging Face model ID to serve (default: `google/gemma-4-26B-A4B-it`)
- `SERVED_AS`: model name exposed by the API, what clients pass as `model` (default: `gemma-4-moe`)
- `VLLM_IMAGE`: vLLM Docker image (default: `vllm/vllm-openai:v0.22.1`)
- `PORT`: port the server listens on (default: `8000`)
- `TARGET`: Daytona region; `us-east-1` is currently the region for GPU sandboxes
- `BOOT_TIMEOUT`: seconds to wait for the server to become healthy (default: `900`)

When changing `MODEL`, also update the `--tool-call-parser` and `--reasoning-parser` flags: parser names must match the model family and your vLLM version, or `vllm serve` won't start.

GPU sandboxes are currently capped at 1 GPU each.

## How It Works

1. **Create sandbox:** Spin up an ephemeral GPU sandbox in `us-east-1` from the official vLLM image
2. **Start the server:** Run `vllm serve` as a background session command; the model downloads from Hugging Face and loads onto the GPU
3. **Wait for health:** Poll `/health` through the preview URL while streaming server logs; if the server process exits, save the log locally and fail fast
4. **Hand off:** Print `export ENDPOINT=... TOKEN=...` lines for the query scripts
5. **Query:** Clients hit the OpenAI-compatible API through the preview URL, authenticating with the `x-daytona-preview-token` header

## License

See the main project LICENSE file for details.

## References

- [vLLM](https://docs.vllm.ai/en/stable/): High-throughput LLM inference engine
- [vLLM OpenAI-compatible server](https://docs.vllm.ai/en/stable/serving/openai_compatible_server.html)
- [Daytona](https://daytona.io)
- [LiteLLM](https://docs.litellm.ai/)
