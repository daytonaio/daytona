# OpenCode Server

## Overview

This example runs an [OpenCode](https://opencode.ai/docs/sdk/) coding agent inside a Daytona sandbox. You can interact with the agent via the CLI to run automations, build apps, and launch web apps or services using [Daytona preview links](https://www.daytona.io/docs/en/preview-and-authentication/#fetching-a-preview-link).

## Features

- **Secure sandbox execution:** The agent operates within a controlled environment, along with code or commands run by the agent.
- **OpenCode integration:** The OpenCode server runs in the sandbox while the host attaches via the SDK, enabling full agent capabilities including reading and editing code files, and running shell commands.
- **Preview deployed apps:** Use Daytona preview links to view and interact with your deployed applications.

## Prerequisites

- **Node.js:** Version 18 or higher is required

## Environment Variables

To run this example, you need to set the following environment variables:

- `DAYTONA_API_KEY`: Required for access to Daytona sandboxes. Get it from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)

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

1. A new Daytona sandbox is created (public so preview links are reachable).
2. OpenCode is installed in the sandbox and the server is started.
3. The host attaches via the OpenCode SDK and enters an interactive loop.
4. User queries are passed to the agent, tool events are streamed, and the result is displayed.
5. When the script is terminated, the sandbox is deleted.

## Example Output

```
Creating sandbox...
Installing OpenCode in sandbox...
Press Ctrl+C at any time to exit.
User: make a lunar lander web app
Thinking...
📝 Add /home/daytona/index.html
📝 Add /home/daytona/style.css
🔨 ✓ Run: ...
Built a playable lunar lander experience...

User:
Cleaning up...
```

## License

See the main project LICENSE file for details.

## References

- [OpenCode SDK](https://opencode.ai/docs/sdk/)
- [Daytona Documentation](https://www.daytona.io/docs)
