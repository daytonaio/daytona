# Serving gpt-oss on GPU Sandboxes (SGLang + Daytona)

## Overview

This guide demonstrates how to serve [gpt-oss-20b](https://huggingface.co/openai/gpt-oss-20b), OpenAI's open-weights reasoning model, with [SGLang](https://github.com/sgl-project/sglang) on a Daytona GPU sandbox and query it from anywhere through a token-authenticated preview URL. The server speaks the OpenAI-compatible API, so any OpenAI client works against it unchanged.

`serve_sglang.py` creates the sandbox, starts `sglang.launch_server`, streams the startup logs, and prints the endpoint once the server is healthy. Four query examples are included: raw `curl` (`query.sh`), the OpenAI SDK with chat, streaming, structured output, reasoning, tool calling, and a prefix-cache demo (`query_openai.py`), LiteLLM (`query_litellm.py`), and a concurrent classification workload over classic-book passages (`classify_passages.py`).

## Features

- **GPU sandbox from the stock SGLang image:** No custom image build, the official `lmsysorg/sglang` image runs as-is
- **GPU type preference:** `gpu_type` requests an H100 first, falling back to an RTX PRO 6000
- **OpenAI-compatible endpoint:** Works with `curl`, the OpenAI SDK, LiteLLM, or anything else that speaks the OpenAI API
- **Token-authenticated preview URL:** The endpoint is reachable from anywhere; requests authenticate with the `x-daytona-preview-token` header
- **Live boot logs and fail-fast startup:** Server logs stream to your terminal while the model loads; if the server dies, the script exits immediately with the full log saved locally
- **Reasoning with effort control:** gpt-oss thinks before it answers; `reasoning_effort` adjusts it per request and the parsed trace comes back in `reasoning_content`
- **Structured output:** `response_format` with a JSON schema constrains decoding, so replies are guaranteed to parse
- **Prefix caching:** RadixAttention is on by default; the `--enable-cache-report` flag exposes per-request cache hits in the usage stats
- **Batched workload:** `classify_passages.py` classifies 273 passages from thirteen classic books by author in one concurrent batch (~825k tokens), scores the result against ground truth, then asks a second question (indoors vs outdoors) over the same passages that reuses the prefix cache

## Requirements

- **Python:** 3.10 or higher
- **Daytona:** Make sure your organization has access to GPU sandboxes

> [!TIP]
> No local GPU is needed; the model runs entirely inside the sandbox.

## Environment Variables

- `DAYTONA_API_KEY`: Required for Daytona sandbox access. Get it from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)
- `HF_TOKEN`: Optional; gpt-oss is not gated, so this only matters for gated models you swap in (Hugging Face recommends a token for faster, less throttled downloads in general)

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

4. Start the server (image pull and model download take a few minutes):

```bash
python serve_sglang.py
```

5. When the server is healthy, the script prints paste-ready exports:

```bash
export ENDPOINT=https://8000-{sandboxId}.{daytonaProxyDomain}
export TOKEN={previewToken}
```

6. Paste them into your shell, then query the endpoint:

```bash
./query.sh                   # raw curl
python query_openai.py       # OpenAI SDK: chat, streaming, structured output, reasoning, tools, prefix cache
python query_litellm.py      # LiteLLM
python classify_passages.py  # concurrent author + setting classification over Gutenberg passages
```

The endpoint authenticates via the token header. Two alternatives: `sb.create_signed_preview_url(PORT, expires_in_seconds=3600)` returns a URL with the token embedded (for clients that can't set headers), and `public=True` at sandbox creation drops proxy auth entirely. Independently, SGLang's `--api-key` flag adds the server's own key check; combined with a public preview, the endpoint takes the standard OpenAI shape of base URL plus `api_key`.

### Budget for thinking

gpt-oss reasons before it answers, and `max_tokens` covers reasoning plus answer combined. If thinking exhausts the budget, the response has `finish_reason: "length"` and `content: null`, which looks like the model returned nothing. Thinking length varies a lot between identical runs, so the examples use generous budgets and turn `reasoning_effort` down to `"low"` for simple or high-volume tasks.

### Querying from inside the sandbox

Code running inside the sandbox can skip the preview URL and token and talk to `http://localhost:8000` directly. The SGLang image ships the `openai` package, so the SDK works there as-is:

```python
from daytona import Daytona, DaytonaConfig

sb = Daytona(DaytonaConfig(target="us-east-1")).get("SANDBOX_ID")
print(sb.process.code_run("""
from openai import OpenAI

client = OpenAI(base_url="http://localhost:8000/v1", api_key="EMPTY")
resp = client.chat.completions.create(
    model="gpt-oss-20b",
    messages=[{"role": "user", "content": "Write a haiku about code that never leaves its sandbox."}],
    max_tokens=4096,
)
print(resp.choices[0].message.content)
""").result)
```

Useful for colocated workloads, like batch inference over data uploaded into the sandbox.

### Cleanup

The sandbox stays up after `serve_sglang.py` exits, so the endpoint keeps working on success and the downloaded weights aren't lost on failure. Delete it when you're done:

```bash
python -c "from daytona import Daytona; Daytona().get('SANDBOX_ID').delete()"
```

The sandbox ID is printed by `serve_sglang.py`.

## Configuration

Constants at the top of `serve_sglang.py`:

- `MODEL`: Hugging Face model ID to serve (default: `openai/gpt-oss-20b`)
- `SERVED_AS`: model name exposed by the API, what clients pass as `model` (default: `gpt-oss-20b`)
- `SGLANG_IMAGE`: SGLang Docker image (default: `lmsysorg/sglang:v0.5.12.post1-cu130`)
- `PORT`: port the server listens on (default: `8000`)
- `TARGET`: Daytona region; `us-east-1` is currently the region for GPU sandboxes
- `BOOT_TIMEOUT`: seconds to wait for the server to become healthy (default: `900`)

When changing `MODEL`, also update the `--tool-call-parser` and `--reasoning-parser` flags: parser names must match the model family and your SGLang version, or tool calls and reasoning come back unparsed in `content`. Both flags also accept `auto` to detect the parser from the model's chat template.

GPU sandboxes are currently capped at 1 GPU each. The larger gpt-oss-120b also fits on a single H100 with extra memory flags; see the guide's "Scaling up" section for the flags and the capacity trade-off.

## How It Works

1. **Create sandbox:** Spin up an ephemeral GPU sandbox in `us-east-1` from the official SGLang image
2. **Start the server:** Run `sglang.launch_server` as a background session command; the model downloads from Hugging Face and loads onto the GPU
3. **Wait for health:** Poll `/health_generate` (a real forward pass, not just a liveness check) through the preview URL while streaming server logs; if the server process exits, save the log locally and fail fast
4. **Hand off:** Print `export ENDPOINT=... TOKEN=...` lines for the query scripts
5. **Query:** Clients hit the OpenAI-compatible API through the preview URL, authenticating with the `x-daytona-preview-token` header

## License

See the main project LICENSE file for details.

## References

- [SGLang](https://docs.sglang.ai/): Fast serving framework for LLMs and VLMs
- [SGLang OpenAI-compatible API](https://docs.sglang.ai/basic_usage/openai_api.html)
- [SGLang structured outputs](https://docs.sglang.ai/advanced_features/structured_outputs.html)
- [gpt-oss-20b](https://huggingface.co/openai/gpt-oss-20b)
- [Daytona](https://daytona.io)
- [LiteLLM](https://docs.litellm.ai/)
