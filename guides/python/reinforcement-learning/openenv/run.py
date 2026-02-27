"""Run a single FinQA episode on a Daytona sandbox using OpenEnv.

Demonstrates the full OpenEnv + Daytona integration pattern:
  1. Create a sandbox from a pre-built snapshot
  2. Connect to the FinQA environment over WebSocket
  3. Run a multi-turn episode: explore tables, query data, submit an answer
  4. Observe the reward signal (1.0 = correct, 0.0 = wrong)
  5. Tear down the sandbox

Prerequisites:
  - DAYTONA_API_KEY set in environment (or .env file)
  - The "openenv-finqa" snapshot already built (see build_snapshot.py)
  - Dependencies installed (pip install -e .)

Usage:
    python run.py
"""

import asyncio
import json

from dotenv import load_dotenv
from finqa_env import CallToolAction, FinQAEnv  # pylint: disable=import-error
from openenv.core.containers.runtime.daytona_provider import DaytonaProvider  # pylint: disable=import-error

load_dotenv()

# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------
SNAPSHOT = "openenv-finqa"
SERVER_CMD = "cd /app/env && uvicorn finqa_env.server.app:app --host 0.0.0.0 --port 8000"

# Available tools in the FinQA environment.
# Note: list_tools() is currently broken in openenv-core (the server rejects
# the WebSocket message with a validation error due to an MCP/JSON-RPC routing
# mismatch). The tools are known from the environment definition:
#   - get_descriptions(company_name)  — list tables for a company
#   - get_table_info(company_name, table_name)  — column names and types
#   - sql_query(query)  — run a SQL query against the company's 10-K data
#   - submit_answer(answer)  — submit a final answer (terminates the episode)


# ---------------------------------------------------------------------------
# Sandbox lifecycle
# ---------------------------------------------------------------------------
def create_sandbox():
    """Create a Daytona sandbox, start the FinQA server, wait for health."""
    provider = DaytonaProvider(auto_stop_interval=0, cmd=SERVER_CMD)

    print(f"Creating sandbox from snapshot '{SNAPSHOT}'...")
    url = provider.start_container(f"snapshot:{SNAPSHOT}")

    print("Waiting for server health check...")
    provider.wait_for_ready(url, 120)
    print("  Server healthy.")

    return provider, url


# ---------------------------------------------------------------------------
# Episode
# ---------------------------------------------------------------------------
async def run_episode(url: str):
    """Connect to the FinQA env and run one full episode.

    Returns (reward, steps_taken, question).
    """
    async with FinQAEnv(base_url=url) as env:

        # 1. Reset — starts a new episode with a random question
        await env.reset()

        # 2. Get the question and company name.
        #    Workaround: obs.metadata["question"] doesn't work because
        #    openenv-core's serialize_observation() strips metadata from
        #    reset observations (exclude={"metadata"}).
        #    The intended public API would be: obs.observation.metadata["question"]
        state = await env._send_and_receive({"type": "state"})
        data = state.get("data", {})
        question = data.get("current_question", "")
        company = data.get("current_company", "")
        print(f"\nQuestion: {question}")
        print(f"Company:  {company}")

        # 3. Discover available tables — call_tool() is a convenience method
        #    that sends a tool call and returns the result directly (no
        #    reward/done tracking). Good for exploration steps.
        descriptions = await env.call_tool("get_descriptions", company_name=company)
        table_names = json.loads(descriptions)  # list of table name strings
        print(f"\nTables:   {table_names}")

        # 4. Inspect the first table's schema
        table_info = await env.call_tool("get_table_info", company_name=company, table_name=table_names[0])
        print(f"Schema:   {table_info[:200]}")

        # 5. Query data — step() wraps the tool call in an RL-style
        #    StepResult with .observation.done and .observation.reward.
        #    Use step() when you need reward/done signals (e.g. in a
        #    training loop). Use call_tool() for simple exploration.
        query = f'SELECT * FROM "{table_names[0]}" LIMIT 5'
        step_result = await env.step(CallToolAction(tool_name="sql_query", arguments={"query": query}))
        obs = step_result.observation
        print(f"\nSQL result (done={obs.done}, reward={obs.reward}):")
        print(f"  {str(obs.result)[:300]}")

        # 6. Submit an answer — this terminates the episode.
        #    A real agent would reason about the data; we just submit a
        #    placeholder to demonstrate the full lifecycle.
        step_result = await env.step(CallToolAction(tool_name="submit_answer", arguments={"answer": "0"}))
        obs = step_result.observation
        print(f"\nSubmitted (done={obs.done}, reward={obs.reward})")

        return obs.reward, 2, question


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
def main():
    provider, url = create_sandbox()

    try:
        reward, steps, question = asyncio.run(run_episode(url))
        print(f"\n{'=' * 60}")
        print("Episode complete")
        print(f"  Question: {question[:100]}")
        print(f"  Reward:   {reward}")
        print(f"  Steps:    {steps}")
        print(f"{'=' * 60}")
    finally:
        print("\nCleaning up sandbox...")
        provider.stop_container()
        print("Done.")


if __name__ == "__main__":
    main()
