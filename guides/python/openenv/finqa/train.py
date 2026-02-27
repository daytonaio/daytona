"""GRPO training for FinQA across 1000 Daytona sandboxes.

Trains Qwen3-14B with LoRA to answer financial questions using tool calls,
with rollouts collected in parallel across Daytona sandboxes running the
FinQA OpenEnv environment.

Uses Group Relative Policy Optimization (GRPO):
1. Collect multi-turn tool-calling episodes via batched vLLM generation (TP=2)
2. Compute group-relative advantages (episodes from same prompt)
3. Policy gradient update (group-relative advantages)
4. Export LoRA adapter, swap actor adapter in vLLM for next iteration

Requires 4 GPUs: vLLM on cuda:0-1 (tensor_parallel=2),
training (base + LoRA) on cuda:2-3 (device_map="auto").

Usage:
    # Build snapshot first (once)
    python build_snapshot.py

    # Train (default: 500 sandboxes, 10 iterations)
    python train.py

    # Quick smoke test
    python train.py --sandboxes 2 --iterations 1 --group-size 2
"""

import argparse
import asyncio
import json
import logging
import os
import re
import tempfile
import time
from collections import Counter, defaultdict, deque
from dataclasses import dataclass, field
from datetime import datetime
from typing import Any

import numpy as np
import torch
import torch.nn.functional as F
from dotenv import load_dotenv
from finqa_env import CallToolAction, FinQAEnv  # pylint: disable=import-error
from openenv.core.containers.runtime.daytona_provider import DaytonaProvider  # pylint: disable=import-error
from peft import LoraConfig, get_peft_model
from transformers import AutoModelForCausalLM, AutoTokenizer
from vllm import LLM, SamplingParams

# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------
SANDBOX_COUNT = 500
EPISODES_PER_GROUP = 6
TRAINING_ITERATIONS = 10
TARGET_GROUPS_PER_ITER = 100  # 0 => use all rollout rounds (bounded by MAX_ROLLOUT_ROUNDS)
MAX_ROLLOUT_ROUNDS = 8
MAX_CONCURRENT_CREATE = 100
MAX_CONCURRENT_PLAY = 200
MAX_PLAY_RETRIES = 3
MAX_EPISODE_STEPS = 20
MAX_GEN_TOKENS = 512
STOP_STRINGS = ["</tool_call>"]
TEMPERATURE = 1.0
ROLLOUT_DISPATCH_WAIT_MS = 500
ROLLOUT_RECONNECT_EVERY_BATCHES = 1
MODEL_NAME = "Qwen/Qwen3-14B"
GPU_MEMORY_UTILIZATION = 0.85
TENSOR_PARALLEL_SIZE = 2  # vLLM tensor parallelism across cuda:0-1
SYNC_EVERY = 1  # Export/switch actor LoRA adapter every N training iterations
GRPO_UPDATE_BATCH_SIZE = 12

# LoRA configuration
LORA_RANK = 16
LORA_ALPHA = 32
LEARNING_RATE = 3e-4
LORA_DROPOUT = 0.0
LORA_TARGET_MODULES = [
    "q_proj",
    "k_proj",
    "v_proj",
    "o_proj",
    "gate_proj",
    "up_proj",
    "down_proj",
]

SERVER_CMD = "cd /app/env && uvicorn finqa_env.server.app:app --host 0.0.0.0 --port 8000"

SYSTEM_PROMPT = """\
You are a financial analyst assistant answering questions about SEC 10-K filings.

Think and reason step by step. Iteratively gather data using the available tools until you have enough information to answer the question.

When submitting your final answer:
- Provide ONLY the numerical value. No explanations, units, or LaTeX formatting.
- Always express percentages, growth rates, and percentage point differences as decimal ratios by dividing by 100 (e.g., 22% → 0.22, -8.9% → -0.089, a 4.5 percentage point difference → 0.045).
- Submit numbers exactly as they appear in the query results. Do not convert units (e.g., if the table shows values in millions, submit the number as-is, not multiplied out).
- For multi-year answers, use: year: value, year: value (e.g., 2022: 0.933, 2023: 0.930, 2024: 0.931)
- For year-over-year changes, use: year to year: value (e.g., 2022 to 2023: 0.189, 2023 to 2024: 0.025)
- For single values, just submit the number (e.g., 0.895 or -77 or 63)
- If the question is yes/no, answer Yes or No"""

# Tool schemas and names — populated at runtime via fetch_tools_from_env().
FINQA_TOOLS: list[dict] = []
TOOL_NAMES: set[str] = set()


async def fetch_tools_from_env(env: FinQAEnv) -> list[dict]:
    """Retrieve tool schemas from a connected env via MCP JSON-RPC over WebSocket.

    Returns tools in OpenAI function-calling format.
    """
    resp = await env._send_and_receive(
        {
            "type": "mcp",
            "data": {"jsonrpc": "2.0", "method": "tools/list", "params": {}, "id": 1},
        }
    )
    mcp_tools = resp["data"]["result"]["tools"]
    openai_tools = []
    for t in mcp_tools:
        schema = t.get("inputSchema") or t.get("input_schema") or {}
        properties = {}
        required = []
        if "properties" in schema:
            for name, prop in schema["properties"].items():
                properties[name] = {
                    "type": prop.get("type", "string"),
                    "description": prop.get("description", ""),
                }
            required = schema.get("required", [])
        openai_tools.append(
            {
                "type": "function",
                "function": {
                    "name": t["name"],
                    "description": t.get("description", ""),
                    "parameters": {
                        "type": "object",
                        "properties": properties,
                        "required": required,
                    },
                },
            }
        )
    return openai_tools


DEBUG_LOG_FILE = "debug.log"
DEBUG_LOG_SAMPLE_LIMIT = 5
DEBUG_SANDBOX_LOG_LIMIT = 6

debug_logger = logging.getLogger("finqa_train_debug")
if not debug_logger.handlers:
    debug_handler = logging.FileHandler(DEBUG_LOG_FILE, mode="a")
    debug_handler.setFormatter(logging.Formatter("%(message)s"))
    debug_logger.addHandler(debug_handler)
debug_logger.setLevel(logging.INFO)
debug_logger.propagate = False


# ---------------------------------------------------------------------------
# Persistent JSONL logging helpers
# ---------------------------------------------------------------------------
def _append_jsonl(filepath: str, obj: dict) -> None:
    """Append one JSON object as a line to a JSONL file (flush-safe)."""
    with open(filepath, "a", encoding="utf-8") as f:
        f.write(json.dumps(obj, default=str) + "\n")
        f.flush()


def _episode_to_dict(ep: "Episode", save_token_ids: bool = False) -> dict:
    """Serialize an Episode to a JSON-friendly dict."""
    turns = []
    for t in ep.turns:
        td: dict[str, Any] = {
            "tool_name": t.tool_name,
            "tool_args": t.tool_args,
            "tool_result": t.tool_result,
        }
        if save_token_ids:
            td["prompt_token_ids"] = t.prompt_token_ids
            td["completion_token_ids"] = t.completion_token_ids
        turns.append(td)
    return {
        "question": ep.question,
        "company": ep.company,
        "question_id": ep.question_id,
        "reward": ep.reward,
        "sandbox_idx": ep.sandbox_idx,
        "num_turns": len(ep.turns),
        "turns": turns,
    }


def save_episodes(filepath: str, episodes: list["Episode"], save_token_ids: bool = False) -> None:
    """Append a batch of episodes to a JSONL file."""
    for ep in episodes:
        _append_jsonl(filepath, _episode_to_dict(ep, save_token_ids))


def summarize_exception_samples(
    failures: list[tuple[int | str, Exception]],
    max_samples: int = 5,
) -> str:
    """Compact summary of best-effort cleanup failures for warning prints."""
    if not failures:
        return ""
    samples = []
    for item_id, exc in failures[:max_samples]:
        msg = str(exc).strip().replace("\n", " ")
        if len(msg) > 80:
            msg = msg[:77] + "..."
        samples.append(f"{item_id}={type(exc).__name__}({msg})")
    suffix = "" if len(failures) <= max_samples else f", +{len(failures) - max_samples} more"
    return ", ".join(samples) + suffix


# ---------------------------------------------------------------------------
# Data structures
# ---------------------------------------------------------------------------
@dataclass
class Turn:
    prompt_token_ids: list[int]
    completion_token_ids: list[int]
    tool_name: str
    tool_args: dict
    tool_result: str


@dataclass
class Episode:
    sandbox_idx: int
    turns: list[Turn]
    reward: float  # binary 0.0 or 1.0
    question: str
    company: str
    question_id: str = ""


@dataclass
class ActiveEpisode:
    """Tracks in-progress episode state during rollout."""

    env: FinQAEnv
    sandbox_idx: int
    chat_history: list[dict]
    turns: list[Turn] = field(default_factory=list)
    question: str = ""
    company: str = ""
    question_id: str = ""
    done: bool = False
    reward: float = 0.0


@dataclass
class PreparedTrainBatch:
    """Rollout/ref-logprob data prepared for one GRPO update."""

    eps_flat: list[Episode]
    advs_flat: list[float]
    groups_count: int
    rolled_eps: int
    rounds_used: int
    carry_in_count: int
    leftover_eps: list[Episode]


# ---------------------------------------------------------------------------
# Sandbox pool management (reused from snake/train.py)
# ---------------------------------------------------------------------------
async def create_sandbox_pool(n: int, snapshot_name: str, semaphore: asyncio.Semaphore):
    """Create n DaytonaProvider instances from a snapshot, bounded by semaphore."""
    pool_by_idx: list[tuple | None] = [None] * n
    created = 0

    async def create_one(idx: int):
        nonlocal created
        async with semaphore:
            provider = DaytonaProvider(auto_stop_interval=0, cmd=SERVER_CMD)
            try:
                url = await asyncio.to_thread(
                    provider.start_container,
                    f"snapshot:{snapshot_name}",
                )
                # Mark sandbox for auto-deletion on stop (provider doesn't expose this yet).
                await asyncio.to_thread(provider._sandbox.set_auto_delete_interval, 0)
                for attempt in range(3):
                    try:
                        await asyncio.to_thread(provider.wait_for_ready, url, 120)
                        break
                    except Exception:
                        if attempt == 2:
                            raise
                        await asyncio.sleep(3)
                # Keep stable index ordering so env <-> pool mapping is deterministic.
                pool_by_idx[idx] = (provider, url)
                created += 1
                if created % 50 == 0 or created == n:
                    print(f"  Sandboxes ready: {created}/{n}")
            except Exception:
                # Avoid leaking partially created containers when readiness fails.
                try:
                    await asyncio.to_thread(provider.stop_container)
                except Exception:
                    pass
                raise

    # Stagger launches to stay under API rate limits (~600 creates/min).
    tasks = []
    for i in range(n):
        tasks.append(asyncio.create_task(create_one(i)))
        if (i + 1) % 10 == 0:
            await asyncio.sleep(1.0)
    results = await asyncio.gather(*tasks, return_exceptions=True)
    errors = [r for r in results if isinstance(r, Exception)]
    if errors:
        print(f"  Warning: {len(errors)}/{n} sandboxes failed to create")
        # Log unique error messages to help diagnose failures.
        err_counts = Counter(f"{type(e).__name__}: {e}" for e in errors)
        for msg, cnt in err_counts.most_common(5):
            print(f"    [{cnt}x] {msg[:200]}")
        if not any(pool_by_idx):
            raise errors[0]
    return [entry for entry in pool_by_idx if entry is not None]


async def destroy_sandbox_pool(pool):
    """Tear down all sandboxes concurrently."""

    async def stop_one(provider):
        try:
            await asyncio.to_thread(provider.stop_container)
        except Exception as e:
            print(f"  Warning during cleanup: {e}")

    await asyncio.gather(*[stop_one(p) for p, _ in pool])


# ---------------------------------------------------------------------------
# Persistent connections
# ---------------------------------------------------------------------------
async def connect_envs(pool, play_sem: asyncio.Semaphore) -> list[FinQAEnv]:
    """Open persistent WebSocket connections to all sandboxes."""
    envs: list[FinQAEnv | None] = [None] * len(pool)

    async def connect_one(i: int, url: str):
        async with play_sem:
            for attempt in range(MAX_PLAY_RETRIES):
                try:
                    env = FinQAEnv(base_url=url)
                    await env.connect()
                    # Increase WebSocket ping timeout to survive long vLLM
                    # generation steps (default 20s is too short).
                    if hasattr(env, "_ws") and env._ws is not None:
                        env._ws.ping_timeout = 300
                    envs[i] = env
                    return
                except (ConnectionError, TimeoutError, OSError):
                    if attempt == MAX_PLAY_RETRIES - 1:
                        raise
                    await asyncio.sleep(2)

    await asyncio.gather(*[connect_one(i, url) for i, (_, url) in enumerate(pool)])
    return [env for env in envs if env is not None]


async def disconnect_envs(envs: list[FinQAEnv]):
    """Close all persistent connections."""
    close_failures: list[tuple[int | str, Exception]] = []
    for idx, env in enumerate(envs):
        try:
            await env.close()
        except Exception as e:
            close_failures.append((idx, e))
    if close_failures:
        print(
            "  Warning: failed to close "
            + f"{len(close_failures)}/{len(envs)} env connections "
            + f"({summarize_exception_samples(close_failures)})"
        )


async def reconnect_envs(
    envs: list[FinQAEnv],
    pool,
    skip_indices: set[int] | None = None,
) -> list[FinQAEnv]:
    """Reconnect any envs whose WebSocket connections have gone stale.

    Tries a lightweight ping (state request) on each env. If it fails,
    closes and reopens the connection. Returns the (mutated) env list.

    Args:
        skip_indices: env indices to skip (e.g. those with in-flight tasks).
            Sending a state ping on an env with an in-flight step request
            causes WebSocket message interleaving — the ping response and
            step response can be received by the wrong awaiter.
    """
    reconnected = 0
    skip = skip_indices or set()

    async def check_and_reconnect(i: int):
        nonlocal reconnected
        env = envs[i]
        try:
            # Quick health check — if the WS is alive this returns fast
            await asyncio.wait_for(env._send_and_receive({"type": "state"}), timeout=5.0)
        except Exception:
            # Connection is dead, reconnect
            try:
                await env.close()
            except Exception as e:
                print(
                    "    Warning: failed to close stale env before reconnect " + f"(idx={i}): {type(e).__name__}: {e}"
                )
            _, url = pool[i]
            new_env = FinQAEnv(base_url=url)
            await new_env.connect()
            if hasattr(new_env, "_ws") and new_env._ws is not None:
                new_env._ws.ping_timeout = 300
            envs[i] = new_env
            reconnected += 1

    await asyncio.gather(
        *[check_and_reconnect(i) for i in range(len(envs)) if i not in skip],
        return_exceptions=True,
    )
    if reconnected:
        print(f"    Reconnected {reconnected}/{len(envs) - len(skip)} checked WebSocket connections.")


# ---------------------------------------------------------------------------
# Prompt building & tool call parsing
# ---------------------------------------------------------------------------
def build_chat_prompt(
    tokenizer,
    chat_history: list[dict],
) -> str:
    """Apply the chat template with tool definitions to produce a prompt string."""
    return tokenizer.apply_chat_template(
        chat_history,
        tools=FINQA_TOOLS,
        tokenize=False,
        add_generation_prompt=True,
        enable_thinking=False,  # Avoid burning tokens on <think> blocks
    )


def iter_json_objects(text: str):
    """Yield JSON objects parsed from arbitrary text by scanning for '{' starts."""
    decoder = json.JSONDecoder()
    i = 0
    n = len(text)
    while i < n:
        start = text.find("{", i)
        if start == -1:
            break
        try:
            obj, end = decoder.raw_decode(text, start)
        except json.JSONDecodeError:
            i = start + 1
            continue
        if isinstance(obj, dict):
            yield obj
        i = end


def extract_call_from_object(data: dict) -> tuple[str, dict] | None:
    """Normalize different tool-call JSON formats into (name, args)."""
    # Direct format: {"name": "...", "arguments": {...}}
    if "name" in data:
        name = data.get("name", "")
        args = data.get("arguments", {})
        if isinstance(args, str):
            args = json.loads(args)
        if name in TOOL_NAMES and isinstance(args, dict):
            return name, args

    # OpenAI-like message format with tool_calls list.
    tool_calls = data.get("tool_calls")
    if isinstance(tool_calls, list):
        for tc in tool_calls:
            if not isinstance(tc, dict):
                continue
            function_data = tc.get("function", {})
            if not isinstance(function_data, dict):
                continue
            name = function_data.get("name", "")
            args = function_data.get("arguments", {})
            if isinstance(args, str):
                args = json.loads(args)
            if name in TOOL_NAMES and isinstance(args, dict):
                return name, args

    return None


def parse_tool_call(text: str) -> tuple[str, dict]:
    """Parse a tool call from model output (Hermes-style XML or JSON).

    Tries several patterns:
    1. Hermes XML: <tool_call>{"name": ..., "arguments": ...}</tool_call>
    2. Raw JSON: {"name": ..., "arguments": ...}
    3. Fallback: submit_answer("unknown")
    """
    # Pattern 1: Hermes-style <tool_call> XML
    m = re.search(r"<tool_call>\s*(\{.*?\})\s*</tool_call>", text, re.DOTALL)
    if m:
        try:
            data = json.loads(m.group(1))
            parsed = extract_call_from_object(data)
            if parsed is not None:
                return parsed
        except (json.JSONDecodeError, TypeError, ValueError):
            pass

    # Pattern 2: raw JSON embedded in free-form text
    for data in iter_json_objects(text):
        try:
            parsed = extract_call_from_object(data)
            if parsed is not None:
                return parsed
        except (json.JSONDecodeError, TypeError, ValueError):
            continue

    # Pattern 3: bare answer after </think> (e.g. "</think>\n\n0.1647")
    m = re.search(r"</think>\s*(.+)", text, re.DOTALL)
    if m:
        bare = m.group(1).strip()
        if bare and bare != "unknown":
            return "submit_answer", {"answer": bare}

    # Pattern 4: text is just a number/short answer (no tool call structure)
    stripped = text.strip()
    if stripped and not stripped.startswith("<") and len(stripped) < 200:
        # Check if it looks like an answer (number, yes/no, etc.)
        if re.match(r"^[\d\.\-\+,: to]+$", stripped) or stripped.lower() in ("yes", "no"):
            return "submit_answer", {"answer": stripped}

    # Fallback
    return "submit_answer", {"answer": "unknown"}


def make_assistant_tool_call_msg(tool_name: str, tool_args: dict) -> dict:
    """Create an assistant message representing a tool call."""
    return {
        "role": "assistant",
        "content": None,
        "tool_calls": [
            {
                "id": "call_0",
                "type": "function",
                "function": {
                    "name": tool_name,
                    "arguments": json.dumps(tool_args),
                },
            }
        ],
    }


def make_tool_result_msg(tool_result: str) -> dict:
    """Create a tool result message."""
    return {
        "role": "tool",
        "tool_call_id": "call_0",
        "content": tool_result or "No result",
    }


# ---------------------------------------------------------------------------
# Batched rollout collection
# ---------------------------------------------------------------------------
async def collect_rollouts(
    envs: list[FinQAEnv],
    pool,
    vllm_model: LLM,
    tokenizer,
    group_size: int,
    lora_request: Any | None = None,
    max_episode_steps: int = MAX_EPISODE_STEPS,
    temperature: float = TEMPERATURE,
    max_gen_tokens: int = MAX_GEN_TOKENS,
    dispatch_wait_ms: int = ROLLOUT_DISPATCH_WAIT_MS,
    reconnect_every_batches: int = ROLLOUT_RECONNECT_EVERY_BATCHES,
) -> list[list[Episode]]:
    """Collect rollouts with dynamic refill to keep sandboxes occupied.

    Returns a list of groups (one group per env index), each containing all
    completed episodes collected from that env during this call.
    """
    n_envs = len(envs)
    groups: list[list[Episode]] = [[] for _ in range(n_envs)]
    if n_envs == 0:
        return groups

    target_episodes = n_envs * max(1, group_size)
    # Auto-tune: favor larger decode batches while still dispatching quickly.
    effective_min_batch = max(1, min(n_envs, max(8, n_envs // 4)))

    wait_s = max(0.0, dispatch_wait_ms / 1000.0)
    reconnect_every_batches = max(1, reconnect_every_batches)
    rollout_t0 = time.time()
    start_reset_timeout_s = 30.0
    start_state_timeout_s = 15.0
    step_timeout_s = 45.0
    loop_wait_timeout_s = max(wait_s, 1.0)

    ready_eps: deque[ActiveEpisode] = deque()
    start_tasks: dict[asyncio.Task, int] = {}
    step_tasks: dict[
        asyncio.Task,
        tuple[ActiveEpisode, list[int], list[int], str, dict, int],
    ] = {}
    force_tasks: dict[asyncio.Task, ActiveEpisode] = {}
    idle_envs: set[int] = set(range(n_envs))
    start_failures = [0] * n_envs
    cooldown_until = [0.0] * n_envs
    completed = 0
    last_completion_ts = time.monotonic()
    ready_since: float | None = None
    batch_idx = 0
    progress_step = max(25, n_envs // 2)
    next_progress = progress_step

    async def start_episode(env_idx: int) -> ActiveEpisode | None:
        env = envs[env_idx]
        await asyncio.wait_for(env.reset(), timeout=start_reset_timeout_s)
        state_res = await asyncio.wait_for(
            env._send_and_receive({"type": "state"}),
            timeout=start_state_timeout_s,
        )
        state_data = state_res.get("data", {})
        question = state_data.get("current_question", "")
        company = state_data.get("current_company", "")
        question_id = str(state_data.get("question_id", "") or "")
        chat_history = [
            {"role": "system", "content": SYSTEM_PROMPT},
            {"role": "user", "content": f"Company: {company}\nQuestion: {question}"},
        ]
        return ActiveEpisode(
            env=envs[env_idx],
            sandbox_idx=env_idx,
            chat_history=chat_history,
            question=question,
            company=company,
            question_id=question_id,
        )

    def complete_episode(ep: ActiveEpisode, reward: float):
        nonlocal completed, last_completion_ts
        ep.done = True
        ep.reward = reward
        groups[ep.sandbox_idx].append(
            Episode(
                sandbox_idx=ep.sandbox_idx,
                turns=ep.turns,
                reward=ep.reward,
                question=ep.question,
                company=ep.company,
                question_id=ep.question_id,
            )
        )
        completed += 1
        last_completion_ts = time.monotonic()
        idle_envs.add(ep.sandbox_idx)

    def maybe_launch_episode_starts():
        if completed >= target_episodes:
            return
        now = time.monotonic()
        for env_idx in list(idle_envs):
            if cooldown_until[env_idx] > now:
                continue
            idle_envs.remove(env_idx)
            start_tasks[asyncio.create_task(start_episode(env_idx))] = env_idx

    while completed < target_episodes:
        maybe_launch_episode_starts()

        now = time.monotonic()
        if ready_eps and ready_since is None:
            ready_since = now

        # Dispatch a generation batch from whichever episodes are ready.
        if ready_eps:
            waited = now - (ready_since or now)
            more_coming = bool(start_tasks or step_tasks or force_tasks)
            should_dispatch = (
                (len(ready_eps) >= effective_min_batch and not more_coming)
                or (len(ready_eps) >= n_envs)
                or (not more_coming)
                or (more_coming and waited >= wait_s)
            )
            if should_dispatch:
                batch = list(ready_eps)
                ready_eps.clear()
                ready_since = None
                batch_idx += 1

                prompts = []
                prompt_token_ids_list = []
                for ep in batch:
                    prompt_str = build_chat_prompt(tokenizer, ep.chat_history)
                    prompts.append(prompt_str)
                    prompt_token_ids_list.append(tokenizer.encode(prompt_str, add_special_tokens=False))

                sampling_params = SamplingParams(
                    temperature=temperature,
                    max_tokens=max_gen_tokens,
                    stop=STOP_STRINGS,
                    include_stop_str_in_output=True,
                )
                generate_kwargs = {
                    "prompts": prompts,
                    "sampling_params": sampling_params,
                }
                if lora_request is not None:
                    generate_kwargs["lora_request"] = lora_request
                # Offload blocking vLLM generation so env I/O tasks can finish.
                outputs = await asyncio.to_thread(vllm_model.generate, **generate_kwargs)

                # Periodically refresh stale WebSocket connections.
                # Skip envs with in-flight tasks to avoid WS message
                # interleaving (state ping response vs step response).
                if batch_idx % reconnect_every_batches == 0:
                    busy_indices = set()
                    for env_idx in start_tasks.values():
                        busy_indices.add(env_idx)
                    for meta in step_tasks.values():
                        busy_indices.add(meta[0].sandbox_idx)
                    for fep in force_tasks.values():
                        busy_indices.add(fep.sandbox_idx)
                    await reconnect_envs(envs, pool, skip_indices=busy_indices)
                for ep in batch:
                    ep.env = envs[ep.sandbox_idx]

                for i, (ep, prompt_ids, output) in enumerate(zip(batch, prompt_token_ids_list, outputs)):
                    model_output = output.outputs[0] if output.outputs else None
                    generated_text = model_output.text if model_output else ""
                    completion_ids = list(model_output.token_ids) if model_output else []
                    tool_name, tool_args = parse_tool_call(generated_text)

                    if i < DEBUG_LOG_SAMPLE_LIMIT:
                        user_msg = ep.chat_history[1]["content"] if len(ep.chat_history) > 1 else "(no user msg)"
                        debug_logger.info(
                            "\n=== batch=%s step=%s sandbox=%s ==="
                            "\nUSER_PROMPT: %s\nGENERATED (%s chars):\n%s"
                            "\nPARSED: %s(%s)",
                            batch_idx,
                            len(ep.turns),
                            ep.sandbox_idx,
                            user_msg[:300],
                            len(generated_text),
                            generated_text[:1000],
                            tool_name,
                            json.dumps(tool_args)[:200],
                        )

                    action = CallToolAction(tool_name=tool_name, arguments=tool_args)
                    task = asyncio.create_task(asyncio.wait_for(ep.env.step(action), timeout=step_timeout_s))
                    step_tasks[task] = (
                        ep,
                        prompt_ids,
                        completion_ids,
                        tool_name,
                        tool_args,
                        len(ep.turns),
                    )
                continue

        pending = list(start_tasks.keys()) + list(step_tasks.keys()) + list(force_tasks.keys())
        if not pending:
            if completed < target_episodes and not ready_eps and (time.monotonic() - last_completion_ts) > 45.0:
                print("    Rollout warning: stalled without completions for 45s; ending round early.")
                break
            if completed >= target_episodes:
                break
            if idle_envs:
                # No active work right now; wait until at least one env cooldown expires.
                soonest = min(cooldown_until[i] for i in idle_envs)
                sleep_for = max(0.01, soonest - time.monotonic())
                await asyncio.sleep(min(sleep_for, 1.0))
                continue
            await asyncio.sleep(max(wait_s, 0.01))
            continue

        done, _ = await asyncio.wait(
            pending,
            return_when=asyncio.FIRST_COMPLETED,
            timeout=loop_wait_timeout_s,
        )
        if not done:
            if (time.monotonic() - last_completion_ts) > 45.0:
                print("    Rollout warning: stalled with in-flight tasks for 45s; ending round early.")
                break
            continue

        for task in done:
            env_idx = start_tasks.pop(task, None)
            if env_idx is not None:
                try:
                    ep = task.result()
                except Exception as e:
                    ep = None
                    if env_idx < DEBUG_SANDBOX_LOG_LIMIT:
                        debug_logger.info(
                            "  START ERROR sandbox=%s: %s: %s",
                            env_idx,
                            type(e).__name__,
                            e,
                        )
                if ep is None:
                    start_failures[env_idx] += 1
                    if start_failures[env_idx] >= MAX_PLAY_RETRIES:
                        # Skip envs with in-flight tasks to avoid WS interleaving.
                        busy = set()
                        for ei in start_tasks.values():
                            busy.add(ei)
                        for m in step_tasks.values():
                            busy.add(m[0].sandbox_idx)
                        for fe in force_tasks.values():
                            busy.add(fe.sandbox_idx)
                        await reconnect_envs(envs, pool, skip_indices=busy)
                        start_failures[env_idx] = 0
                    fail_count = max(1, start_failures[env_idx])
                    backoff = min(5.0, 0.25 * (2 ** (fail_count - 1)))
                    cooldown_until[env_idx] = time.monotonic() + backoff
                    idle_envs.add(env_idx)
                else:
                    start_failures[env_idx] = 0
                    cooldown_until[env_idx] = 0.0
                    ready_eps.append(ep)
                    if ready_since is None:
                        ready_since = time.monotonic()
                continue

            step_meta = step_tasks.pop(task, None)
            if step_meta is None:
                force_ep = force_tasks.pop(task, None)
                if force_ep is None:
                    continue
                try:
                    force_result = task.result()
                    force_reward = force_result.observation.reward or 0.0
                except Exception as e:
                    if force_ep.sandbox_idx < DEBUG_SANDBOX_LOG_LIMIT:
                        debug_logger.info(
                            "  FORCE ERROR sandbox=%s: %s: %s",
                            force_ep.sandbox_idx,
                            type(e).__name__,
                            e,
                        )
                    force_reward = 0.0
                complete_episode(force_ep, force_reward)
                if completed >= next_progress:
                    active_eps = len(ready_eps) + len(step_tasks) + len(start_tasks) + len(force_tasks)
                    print(
                        f"    Rollout progress: {completed}/{target_episodes} done, "
                        + f"active={active_eps}, ready={len(ready_eps)}"
                    )
                    next_progress += progress_step
                continue
            ep, prompt_ids, completion_ids, tool_name, tool_args, step_idx = step_meta

            try:
                result = task.result()
            except Exception as e:
                tool_result_text = f"Error: {e}"
                reward = 0.0
                # Connection errors are unrecoverable mid-episode — mark done.
                # submit_answer is always terminal regardless of error type.
                is_conn_error = "Connection" in type(e).__name__ or "closed" in str(e).lower()
                done_flag = is_conn_error or tool_name == "submit_answer"
                if ep.sandbox_idx < DEBUG_SANDBOX_LOG_LIMIT:
                    debug_logger.info(
                        "  STEP ERROR sandbox=%s (done=%s): %s: %s",
                        ep.sandbox_idx,
                        done_flag,
                        type(e).__name__,
                        e,
                    )
            else:
                obs = result.observation
                raw_result = obs.result if hasattr(obs, "result") and obs.result else obs.metadata
                # obs.result may be a dict with MCP content structure;
                # extract the text for chat history.
                if isinstance(raw_result, dict):
                    sc = raw_result.get("structured_content", {})
                    if "result" in sc:
                        tool_result_text = str(sc["result"])
                    else:
                        # Fallback: grab text from content array
                        parts = raw_result.get("content", [])
                        tool_result_text = " ".join(p.get("text", "") for p in parts if isinstance(p, dict)) or str(
                            raw_result
                        )
                else:
                    tool_result_text = str(raw_result) if raw_result else "No result"
                done_flag = obs.done
                reward = obs.reward or 0.0
                # submit_answer is always terminal — guard against the
                # server returning done=False due to WebSocket message
                # interleaving (reconnect_envs ping vs in-flight step).
                if tool_name == "submit_answer" and not done_flag:
                    done_flag = True
                    if ep.sandbox_idx < DEBUG_SANDBOX_LOG_LIMIT:
                        debug_logger.info(
                            "  WARN: submit_answer got done=False on sandbox=%s, "
                            + "forcing done=True (likely WS interleave)",
                            ep.sandbox_idx,
                        )
                if ep.sandbox_idx < DEBUG_SANDBOX_LOG_LIMIT:
                    debug_logger.info(
                        "  RESULT sandbox=%s step=%s tool=%s: done=%s reward=%s result=%s",
                        ep.sandbox_idx,
                        step_idx,
                        tool_name,
                        done_flag,
                        reward,
                        str(tool_result_text)[:200],
                    )

            ep.turns.append(
                Turn(
                    prompt_token_ids=prompt_ids,
                    completion_token_ids=completion_ids,
                    tool_name=tool_name,
                    tool_args=tool_args,
                    tool_result=tool_result_text,
                )
            )

            if done_flag:
                complete_episode(ep, reward)
            elif len(ep.turns) >= max_episode_steps:
                # Force terminal action once turn budget is exhausted.
                force_task = asyncio.create_task(
                    asyncio.wait_for(
                        ep.env.step(
                            CallToolAction(
                                tool_name="submit_answer",
                                arguments={"answer": "unknown"},
                            )
                        ),
                        timeout=step_timeout_s,
                    )
                )
                force_tasks[force_task] = ep
            else:
                ep.chat_history.append(make_assistant_tool_call_msg(tool_name, tool_args))
                ep.chat_history.append(make_tool_result_msg(tool_result_text))
                ready_eps.append(ep)
                if ready_since is None:
                    ready_since = time.monotonic()

            if completed >= next_progress:
                active_eps = len(ready_eps) + len(step_tasks) + len(start_tasks) + len(force_tasks)
                print(
                    f"    Rollout progress: {completed}/{target_episodes} done, "
                    + f"active={active_eps}, ready={len(ready_eps)}"
                )
                next_progress += progress_step

    # Cancel any excess in-flight work once target sample count is reached.
    pending_cancel = list(start_tasks.keys()) + list(step_tasks.keys()) + list(force_tasks.keys())
    # Track envs with in-flight WS requests — cancellation leaves stale
    # responses queued on the socket, corrupting subsequent communication.
    stale_env_indices = set()
    for env_idx in start_tasks.values():
        stale_env_indices.add(env_idx)
    for meta in step_tasks.values():
        stale_env_indices.add(meta[0].sandbox_idx)
    for fep in force_tasks.values():
        stale_env_indices.add(fep.sandbox_idx)
    for task in pending_cancel:
        task.cancel()
    if pending_cancel:
        await asyncio.gather(*pending_cancel, return_exceptions=True)
    # Force-disconnect envs whose WebSocket has stale responses from
    # cancelled tasks.  The next reconnect_envs() will reopen them cleanly.
    disconnect_failures: list[tuple[int | str, Exception]] = []
    for idx in stale_env_indices:
        try:
            await envs[idx].disconnect()
        except Exception as e:
            disconnect_failures.append((idx, e))
    if disconnect_failures:
        print(
            "    Warning: failed to disconnect "
            + f"{len(disconnect_failures)}/{len(stale_env_indices)} stale envs "
            + f"({summarize_exception_samples(disconnect_failures)})"
        )

    flat_eps = [ep for per_env in groups for ep in per_env]
    elapsed = time.time() - rollout_t0
    n_correct = sum(1 for ep in flat_eps if ep.reward > 0)
    avg_turns = np.mean([len(ep.turns) for ep in flat_eps]) if flat_eps else 0.0
    print(
        f"    Rollout done: {len(flat_eps)} eps, {n_correct}/{len(flat_eps)} correct, "
        + f"avg_turns={avg_turns:.1f}, {elapsed:.1f}s"
    )
    return groups


# ---------------------------------------------------------------------------
# GRPO: strict same-prompt grouping, advantages, update
# ---------------------------------------------------------------------------
def episode_prompt_key(ep: Episode) -> tuple[str, str]:
    """Stable prompt identity for grouping rollouts."""
    if ep.question_id:
        return ("id", ep.question_id)
    return ("text", f"{ep.company}\n{ep.question}")


def build_strict_prompt_groups(
    episodes: list[Episode],
    group_size: int,
) -> tuple[list[list[Episode]], list[Episode]]:
    """Form exact-size same-prompt groups, returning (groups, leftovers)."""
    buckets: dict[tuple[str, str], list[Episode]] = defaultdict(list)
    for ep in episodes:
        buckets[episode_prompt_key(ep)].append(ep)

    groups: list[list[Episode]] = []
    leftovers: list[Episode] = []
    for bucket in buckets.values():
        n_full = len(bucket) // group_size
        n_take = n_full * group_size
        for i in range(n_full):
            start = i * group_size
            groups.append(bucket[start : start + group_size])
        leftovers.extend(bucket[n_take:])

    return groups, leftovers


def compute_group_advantages(groups: list[list[Episode]]) -> list[list[float]]:
    """Compute GRPO advantages strictly within each provided prompt group."""
    all_advantages: list[list[float]] = []
    for group in groups:
        if not group:
            all_advantages.append([])
            continue
        rewards = np.array([ep.reward for ep in group], dtype=np.float32)
        std = float(np.std(rewards))
        if len(group) > 1 and std > 1e-8:
            mean = float(np.mean(rewards))
            advs = (rewards - mean) / (std + 1e-8)
        else:
            advs = np.zeros_like(rewards)
        all_advantages.append([float(a) for a in advs])
    return all_advantages


def grpo_update(
    train_model,
    optimizer: torch.optim.Optimizer,
    episodes_flat: list[Episode],
    advantages_flat: list[float],
    batch_size: int = GRPO_UPDATE_BATCH_SIZE,
) -> float:
    """GRPO policy gradient update.

    For each episode turn:
      loss_turn = -(advantage * policy_logprob)

    Micro-batches by episode to fit in GPU memory.
    Device is derived from the model parameters (supports device_map="auto").
    """
    train_model.train()
    optimizer.zero_grad(set_to_none=True)
    total_loss_t = None
    total_tokens_t = None
    input_device = next(train_model.parameters()).device
    batch_size = max(1, int(batch_size))

    # Precompute per-episode token counts and flatten contributing turns.
    ep_token_counts = [0 for _ in episodes_flat]
    turn_samples: list[dict[str, Any]] = []
    for ep_idx, (ep, adv) in enumerate(zip(episodes_flat, advantages_flat)):
        if not ep.turns or adv == 0.0:
            continue
        for _, turn in enumerate(ep.turns):
            if not turn.completion_token_ids:
                continue
            comp_len = len(turn.completion_token_ids)
            ep_token_counts[ep_idx] += comp_len
            turn_samples.append(
                {
                    "ep_idx": ep_idx,
                    "adv": adv,
                    "prompt_len": len(turn.prompt_token_ids),
                    "input_ids": turn.prompt_token_ids + turn.completion_token_ids,
                    "completion_ids": turn.completion_token_ids,
                }
            )

    n_contributing = sum(1 for t in ep_token_counts if t > 0)
    if n_contributing == 0 or not turn_samples:
        return 0.0

    # Bucket by length to reduce padding waste.
    turn_samples.sort(key=lambda s: len(s["input_ids"]))

    pad_token_id = train_model.config.pad_token_id
    if pad_token_id is None:
        pad_token_id = train_model.config.eos_token_id
    if isinstance(pad_token_id, list):
        pad_token_id = pad_token_id[0]
    if pad_token_id is None:
        pad_token_id = 0

    for start in range(0, len(turn_samples), batch_size):
        chunk = turn_samples[start : start + batch_size]
        max_len = max(len(s["input_ids"]) for s in chunk)
        bs = len(chunk)

        input_t = torch.full((bs, max_len), pad_token_id, dtype=torch.long, device=input_device)
        attn_mask = torch.zeros((bs, max_len), dtype=torch.long, device=input_device)

        for i, sample in enumerate(chunk):
            seq = sample["input_ids"]
            seq_len = len(seq)
            input_t[i, :seq_len] = torch.tensor(seq, dtype=torch.long, device=input_device)
            attn_mask[i, :seq_len] = 1

        outputs = train_model(input_ids=input_t, attention_mask=attn_mask)
        logits = outputs.logits
        prompt_lens = [int(sample["prompt_len"]) for sample in chunk]
        comp_lens = [len(sample["completion_ids"]) for sample in chunk]
        max_comp_len = max(comp_lens)

        comp_targets = torch.full((bs, max_comp_len), -100, dtype=torch.long, device=logits.device)

        sample_advs = []
        sample_scales = []
        for i, sample in enumerate(chunk):
            comp_len = comp_lens[i]
            if comp_len > 0:
                comp_targets[i, :comp_len] = torch.tensor(
                    sample["completion_ids"], dtype=torch.long, device=logits.device
                )
            sample_advs.append(float(sample["adv"]))
            sample_scales.append(1.0 / (n_contributing * ep_token_counts[sample["ep_idx"]]))

        prompt_lens_t = torch.tensor(prompt_lens, dtype=torch.long, device=logits.device)
        comp_lens_t = torch.tensor(comp_lens, dtype=torch.long, device=logits.device)
        batch_idx = torch.arange(bs, device=logits.device)[:, None]
        token_offsets = torch.arange(max_comp_len, device=logits.device)[None, :]
        valid_prompt_mask = prompt_lens_t > 0
        valid_mask = (token_offsets < comp_lens_t[:, None]) & valid_prompt_mask[:, None]

        # Clamp invalid prompt rows to a safe index; valid_mask zeroes them out.
        gather_pos = torch.clamp(prompt_lens_t - 1, min=0)[:, None] + token_offsets
        gather_pos = torch.clamp(gather_pos, max=logits.shape[1] - 1)
        completion_logits = logits[batch_idx, gather_pos]

        vocab_size = completion_logits.shape[-1]
        nll = F.cross_entropy(
            completion_logits.reshape(-1, vocab_size),
            comp_targets.reshape(-1),
            reduction="none",
            ignore_index=-100,
        ).reshape(bs, max_comp_len)
        policy_lps = -nll

        valid_f = valid_mask.to(dtype=policy_lps.dtype)
        adv_t = torch.tensor(sample_advs, device=logits.device, dtype=policy_lps.dtype)[:, None]
        scale_t = torch.tensor(sample_scales, device=logits.device, dtype=policy_lps.dtype)[:, None]

        token_loss = (-adv_t * policy_lps) * valid_f

        if total_loss_t is None:
            total_loss_t = torch.zeros((), device=logits.device, dtype=torch.float32)
            total_tokens_t = torch.zeros((), device=logits.device, dtype=torch.long)

        total_loss_t = total_loss_t + token_loss.detach().sum().to(torch.float32)
        total_tokens_t = total_tokens_t + valid_mask.sum()

        # Match prior objective: mean per-token per episode, then mean across episodes.
        batch_loss = (token_loss * scale_t).sum()

        batch_loss.backward()

    torch.nn.utils.clip_grad_norm_(train_model.parameters(), max_norm=1.0)
    optimizer.step()

    if total_loss_t is None or total_tokens_t is None:
        return 0.0
    total_tokens = int(total_tokens_t.item())
    total_loss = float(total_loss_t.item())
    return total_loss / max(total_tokens, 1)


def resolve_lora_request_class():
    """Load LoRARequest from whichever module path the installed vLLM exposes."""
    try:
        from vllm.lora.request import LoRARequest as _cls  # pylint: disable=import-outside-toplevel

        return _cls
    except Exception:
        try:
            from vllm import LoRARequest as _cls  # pylint: disable=import-outside-toplevel,no-name-in-module

            return _cls
        except Exception as e:
            raise ImportError(
                "Could not import LoRARequest from vLLM. Upgrade vLLM or use a version with LoRA runtime support."
            ) from e


def export_lora_adapter(
    train_model,
    export_root: str,
    iteration: int,
) -> str:
    """Save adapter-only checkpoint for vLLM LoRA runtime loading."""
    out_dir = os.path.join(export_root, f"iter_{iteration:04d}")
    train_model.save_pretrained(out_dir)
    return out_dir


def run_grpo_update_and_maybe_export(
    train_model,
    optimizer: torch.optim.Optimizer,
    prepared_batch: PreparedTrainBatch,
    batch_size: int,
    export_root: str,
    iteration: int,
    do_export: bool,
) -> tuple[float, str | None]:
    """Run GRPO update on training GPUs and optionally export adapter weights."""
    loss = grpo_update(
        train_model,
        optimizer,
        prepared_batch.eps_flat,
        prepared_batch.advs_flat,
        batch_size=batch_size,
    )
    new_lora_dir = export_lora_adapter(train_model, export_root, iteration) if do_export else None
    return loss, new_lora_dir


async def prepare_train_batch(
    *,
    iter_idx: int,
    total_iters: int,
    envs: list[FinQAEnv],
    pool,
    vllm_model: LLM,
    tokenizer,
    args,
    lora_request: Any | None,
    reconnect_before_collect: bool,
    carry_pending_eps: list[Episode] | None = None,
) -> PreparedTrainBatch:
    """Prepare one training batch: rollouts, grouping, advantages, ref logprobs."""
    if reconnect_before_collect:
        await reconnect_envs(envs, pool)

    print(f"\n  Iteration {iter_idx + 1}/{total_iters}: collecting rollouts...")
    groups: list[list[Episode]] = []
    pending_eps: list[Episode] = list(carry_pending_eps or [])
    carry_in_count = len(pending_eps)
    rolled_eps = 0
    rounds_used = 0

    for rr in range(args.max_rollout_rounds):
        if rr > 0:
            await reconnect_envs(envs, pool)

        round_t0 = time.time()
        round_groups = await collect_rollouts(
            envs,
            pool,
            vllm_model,
            tokenizer,
            1,
            lora_request=lora_request,
            max_episode_steps=args.max_steps,
            temperature=args.temperature,
            max_gen_tokens=args.max_gen_tokens,
            dispatch_wait_ms=args.rollout_dispatch_wait_ms,
        )
        round_eps = [ep for g in round_groups for ep in g]
        rolled_eps += len(round_eps)
        pending_eps.extend(round_eps)

        # Persist trajectories immediately (crash-safe)
        save_episodes(
            os.path.join(args.run_dir, "trajectories.jsonl"),
            round_eps,
            args.save_token_ids,
        )

        new_groups, pending_eps = build_strict_prompt_groups(pending_eps, args.group_size)
        groups.extend(new_groups)
        rounds_used = rr + 1

        grouped_eps = len(groups) * args.group_size
        elapsed_round = time.time() - round_t0
        print(
            f"    Round {rr + 1}/{args.max_rollout_rounds}: "
            + f"rolled={len(round_eps)} total_rolled={rolled_eps} "
            + f"groups={len(groups)} grouped_eps={grouped_eps} "
            + f"pending={len(pending_eps)} "
            + f"{elapsed_round:.1f}s"
        )

        # Persist rollout round summary
        _append_jsonl(
            os.path.join(args.run_dir, "rollouts.jsonl"),
            {
                "iteration": iter_idx + 1,
                "round": rr + 1,
                "round_eps": len(round_eps),
                "correct": sum(1 for ep in round_eps if ep.reward > 0),
                "avg_turns": (float(np.mean([len(ep.turns) for ep in round_eps])) if round_eps else 0.0),
                "elapsed_s": round(elapsed_round, 2),
                "groups": len(groups),
                "grouped_eps": grouped_eps,
                "pending": len(pending_eps),
            },
        )

        if args.target_groups_per_iter > 0 and len(groups) >= args.target_groups_per_iter:
            break

    if not groups:
        print("  No strict same-question groups collected, skipping update.")
        return PreparedTrainBatch(
            eps_flat=[],
            advs_flat=[],
            groups_count=0,
            rolled_eps=rolled_eps,
            rounds_used=rounds_used,
            carry_in_count=carry_in_count,
            leftover_eps=pending_eps,
        )

    advs = compute_group_advantages(groups)
    eps_flat = [ep for g in groups for ep in g]
    advs_flat = [a for ag in advs for a in ag]
    if not eps_flat:
        print("  No episodes collected, skipping update.")
        return PreparedTrainBatch(
            eps_flat=[],
            advs_flat=[],
            groups_count=len(groups),
            rolled_eps=rolled_eps,
            rounds_used=rounds_used,
            carry_in_count=carry_in_count,
            leftover_eps=pending_eps,
        )

    return PreparedTrainBatch(
        eps_flat=eps_flat,
        advs_flat=advs_flat,
        groups_count=len(groups),
        rolled_eps=rolled_eps,
        rounds_used=rounds_used,
        carry_in_count=carry_in_count,
        leftover_eps=pending_eps,
    )


# ---------------------------------------------------------------------------
# Main training loop
# ---------------------------------------------------------------------------
async def train(args):
    load_dotenv()

    # --- Set up persistent run directory for JSONL logs ---
    if args.run_dir is None:
        args.run_dir = os.path.join("runs", datetime.now().strftime("%Y%m%d_%H%M%S"))
    os.makedirs(args.run_dir, exist_ok=True)
    with open(os.path.join(args.run_dir, "config.json"), "w", encoding="utf-8") as f:
        json.dump(vars(args), f, indent=2, default=str)
    print(f"Run directory: {args.run_dir}")

    tp_size = args.tensor_parallel_size
    n_gpus = torch.cuda.device_count()
    print(f"4-GPU LoRA mode: vLLM on cuda:0-{tp_size - 1} (TP={tp_size}), training on cuda:{tp_size}-{n_gpus - 1}")

    print(f"Model: {args.model}")
    print(f"GPUs: {n_gpus} x {torch.cuda.get_device_name(0)}")
    print(f"LoRA: rank={args.lora_rank}, alpha={args.lora_alpha}, dropout={args.lora_dropout}")
    print(f"Sandboxes: {args.sandboxes}")
    print(f"Iterations: {args.iterations}")
    print(f"GRPO group size K: {args.group_size}")
    print(f"Target groups/iter: {args.target_groups_per_iter} (0=use all rollout rounds)")
    print(f"Max rollout rounds/iter: {args.max_rollout_rounds}")
    print(f"Rollout dispatch wait: {args.rollout_dispatch_wait_ms} ms")
    print(f"Actor LoRA sync every: {args.sync_every} iter(s)")
    print()

    # --- Load tokenizer ---
    print("Loading tokenizer...")
    tokenizer = AutoTokenizer.from_pretrained(args.model)
    if tokenizer.pad_token is None:
        tokenizer.pad_token = tokenizer.eos_token

    # --- Load vLLM model (for fast batched generation on cuda:0-1, TP=2) ---
    print("Loading vLLM model...")
    vllm_kwargs = {
        "model": args.model,
        "gpu_memory_utilization": args.gpu_memory_utilization,
        "tensor_parallel_size": tp_size,
        "dtype": "auto",
        "enforce_eager": True,
        "enable_lora": True,
        "max_loras": 1,
        "max_lora_rank": args.lora_rank,
    }
    try:
        vllm_model = LLM(**vllm_kwargs)
    except TypeError:
        # Backward compatibility for older vLLM signatures.
        vllm_kwargs.pop("max_loras", None)
        vllm_model = LLM(**vllm_kwargs)

    # --- Load base model on cuda:2-3 with device_map, then wrap with LoRA ---
    print("Loading training model (base + LoRA)...")
    # Reserve vLLM GPUs (0..tp_size-1) by giving them 0 memory
    max_memory = {i: "0GiB" for i in range(tp_size)}
    for i in range(tp_size, n_gpus):
        max_memory[i] = "75GiB"
    base_model = AutoModelForCausalLM.from_pretrained(
        args.model,
        torch_dtype=torch.bfloat16,
        device_map="auto",
        max_memory=max_memory,
    )

    lora_config = LoraConfig(
        r=args.lora_rank,
        lora_alpha=args.lora_alpha,
        lora_dropout=args.lora_dropout,
        target_modules=args.lora_target_modules,
        task_type="CAUSAL_LM",
    )
    train_model = get_peft_model(base_model, lora_config)
    train_model.print_trainable_parameters()
    train_model.enable_input_require_grads()
    # Training uses full-sequence teacher-forced forwards; KV cache is not useful
    # here and is incompatible with gradient checkpointing.
    train_model.config.use_cache = False
    if hasattr(train_model, "base_model") and hasattr(train_model.base_model, "config"):
        train_model.base_model.config.use_cache = False

    if args.disable_gradient_checkpointing:
        train_model.gradient_checkpointing_disable()
    else:
        train_model.gradient_checkpointing_enable()
    train_model.train()

    optimizer = torch.optim.AdamW(
        [p for p in train_model.parameters() if p.requires_grad],
        lr=args.lr,
    )

    # --- Create sandbox pool ---
    pool = []
    envs = []
    lora_export_root = tempfile.mkdtemp(prefix="grpo_lora_")
    active_lora_dir = ""
    active_lora_request = None
    lora_request_seq = 1
    lora_request_cls = None
    inflight_update_task = None
    try:
        lora_request_cls = resolve_lora_request_class()
        # Seed the actor with the initial adapter so rollout-1 matches
        # train_model state instead of using the base model.
        active_lora_dir = export_lora_adapter(train_model, lora_export_root, 0)
        active_lora_request = lora_request_cls("grpo_init", lora_request_seq, active_lora_dir)
        lora_request_seq += 1

        print(f"\nCreating {args.sandboxes} sandboxes from snapshot '{args.snapshot}' ...")
        semaphore = asyncio.Semaphore(MAX_CONCURRENT_CREATE)
        pool = await create_sandbox_pool(args.sandboxes, args.snapshot, semaphore)
        print(f"All {len(pool)} sandboxes ready.\n")

        play_sem = asyncio.Semaphore(MAX_CONCURRENT_PLAY)
        print("Connecting to sandboxes ...")
        envs = await connect_envs(pool, play_sem)
        print(f"All {len(envs)} connections ready.\n")

        # --- Fetch tool schemas from the env ---
        global FINQA_TOOLS, TOOL_NAMES  # noqa: PLW0603  # pylint: disable=global-statement
        FINQA_TOOLS = await fetch_tools_from_env(envs[0])
        TOOL_NAMES = {t["function"]["name"] for t in FINQA_TOOLS}
        print(f"Tools: {sorted(TOOL_NAMES)}\n")

        # --- Training loop ---
        header = (
            f"{'iter':>6}  {'accuracy':>9}  {'avg_steps':>10}  "
            + f"{'loss':>9}  {'groups':>7}  {'eps/s':>8}  {'time':>7}"
        )
        print(header)
        print("-" * len(header))

        # Warm up one prepared batch so subsequent iterations can overlap
        # (update/export on train GPUs) with (next rollouts + ref logprobs on vLLM).
        prepared_batch = None
        pending_eps_carry: list[Episode] = []
        if args.iterations > 0:
            prepared_batch = await prepare_train_batch(
                iter_idx=0,
                total_iters=args.iterations,
                envs=envs,
                pool=pool,
                vllm_model=vllm_model,
                tokenizer=tokenizer,
                args=args,
                lora_request=active_lora_request,
                reconnect_before_collect=False,
                carry_pending_eps=pending_eps_carry,
            )
            pending_eps_carry = prepared_batch.leftover_eps

        for it in range(args.iterations):
            t0 = time.time()
            batch = prepared_batch
            if batch is None:
                break

            # 1. Start GRPO update/export for the current prepared batch.
            update_task = None
            if batch.eps_flat:
                print("  Running GRPO update...")
                if (it + 1) % args.sync_every == 0:
                    print("  Exporting LoRA adapter for vLLM actor...")
                update_task = asyncio.create_task(
                    asyncio.to_thread(
                        run_grpo_update_and_maybe_export,
                        train_model,
                        optimizer,
                        batch,
                        args.grpo_update_batch_size,
                        lora_export_root,
                        it + 1,
                        (it + 1) % args.sync_every == 0,
                    )
                )
                inflight_update_task = update_task

            # 2. While the train GPUs update iteration t, use vLLM + envs to
            # prepare iteration t+1 (lag-1 actor pipeline).
            try:
                if it + 1 < args.iterations:
                    prepared_batch = await prepare_train_batch(
                        iter_idx=it + 1,
                        total_iters=args.iterations,
                        envs=envs,
                        pool=pool,
                        vllm_model=vllm_model,
                        tokenizer=tokenizer,
                        args=args,
                        lora_request=active_lora_request,
                        reconnect_before_collect=True,
                        carry_pending_eps=pending_eps_carry,
                    )
                    pending_eps_carry = prepared_batch.leftover_eps
                else:
                    prepared_batch = None
            except Exception:
                if inflight_update_task is not None:
                    try:
                        await asyncio.shield(inflight_update_task)
                    except Exception as e:
                        print(f"  Warning: in-flight update failed during prep error: {type(e).__name__}: {e}")
                    inflight_update_task = None
                raise

            if update_task is None:
                continue

            loss, new_lora_dir = await update_task
            inflight_update_task = None

            # 3. Publish the newly exported adapter to vLLM for future rollouts.
            if new_lora_dir:
                active_lora_request = lora_request_cls(
                    f"grpo_iter_{it + 1}",
                    lora_request_seq,
                    new_lora_dir,
                )
                lora_request_seq += 1
                active_lora_dir = new_lora_dir

            # Metrics
            elapsed = time.time() - t0
            rewards = [ep.reward for ep in batch.eps_flat]
            accuracy = np.mean(rewards) if rewards else 0.0
            avg_steps = np.mean([len(ep.turns) for ep in batch.eps_flat]) if batch.eps_flat else 0
            grouped_eps = len(batch.eps_flat)
            eps_sec = grouped_eps / elapsed if elapsed > 0 else 0

            print(
                f"  Rollout summary: rounds={batch.rounds_used}, rolled={batch.rolled_eps}, "
                + f"grouped={grouped_eps}, groups={batch.groups_count}, "
                + f"carry_in={batch.carry_in_count}, carry_out={len(batch.leftover_eps)}"
            )

            print(
                f"{it + 1:>4}/{args.iterations}  {accuracy:>9.3f}  "
                + f"{avg_steps:>10.1f}  {loss:>9.4f}  {batch.groups_count:>7d}  "
                + f"{eps_sec:>8.1f}  {elapsed:>6.0f}s"
            )

            # Persist iteration metrics
            _append_jsonl(
                os.path.join(args.run_dir, "metrics.jsonl"),
                {
                    "iteration": it + 1,
                    "accuracy": round(float(accuracy), 4),
                    "avg_steps": round(float(avg_steps), 2),
                    "loss": round(float(loss), 6),
                    "groups_count": batch.groups_count,
                    "eps_sec": round(float(eps_sec), 2),
                    "elapsed_s": round(elapsed, 2),
                    "rounds_used": batch.rounds_used,
                    "rolled_eps": batch.rolled_eps,
                    "grouped_eps": grouped_eps,
                    "carry_in": batch.carry_in_count,
                    "carry_out": len(batch.leftover_eps),
                },
            )

        if pending_eps_carry:
            print(f"\nDropping {len(pending_eps_carry)} unmatched leftover episodes at end of training.")

    finally:
        if inflight_update_task is not None:
            try:
                await asyncio.shield(inflight_update_task)
            except Exception as e:
                print(f"  Warning: in-flight update failed during shutdown: {type(e).__name__}: {e}")
        print(f"\nCleaning up {len(pool)} sandboxes ...")
        await disconnect_envs(envs)
        await destroy_sandbox_pool(pool)
        try:
            vllm_model.llm_engine.engine_core.shutdown()
        except Exception as e:
            print(f"  Warning during vLLM shutdown: {type(e).__name__}: {e}")
        print("Done.")


def main():
    parser = argparse.ArgumentParser(description="GRPO training for FinQA on Daytona sandboxes.")
    parser.add_argument(
        "--sandboxes",
        type=int,
        default=SANDBOX_COUNT,
        help=f"Number of concurrent sandboxes (default: {SANDBOX_COUNT}).",
    )
    parser.add_argument(
        "--iterations",
        type=int,
        default=TRAINING_ITERATIONS,
        help=f"Training iterations (default: {TRAINING_ITERATIONS}).",
    )
    parser.add_argument(
        "--group-size",
        type=int,
        default=EPISODES_PER_GROUP,
        help=f"Strict GRPO group size K (same-question episodes per group, default: {EPISODES_PER_GROUP}).",
    )
    parser.add_argument(
        "--target-groups-per-iter",
        type=int,
        default=TARGET_GROUPS_PER_ITER,
        help=(
            "Stop rollout collection once this many strict same-question groups "
            + "are formed for the update (default: "
            + f"{TARGET_GROUPS_PER_ITER}). Set <=0 to use all rollout rounds."
        ),
    )
    parser.add_argument(
        "--max-rollout-rounds",
        type=int,
        default=MAX_ROLLOUT_ROUNDS,
        help=(
            "Maximum rollout rounds per training iteration while forming "
            + f"strict same-question groups (default: {MAX_ROLLOUT_ROUNDS})."
        ),
    )
    parser.add_argument(
        "--snapshot",
        default="openenv-finqa",
        help="Snapshot name (default: openenv-finqa).",
    )
    parser.add_argument(
        "--model",
        default=MODEL_NAME,
        help=f"Model name (default: {MODEL_NAME}).",
    )
    parser.add_argument(
        "--lr",
        type=float,
        default=LEARNING_RATE,
        help=f"Learning rate (default: {LEARNING_RATE}).",
    )
    parser.add_argument(
        "--temperature",
        type=float,
        default=TEMPERATURE,
        help=f"Sampling temperature (default: {TEMPERATURE}).",
    )
    parser.add_argument(
        "--max-steps",
        type=int,
        default=MAX_EPISODE_STEPS,
        help=f"Max steps per episode (default: {MAX_EPISODE_STEPS}).",
    )
    parser.add_argument(
        "--max-gen-tokens",
        type=int,
        default=MAX_GEN_TOKENS,
        help=f"Max generation tokens (default: {MAX_GEN_TOKENS}).",
    )
    parser.add_argument(
        "--rollout-dispatch-wait-ms",
        type=int,
        default=ROLLOUT_DISPATCH_WAIT_MS,
        help=(
            "Max wait to accumulate ready episodes before generation "
            + f"dispatch (default: {ROLLOUT_DISPATCH_WAIT_MS} ms)."
        ),
    )
    parser.add_argument(
        "--gpu-memory-utilization",
        type=float,
        default=GPU_MEMORY_UTILIZATION,
        help=f"vLLM GPU memory utilization (default: {GPU_MEMORY_UTILIZATION}).",
    )
    parser.add_argument(
        "--tensor-parallel-size",
        type=int,
        default=TENSOR_PARALLEL_SIZE,
        help=f"vLLM tensor parallel size (default: {TENSOR_PARALLEL_SIZE}).",
    )
    parser.add_argument(
        "--lora-rank",
        type=int,
        default=LORA_RANK,
        help=f"LoRA rank (default: {LORA_RANK}).",
    )
    parser.add_argument(
        "--lora-alpha",
        type=int,
        default=LORA_ALPHA,
        help=f"LoRA alpha (default: {LORA_ALPHA}).",
    )
    parser.add_argument(
        "--lora-dropout",
        type=float,
        default=LORA_DROPOUT,
        help=f"LoRA dropout (default: {LORA_DROPOUT}).",
    )
    parser.add_argument(
        "--lora-target-modules",
        nargs="+",
        default=LORA_TARGET_MODULES,
        help="LoRA target modules.",
    )
    parser.add_argument(
        "--sync-every",
        type=int,
        default=SYNC_EVERY,
        help=f"Export/switch actor LoRA adapter every N training iterations (default: {SYNC_EVERY}).",
    )
    parser.add_argument(
        "--disable-gradient-checkpointing",
        action="store_true",
        help="Disable gradient checkpointing for faster training if memory allows.",
    )
    parser.add_argument(
        "--grpo-update-batch-size",
        type=int,
        default=GRPO_UPDATE_BATCH_SIZE,
        help=(
            "Turns per GRPO update micro-batch. Higher is faster but uses "
            + f"more memory (default: {GRPO_UPDATE_BATCH_SIZE})."
        ),
    )
    parser.add_argument(
        "--run-dir",
        default=None,
        help="Directory for persistent JSONL logs (default: runs/YYYYMMDD_HHMMSS/).",
    )
    parser.add_argument(
        "--save-token-ids",
        action="store_true",
        help="Include prompt/completion token ID lists in trajectory logs (large).",
    )
    args = parser.parse_args()
    if args.group_size <= 0:
        raise ValueError("--group-size must be >= 1.")
    if args.max_rollout_rounds <= 0:
        raise ValueError("--max-rollout-rounds must be >= 1.")
    if args.target_groups_per_iter < 0:
        raise ValueError("--target-groups-per-iter must be >= 0.")
    if args.rollout_dispatch_wait_ms < 0:
        raise ValueError("--rollout-dispatch-wait-ms must be >= 0.")
    if args.sync_every <= 0:
        raise ValueError("--sync-every must be >= 1.")
    if args.grpo_update_batch_size <= 0:
        raise ValueError("--grpo-update-batch-size must be >= 1.")
    asyncio.run(train(args))


if __name__ == "__main__":
    main()
