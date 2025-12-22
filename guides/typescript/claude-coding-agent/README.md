# Claude Coding Agent - Two-Agent System

## Overview

This example demonstrates a **two-agent system** where:

- A **Project Manager Agent** runs locally with its own Claude instance
- A **Developer Agent** runs inside a Daytona sandbox with the full capabilities of Claude Code

The Project Manager receives user requests, breaks them down into tasks, delegates coding work to the Developer Agent, reviews the outputs, and communicates results back to the user. This creates a hierarchical agent architecture where high-level planning is separated from low-level code execution.

> Note: Both agents use Anthropic API keys. The Developer Agent's key is passed into the sandbox environment and may be accessible to any code executed within it.

## Features

- **Dual-agent architecture:** Project Manager oversees tasks while Developer Agent executes code
- **Color-coded output:** Project Manager messages in green, Developer Agent in white for easy differentiation
- **Secure sandbox execution:** Developer Agent operates within a controlled Daytona environment
- **Claude Agent SDK integration:** Developer Agent has full abilities including reading/editing files and running shell commands
- **Preview deployed apps:** Use Daytona preview links to view and interact with deployed applications
- **Intelligent delegation:** Project Manager analyzes outputs and decides if more work is needed

## Prerequisites

- **Node.js:** Version 18 or higher is required

## Environment Variables

To run this example, you need to set the following environment variables:

- `DAYTONA_API_KEY`: Required for access to Daytona sandboxes. Get it from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)
- `ANTHROPIC_API_KEY`: Required for the **Project Manager Agent** (runs locally). Get it from [Claude Developer Platform](https://console.anthropic.com/settings/keys)
- `SANDBOX_ANTHROPIC_API_KEY`: Required for the **Developer Agent** (runs in sandbox). Get it from [Claude Developer Platform](https://console.anthropic.com/settings/keys)

Create a `.env` file in the project directory with these variables.

**Note:** You can use the same Anthropic API key for both `ANTHROPIC_API_KEY` and `SANDBOX_ANTHROPIC_API_KEY`, or use separate keys for billing/tracking purposes.

## Getting Started

### Setup and Run

1. Install dependencies:

   ```bash
   npm install
   ```

2. Run the example:

   ```bash
   npm run start
   ```

## How It Works

The two-agent system follows this workflow:

1. **Initialization:**
   - A new Daytona sandbox is created for the Developer Agent
   - The Claude Agent SDK is installed in the sandbox
   - The Project Manager Agent is initialized locally

2. **User interaction:**
   - User sends a request (e.g., "build a web app")
   - Project Manager Agent analyzes the request and plans the approach

3. **Task delegation:**
   - Project Manager uses `<developer_task>` tags to delegate specific tasks
   - Developer Agent receives the task and executes it in the sandbox
   - Developer Agent's output is streamed back in real-time

4. **Review and iteration:**
   - Project Manager reviews the Developer Agent's output
   - Decides if more work is needed or if the task is complete
   - Can delegate additional tasks if necessary

5. **Cleanup:**
   - When the script is terminated, the sandbox is automatically deleted

## Architecture

```
┌─────────────────────────┐
│        User             │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│  Project Manager Agent  │ (Local - Claude Sonnet 4) [GREEN OUTPUT]
│  - Plans tasks          │
│  - Delegates work       │
│  - Reviews outputs      │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│   Developer Agent       │ (Daytona Sandbox) [WHITE OUTPUT]
│  - Executes code        │
│  - Manages files        │
│  - Starts services      │
│  - Provides previews    │
└─────────────────────────┘
```

**Color Coding:**

- 🟢 **Green text** = Project Manager Agent (high-level planning and coordination)
- ⚪ **White text** = Developer Agent (code execution and output)

## Example Output

```
Creating Developer Agent sandbox...
Installing Developer Agent SDK...
Initializing Developer Agent...
Initializing Project Manager Agent...

=== Two-Agent System Ready ===
Project Manager Agent: Manages tasks and delegates to Developer (Green text)
Developer Agent: Executes code in Daytona sandbox (White text)
Press Ctrl+C at any time to exit.

User: make a lunar lander web app

[Project Manager] Processing your request...

[Project Manager]: I'll help you create a lunar lander web app! Let me delegate this task to the Developer Agent.

<developer_task>
Create a lunar lander web game with the following features:
- HTML/CSS/JavaScript based
- Canvas graphics for the game
- Physics simulation (gravity, thrust, velocity)
- Keyboard controls (arrow keys)
- Landing detection (safe landing vs crash)
- Fuel management system
- Start the game on port 80 and provide the preview URL
</developer_task>

[Delegating to Developer Agent]...

I'll help you create a lunar lander web app! Let me build a complete game with physics simulation, graphics, and controls.
🔨 Write
🔨 Write
🔨 Bash
Perfect! The web server is now running.

Your Lunar Lander game is live at:
🌐 https://80-17ac1c0f-d684-4122-93b5-8f52fd5393f8.proxy.daytona.works

[Project Manager]: Excellent! The Developer Agent has successfully created your lunar lander web app. The game is now running and accessible at the preview URL above. The implementation includes:

✓ Physics simulation with gravity and thrust
✓ Canvas-based graphics
✓ Keyboard controls for gameplay
✓ Landing detection system
✓ Fuel management
✓ Web server running on port 80

TASK_COMPLETE

[Project Manager] All tasks completed!
```

## Customization

You can customize the Project Manager Agent's behavior by modifying the system prompt in `src/index.ts`. The current implementation:

- Uses `<developer_task>` tags for delegation
- Automatically reviews Developer Agent outputs
- Says "TASK_COMPLETE" when finished

## License

See the main project LICENSE file for details.

## References

- [Claude Agent SDK Overview](https://platform.claude.com/docs/en/agent-sdk/overview)
- [Claude Agent SDK Reference (Python)](https://platform.claude.com/docs/en/agent-sdk/python)
- [Daytona Documentation](https://www.daytona.io/docs)
