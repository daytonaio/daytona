# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""Parse agent log files into structured JSON for the viewer."""

import json
import re
import sys
from dataclasses import dataclass, field


@dataclass
class Agent:
    agent_id: str
    depth: int
    task: str = ""
    sandbox_id: str = ""
    iterations: list = field(default_factory=list)
    result: str = ""
    error: str = ""
    spawned_agents: list = field(default_factory=list)
    start_time: str = ""
    end_time: str = ""


def parse_log(log_path: str) -> dict:
    """Parse a log file and extract agent tree."""

    with open(log_path, "r") as f:
        content = f.read()

    agents = {}  # agent_id -> Agent
    agent_tasks = {}  # agent_id -> task (from spawning)

    lines = content.split("\n")

    for i, line in enumerate(lines):
        # Agent created
        match = re.search(r"Agent (\w+) \(depth=(\d+)\) created sandbox ([\w-]+)", line)
        if match:
            agent_id, depth, sandbox_id = match.groups()
            agents[agent_id] = Agent(
                agent_id=agent_id,
                depth=int(depth),
                sandbox_id=sandbox_id,
                task=agent_tasks.get(agent_id, "ROOT AGENT"),
            )
            # Extract timestamp
            ts_match = re.match(r"(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})", line)
            if ts_match:
                agents[agent_id].start_time = ts_match.group(1)
            continue

        # Agent spawning sub-agent
        match = re.search(r"Agent (\w+) spawning sub-agent for: (.+)", line)
        if match:
            parent_id, task = match.groups()
            # Store task for when agent is created
            # We'll need to match this to the next agent created
            if parent_id not in agents:
                continue
            # Task will be assigned when agent is created
            continue

        # Spawning with task (capture the task)
        match = re.search(r"Agent (\w+) spawning (\d+) sub-agents", line)
        if match:
            parent_id, count = match.groups()
            continue

        # Agent spawning specific sub-agent
        match = re.search(r"Agent (\w+) spawning sub-agent for: (.{1,200})", line)
        if match:
            parent_id, task = match.groups()
            # Store for next agent creation
            if "pending_tasks" not in dir():
                pending_tasks = []
            pending_tasks = getattr(parse_log, "pending_tasks", [])
            pending_tasks.append((parent_id, task))
            parse_log.pending_tasks = pending_tasks

        # Agent iteration
        match = re.search(r"Agent (\w+) - Iteration (\d+)", line)
        if match:
            agent_id, iteration = match.groups()
            if agent_id in agents:
                # Look for code block after this
                pass
            continue

        # Agent returned final answer
        match = re.search(r"Agent (\w+) returned final answer", line)
        if match:
            agent_id = match.group(1)
            if agent_id in agents:
                ts_match = re.match(r"(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})", line)
                if ts_match:
                    agents[agent_id].end_time = ts_match.group(1)
            continue

        # Agent error
        match = re.search(r"Agent (\w+) error: (.+)", line)
        if match:
            agent_id, error = match.groups()
            if agent_id in agents:
                agents[agent_id].error = error
                ts_match = re.match(r"(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})", line)
                if ts_match:
                    agents[agent_id].end_time = ts_match.group(1)
            continue

    # Build tree structure
    # Find root (depth=0)
    root = None
    for agent in agents.values():
        if agent.depth == 0:
            root = agent
            break

    if not root:
        return {"error": "No root agent found", "agents": list(agents.keys())}

    # Extract tasks from log more carefully
    task_pattern = re.compile(r"spawning sub-agent for: (.+?)(?:\.\.\.|$)", re.MULTILINE)
    tasks = task_pattern.findall(content)

    # Match tasks to agents by order
    depth1_agents = sorted([a for a in agents.values() if a.depth == 1], key=lambda x: x.start_time)
    depth2_agents = sorted([a for a in agents.values() if a.depth == 2], key=lambda x: x.start_time)

    # Assign tasks (simplified)
    task_idx = 0
    for agent in depth1_agents:
        if task_idx < len(tasks):
            task_text = tasks[task_idx]
            agent.task = task_text[:200] + "..." if len(task_text) > 200 else task_text
            task_idx += 1

    for agent in depth2_agents:
        if task_idx < len(tasks):
            task_text = tasks[task_idx]
            agent.task = task_text[:200] + "..." if len(task_text) > 200 else task_text
            task_idx += 1

    # Build parent-child relationships based on timing
    for agent in depth1_agents:
        root.spawned_agents.append(agent.agent_id)

    # For depth 2, find their parent (depth 1 agent that was active when they were created)
    for d2_agent in depth2_agents:
        # Find depth-1 agent that spawned around that time
        for d1_agent in depth1_agents:
            if d1_agent.start_time <= d2_agent.start_time:
                if d2_agent.agent_id not in d1_agent.spawned_agents:
                    d1_agent.spawned_agents.append(d2_agent.agent_id)
                    break

    # Convert to serializable format
    def agent_to_dict(agent_id):
        if agent_id not in agents:
            return {"agent_id": agent_id, "error": "not found"}
        agent = agents[agent_id]
        return {
            "agent_id": agent.agent_id,
            "depth": agent.depth,
            "task": agent.task,
            "sandbox_id": agent.sandbox_id,
            "result": agent.result,
            "error": agent.error,
            "start_time": agent.start_time,
            "end_time": agent.end_time,
            "children": [agent_to_dict(child_id) for child_id in agent.spawned_agents],
        }

    return {"log_file": log_path, "total_agents": len(agents), "tree": agent_to_dict(root.agent_id)}


def main():
    if len(sys.argv) < 2:
        print("Usage: parse_log.py <log_file> [output.json]")
        sys.exit(1)

    log_path = sys.argv[1]
    output_path = sys.argv[2] if len(sys.argv) > 2 else "parsed_agents.json"

    result = parse_log(log_path)

    with open(output_path, "w") as f:
        json.dump(result, f, indent=2)

    print(f"Parsed {result.get('total_agents', 0)} agents")
    print(f"Output saved to {output_path}")


if __name__ == "__main__":
    main()
