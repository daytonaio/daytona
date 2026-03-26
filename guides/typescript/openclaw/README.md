# OpenClaw Daytona Sandbox

## Overview

This example runs [OpenClaw](https://openclaw.ai/), a general purpose AI assistant, inside a Daytona sandbox. You can interact with OpenClaw via its Control UI using a [Daytona preview link](https://www.daytona.io/docs/en/preview-and-authentication/#fetching-a-preview-link).

## Features

- **Secure sandbox execution:** OpenClaw runs in a controlled environment, along with any code or commands run by agents.
- **Multi-channel gateway:** Can connect to WhatsApp, Telegram, Discord, and more simultaneously.
- **Preview Control UI:** Use Daytona preview links to access the OpenClaw web dashboard with no local install.
- **Flexible LLM support:** Connect to Anthropic, OpenAI, and other providers; configure models via `openclaw.json` and `.env.sandbox`.

## Prerequisites

- **Node.js:** Version 18 or higher is required

## Environment Variables

To run this example, you need to set the following environment variables:

**`.env`** (used by the main script only):

- `DAYTONA_API_KEY`: Required for access to Daytona sandboxes. Get it from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)

**`.env.sandbox`** (available inside the OpenClaw sandbox):

- `ANTHROPIC_API_KEY`: Required for Claude. Get it from [Anthropic Console](https://console.anthropic.com/)
- Any other variables you add here are loaded into the sandbox environment

Create these files in the project directory (copy from `.env.example` and `.env.sandbox.example`).

## Getting Started

### Setup and Run

1. Install dependencies:

   ```bash
   npm install
   ```

2. Run the example:

   ```bash
   npm start
   ```

## How It Works

When this example is run, the agent follows the following workflow:

1. A new Daytona sandbox is created (using the `daytona-medium` snapshot with OpenClaw preinstalled).
2. OpenClaw is configured with your `openclaw.json` and `.env.sandbox` secrets.
3. The OpenClaw gateway starts inside the sandbox.
4. A Daytona preview link is shown pointing to the OpenClaw Control UI.
5. When the script is terminated (Ctrl+C), the sandbox is deleted—unless `PERSIST_SANDBOX` is set to `true`, in which case the sandbox is left running.

## Example Output

```
Creating Daytona sandbox...
Configuring OpenClaw...
Starting OpenClaw...
(Ctrl+C to shut down and delete the sandbox)

🔗 Secret link to Control UI: https://18789-898f722f-76fc-4ec6-85ca-a82bb30f3d72.proxy.daytona.works?token=7e38c7347437c5642c57bc769f630e53fe118e001d7b6c6c

OpenClaw logs:
--------------------------------
(node:131) [DEP0040] DeprecationWarning: The `punycode` module is deprecated. Please use a userland alternative instead.
(Use `node --trace-deprecation ...` to show where the warning was created)
│
◇  Doctor changes ────────────────────────╮
│                                         │
│  WhatsApp configured, not enabled yet.  │
│                                         │
├─────────────────────────────────────────╯
```

Open the provided URL in your browser to interact with the OpenClaw agent via the Control UI.

## Configuration

### Script configuration

You will find several constants in `src/index.ts` which control the behavior of the script:

| Constant | Default | Description |
|----------|---------|-------------|
| `OPENCLAW_PORT` | 18789 | OpenClaw Gateway and Control UI port |
| `SHOW_LOGS` | true | Stream OpenClaw stdout/stderr to the terminal. |
| `MAKE_PUBLIC` | true | Expose the sandbox for public internet access. |
| `PERSIST_SANDBOX` | true | When true, the sandbox is not deleted when the script exits. |
| `DAYTONA_SNAPSHOT` | daytona-medium | Sandbox image with OpenClaw preinstalled. |

### OpenClaw Configuration

You can tailor OpenClaw to your setup by editing `openclaw.json`. The script combines this file with built-in defaults and an authorization token, and writes the result to `~/.openclaw/openclaw.json` inside the sandbox.

The default configuration is:

```json
{
  "agents": {
    "defaults": {
      "model": { "primary": "anthropic/claude-sonnet-4-5" }
    }
  },
  "auth": {
    "profiles": {
      "anthropic:api": { "provider": "anthropic", "mode": "api_key" }
    },
    "order": { "anthropic": ["anthropic:api"] }
  },
  "channels": {
    "whatsapp": { "allowFrom": [] }
  }
}
```

In order to accept WhatsApp messages, the numbers of the allowed senders need to be added to the allowFrom list.

You can extend it with additional sections:

| Section | Purpose |
|--------|---------|
| `agents.defaults` | [Model, workspace path, timeouts, sandbox](https://docs.openclaw.ai/gateway/configuration-reference#agent-defaults) |
| `agents.list` | [Multiple agents with different names and tools](https://docs.openclaw.ai/gateway/configuration-reference#agents-list-per-agent-overrides) |
| `auth` | [API keys and OAuth for Claude, GPT, etc.](https://docs.openclaw.ai/gateway/configuration-reference#auth-storage) |
| `channels` | [Connect messaging apps and control who can message](https://docs.openclaw.ai/gateway/configuration-reference#channels) |
| `gateway` | [Port, authentication, Control UI access](https://docs.openclaw.ai/gateway/configuration-reference#gateway) |
| `models` | [Add OpenRouter, local models, other providers](https://docs.openclaw.ai/gateway/configuration-reference#custom-providers-and-base-urls) |
| `session` | [How conversations are grouped and reset](https://docs.openclaw.ai/gateway/configuration-reference#session) |
| `tools` | [What the agent can do (code, web, browser)](https://docs.openclaw.ai/gateway/configuration-reference#tools) |

For a full reference see [Configuration Reference](https://docs.openclaw.ai/gateway/configuration-reference) and [Configuration Examples](https://docs.openclaw.ai/gateway/configuration-examples).

## License

See the main project LICENSE file for details.

## References

- [OpenClaw Documentation](https://docs.openclaw.ai/)
- [Daytona Documentation](https://www.daytona.io/docs)
