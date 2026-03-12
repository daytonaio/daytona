# FinQA Evaluation & Training on Sandboxes (OpenEnv + Daytona)

## Overview

This guide demonstrates how to use [OpenEnv](https://github.com/meta-pytorch/OpenEnv) with Daytona sandboxes to evaluate and train models on [FinQA](https://huggingface.co/datasets/snorkelai/finqa-data), a financial question-answering dataset built from SEC 10-K filings. An agent interacts with company financial tables through SQL queries, then submits a final numerical answer.

Two modes are included: a lightweight single-episode demo (`run.py`) and a full GRPO training loop (`train.py`) that runs parallel rollouts across hundreds of sandboxes.

## Features

- **Snapshot-based sandboxes:** Pre-built snapshots with the FinQA environment and dataset baked in
- **OpenEnv WebSocket protocol:** Connects to FinQA environment servers running inside Daytona sandboxes over WebSocket using the standard OpenEnv client
- **Multi-turn tool-calling episodes:** Agents explore tables, run SQL queries, and submit answers across multiple interaction turns
- **RL reward signal:** Exact-match answer checking provides a binary reward (1.0 = correct, 0.0 = wrong)
- **GRPO + LoRA training:** Trains `Qwen3-14B` with LoRA using Group Relative Policy Optimization
- **Parallel rollouts at scale:** Collects episodes across up to 500 concurrent Daytona sandboxes
- **Batched vLLM generation:** Uses vLLM with tensor parallelism for fast inference during rollout collection

## Requirements

- **Python:** 3.10 or higher
- **GPU (training only):** 4 GPUs recommended — 2 for vLLM (tensor parallel), 2 for training

> [!TIP]
> `run.py` does not require a GPU. Only `train.py` needs GPU resources.

## Environment Variables

- `DAYTONA_API_KEY`: Required for Daytona sandbox access. Get it from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)

## Getting Started

### Setup and Run

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

4. Build the FinQA snapshot (one-time, takes 3-5 minutes):

```bash
python build_snapshot.py
```

5. Run a single episode:

```bash
python run.py
```

### Training

Install the training extras and run:

```bash
pip install -e ".[train]"
python train.py
```

Quick smoke test (no real training, just verifies the pipeline):

```bash
python train.py --sandboxes 2 --iterations 1 --group-size 2
```

### CLI Options

`train.py` accepts the following arguments:

- `--sandboxes` — Number of Daytona sandboxes (default: 500)
- `--iterations` — Number of training iterations (default: 10)
- `--group-size` — Episodes per prompt group for GRPO (default: 6)
- `--target-groups-per-iter` — Stop rollout collection once this many groups are formed (default: 100, set 0 to use all rounds)
- `--max-rollout-rounds` — Maximum rollout rounds per training iteration (default: 8)
- `--model` — HuggingFace model ID (default: `Qwen/Qwen3-14B`)
- `--snapshot` — Daytona snapshot name (default: `openenv-finqa`)
- `--lr` — Learning rate (default: 8e-5)
- `--temperature` — Sampling temperature (default: 1.0)
- `--max-steps` — Max episode steps (default: 20)
- `--max-gen-tokens` — Max tokens per generation (default: 512)
- `--rollout-dispatch-wait-ms` — Max wait to accumulate ready episodes before generation dispatch (default: 500)
- `--tensor-parallel-size` — vLLM tensor parallelism (default: 2)
- `--gpu-memory-utilization` — vLLM GPU memory fraction (default: 0.85)
- `--lora-rank` — LoRA rank (default: 16)
- `--lora-alpha` — LoRA alpha (default: 32)
- `--lora-dropout` — LoRA dropout (default: 0.0)
- `--lora-target-modules` — LoRA target modules (default: all attention + MLP projections)
- `--sync-every` — Export LoRA adapter to vLLM every N iterations (default: 1)
- `--grpo-update-batch-size` — Batch size for GRPO updates (default: 12)
- `--disable-gradient-checkpointing` — Disable gradient checkpointing
- `--run-dir` — Directory for persistent JSONL logs (default: `runs/YYYYMMDD_HHMMSS/`)
- `--save-token-ids` — Include prompt/completion token ID lists in trajectory logs

`build_snapshot.py` accepts:

- `--snapshot-name` — Name for the snapshot (default: `openenv-finqa`)

## How It Works

### Single Episode (run.py)

1. **Create sandbox:** Spin up a Daytona sandbox from the `openenv-finqa` snapshot and wait for the server health check
2. **Reset:** Connect over WebSocket and call `reset()` to start a new episode with a random question and company
3. **Explore:** Use `get_descriptions` and `get_table_info` to discover available tables and their schemas
4. **Query and submit:** Run SQL queries with `sql_query`, then submit a final answer with `submit_answer`
5. **Teardown:** Stop the sandbox and report the reward

### Training Loop (train.py)

1. **Initialize:** Load `Qwen3-14B` with LoRA on training GPUs, start vLLM on inference GPUs, create sandbox pool
2. **Collect rollouts:** Run multi-turn episodes in parallel — vLLM generates tool calls, sandboxes execute them and return observations
3. **Group episodes:** Group completed episodes by prompt (question); compute group-relative advantages
4. **Policy update:** Policy gradient step with group-relative advantages
5. **Sync actor:** Export the updated LoRA adapter and hot-swap it into vLLM for the next iteration

## Available Tools

The FinQA environment exposes four tools via the OpenEnv protocol:

| Tool | Arguments | Description |
|------|-----------|-------------|
| `get_descriptions` | `company_name` | List available table names for a company |
| `get_table_info` | `company_name`, `table_name` | Get table metadata: columns, types, unique values |
| `sql_query` | `query` | Execute a SQL query against the company's 10-K data |
| `submit_answer` | `answer` | Submit a final numerical answer (terminates the episode) |

## Configuration

### Sandbox Settings

- `SANDBOX_COUNT` — Number of sandboxes in the pool (default: 500)
- `MAX_CONCURRENT_CREATE` — Max concurrent sandbox creations (default: 100)
- `MAX_CONCURRENT_PLAY` — Max concurrent active episodes (default: 200)
- `MAX_PLAY_RETRIES` — Retries for failed episodes (default: 3)

### Model Settings

- `MODEL_NAME` — HuggingFace model ID (default: `Qwen/Qwen3-14B`)
- `TENSOR_PARALLEL_SIZE` — vLLM tensor parallelism (default: 2)
- `GPU_MEMORY_UTILIZATION` — vLLM GPU memory fraction (default: 0.85)
- `MAX_GEN_TOKENS` — Max tokens per generation (default: 512)
- `TEMPERATURE` — Sampling temperature (default: 1.0)

### Training Settings

- `LEARNING_RATE` — Adam learning rate for LoRA (default: 8e-5)
- `LORA_RANK` / `LORA_ALPHA` — LoRA hyperparameters (default: 16 / 32)
- `LORA_DROPOUT` — LoRA dropout (default: 0.0)
- `LORA_TARGET_MODULES` — Linear layers targeted by LoRA (default: all attention + MLP projections)
- `EPISODES_PER_GROUP` — Group size for GRPO advantage computation (default: 6)
- `TRAINING_ITERATIONS` — Number of outer training loop iterations (default: 10)
- `TARGET_GROUPS_PER_ITER` — Stop rollout collection once this many groups are formed (default: 100)
- `MAX_ROLLOUT_ROUNDS` — Maximum rollout rounds per training iteration (default: 8)
- `ROLLOUT_DISPATCH_WAIT_MS` — Max wait to accumulate ready episodes before generation dispatch (default: 500)
- `GRPO_UPDATE_BATCH_SIZE` — Batch size for gradient updates (default: 12)
- `SYNC_EVERY` — Export LoRA adapter to vLLM every N iterations (default: 1)

## License

See the main project LICENSE file for details.

## References

- [OpenEnv](https://github.com/meta-pytorch/OpenEnv) — Meta's open-source RL environment framework
- [FinQA Dataset](https://huggingface.co/datasets/snorkelai/finqa-data) — Financial QA from SEC 10-K filings
- [GRPO (DeepSeek-R1)](https://arxiv.org/abs/2501.12948) — Group Relative Policy Optimization
- [Qwen3](https://huggingface.co/Qwen/Qwen3-14B) — Qwen3-14B base model
- [vLLM](https://github.com/vllm-project/vllm) — High-throughput LLM inference engine
