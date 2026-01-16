# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""Core RLM agent implementation."""

import logging
import threading
import time
import uuid
from concurrent.futures import ThreadPoolExecutor, as_completed
from typing import TYPE_CHECKING, Callable

from rlm.client import BaseLMClient
from rlm.prompts import build_system_prompt, build_user_prompt, format_execution_result
from rlm.repl import DaytonaREPL
from rlm.sandbox import SandboxExecutor, SandboxManager
from rlm.types import (
    AgentResult,
    Config,
    Iteration,
    UsageStats,
)

if TYPE_CHECKING:
    from daytona import Sandbox

    from output_logging.console import ConsoleOutput


logger = logging.getLogger(__name__)


class RLMAgent:
    """
    Recursive Language Model agent.

    Executes in a Daytona sandbox, can spawn sub-agents via rlm_query().
    """

    def __init__(
        self,
        client: BaseLMClient,
        sandbox_manager: SandboxManager,
        config: Config,
        problem_statement: str,
        instance_id: str,
        repo_url: str,
        depth: int = 0,
        task: str | None = None,
        on_iteration: Callable[[Iteration], None] | None = None,
        global_start_time: float | None = None,
        output: "ConsoleOutput | None" = None,
        existing_sandbox: "Sandbox | None" = None,
    ):
        """
        Initialize the agent.

        Args:
            client: LLM client for completions
            sandbox_manager: Manager for creating sandboxes
            config: RLM configuration
            problem_statement: Problem statement / task description
            instance_id: Instance identifier (e.g. repo name)
            repo_url: GitHub repository URL for creating sub-agent sandboxes
            depth: Recursion depth (0 = root)
            task: Task delegated by parent (None for root)
            on_iteration: Callback for iteration logging
            global_start_time: Start time for timeout tracking
            output: Console output for progress display
            existing_sandbox: Pre-created sandbox to use
        """
        self.client = client
        self.sandbox_manager = sandbox_manager
        self.config = config
        self.problem_statement = problem_statement
        self.instance_id = instance_id
        self.repo_url = repo_url
        self.depth = depth
        self.task = task
        self.on_iteration = on_iteration
        self.global_start_time = global_start_time or time.time()
        self.output = output
        self.existing_sandbox = existing_sandbox

        self.agent_id = f"agent_{uuid.uuid4().hex[:8]}"
        self.sandbox: "Sandbox | None" = None
        self.sandbox_id: str | None = None
        self.repl: DaytonaREPL | None = None

        self._iterations: list[Iteration] = []
        self._spawned_agents: list[AgentResult] = []
        self._usage = UsageStats()
        self._result: str | None = None
        self._lock = threading.Lock()

    def run(self) -> AgentResult:
        """
        Run the agent to completion.

        Returns:
            AgentResult with final answer, iterations, usage, etc.
        """
        start_time = time.time()

        try:
            # Create or use existing sandbox
            if self.existing_sandbox is not None:
                self.sandbox = self.existing_sandbox
                self.sandbox_id = self.sandbox.id
                logger.info(
                    f"Agent {self.agent_id} (depth={self.depth}) using existing sandbox {self.sandbox_id}"
                )
            else:
                self.sandbox, info = self.sandbox_manager.create_sandbox_from_repo(
                    repo_url=self._get_repo_url(),
                )
                self.sandbox_id = info.sandbox_id
                logger.info(
                    f"Agent {self.agent_id} (depth={self.depth}) created sandbox {self.sandbox_id}"
                )

            # Determine task content for REPL variable
            # Root agents get the problem statement, sub-agents get their assigned task
            task_content = self.problem_statement if self.depth == 0 else self.task

            # Initialize REPL with handlers and task variable
            self.repl = DaytonaREPL(
                sandbox=self.sandbox,
                rlm_query_handler=self._handle_rlm_query,
                rlm_query_batched_handler=self._handle_rlm_query_batched,
                initial_variables={"task": task_content},
                conda_env=None,
            )

            # Run iteration loop
            self._run_loop()

        except Exception as e:
            logger.error(f"Agent {self.agent_id} error: {e}")
            # Notify console of completion (with error)
            if self.output:
                self.output.agent_complete(self.agent_id, self.depth, None)
            # Root agents use problem_statement, sub-agents use task
            task_for_log = self.problem_statement if self.depth == 0 else self.task
            return AgentResult(
                agent_id=self.agent_id,
                depth=self.depth,
                sandbox_id=self.sandbox_id or "",
                task=task_for_log,
                iterations=self._iterations,
                spawned_agents=self._spawned_agents,
                result=None,
                usage=self._usage,
                execution_time=time.time() - start_time,
                error=str(e),
            )
        finally:
            # Cleanup REPL (stops broker and poller thread)
            if self.repl:
                self.repl.cleanup()

            # Cleanup sandbox - but NOT for root agent (may be needed after)
            if self.sandbox_id and self.depth > 0:
                self.sandbox_manager.delete_sandbox(self.sandbox_id)

        # Notify console of completion
        if self.output:
            self.output.agent_complete(self.agent_id, self.depth, self._result)

        # Root agents use problem_statement, sub-agents use task
        task_for_log = self.problem_statement if self.depth == 0 else self.task
        return AgentResult(
            agent_id=self.agent_id,
            depth=self.depth,
            sandbox_id=self.sandbox_id or "",
            task=task_for_log,
            iterations=self._iterations,
            spawned_agents=self._spawned_agents,
            result=self._result,
            result_truncated=len(self._result or "") > self.config.rlm.result_truncation_limit,
            usage=self._usage,
            execution_time=time.time() - start_time,
        )

    def _get_repo_url(self) -> str:
        """Get the repository URL for creating sub-agent sandboxes."""
        return self.repo_url

    def _run_loop(self) -> None:
        """Run the main iteration loop."""
        # Build system prompt
        system_prompt = build_system_prompt(depth=self.depth)

        messages = [{"role": "system", "content": system_prompt}]
        execution_result = None

        for iteration in range(self.config.rlm.max_iterations):
            # Check global timeout
            if self._is_timeout():
                logger.warning(f"Agent {self.agent_id} hit global timeout")
                break

            # Build user prompt
            user_prompt = build_user_prompt(iteration, execution_result)
            messages.append({"role": "user", "content": user_prompt})

            # Get model completion
            logger.debug(f"Agent {self.agent_id} iteration {iteration}")
            response = self.client.completion(messages)
            self._usage += self.client.last_usage

            # Add assistant response to history
            messages.append({"role": "assistant", "content": response})

            # Execute code blocks
            repl_result = self.repl.execute_response(response)

            # Create iteration record
            iter_record = Iteration(
                iteration=iteration,
                prompt=messages[-2]["content"],  # User prompt
                raw_response=response,
                parsed_code_blocks=repl_result.code_blocks,
                spawned_agents=[],  # Will be populated below
            )
            self._iterations.append(iter_record)

            # Call iteration callback
            if self.on_iteration:
                self.on_iteration(iter_record)

            # Check for final answer (FINAL_VAR is resolved to final_answer in execution script)
            if repl_result.final_answer is not None:
                self._result = repl_result.final_answer
                logger.info(f"Agent {self.agent_id} returned final answer")
                break

            # Format execution result for next iteration
            if repl_result.code_blocks:
                result_parts = []
                for block in repl_result.code_blocks:
                    result_parts.append(
                        format_execution_result(block.code, block.stdout, block.stderr, block.error)
                    )
                execution_result = "\n\n".join(result_parts)
            else:
                execution_result = None

        # If no final answer, try to extract patch from sandbox (root only)
        if self._result is None and self.depth == 0:
            self._result = self._extract_patch()

    def _handle_rlm_query(self, task: str) -> str:
        """
        Handle an rlm_query() call by spawning a sub-agent.

        Args:
            task: Task for the sub-agent

        Returns:
            Sub-agent's result string
        """
        logger.info(f"Agent {self.agent_id} spawning sub-agent for: {task[:50]}...")

        # Check budget
        if not self.sandbox_manager.budget.can_acquire():
            return (
                f"Error: Cannot spawn sub-agent - sandbox budget exhausted "
                f"({self.sandbox_manager.budget.status.created}/{self.sandbox_manager.budget.max} used). "
                f"Complete the task with the resources you have."
            )

        # Create sub-agent
        sub_agent = RLMAgent(
            client=self.client,
            sandbox_manager=self.sandbox_manager,
            config=self.config,
            problem_statement=self.problem_statement,
            instance_id=self.instance_id,
            repo_url=self.repo_url,
            depth=self.depth + 1,
            task=task,
            on_iteration=self.on_iteration,
            global_start_time=self.global_start_time,
            output=self.output,
        )

        # Notify console of spawn
        if self.output:
            self.output.agent_spawn(
                parent_id=self.agent_id,
                child_id=sub_agent.agent_id,
                task=task,
                depth=self.depth + 1,
            )

        # Run sub-agent
        result = sub_agent.run()
        self._spawned_agents.append(result)
        self._usage += result.usage

        # Return result, truncated if necessary
        if result.error:
            return f"Sub-agent error: {result.error}"

        answer = result.result or "Sub-agent did not return a result."

        # Truncate if too long
        limit = self.config.rlm.result_truncation_limit
        if len(answer) > limit:
            answer = answer[:limit] + f"\n\n[TRUNCATED - {len(answer) - limit} chars omitted]"

        return answer

    def _handle_rlm_query_batched(self, tasks: list[str]) -> list[str]:
        """
        Handle rlm_query_batched() by spawning multiple sub-agents in parallel.

        Args:
            tasks: List of tasks for sub-agents

        Returns:
            List of result strings
        """
        logger.info(f"Agent {self.agent_id} spawning {len(tasks)} sub-agents in parallel")

        # Check budget upfront
        if not self.sandbox_manager.budget.can_acquire(len(tasks)):
            remaining = self.sandbox_manager.budget.remaining
            return [
                f"Error: Cannot spawn {len(tasks)} sub-agents - only {remaining} sandbox slots remaining. "
                f"Reduce batch size or use sequential rlm_query() calls."
            ] * len(tasks)

        results = [""] * len(tasks)

        # Run sub-agents in parallel using thread pool
        with ThreadPoolExecutor(max_workers=min(len(tasks), 10)) as executor:
            future_to_idx = {
                executor.submit(self._handle_rlm_query, task): i for i, task in enumerate(tasks)
            }

            for future in as_completed(future_to_idx):
                idx = future_to_idx[future]
                try:
                    results[idx] = future.result()
                except Exception as e:
                    results[idx] = f"Error spawning sub-agent: {e}"

        return results

    def _extract_patch(self) -> str | None:
        """
        Extract git diff from the sandbox (fallback if no FINAL() called).

        Returns:
            Git diff string or None
        """
        if self.sandbox is None:
            return None

        try:
            executor = SandboxExecutor(self.sandbox)

            # Stage all changes
            executor.execute("git add -A")

            # Get diff
            result = executor.execute("git diff --cached HEAD")
            if result["returncode"] == 0 and result["output"].strip():
                return result["output"].strip()

            # Try unstaged diff as fallback
            result = executor.execute("git diff HEAD")
            if result["returncode"] == 0 and result["output"].strip():
                return result["output"].strip()

        except Exception as e:
            logger.warning(f"Failed to extract patch: {e}")

        return None

    def _is_timeout(self) -> bool:
        """Check if global timeout has been exceeded."""
        elapsed = time.time() - self.global_start_time
        return elapsed >= self.config.rlm.global_timeout
