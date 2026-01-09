# Copyright 2026 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import asyncio
import json
import os
from typing import Any, Awaitable, Dict, List, TypedDict

from datasets import Dataset
from dotenv import load_dotenv
from trl import GRPOConfig, GRPOTrainer

from daytona import AsyncDaytona, AsyncSandbox
from daytona.common.errors import DaytonaTimeoutError

load_dotenv()

EFFECTIVE_BATCH_SIZE = 500
# We evaluate each completion concurrently, in its own sandbox,
# so we spawn EFFECTIVE_BATCH_SIZE number of sandboxes.
MAX_TIMEOUT_SECONDS = 1
MODEL_NAME = "Qwen/Qwen3-1.7B-Base"

SORTING_PROMPT = """# I've been fiddling with different ways to sort numbers in Python.
# At first I just used sorted() and list.sort(), but then I decided to try
# my hand at writing some original sorting functions. And I succeeded!
# I don't call sorted(), list.sort(), heapq, or use any imports here - just plain
# Python and an original algorithm.
def sort_numbers(xs: list[int]) -> list[int]:
    \"\"\"Sort a list of integers in ascending order.

    Args:
        xs: A list of integers to be sorted.

    Returns:
        A new list containing the same integers, sorted from smallest to largest.
    \"\"\"
"""

MAX_SUBARRAY_PROMPT = """# I've been exploring different ways to compute the maximum sum of a contiguous
# subarray in Python. At first I wrote a straightforward brute-force version
# with nested loops, but now I'm trying to come up with my own cleaner
# implementation. There are lots of possible approaches here, and this function
# is just my original take on the problem.
def max_subarray_sum(xs: list[int]) -> int:
    \"\"\"Return the maximum sum of a non-empty contiguous subarray.

    Args:
        xs: A non-empty list of integers.

    Returns:
        The largest possible sum of any contiguous subarray of xs.
    \"\"\"
"""

TASKS = {
    "sorting": {
        "prompt": SORTING_PROMPT,
        "func_name": "sort_numbers",
        "banned_patterns": ["sorted(", ".sort(", "heapq", "import ", "__import__"],
        "tests": [
            "[]",
            "[1, 3, 2]",
            "[random.randint(-1000, 1000) for _ in range(200)]",
            "[random.randint(-100, 100) for _ in range(1000)]",
            "list(range(0, 100)) + list(range(200, 100, -1)) + list(range(200, 300))",
        ],
        "reference": "sorted",
    },
    "max_subarray": {
        "prompt": MAX_SUBARRAY_PROMPT,
        "func_name": "max_subarray_sum",
        "banned_patterns": [],
        "tests": [
            "[5]",
            "[-3]",
            "[-2, -3, -1, -4]",
            "[-2, 1, -3, 4, -1, 2, 1, -5, 4]",
            "[1, 2, 3, 4]",
            "[random.randint(-1000, 1000) for _ in range(200)]",
            "[random.randint(-100, 100) for _ in range(1000)]",
        ],
        "reference": "_kadane",
    },
}

PROMPT_TO_TASK = {task["prompt"]: task for task in TASKS.values()}


async def _create_sandbox_pool_async(daytona: AsyncDaytona, n: int = 10) -> List[AsyncSandbox]:
    print(f"Creating {n} sandboxes...")
    tasks = [daytona.create() for _ in range(n)]
    sandboxes = await asyncio.gather(*tasks)
    print(f"Successfully created all {len(sandboxes)} sandboxes")
    return list(sandboxes)


async def _cleanup_sandbox_pool_async(sandbox_pool: List[AsyncSandbox]) -> None:
    if not sandbox_pool:
        return
    print("Cleaning up sandboxes...")
    tasks = [sandbox.delete() for sandbox in sandbox_pool]
    results = await asyncio.gather(*tasks, return_exceptions=True)
    for r in results:
        if isinstance(r, Exception):
            print(f"  Sandbox delete error: {type(r).__name__}: {r}")
    print("All sandboxes cleaned up")


class EvalResult(TypedDict):
    no_error: bool
    num_passed: int
    num_tests: int


def _fail_result(num_tests: int) -> EvalResult:
    return {"no_error": False, "num_passed": 0, "num_tests": num_tests}


def sanitize_completion(text: str) -> str:
    # Since the model continues the body of the function,
    # we take lines until the first unindented line.
    lines = text.splitlines()
    kept: List[str] = []
    for line in lines:
        if line and (not line.startswith("    ")):
            break
        kept.append(line)
    return "\n".join(kept).rstrip()


def has_banned_pattern(text: str, task: Dict[str, Any]) -> bool:
    banned = task.get("banned_patterns", [])
    if not banned:
        return False
    lowered = text.lower()
    return any(p.lower() in lowered for p in banned)


def build_test_harness(task: Dict[str, Any], function_body: str) -> str:
    prompt = task["prompt"]
    func_name = task["func_name"]
    reference_function = task["reference"]
    tests = task["tests"]

    tests_tuple = ",\n        ".join(tests)

    return f"""{prompt}
{function_body}

import json
import random
random.seed(0)

def _kadane(xs):
    max_sum = current = xs[0]
    for x in xs[1:]:
        current = max(x, current + x)
        max_sum = max(max_sum, current)
    return max_sum

def _run_tests():
    tests = (
        {tests_tuple}
    )
    results = []
    for xs in tests:
        try:
            out = {func_name}(xs.copy())
            expected = {reference_function}(xs.copy())
            results.append(out == expected)
        except Exception:
            results.append(False)
    print(json.dumps({{"results": results}}))

if __name__ == "__main__":
    _run_tests()
"""


async def evaluate_single_completion_async(
    sandbox: AsyncSandbox,
    raw_completion: str,
    prompt: str,
) -> EvalResult:
    task = PROMPT_TO_TASK[prompt]
    num_task_tests = len(task["tests"])
    body = sanitize_completion(raw_completion)

    if not body.strip():
        return _fail_result(num_task_tests)
    if has_banned_pattern(body, task):
        return _fail_result(num_task_tests)

    code = build_test_harness(task, body)

    try:
        response = await sandbox.code_interpreter.run_code(code, timeout=MAX_TIMEOUT_SECONDS)
    except DaytonaTimeoutError:
        print(f"Completion timed out after {MAX_TIMEOUT_SECONDS}s " f"in sandbox {getattr(sandbox, 'id', '?')}")
        return _fail_result(num_task_tests)
    except Exception as e:
        print(
            f"Error evaluating completion in sandbox {getattr(sandbox, 'id', '?')}: " f"{type(e).__name__}: {e}",
        )
        return _fail_result(num_task_tests)

    if response.error is not None:
        return _fail_result(num_task_tests)
    raw_output = response.stdout.strip()
    if not raw_output:
        return _fail_result(num_task_tests)
    last_line = raw_output.splitlines()[-1]
    try:
        results = json.loads(last_line)
    except Exception:
        return _fail_result(num_task_tests)
    correct = results.get("results", [])

    return {
        "no_error": True,
        "num_passed": sum(bool(x) for x in correct),
        "num_tests": len(correct),
    }


async def _evaluate_batch_async(
    sandbox_pool: List[AsyncSandbox], completions: List[str], prompts: List[str]
) -> List[EvalResult]:
    print(f"Evaluating {len(completions)} completions in parallel across " f"{len(sandbox_pool)} sandboxes...")

    async def run_one(i: int, sandbox: AsyncSandbox, completion: str, prompt: str) -> EvalResult:
        task = PROMPT_TO_TASK[prompt]
        num_task_tests = len(task["tests"])
        try:
            stats = await evaluate_single_completion_async(sandbox, completion, prompt)
            print(f"  Completion {i + 1}/{len(completions)} done")
            return stats
        except Exception as e:
            print(f"  Completion {i + 1}/{len(completions)} failed: " f"{type(e).__name__}: {e}")
            return _fail_result(num_task_tests)

    tasks = [
        run_one(i, sandbox_pool[i % len(sandbox_pool)], completion, prompt)
        for i, (completion, prompt) in enumerate(zip(completions, prompts))
    ]

    stats_list = await asyncio.gather(*tasks)
    print(f"  Done: {len(completions)}/{len(completions)} completions evaluated")

    return stats_list


def main():
    # Create local event loop for mixing sync training library with async sandbox API
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)

    def run_async(coro: Awaitable[Any]) -> Any:
        """Helper to run async code from sync context (e.g., reward functions)."""
        return loop.run_until_complete(coro)

    daytona = AsyncDaytona()

    sandbox_pool: List[AsyncSandbox] = []

    training_args = GRPOConfig(
        output_dir="training_results",
        per_device_train_batch_size=20,
        # batch size chosen so the training runs comfortably on a single 80GB GPU,
        # if running this on a GPU with less memory, reduce the batch size accordingly
        gradient_accumulation_steps=25,
        num_generations=EFFECTIVE_BATCH_SIZE // len(TASKS),
        max_prompt_length=256,
        max_completion_length=512,
        learning_rate=8e-6,
        num_train_epochs=1,
        logging_steps=1,
        report_to="none",
        max_steps=8,
        bf16=True,
        use_vllm=True,
        vllm_mode="colocate",
        vllm_gpu_memory_utilization=0.15,
        gradient_checkpointing=True,
        loss_type="dapo",
        beta=0.01,
    )
    assert EFFECTIVE_BATCH_SIZE % len(TASKS) == 0, "EFFECTIVE_BATCH_SIZE must be divisible by number of tasks."
    assert (
        training_args.per_device_train_batch_size * training_args.gradient_accumulation_steps
    ) == EFFECTIVE_BATCH_SIZE, "The total batch size must equal the sandbox pool size."

    try:
        sandbox_pool = run_async(_create_sandbox_pool_async(daytona, n=EFFECTIVE_BATCH_SIZE))

        train_dataset = Dataset.from_dict({"prompt": [task["prompt"] for task in TASKS.values()]})

        def reward_func(prompts, completions, **_kwargs):
            stats_list = run_async(_evaluate_batch_async(sandbox_pool, completions, prompts))
            rewards = []
            for s in stats_list:
                if not s["no_error"]:
                    rewards.append(-1.0)
                elif s["num_tests"] == 0:
                    rewards.append(0.0)
                else:
                    rewards.append(s["num_passed"] / s["num_tests"])
            return rewards

        trainer = GRPOTrainer(
            model=MODEL_NAME,
            args=training_args,
            train_dataset=train_dataset,
            reward_funcs=[reward_func],
        )

        trainer.train()

        os.makedirs(training_args.output_dir, exist_ok=True)
        log_path = os.path.join(training_args.output_dir, "metrics.jsonl")
        with open(log_path, "w") as f:
            for rec in trainer.state.log_history:
                f.write(json.dumps(rec) + "\n")
        print("wrote logs to", log_path)

    finally:
        if sandbox_pool:
            run_async(_cleanup_sandbox_pool_async(sandbox_pool))

        try:
            run_async(daytona.close())
            print("Daytona client closed")
        except Exception as e:
            print(f"Error closing Daytona client: {type(e).__name__}: {e}")

        loop.close()


if __name__ == "__main__":
    main()
