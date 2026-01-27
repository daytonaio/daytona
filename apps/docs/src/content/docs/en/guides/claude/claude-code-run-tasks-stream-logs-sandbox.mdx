---
title: Running Claude Code with Daytona
description: Step-by-step guide to running Claude Code with Daytona sandboxes.
---

import { TabItem, Tabs } from '@astrojs/starlight/components'

Claude Code allows you to automate and orchestrate tasks using natural language and code. With Daytona, you can easily run Claude Code inside isolated sandboxes, making it simple to experiment and execute tasks securely.

## Running Claude Code in a Daytona Sandbox

You can run Claude Code and execute tasks with it directly inside a Daytona sandbox. The following examples show how to set up a sandbox, install Claude Code, run tasks programmatically, and stream logs in real time.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">
> **Note:** While both sync and async modes support streaming PTY output, `AsyncDaytona` is recommended as it provides automatic background callbacks via `on_data`. The synchronous API requires blocking iteration or manual threading to handle output.
```python
import os
import asyncio
from daytona import AsyncDaytona

async def run_claude_code():
    async with AsyncDaytona() as daytona:
        sandbox = await daytona.create()

        # Define the Claude Code command to be executed
        claude_command = "claude --dangerously-skip-permissions -p 'write a dad joke about penguins' --output-format stream-json --verbose"

        # Install Claude Code in the sandbox
        await sandbox.process.exec("npm install -g @anthropic-ai/claude-code")

        pty_handle = await sandbox.process.create_pty_session(
            id="claude", on_data=lambda data: print(data.decode(), end="")
        )

        await pty_handle.wait_for_connection()

        # Run the Claude Code command inside the sandbox
        await pty_handle.send_input(
            f"ANTHROPIC_API_KEY={os.environ['ANTHROPIC_API_KEY']} {claude_command}\n"
        )

        # Use this to close the terminal session if no more commands will be executed
        # await pty_handle.send_input("exit\n")

        await pty_handle.wait()

        # If you are done and have closed the PTY terminal, it is recommended to clean up resources by deleting the sandbox
        # await sandbox.delete()

if __name__ == "__main__":
    asyncio.run(run_claude_code())

````
</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">
```typescript
import { Daytona } from "@daytonaio/sdk";

const daytona = new Daytona();

try {
    const sandbox = await daytona.create();

    // Define the Claude Code command to be executed
    const claudeCommand =
    "claude --dangerously-skip-permissions -p 'write a dad joke about penguins' --output-format stream-json --verbose";

    // Install Claude Code in the sandbox
    await sandbox.process.executeCommand("npm install -g @anthropic-ai/claude-code");

    const ptyHandle = await sandbox.process.createPty({
        id: "claude",
        onData: (data) => {
            process.stdout.write(data);
        },
    });

    await ptyHandle.waitForConnection();

    // Run the Claude Code command inside the sandbox
    ptyHandle.sendInput(
    `ANTHROPIC_API_KEY=${process.env.ANTHROPIC_API_KEY} ${claudeCommand}\n`
    );

    // Use this to close the terminal session if no more commands will be executed
    // ptyHandle.sendInput("exit\n")

    await ptyHandle.wait();

    // If you are done and have closed the PTY terminal, it is recommended to clean up resources by deleting the sandbox
    // await sandbox.delete();
} catch (error) {
    console.error("Failed to run Claude Code in Daytona sandbox:", error);
}
````

</TabItem>
</Tabs>
