# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""Daytona sandbox management for deeper-rlm."""

import logging
from dataclasses import dataclass
from typing import TYPE_CHECKING, Any

from daytona import (
    CreateSandboxFromImageParams,
    Daytona,
    DaytonaConfig,
    Image,
)

from rlm.budget import SandboxBudget

if TYPE_CHECKING:
    from daytona import Sandbox


logger = logging.getLogger(__name__)


@dataclass
class SandboxInfo:
    """Information about a sandbox."""

    sandbox_id: str
    instance_id: str


class SandboxManager:
    """
    Manages Daytona sandboxes for RLM agents.

    Handles:
    - Sandbox creation from GitHub repos
    - Budget tracking
    - Sandbox cleanup
    """

    def __init__(
        self,
        api_key: str,
        budget: SandboxBudget,
    ):
        """
        Initialize the sandbox manager.

        Args:
            api_key: Daytona API key
            budget: Shared sandbox budget tracker
        """
        self.budget = budget

        # Initialize Daytona client
        daytona_config = DaytonaConfig(api_key=api_key)
        self.daytona = Daytona(daytona_config)

        # Track active sandboxes
        self._active_sandboxes: dict[str, "Sandbox"] = {}

    def create_sandbox_from_repo(
        self,
        repo_url: str,
        branch: str | None = None,
        commit: str | None = None,
    ) -> tuple["Sandbox", SandboxInfo]:
        """
        Create a sandbox from a GitHub repository.

        Args:
            repo_url: GitHub repository URL
            branch: Optional branch name (default: repo's default branch)
            commit: Optional commit SHA to checkout

        Returns:
            Tuple of (Sandbox, SandboxInfo)

        Raises:
            RuntimeError: If budget exhausted or sandbox creation fails
        """
        # Check budget
        if not self.budget.try_acquire():
            raise RuntimeError(
                f"Cannot create sandbox - budget exhausted " f"({self.budget.status.created}/{self.budget.max} used)"
            )

        try:
            # Create a base image with git installed (needed for agent to produce diffs)
            base_image = (
                Image.debian_slim("3.11").run_commands("apt-get update && apt-get install -y git").workdir("/workspace")
            )

            logger.info(f"Creating sandbox from base image for repo: {repo_url}")
            sandbox = self.daytona.create(
                CreateSandboxFromImageParams(image=base_image),
                timeout=0,  # No timeout for image build
                on_snapshot_create_logs=lambda msg: logger.debug(f"Image build: {msg}"),
            )

            # Clone the repository
            logger.info(f"Cloning repository: {repo_url}")
            sandbox.git.clone(
                url=repo_url,
                path="/workspace",
                branch=branch,
            )

            # Checkout specific commit if provided
            if commit:
                logger.info(f"Checking out commit: {commit}")
                result = sandbox.process.exec(
                    f"git checkout {commit}",
                    cwd="/workspace",
                    timeout=60,
                )
                if result.exit_code != 0:
                    raise RuntimeError(f"Failed to checkout commit {commit}: {result.result}")

            # Extract repo name for info
            repo_name = repo_url.rstrip("/").split("/")[-1].replace(".git", "")

            info = SandboxInfo(
                sandbox_id=sandbox.id,
                instance_id=repo_name,
            )

            self._active_sandboxes[sandbox.id] = sandbox
            logger.info(f"Sandbox created: {sandbox.id}")

            return sandbox, info

        except Exception as e:
            # Release budget on failure
            self.budget.release()
            raise RuntimeError(f"Failed to create sandbox from repo: {e}") from e

    def delete_sandbox(self, sandbox_id: str) -> None:
        """
        Delete a sandbox and release budget.

        Args:
            sandbox_id: ID of sandbox to delete
        """
        sandbox = self._active_sandboxes.pop(sandbox_id, None)
        if sandbox is not None:
            try:
                sandbox.delete()
                logger.info(f"Sandbox deleted: {sandbox_id}")
            except Exception as e:
                logger.warning(f"Error deleting sandbox {sandbox_id}: {e}")
            finally:
                self.budget.release()

    def get_sandbox(self, sandbox_id: str) -> "Sandbox | None":
        """Get an active sandbox by ID."""
        return self._active_sandboxes.get(sandbox_id)

    def cleanup_all(self) -> None:
        """Clean up all active sandboxes."""
        sandbox_ids = list(self._active_sandboxes.keys())
        for sandbox_id in sandbox_ids:
            self.delete_sandbox(sandbox_id)


class SandboxExecutor:
    """
    Executes commands in a Daytona sandbox.

    Wraps commands with conda environment activation.
    """

    def __init__(self, sandbox: "Sandbox", cwd: str = "/workspace", conda_env: str = "testbed"):
        """
        Initialize the executor.

        Args:
            sandbox: Daytona sandbox instance
            cwd: Default working directory
            conda_env: Conda environment to activate
        """
        self.sandbox = sandbox
        self.cwd = cwd
        self.conda_env = conda_env

    def execute(
        self,
        command: str,
        cwd: str | None = None,
        timeout: int = 120,
    ) -> dict[str, Any]:
        """
        Execute a command in the sandbox.

        Args:
            command: Shell command to execute
            cwd: Working directory (defaults to self.cwd)
            timeout: Timeout in seconds

        Returns:
            Dict with 'output' and 'returncode'
        """
        effective_cwd = cwd or self.cwd
        wrapped_command = self._wrap_command(command)

        logger.debug(f"Executing: {command}")
        logger.debug(f"Wrapped: {wrapped_command}")

        try:
            response = self.sandbox.process.exec(
                wrapped_command,
                cwd=effective_cwd,
                timeout=timeout,
            )

            return {
                "output": response.result or "",
                "returncode": response.exit_code,
            }
        except Exception as e:
            logger.error(f"Command execution error: {e}")
            return {
                "output": str(e),
                "returncode": 1,
            }

    def _wrap_command(self, command: str) -> str:
        """Wrap command with conda environment activation."""
        quoted = self._shell_quote(command)
        return f"conda run -n {self.conda_env} --no-capture-output bash -c {quoted}"

    def _shell_quote(self, s: str) -> str:
        """Safely quote a string for shell execution."""
        return "'" + s.replace("'", "'\"'\"'") + "'"
