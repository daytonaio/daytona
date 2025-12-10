# Claude Coding Agent

## Overview

This example runs a coding agent with the capabilities of Claude Code inside a Daytona sandbox. You can interact with the agent via the CLI to run automations, build apps, and launch web apps or services using [Daytona preview links](https://www.daytona.io/docs/en/preview-and-authentication/#fetching-a-preview-link).

> Note: In this example, your Anthropic API key is passed into the sandbox environment and may be accessible to any code executed within it.

## Features

- **Secure sandbox execution:** The agent operates within a controlled environment, along with code or commands run by the agent.
- **Claude Agent integration:** Includes the full abilities of the Claude Agent SDK, including reading and editing code files, and running shell commands.
- **Preview deployed apps:** Use Daytona preview links to view and interact with your deployed applications.

## Prerequisites

- **Node.js:** Version 18 or higher is required

## Environment Variables

To run this example, you need to set the following environment variables:

- `DAYTONA_API_KEY`: Required for access to Daytona sandboxes. Get it from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)
- `SANDBOX_ANTHROPIC_API_KEY`: Required to run Claude Code. Get it from [Claude Developer Platform](https://console.anthropic.com/settings/keys)

Create a `.env` file in the project directory with these variables.

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

When this example is run, the agent follows the following workflow:

1. A new Daytona sandbox is created.
2. The coding agent is installed and launched inside the sandbox.
3. User queries are passed to the agent, and the result is displayed to the user.
4. When the script is terminated, the sandbox is deleted.

## Example Output

```
Creating sandbox...
Installing Agent SDK...
Initializing Agent SDK...
Press Ctrl+C at any time to exit.
User: make a lunar lander web app
Thinking...
I'll help you create a lunar lander web app! This is a fun game where players control a lunar module trying to land safely on the moon's surface. Let me plan the implementation first to make sure we build something great.
🔨 EnterPlanMode
Let me create a lunar lander web app for you! I'll build a complete game with:
- Physics simulation (gravity, thrust, velocity)
- Canvas-based graphics
- Keyboard controls (arrow keys for thrust)
- Landing detection (safe vs crash)
- Fuel management
- Visual feedback and game states

Let me start by creating the necessary files:
> 🔨 Write
🔨 Write
> 🔨 Write
...
Perfect! The web server is now running successfully. 🚀

Your Lunar Lander game is live at:
🌐 https://80-17ac1c0f-d684-4122-93b5-8f52fd5393f8.proxy.daytona.works

The server is serving the files from /home/daytona/. Click the link above to start playing!

Objective: Land gently on the green platform with safe speeds. Good luck! 🌙
```

## License

See the main project LICENSE file for details.

## References

- [Claude Agent SDK Overview](https://platform.claude.com/docs/en/agent-sdk/overview)
- [Claude Agent SDK Reference (Python)](https://platform.claude.com/docs/en/agent-sdk/python)
- [Daytona Documentation](https://www.daytona.io/docs)
