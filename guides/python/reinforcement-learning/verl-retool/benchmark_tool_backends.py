"""Benchmark veRL ReTool tool backends: Daytona, Docker, and SandboxFusion."""

from __future__ import annotations

import argparse
import asyncio
import json
import os
import platform
import statistics
import sys
from dataclasses import dataclass
from datetime import UTC, datetime
from pathlib import Path
from time import perf_counter
from typing import Any

DEFAULT_CONCURRENCY = [1, 4, 8, 16, 32, 64, 128]
DEFAULT_WARMUPS = 3
DEFAULT_ITERATIONS = 20


@dataclass(frozen=True)
class Scenario:
    """Describe one benchmark scenario."""

    name: str
    label: str
    code: str
    expected_substring: str | None
    expects_error: bool = False


SCENARIOS = [
    Scenario(
        name="simple_stdout",
        label="Executing print(2 + 2) - minimal round-trip latency",
        code="print(2 + 2)",
        expected_substring="4",
    ),
    Scenario(
        name="cpu_bound_stdout",
        label="Executing sum(i*i for 20k iterations) - CPU-bound compute",
        code="total = 0\nfor i in range(20000):\n    total += i * i\nprint(total)\n",
        expected_substring="2666466670000",
    ),
    Scenario(
        name="runtime_error",
        label="Executing raise ValueError - error propagation",
        code="raise ValueError('boom from benchmark')",
        expected_substring="ValueError",
        expects_error=True,
    ),
]


def percentile(values: list[float], pct: float) -> float:
    """Return a simple percentile for a non-empty sample."""
    if not values:
        raise ValueError("percentile() requires at least one value")

    ordered = sorted(values)
    index = round((len(ordered) - 1) * pct)
    return ordered[index]


def resolve_verl_root(verl_root: str | None) -> Path | None:
    """Add a local veRL checkout to `sys.path` when provided."""
    root_value = verl_root or os.getenv("VERL_ROOT")
    if root_value is None:
        return None

    root = Path(root_value).expanduser().resolve()
    if not root.exists():
        raise SystemExit(f"--verl-root path does not exist: {root}")

    if str(root) not in sys.path:
        sys.path.insert(0, str(root))

    return root


def build_code_interpreter_schema():
    """Return the code interpreter schema used by the benchmarked tools."""
    from verl.tools.schemas import (
        OpenAIFunctionParametersSchema,
        OpenAIFunctionPropertySchema,
        OpenAIFunctionSchema,
        OpenAIFunctionToolSchema,
    )

    return OpenAIFunctionToolSchema(
        type="function",
        function=OpenAIFunctionSchema(
            name="code_interpreter",
            description="A tool for executing Python code.",
            parameters=OpenAIFunctionParametersSchema(
                type="object",
                properties={
                    "code": OpenAIFunctionPropertySchema(
                        type="string",
                        description="The Python code to execute.",
                    )
                },
                required=["code"],
            ),
        ),
    )


def build_daytona_config(args: argparse.Namespace) -> dict[str, Any]:
    """Build the Daytona tool config for the benchmark."""
    config = {
        "type": "native",
        "rate_limit": args.rate_limit or max(args.concurrency),
        "enable_global_rate_limit": True,
        "create_timeout": args.create_timeout,
        "default_timeout": args.default_timeout,
        "delete_timeout": args.delete_timeout,
        "auto_stop_interval": args.auto_stop_interval,
        "auto_delete_interval": args.auto_delete_interval,
        "name_prefix": "verl-daytona-bench",
        "language": "python",
    }

    for key, value in {
        "snapshot": args.daytona_snapshot,
        "api_url": args.daytona_api_url,
        "target": args.daytona_target,
        "organization_id": args.daytona_organization_id,
    }.items():
        if value is not None:
            config[key] = value

    return config


def build_sandboxfusion_config(args: argparse.Namespace) -> dict[str, Any]:
    """Build the SandboxFusion tool config for the benchmark."""
    if args.sandbox_fusion_url is None:
        raise SystemExit("--sandbox-fusion-url is required when --backend=sandboxfusion")

    return {
        "type": "native",
        "sandbox_fusion_url": args.sandbox_fusion_url,
        "num_workers": args.num_workers or max(args.concurrency),
        "enable_global_rate_limit": True,
        "rate_limit": args.rate_limit or max(args.concurrency),
        "default_timeout": args.default_timeout,
        "default_language": "python",
        "memory_limit_mb": args.memory_limit_mb,
    }


def check_backend_prereqs(args: argparse.Namespace) -> None:
    """Fail fast when required backend configuration is missing."""
    if args.backend == "daytona":
        if not os.environ.get("DAYTONA_API_KEY") and not os.environ.get("DAYTONA_JWT_TOKEN"):
            raise SystemExit(
                "DAYTONA_API_KEY (or DAYTONA_JWT_TOKEN) is not set. Export it before running the Daytona benchmark."
            )
    elif args.backend == "docker":
        import shutil

        if shutil.which("docker") is None:
            raise SystemExit("Docker is not installed or not on PATH.")


class DockerContainerTool:
    """Minimal tool that runs code in a fresh Docker container per execution.

    Implements the same create/execute/release interface as veRL's BaseTool
    so the benchmark harness can treat all backends uniformly.
    """

    def __init__(self, config: dict[str, Any]):
        self._image = config.get("image", "python:3.11-slim")
        self._memory = config.get("memory", "256m")
        self._timeout = config.get("default_timeout", 30)
        self._instances: dict[str, bool] = {}

    async def create(self, instance_id: str | None = None) -> tuple[str, None]:
        iid = instance_id or f"docker-{id(self)}-{len(self._instances)}"
        self._instances[iid] = True
        return iid, None

    async def execute(self, instance_id: str, parameters: dict[str, Any], **kwargs) -> tuple[Any, float, dict]:
        code = parameters.get("code", "")
        timeout = parameters.get("timeout", self._timeout)

        proc = await asyncio.create_subprocess_exec(
            "docker", "run", "--rm", "--network=none",
            f"--memory={self._memory}", "--cpus=1",
            self._image, "python3", "-c", code,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
        )

        try:
            stdout, stderr = await asyncio.wait_for(proc.communicate(), timeout=timeout)
        except asyncio.TimeoutError:
            proc.kill()
            await proc.wait()

            class _Response:
                def __init__(self, text):
                    self.text = text

            return _Response("TimeoutError: execution exceeded timeout"), 0.0, {"had_error": True}

        output = stdout.decode() + stderr.decode()
        had_error = proc.returncode != 0

        class _Response:
            def __init__(self, text):
                self.text = text

        return _Response(output), 0.0, {"had_error": had_error}

    async def release(self, instance_id: str, **kwargs) -> None:
        self._instances.pop(instance_id, None)

    async def close(self) -> None:
        pass


def make_tool(args: argparse.Namespace):
    """Instantiate the requested backend from the local veRL checkout."""
    if args.backend == "docker":
        return DockerContainerTool({
            "image": args.docker_image,
            "memory": args.docker_memory,
            "default_timeout": args.default_timeout,
        })

    resolve_verl_root(args.verl_root)
    schema = build_code_interpreter_schema()

    # Import veRL modules only after the optional checkout path is on sys.path.
    try:
        if args.backend == "daytona":
            from recipe.retool.daytona_sandbox_tool import CustomDaytonaSandboxTool

            return CustomDaytonaSandboxTool(build_daytona_config(args), schema)

        from recipe.retool.retool import CustomSandboxFusionTool

        return CustomSandboxFusionTool(build_sandboxfusion_config(args), schema)
    except ImportError as exc:
        raise SystemExit(
            "Failed to import veRL ReTool modules. Install veRL in the active environment or pass --verl-root."
        ) from exc


def ensure_output_dir(output_root: Path, backend: str) -> Path:
    """Create a timestamped output directory for one backend run."""
    timestamp = datetime.now(UTC).strftime("%Y%m%dT%H%M%SZ")
    output_dir = output_root / backend / timestamp
    output_dir.mkdir(parents=True, exist_ok=True)
    return output_dir


async def _create_one(tool) -> tuple[str, float]:
    """Create one sandbox and return the instance id plus elapsed time."""
    started_at = perf_counter()
    instance_id, _ = await tool.create()
    return instance_id, perf_counter() - started_at


async def measure_setup(tool, count: int) -> dict[str, Any]:
    """Create `count` sandboxes in parallel and collect setup timings."""
    total_started_at = perf_counter()
    results = await asyncio.gather(*[_create_one(tool) for _ in range(count)], return_exceptions=True)
    total_wall = perf_counter() - total_started_at

    succeeded = [(iid, elapsed) for result in results if not isinstance(result, BaseException) for iid, elapsed in [result]]
    failures = [result for result in results if isinstance(result, BaseException)]

    if failures:
        created_ids = [instance_id for instance_id, _ in succeeded]
        if created_ids:
            await force_cleanup(tool, created_ids)
        raise failures[0]

    instance_ids = [instance_id for instance_id, _ in succeeded]
    create_times = [elapsed for _, elapsed in succeeded]

    return {
        "sandbox_count": count,
        "create_times_s": create_times,
        "p50_create_s": percentile(create_times, 0.50),
        "p95_create_s": percentile(create_times, 0.95),
        "max_create_s": max(create_times),
        "mean_create_s": statistics.fmean(create_times),
        "total_wall_s": total_wall,
        "instance_ids": instance_ids,
    }


async def run_single_execution(tool, instance_id: str, scenario: Scenario, timeout: int) -> dict[str, Any]:
    """Execute one snippet on one sandbox and capture latency plus outcome."""
    started_at = perf_counter()

    try:
        response, _, metrics = await tool.execute(instance_id, {"code": scenario.code, "timeout": timeout})
    except Exception as exc:
        return {
            "latency_s": perf_counter() - started_at,
            "success": False,
            "error": str(exc),
            "response_text": "",
            "metrics": {},
        }

    response_text = response.text if hasattr(response, "text") else str(response)
    metrics = metrics or {}
    had_error = bool(metrics.get("had_error"))
    finished_at = perf_counter()

    if scenario.expects_error:
        success = scenario.expected_substring in response_text
        if "had_error" in metrics:
            success = success and had_error
    else:
        success = scenario.expected_substring in response_text
        if "had_error" in metrics:
            success = success and not had_error

    return {
        "latency_s": finished_at - started_at,
        "success": success,
        "error": None,
        "response_text": response_text,
        "metrics": metrics,
    }


async def measure_execution(
    tool,
    instance_ids: list[str],
    scenario: Scenario,
    concurrency: int,
    warmups: int,
    iterations: int,
    timeout: int,
) -> dict[str, Any]:
    """Measure execution latency across an existing sandbox pool."""
    pool = instance_ids[:concurrency]
    measured_calls = []
    total_measured_time = 0.0

    for phase_name, batch_count in (("warmup", warmups), ("measured", iterations)):
        for _ in range(batch_count):
            batch_started_at = perf_counter()

            # Spread one request across one sandbox when measuring concurrency.
            tasks = [run_single_execution(tool, pool[i % len(pool)], scenario, timeout) for i in range(concurrency)]
            results = await asyncio.gather(*tasks)
            batch_elapsed = perf_counter() - batch_started_at

            if phase_name == "measured":
                total_measured_time += batch_elapsed
                measured_calls.extend(results)

    latencies = [call["latency_s"] for call in measured_calls]
    successes = sum(1 for call in measured_calls if call["success"])

    return {
        "scenario": scenario.name,
        "concurrency": concurrency,
        "measured_call_count": len(measured_calls),
        "success_count": successes,
        "failure_count": len(measured_calls) - successes,
        "p50_latency_s": percentile(latencies, 0.50),
        "p95_latency_s": percentile(latencies, 0.95),
        "max_latency_s": max(latencies),
        "mean_latency_s": statistics.fmean(latencies),
        "throughput_calls_per_s": len(measured_calls) / total_measured_time if total_measured_time else 0.0,
    }


async def _release_one(tool, instance_id: str) -> dict[str, Any]:
    """Release one sandbox and record the result."""
    started_at = perf_counter()

    try:
        await tool.release(instance_id)
    except Exception as exc:
        return {
            "instance_id": instance_id,
            "elapsed_s": perf_counter() - started_at,
            "error": str(exc),
        }

    return {
        "instance_id": instance_id,
        "elapsed_s": perf_counter() - started_at,
        "error": None,
    }


async def force_cleanup(tool, instance_ids: list[str]) -> None:
    """Best-effort cleanup used after partial setup failures."""
    await asyncio.gather(*[_release_one(tool, instance_id) for instance_id in instance_ids], return_exceptions=True)


async def measure_teardown(tool, instance_ids: list[str]) -> dict[str, Any]:
    """Release all sandboxes in parallel and collect teardown timings."""
    total_started_at = perf_counter()
    release_results = await asyncio.gather(*[_release_one(tool, instance_id) for instance_id in instance_ids])
    total_wall = perf_counter() - total_started_at

    failures = [result for result in release_results if result["error"] is not None]
    if failures:
        summary = ", ".join(f"{item['instance_id']}: {item['error']}" for item in failures[:3])
        raise RuntimeError(f"Failed to release {len(failures)} sandboxes: {summary}")

    release_times = [result["elapsed_s"] for result in release_results]

    return {
        "sandbox_count": len(instance_ids),
        "release_times_s": release_times,
        "mean_release_s": statistics.fmean(release_times) if release_times else 0.0,
        "total_wall_s": total_wall,
    }


def build_terminal_summary(summary: dict[str, Any]) -> str:
    """Render a fixed-width terminal summary for the measured rows."""
    headers = ["Scenario", "Conc", "p50 (s)", "p95 (s)", "Thru (c/s)", "OK", "Fail"]
    widths = [16, 5, 8, 8, 10, 6, 4]
    ok_rows = [row for row in summary["results"] if row.get("status") == "ok"]

    if not ok_rows:
        return ""

    def fmt_row(values: list[str]) -> str:
        return "  ".join(value.rjust(width) for value, width in zip(values, widths, strict=False))

    lines = ["", fmt_row(headers), "  ".join("-" * width for width in widths)]

    for row in ok_rows:
        lines.append(
            fmt_row(
                [
                    row["scenario"],
                    str(row["concurrency"]),
                    f"{row['p50_latency_s']:.4f}",
                    f"{row['p95_latency_s']:.4f}",
                    f"{row['throughput_calls_per_s']:.2f}",
                    str(row["success_count"]),
                    str(row["failure_count"]),
                ]
            )
        )

    return "\n".join(lines) + "\n"


def write_json(path: Path, payload: Any) -> None:
    """Write stable JSON output."""
    path.write_text(json.dumps(payload, indent=2, sort_keys=True, default=str) + "\n")


def write_csv(path: Path, result_rows: list[dict[str, Any]]) -> None:
    """Write the flat benchmark results CSV."""
    import csv

    fieldnames = [
        "scenario",
        "concurrency",
        "p50_s",
        "p95_s",
        "max_s",
        "mean_s",
        "throughput_calls_per_s",
        "success_count",
        "failure_count",
    ]

    with path.open("w", newline="", encoding="utf-8") as handle:
        writer = csv.DictWriter(handle, fieldnames=fieldnames)
        writer.writeheader()

        for row in result_rows:
            if row.get("status") != "ok":
                continue

            writer.writerow(
                {
                    "scenario": row["scenario"],
                    "concurrency": row["concurrency"],
                    "p50_s": f"{row['p50_latency_s']:.4f}",
                    "p95_s": f"{row['p95_latency_s']:.4f}",
                    "max_s": f"{row['max_latency_s']:.4f}",
                    "mean_s": f"{row['mean_latency_s']:.4f}",
                    "throughput_calls_per_s": f"{row['throughput_calls_per_s']:.2f}",
                    "success_count": row["success_count"],
                    "failure_count": row["failure_count"],
                }
            )


async def maybe_close_tool(tool) -> None:
    """Close the tool when it exposes an async close hook."""
    close_method = getattr(tool, "close", None)
    if close_method is None:
        return

    result = close_method()
    if asyncio.iscoroutine(result):
        await result


async def run_benchmarks(args: argparse.Namespace) -> tuple[list[dict[str, Any]], dict[str, Any] | None]:
    """Run setup, execution, and teardown for the selected backend."""
    tool = make_tool(args)
    max_concurrency = max(args.concurrency)
    result_rows: list[dict[str, Any]] = []
    setup_result: dict[str, Any] | None = None
    setup: dict[str, Any] | None = None
    instance_ids: list[str] = []

    try:
        print(f"\n{'=' * 60}")
        print(f"  SETUP - Creating {max_concurrency} sandboxes with {args.backend}")
        print(f"{'=' * 60}")

        setup = await measure_setup(tool, max_concurrency)
        instance_ids = setup.pop("instance_ids")
        print(
            f"  {max_concurrency} sandboxes created in {setup['total_wall_s']:.2f}s "
            f"(p50={setup['p50_create_s']:.3f}s)"
        )

        for scenario in SCENARIOS:
            print(f"\n{'=' * 60}")
            print(f"  {scenario.label}")
            print(f"{'=' * 60}")

            for concurrency in args.concurrency:
                print(f"  concurrency={concurrency}...", end=" ", flush=True)

                result = await measure_execution(
                    tool,
                    instance_ids,
                    scenario,
                    concurrency,
                    args.warmups,
                    args.iterations,
                    args.default_timeout,
                )
                result["status"] = "ok"
                result_rows.append(result)

                ok_count = result["success_count"]
                fail_count = result["failure_count"]
                print(
                    f"p50={result['p50_latency_s']:.3f}s  "
                    f"throughput={result['throughput_calls_per_s']:.1f}/s  "
                    f"{ok_count}/{ok_count + fail_count} ok"
                )
    finally:
        if instance_ids:
            print(f"\n{'=' * 60}")
            print(f"  TEARDOWN - Releasing {len(instance_ids)} sandboxes")
            print(f"{'=' * 60}")

            teardown = await measure_teardown(tool, instance_ids)
            print(f"  {len(instance_ids)} sandboxes released in {teardown['total_wall_s']:.2f}s")
            setup_result = {"setup": setup, "teardown": teardown}

        await maybe_close_tool(tool)

    return result_rows, setup_result


def parse_args() -> argparse.Namespace:
    """Parse the benchmark CLI arguments."""
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--backend", choices=["daytona", "docker", "sandboxfusion"], default="daytona")
    parser.add_argument("--verl-root", default=None, help="Path to a local veRL checkout if it is not installed.")
    parser.add_argument("--output-root", default="outputs")
    parser.add_argument(
        "--concurrency",
        nargs="+",
        type=int,
        default=DEFAULT_CONCURRENCY,
        help="Concurrent tool calls per batch. Sandboxes are pre-created for the max level.",
    )
    parser.add_argument("--warmups", type=int, default=DEFAULT_WARMUPS)
    parser.add_argument("--iterations", type=int, default=DEFAULT_ITERATIONS)
    parser.add_argument("--default-timeout", type=int, default=30)
    parser.add_argument("--create-timeout", type=int, default=60)
    parser.add_argument("--delete-timeout", type=int, default=60)
    parser.add_argument("--auto-stop-interval", type=int, default=15)
    parser.add_argument("--auto-delete-interval", type=int, default=30)
    parser.add_argument("--rate-limit", type=int, default=None)
    parser.add_argument("--num-workers", type=int, default=None)
    parser.add_argument("--memory-limit-mb", type=int, default=1024)
    parser.add_argument("--sandbox-fusion-url", default=None)
    parser.add_argument("--daytona-api-url", default=None)
    parser.add_argument("--daytona-target", default=None)
    parser.add_argument("--daytona-organization-id", default=None)
    parser.add_argument("--daytona-snapshot", default=None)
    parser.add_argument("--docker-image", default="python:3.11-slim")
    parser.add_argument("--docker-memory", default="256m")
    return parser.parse_args()


def validate_args(args: argparse.Namespace) -> None:
    """Validate numeric arguments are positive."""
    for c in args.concurrency:
        if c < 1:
            raise SystemExit(f"--concurrency values must be >= 1, got {c}")
    if args.warmups < 0:
        raise SystemExit(f"--warmups must be >= 0, got {args.warmups}")
    if args.iterations < 1:
        raise SystemExit(f"--iterations must be >= 1, got {args.iterations}")
    if args.default_timeout < 1:
        raise SystemExit(f"--default-timeout must be >= 1, got {args.default_timeout}")


def main() -> int:
    """Run the benchmark and save the JSON plus CSV artifacts."""
    args = parse_args()
    validate_args(args)
    check_backend_prereqs(args)

    output_dir = ensure_output_dir(Path(args.output_root), args.backend)
    result_rows, setup_result = asyncio.run(run_benchmarks(args))

    summary = {
        "backend": args.backend,
        "created_at": datetime.now(UTC).isoformat(),
        "host": platform.platform(),
        "python_version": platform.python_version(),
        "results": result_rows,
        "setup_result": setup_result,
    }

    write_json(output_dir / "summary.json", summary)
    write_csv(output_dir / "results.csv", result_rows)

    print(f"\nSaved benchmark artifacts to {output_dir}")
    print(build_terminal_summary(summary))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
