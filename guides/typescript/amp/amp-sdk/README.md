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
3. Start an interactive prompt loop using CLI-per-turn mode with thread continuity

Example session:

```
$ npm run start
Creating sandbox...
Installing Amp CLI...
Starting Amp Code...
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
- Thread-based session continuity across multiple turns
- Automatic cleanup on exit

## How It Works

This example uses a PTY (pseudo-terminal) to stream output from Amp CLI, running one command per user turn with thread-based session continuity:

1. A PTY is created for streaming output from Amp commands
2. First prompt: Run `amp --dangerously-allow-all --stream-json -m smart -x "prompt"` and capture the thread ID from the streaming JSON init message
3. Follow-up prompts: Run `amp --dangerously-allow-all --stream-json -m smart -x "prompt" threads continue <thread-id>`
4. If the thread ID isn't captured from the stream, fall back to parsing `amp threads list` text output
5. Each command streams JSON output for real-time display of assistant messages and tool usage

## Learn More

- [Amp Manual - CLI](https://ampcode.com/manual#cli)
- [Amp Manual - Streaming JSON](https://ampcode.com/manual#cli-streaming-json)
- [Daytona Documentation](https://www.daytona.io/docs)
