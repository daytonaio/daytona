# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# pylint: disable=no-member

"""
Create a long-lived agent.

Fails if an agent with the given name already exists, so it is safe to run
more than once.
"""
import argparse
import os
import sys

import anthropic
import dotenv

dotenv.load_dotenv(override=True)

SANDBOX_TOOLS = ["bash", "read", "write", "edit", "glob", "grep"]
WEB_TOOLS = ["web_fetch", "web_search"]

parser = argparse.ArgumentParser(description="Create a long-lived agent.")
parser.add_argument("name", help="agent name; must be unique among non-archived agents")
args = parser.parse_args()
name = args.name

client = anthropic.Anthropic(api_key=os.environ["ANTHROPIC_API_KEY"])

for existing in client.beta.agents.list():
    if existing.name == name and existing.archived_at is None:
        print(f"agent named {name!r} already exists: {existing.id}", file=sys.stderr)
        sys.exit(1)

agent = client.beta.agents.create(
    name=name,
    model="claude-sonnet-4-6",
    system="You have a working sandbox. Use your tools to do what is asked. Be terse.",
    tools=[
        {
            "type": "agent_toolset_20260401",
            "default_config": {"enabled": False, "permission_policy": {"type": "always_allow"}},
            "configs": [
                {"name": tool_name, "enabled": True, "permission_policy": {"type": "always_allow"}}
                for tool_name in SANDBOX_TOOLS + WEB_TOOLS
            ],
        }
    ],
)

print(f"created agent {agent.id} (name: {agent.name}, version: {agent.version})")
print()
print(f"pass to scripts as: --agent-id {agent.id}")
