# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""Simple console output for deeper-rlm."""

from rlm.types import AgentResult, Iteration


class ConsoleOutput:
    """Console output for real-time progress display."""

    def __init__(self, verbose: bool = True):
        self.verbose = verbose

    def iteration(self, agent_id: str, depth: int, iteration: Iteration) -> None:
        """Display an iteration's details."""
        if not self.verbose:
            return

        prefix = "  " * depth
        print(f"{prefix}Agent {agent_id} - Iteration {iteration.iteration}")

        for block in iteration.parsed_code_blocks:
            print(f"{prefix}Code:")
            # Indent code block
            code_preview = block.code[:500]
            for line in code_preview.split("\n"):
                print(f"{prefix}  {line}")

            if block.stdout:
                truncated = block.stdout[:300] + "..." if len(block.stdout) > 300 else block.stdout
                print(f"{prefix}Output: {truncated}")

            if block.error:
                print(f"{prefix}Error: {block.error[:200]}")

    def agent_spawn(self, _parent_id: str, child_id: str, task: str, depth: int) -> None:
        """Display sub-agent spawn."""
        if not self.verbose:
            return

        prefix = "  " * depth
        truncated_task = task[:80] + "..." if len(task) > 80 else task
        print(f"{prefix}Spawning sub-agent {child_id}: {truncated_task}")

    def agent_complete(self, agent_id: str, depth: int, result: str | None) -> None:
        """Display agent completion."""
        if not self.verbose:
            return

        prefix = "  " * depth
        if result:
            truncated = result[:100] + "..." if len(result) > 100 else result
            print(f"{prefix}Agent {agent_id} complete: {truncated}")
        else:
            print(f"{prefix}Agent {agent_id} complete (no result)")

    def tree_view(self, agent: AgentResult) -> None:
        """Display agent tree structure."""
        print("\nAgent Tree:")
        self._print_tree(agent, prefix="")

    def _print_tree(self, agent: AgentResult, prefix: str, connector: str = "") -> None:
        """Recursively print tree structure."""
        task_preview = ""
        if agent.task:
            task_preview = f": {agent.task[:40]}..." if len(agent.task) > 40 else f": {agent.task}"
        iters = len(agent.iterations)
        print(f"{prefix}{connector}{agent.agent_id} (depth={agent.depth}, iters={iters}){task_preview}")

        # Calculate the base prefix for children based on our connector
        if connector == "└─":
            child_base = prefix + "  "  # Last child, use spaces
        elif connector == "├─":
            child_base = prefix + "| "  # More siblings, continue line
        else:
            child_base = prefix  # Root node

        for i, sub in enumerate(agent.spawned_agents):
            is_last = i == len(agent.spawned_agents) - 1
            child_connector = "└─" if is_last else "├─"
            self._print_tree(sub, child_base, child_connector)
