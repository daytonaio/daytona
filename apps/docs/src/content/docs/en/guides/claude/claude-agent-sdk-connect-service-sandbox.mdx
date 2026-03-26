---
title: Build a Two-Agent Coding System with Claude and Daytona
description: Step-by-step guide to building a dual-agent coding system using Claude Agent SDK and Daytona sandboxes.
---

This guide demonstrates how to run a two-agent autonomous coding system using the [Claude Agent SDK](https://platform.claude.com/docs/en/agent-sdk/overview) and Daytona sandboxes. The system consists of a **Project Manager Agent** (local) and a **Developer Agent** (in-sandbox), enabling advanced delegation, planning, and secure code execution.

The Project Manager Agent runs locally and uses the basic Anthropic interface with the `claude-sonnet-4-20250514` model for high-level planning and task delegation. The Developer Agent runs inside the Daytona sandbox and is created using the Claude Agent SDK, which leverages Claude Code for advanced coding and automation capabilities. This architecture separates high-level planning from low-level code execution for more robust automation.

A key advantage of this approach is its **extensibility**: you can easily replace the Project Manager Agent with your own custom orchestrator logic, or even another agent, making the system highly reusable and adaptable to a wide range of advanced automation and coordination use cases.

---

### 1. Workflow Overview

When the main module is launched, a Daytona sandbox is created for the Developer Agent, and a Project Manager Agent is initialized locally. Interaction with the system occurs via a command line chat interface. The Project Manager Agent receives prompts, plans the workflow, and delegates coding tasks to the Developer Agent. The Developer Agent executes tasks in the sandbox and streams results back to the Project Manager, who reviews and coordinates further actions. All logs and outputs from both agents are streamed in real time to the terminal, providing full visibility into the process as it is managed by the Project Manager Agent.

The Developer Agent can also host web apps and provide preview links using [Daytona Preview Links](https://www.daytona.io/docs/en/preview-and-authentication/). The Project Manager Agent will present these links and summarize the results for you.

You can continue interacting with the system until you are finished. When you exit the program, the sandbox is deleted automatically.

---

### 2. Project Setup

#### Clone the Repository

First, clone the daytona [repository](https://github.com/daytonaio/daytona.git) and navigate to the example directory:

```bash
git clone https://github.com/daytonaio/daytona.git
cd daytona/guides/typescript/anthropic/multi-agent-claude-sdk
```

#### Configure Environment

To run this example, you need to set the following environment variables:

- `DAYTONA_API_KEY`: Required for access to Daytona sandboxes. Get it from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)
- `ANTHROPIC_API_KEY`: Required for the **Project Manager Agent** (runs locally). Get it from [Claude Developer Platform](https://console.anthropic.com/settings/keys)
- `SANDBOX_ANTHROPIC_API_KEY`: **Optional** for the **Developer Agent** (runs in sandbox). If not provided, defaults to using `ANTHROPIC_API_KEY`. Get it from [Claude Developer Platform](https://console.anthropic.com/settings/keys)

Copy `.env.example` to `.env` and add your keys:

```bash
DAYTONA_API_KEY=your_daytona_key
ANTHROPIC_API_KEY=your_anthropic_key
SANDBOX_ANTHROPIC_API_KEY=your_anthropic_key
```

:::tip[Agent API Key Options]
You can use a single `ANTHROPIC_API_KEY` for both agents, or provide a separate `SANDBOX_ANTHROPIC_API_KEY` for billing/tracking purposes.
:::

:::caution[API Key Security]
The `SANDBOX_ANTHROPIC_API_KEY` is passed into the Daytona sandbox environment and is accessible to any code executed inside the sandbox.
:::

#### Local Usage

:::note[Node.js Version]
Node.js 18 or newer is required to run this example. Please ensure your environment meets this requirement before proceeding.
:::

Install dependencies:

```bash
npm install
```

Run the two-agent system:

```bash
npm run start
```

The system will start and wait for your prompt.

---

### 3. Example Usage

Example chat session:

```
$ npm run start
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

---


### 4. Understanding the Agent Architecture

This system is composed of two collaborating agents, each with a distinct role and implementation. Below is a step-by-step guide to how each agent operates and interacts within the workflow.

#### Project Manager Agent (Terminal Orchestration)

1. **User Interaction:**
  - All user interaction occurs via the terminal with the Project Manager Agent.
  - The Project Manager Agent is configured with a system prompt that defines its role and maintains the full conversation history.
2. **Awareness of Developer Agent:**
  - The Project Manager Agent knows that a Developer Agent is available inside a Daytona sandbox and can be invoked as needed.
3. **Task Delegation:**
  - When the Project Manager Agent determines that a coding task should be delegated, it encapsulates the task within `<developer_task>` tags in its response.
  - The system parses these tags and, when present, invokes the Developer Agent with the specified task.
4. **Iterative Workflow:**
  - This process can repeat multiple times, with the Project Manager Agent reasoning about progress and delegating further tasks as needed.
5. **Session Completion:**
  - When the Project Manager Agent determines the overall task is complete, it outputs `TASK_COMPLETE`, which signals the system to terminate the session.

#### Developer Agent (Sandbox Execution)

1. **Provisioning:**
  - The Developer Agent is provisioned inside a Daytona sandbox and is responsible for executing coding tasks.
2. **SDK Installation:**
  - The system installs the Claude Agent SDK in the sandbox by running `pip install` (see [process execution](/docs/en/process-code-execution#process-execution)).
3. **Interpreter Context:**
  - A new [code interpreter context](/docs/en/process-code-execution#stateful-code-interpreter) is created for isolated execution.
4. **Script Upload:**
  - The coding agent script is uploaded to the sandbox using [file uploading](/docs/file-system-operations#uploading-a-single-file).
5. **SDK Initialization:**
  - The Claude Agent SDK is initialized in the interpreter context (e.g., `import coding_agent`).
6. **Task Execution:**
  - When a `<developer_task>` is received, the system sends the task to the Developer Agent by running a Python command in the interpreter context:
    ```typescript
    const result = await sandbox.codeInterpreter.runCode(
     `coding_agent.run_query_sync(os.environ.get('PROMPT', ''))`,
     {
      context: ctx,
      envs: { PROMPT: task },
      onStdout,
      onStderr,
     }
    );
    ```
  - The Developer Agent executes the task, streams output, and returns results to the Project Manager Agent for review and further coordination.

---

### 5. Customization

You can customize the Project Manager Agent's behavior by modifying the system prompt in `src/index.ts`. The current implementation:

- Uses `<developer_task>` tags for delegation
- Automatically reviews Developer Agent outputs
- Says "TASK_COMPLETE" when finished

---

### 6. Cleanup

When you exit the main program, the Daytona sandbox and all files are automatically deleted.

---

**Key advantages:**

- Secure, isolated execution in Daytona sandboxes
- Hierarchical agent architecture for robust automation
- Extensible and reusable architecture
- Automatic dev server detection and live preview links
- Multi-language and full-stack support
- Simple setup and automatic cleanup
