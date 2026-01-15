"""Tree-structured JSON logger for deeper-rlm."""

import json
from pathlib import Path
from typing import Any

from rlm.types import AgentResult


class TreeLogger:
    """
    Logger that saves results in nested JSON tree format.

    Creates a detail file per run with full agent tree and all iterations.
    """

    def __init__(self, output_dir: str | Path):
        """
        Initialize the logger.

        Args:
            output_dir: Directory for output files
        """
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(parents=True, exist_ok=True)

    def log_agent_result(self, agent: AgentResult, run_id: str) -> Path:
        """
        Log an agent result.

        Args:
            agent: Agent result to log
            run_id: Unique run identifier

        Returns:
            Path to the detail file
        """
        # Write detail
        detail_path = self.output_dir / f"{run_id}.detail.json"
        detail = {
            "run_id": run_id,
            "root_agent": self._serialize_agent(agent),
        }
        with open(detail_path, "w") as f:
            json.dump(detail, f, indent=2, default=str)

        # Update index.json for viewer
        self._update_index()

        return detail_path

    def _update_index(self) -> None:
        """Update index.json with list of available detail files."""
        files = sorted(
            [f.name for f in self.output_dir.glob("*.detail.json")],
            key=lambda f: self.output_dir.joinpath(f).stat().st_mtime,
            reverse=True,
        )
        index_path = self.output_dir / "index.json"
        with open(index_path, "w") as f:
            json.dump(files, f, indent=2)

    def _serialize_agent(self, agent: AgentResult) -> dict[str, Any]:
        """Serialize an agent result to dict."""
        return {
            "agent_id": agent.agent_id,
            "depth": agent.depth,
            "sandbox_id": agent.sandbox_id,
            "task": agent.task,
            "iterations": [
                {
                    "iteration": it.iteration,
                    "prompt": it.prompt,
                    "raw_response": it.raw_response,
                    "parsed_code_blocks": [
                        {
                            "code": block.code,
                            "stdout": block.stdout,
                            "stderr": block.stderr,
                            "execution_time": block.execution_time,
                            "error": block.error,
                        }
                        for block in it.parsed_code_blocks
                    ],
                    "spawned_agents": [self._serialize_agent(sub) for sub in it.spawned_agents],
                }
                for it in agent.iterations
            ],
            "spawned_agents": [self._serialize_agent(sub) for sub in agent.spawned_agents],
            "result": agent.result,
            "result_truncated": agent.result_truncated,
            "usage": {
                "input_tokens": agent.usage.input_tokens,
                "output_tokens": agent.usage.output_tokens,
                "cost": agent.usage.cost,
            },
            "execution_time": agent.execution_time,
            "error": agent.error,
        }
