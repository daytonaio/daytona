# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# Note: This module is uploaded to the Daytona sandbox and used inside of the code interpreter.

import os
import logging
import sys
import asyncio

# Import the Claude Agent SDK
from claude_agent_sdk import ClaudeSDKClient, ClaudeAgentOptions, AssistantMessage, TextBlock, ToolUseBlock # type: ignore

# Suppress INFO level logging from claude_agent_sdk
logging.getLogger('claude_agent_sdk').setLevel(logging.WARNING)

# Helper to run async functions synchronously
def run_sync(coro):
    loop = asyncio.get_event_loop()
    return loop.run_until_complete(coro)

# Set up a global event loop
loop = asyncio.new_event_loop()
asyncio.set_event_loop(loop)

# Generate a system prompt for the agent
system_prompt = """
You are running in a Daytona sandbox.
Use the /home/daytona directory instead of /workspace for file operations.
Your public preview URL for port 80 is: {}.
This is an example of the preview URL format.
When you start other services, they will follow the same pattern on other ports.
""".format(os.environ.get('PREVIEW_URL', ''))

# Create an agent instance
client = ClaudeSDKClient(
  options=ClaudeAgentOptions(
    allowed_tools=["Read", "Edit", "Glob", "Grep", "Bash"],
    permission_mode="acceptEdits",
    system_prompt=system_prompt
  )
)

# Initialize the client
async def init_client():
  await client.__aenter__()
  print("Agent SDK is ready.")

run_sync(init_client())

# Run a query and stream the response
async def run_query(prompt):
  await client.query(prompt)
  async for message in client.receive_response():
    if isinstance(message, AssistantMessage):
      for block in message.content:
        if isinstance(block, TextBlock):
          text = block.text
          if not text.endswith("\n"):
            text = text + "\n"
          sys.stdout.write(text)
          sys.stdout.flush()
        elif isinstance(block, ToolUseBlock):
          sys.stdout.write(f"🔨 {block.name}\n")
          sys.stdout.flush()

# Synchronous wrapper for run_query
def run_query_sync(prompt):
  return run_sync(run_query(prompt))