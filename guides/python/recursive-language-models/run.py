#!/usr/bin/env python3
"""Main entry point for deeper-rlm."""

import logging
import os
from datetime import datetime
from pathlib import Path

import typer
from dotenv import load_dotenv

from output_logging.console import ConsoleOutput
from output_logging.tree_logger import TreeLogger
from rlm.agent import RLMAgent
from rlm.budget import SandboxBudget
from rlm.client import create_client
from rlm.sandbox import SandboxManager
from rlm.types import Config

# Load environment variables
load_dotenv()

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)

def main(
    repo: str = typer.Argument(..., help="GitHub repository URL"),
    prompt: str = typer.Option(..., "--prompt", "-p", help="Task prompt for the agent"),
    branch: str = typer.Option(None, "--branch", "-b", help="Branch name"),
    commit: str = typer.Option(None, "--commit", help="Specific commit SHA"),
    config_path: Path = typer.Option(
        Path("config.yaml"),
        "--config",
        "-c",
        help="Path to YAML configuration file",
    ),
    output_file: Path = typer.Option(
        None,
        "--output",
        "-o",
        help="Output file for the patch (default: stdout)",
    ),
    verbose: bool = typer.Option(
        True,
        "--verbose/--quiet",
        help="Enable verbose output",
    ),
):
    """Run deeper-rlm agent on a GitHub repository."""
    print("deeper-rlm")
    print()

    # Load configuration
    try:
        config = Config.from_yaml(config_path)
    except Exception as e:
        print(f"Error loading config: {e}")
        raise typer.Exit(1)

    # Validate config
    if config.model is None:
        print("Error: No model configured")
        raise typer.Exit(1)

    print(f"Repository: {repo}")
    if branch:
        print(f"Branch: {branch}")
    if commit:
        print(f"Commit: {commit}")
    print(f"Model: {config.model.name}")
    print(f"Max sandboxes: {config.rlm.max_sandboxes}")
    print()

    # Setup output
    output_handler = ConsoleOutput(verbose=verbose)

    try:
        # Get API keys from environment
        model_api_key = os.environ["LLM_API_KEY"]
        daytona_api_key = os.environ["DAYTONA_API_KEY"]

        # Create LLM client
        client = create_client(
            model_name=config.model.name,
            api_key=model_api_key,
        )

        # Create sandbox budget and manager
        budget = SandboxBudget(config.rlm.max_sandboxes)
        sandbox_manager = SandboxManager(
            api_key=daytona_api_key,
            budget=budget,
        )

        # Create sandbox from GitHub repo
        print("Creating sandbox from repository...")
        sandbox, info = sandbox_manager.create_sandbox_from_repo(
            repo_url=repo,
            branch=branch,
            commit=commit,
        )
        print(f"Sandbox created: {sandbox.id}")
        print()

        # Extract repo name for logging
        repo_name = repo.rstrip("/").split("/")[-1].replace(".git", "")

        # Create agent
        agent = RLMAgent(
            client=client,
            sandbox_manager=sandbox_manager,
            config=config,
            problem_statement=prompt,
            instance_id=repo_name,
            repo_url=repo,
            depth=0,
            on_iteration=lambda it: output_handler.iteration(agent.agent_id, 0, it),
            output=output_handler,
            existing_sandbox=sandbox,
        )

        # Run agent
        print(f"Starting agent with prompt: {prompt[:80]}...")
        print()
        agent_result = agent.run()

        # Get the patch
        patch = agent_result.result

        # Display results
        print()
        print("=" * 40)
        print("Results")
        print("=" * 40)

        if patch:
            print(f"Patch generated ({len(patch)} chars)")
            print()

            if output_file:
                output_file.write_text(patch)
                print(f"Patch saved to: {output_file}")
            else:
                print("--- Patch ---")
                print(patch)
                print("--- End Patch ---")
        else:
            print("No patch was generated")

        # Show stats
        print()
        print(f"Total iterations: {len(agent_result.iterations)}")
        print(f"Sub-agents spawned: {len(agent_result.spawned_agents)}")
        print(f"Execution time: {agent_result.execution_time:.1f}s")
        print(f"Total cost: ${agent_result.usage.cost:.4f}")

        # Show tree view
        if agent_result:
            output_handler.tree_view(agent_result)

        # Save results for viewer
        tree_logger = TreeLogger("results")
        run_id = f"{repo_name}_{datetime.now().strftime('%Y%m%d_%H%M%S')}"
        result_path = tree_logger.log_agent_result(agent_result, run_id)
        print(f"\nResults saved to: {result_path}")

        # Cleanup
        sandbox_manager.cleanup_all()

    except Exception as e:
        logger.exception("Error running agent")
        print(f"Error: {e}")
        raise typer.Exit(1)


if __name__ == "__main__":
    typer.run(main)
