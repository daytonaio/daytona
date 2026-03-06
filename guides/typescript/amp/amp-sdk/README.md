# Amp Code Coding Agent with Daytona

A coding agent powered by the [Amp Code CLI](https://ampcode.com/) running inside secure [Daytona sandboxes](https://www.daytona.io/).

## Prerequisites

- Node.js 18 or newer
- A Daytona API key from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)
- An Amp API key from [Amp Settings](https://ampcode.com/settings)
- **Amp paid credits** - Execute mode requires paid credits. [Add credits here](https://ampcode.com/pay)

## Setup

1. Install dependencies:

   ```bash
   npm install
   ```

2. Copy `.env.example` to `.env` and add your API keys:

   ```bash
   DAYTONA_API_KEY=your_daytona_key
   SANDBOX_AMP_API_KEY=your_amp_key
   ```

## Usage

Run the agent:

```bash
npm run start
```

The agent gets a Daytona-aware system prompt: sandbox context, the preview URL pattern, and instructions to always run server commands in the background with `&` as the last action so the user gets control back.

The agent will:

1. Create a Daytona sandbox
2. Install the Amp CLI in the sandbox
3. Start an interactive prompt loop using streaming JSON mode

Example session:

```
$ npm run start
Creating sandbox...
Installing Amp CLI...
Starting Amp Code...
Initializing agent...
Thinking...
Got it! I'm ready to help. What would you like to build or work on?

Agent ready. Press Ctrl+C at any time to exit.

User: say hello
Thinking...
Hello! 👋 How can I help you today?

User:
```

## Features

- Secure, isolated execution in Daytona sandboxes
- Amp CLI with streaming JSON output for real-time updates
- Automatic cleanup on exit

## How It Works

This example runs a single Amp process in the sandbox with `--execute`, `--stream-json`, and `--stream-json-input` for bidirectional PTY control:

1. Amp is started with `amp --dangerously-allow-all --execute --stream-json --stream-json-input -m smart` and kept running
2. User prompts are sent as JSON lines on stdin (Amp's stream-json-input format)
3. Amp outputs JSON lines for system, assistant, user (tool results), and result messages; tool usage and text are displayed in real time
4. Control returns when the assistant sends `end_turn`, Amp sends a `result` message, or a command ends with `&`. A 120s timeout prevents indefinite blocking

## Learn More

- [Amp Manual - CLI](https://ampcode.com/manual#cli)
- [Amp Manual - Streaming JSON](https://ampcode.com/manual#cli-streaming-json)
- [Daytona Documentation](https://www.daytona.io/docs)
