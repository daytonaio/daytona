---
title: Generate Verified Code With Google ADK Agent
description: Build Google ADK agents that generate and test code using Daytona's isolated sandbox environment.
---

import { TabItem, Tabs } from '@astrojs/starlight/components'

This guide demonstrates how to use the `DaytonaPlugin` for Google ADK to build an agent that generates, tests, and verifies code in a secure sandbox environment. The plugin enables agents to execute Python, JavaScript, and TypeScript code, run shell commands, and manage files within isolated Daytona sandboxes.

In this example, we build a code generator agent that takes a natural language description of a function, generates the implementation in TypeScript, creates test cases, executes them in the sandbox, and iterates until all tests pass before returning the verified code.

---

### 1. Workflow Overview

You describe the function you want in plain English, specifying the language (Python, JavaScript, or TypeScript). The agent generates the implementation, writes tests for it, and executes everything in a Daytona sandbox. If tests fail, the agent automatically fixes the code and re-runs until all tests pass. Only then does it return the verified, working code.

The key benefit: you receive code that has already been tested and verified, not just generated.

### 2. Project Setup

#### Clone the Repository

Clone the Daytona repository and navigate to the example directory:

```bash
git clone https://github.com/daytonaio/daytona
cd daytona/guides/python/google-adk/code-generator-agent/gemini
```

#### Install Dependencies

:::note[Python Version Requirement]
This example requires **Python 3.10 or higher**. It's recommended to use a virtual environment (e.g., `venv` or `poetry`) to isolate project dependencies.
:::

Install the required packages for this example:

<Tabs>
  <TabItem label="Python" icon="seti:python">
    ```bash
    pip install -U google-adk daytona-adk python-dotenv
    ```

    The packages include:
    - `google-adk`: Google's Agent Development Kit for building AI agents
    - `daytona-adk`: Provides the `DaytonaPlugin` that enables secure code execution in Daytona sandboxes
    - `python-dotenv`: Used for loading environment variables from `.env` file
  </TabItem>
</Tabs>

#### Configure Environment

Get your API keys and configure your environment:

1. **Daytona API key:** Get it from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)
2. **Google API key:** Get it from [Google AI Studio](https://aistudio.google.com/apikey)

Create a `.env` file in your project:

```bash
DAYTONA_API_KEY=dtn_***
GOOGLE_API_KEY=***
```

### 3. Understanding the Core Components

Before diving into the implementation, let's understand the key components we'll use:

#### Google ADK Components

- **Agent**: The AI model wrapper that processes requests and decides which tools to use. It receives instructions, has access to tools, and generates responses.
- **App**: A top-level container that bundles agents with plugins into a single configuration unit. It provides centralized management for shared resources and defines the root agent for your workflow.
- **InMemoryRunner**: The execution engine that runs agents and manages conversation state. It orchestrates the event-driven execution loop, handles message processing, and manages services like session history and artifact storage.

:::note[Running the Agent]
There are two ways to run Google ADK agents: using the `App` class with `InMemoryRunner`, or using `InMemoryRunner` directly with just an agent. The `App` serves as a configuration container that bundles agents with plugins, while the `Runner` handles actual execution and lifecycle management. This guide uses the `App` approach for cleaner organization of agents and plugins.
:::

#### Daytona Plugin

The `DaytonaPlugin` provides tools that allow the agent to:
- Execute code in Python, JavaScript, or TypeScript
- Run shell commands
- Upload and read files
- Start long-running background processes

All operations happen in an isolated sandbox that is automatically cleaned up when done.

### 4. Initialize Environment and Imports

First, we set up our imports and load environment variables:

<Tabs>
  <TabItem label="Python" icon="seti:python">
    ```python
    import asyncio
    import logging

    from dotenv import load_dotenv
    from google.adk.agents import Agent
    from google.adk.apps import App
    from google.adk.runners import InMemoryRunner

    from daytona_adk import DaytonaPlugin

    load_dotenv()

    logging.basicConfig(level=logging.DEBUG)
    ```

    **What each import does:**
    - `asyncio`: Required for running the async ADK runner
    - `logging`: Enables debug output to see agent reasoning
    - `load_dotenv`: Loads API keys from your `.env` file
    - `Agent`, `App`, `InMemoryRunner`: Core Google ADK components
    - `DaytonaPlugin`: Provides sandbox execution tools to the agent

    **Logging configuration:**
    The `logging.basicConfig(level=logging.DEBUG)` line configures Python's logging to show detailed debug output. You can adjust the logging level by passing different values:
    - `logging.DEBUG`: Most verbose, shows all internal operations including DaytonaPlugin sandbox creation and tool invocations
    - `logging.INFO`: Shows informational messages about agent progress
    - `logging.WARNING`: Shows only warnings and errors
    - `logging.ERROR`: Shows only errors

    :::tip[Behind the Scenes]
    With `DEBUG` level logging enabled, you can see the DaytonaPlugin's internal operations, including when the sandbox is created, when the `execute_code_in_daytona` tool is invoked, and when cleanup occurs. The plugin's `plugin_name` (configurable, defaults to `daytona_plugin`) appears in these log messages, making it easy to trace plugin activity.
    :::
  </TabItem>
</Tabs>

### 5. Define the Response Extractor

The ADK runner returns a list of events from the agent's execution. We need a helper function to extract the final text response:

<Tabs>
  <TabItem label="Python" icon="seti:python">
    ```python
    def extract_final_response(response: list) -> str:
        """Extract the final text response from a list of ADK events."""
        for event in reversed(response):
            text_parts = []

            if hasattr(event, "text") and event.text:
                return event.text
            if hasattr(event, "content") and event.content:
                content = event.content
                if hasattr(content, "parts") and content.parts:
                    for part in content.parts:
                        if hasattr(part, "text") and part.text:
                            text_parts.append(part.text)
                    if text_parts:
                        return "".join(text_parts)
                if hasattr(content, "text") and content.text:
                    return content.text
            if isinstance(event, dict):
                text = event.get("text") or event.get("content", {}).get("text")
                if text:
                    return text

        return ""
    ```

    This function iterates through events in reverse order to find the last text response. It handles multiple possible event structures that the ADK may return.
  </TabItem>
</Tabs>

### 6. Define the Agent Instruction

The instruction is critical - it defines how the agent behaves. Our instruction enforces a test-driven workflow:

<Tabs>
  <TabItem label="Python" icon="seti:python">
    ```python
    AGENT_INSTRUCTION = """You are a code generator agent that writes verified, working code.
    You support Python, JavaScript, and TypeScript.

    Your workflow for every code request:
    1. Write the function
    2. Write tests for it
    3. EXECUTE the code in the sandbox to verify it works - do not skip this step
    4. If execution fails, fix and re-execute until tests pass
    5. Once verified, respond with ONLY the function (no tests)

    You must always execute code before responding. Never return untested code.
    Only include tests in your response if the user explicitly asks for them.
    """
    ```

    **Key aspects of this instruction:**
    - **Enforces execution**: The agent must run code in the sandbox before responding
    - **Iterative fixing**: If tests fail, the agent fixes and retries
    - **Controlled output**: By default, the final response contains only the working function. If you want to see the tests, include an instruction to return them in your prompt.
    - **Multi-language**: Supports Python, JavaScript, and TypeScript
  </TabItem>
</Tabs>

### 7. Configure the Daytona Plugin

Initialize the plugin that provides sandbox execution capabilities:

<Tabs>
  <TabItem label="Python" icon="seti:python">
    ```python
    plugin = DaytonaPlugin(
        labels={"example": "code-generator"},
    )
    ```

    **Configuration options:**
    - `labels`: Custom metadata tags for the sandbox (useful for tracking/filtering)
    - `api_key`: Daytona API key (defaults to `DAYTONA_API_KEY` env var)
    - `sandbox_name`: Custom name for the sandbox
    - `plugin_name`: Name displayed in logs when the plugin logs messages (defaults to `daytona_plugin`)
    - `env_vars`: Environment variables to set in the sandbox
    - `auto_stop_interval`: Minutes before auto-stop (default: 15)
    - `auto_delete_interval`: Minutes before auto-delete (disabled by default)
  </TabItem>
</Tabs>

### 8. Create the Agent

Create the agent with the Gemini model, our instruction, and the Daytona tools:

<Tabs>
  <TabItem label="Python" icon="seti:python">
    ```python
    agent = Agent(
        model="gemini-2.5-pro",
        name="code_generator_agent",
        instruction=AGENT_INSTRUCTION,
        tools=plugin.get_tools(),
    )
    ```

    **Parameters explained:**
    - `model`: The Gemini model to use for reasoning and code generation
    - `name`: Identifier for the agent
    - `instruction`: The behavioral guidelines we defined
    - `tools`: List of tools from the Daytona plugin that the agent can use
  </TabItem>
</Tabs>

### 9. Create the App and Runner

Bundle the agent and plugin into an App, then run it:

<Tabs>
  <TabItem label="Python" icon="seti:python">
    ```python
    app = App(
        name="code_generator_app",
        root_agent=agent,
        plugins=[plugin],
    )

    async with InMemoryRunner(app=app) as runner:
        prompt = "Write a TypeScript function called 'groupBy' that takes an array and a key function, and groups array elements by the key. Use proper type annotations."

        response = await runner.run_debug(prompt)

        final_response = extract_final_response(response)
        print(final_response)
    ```
  </TabItem>
</Tabs>

**What happens here:**
1. The `App` bundles the agent with the plugin for proper lifecycle management
2. `InMemoryRunner` is used as an async context manager (the `async with` statement). A context manager in Python automatically handles setup and cleanup - when the code enters the `async with` block, the runner initializes; when it exits (either normally or due to an error), the runner cleans up resources.
3. `run_debug` sends the prompt and returns all execution events
4. The sandbox is automatically deleted when the `async with` block exits - this cleanup happens regardless of whether the code completed successfully or raised an exception

### 10. Running the Example

Run the complete example:

<Tabs>
  <TabItem label="Python" icon="seti:python">
    ```bash
    python main.py
    ```
  </TabItem>
</Tabs>

#### Understanding the Agent's Execution Flow

When you run the code, the agent works through your request step by step. With `logging.DEBUG` enabled, you'll see detailed output including:

- **DaytonaPlugin operations**: Sandbox creation, tool invocations (`execute_code_in_daytona`), and cleanup
- **LLM requests and responses**: The prompts sent to Gemini and the responses received
- **Plugin registration**: Confirmation that the `daytona_plugin` was registered with the agent

Here's what the debug output reveals about each step:

**Step 1: Sandbox Creation**

```
DEBUG:daytona_adk.plugin:Daytona sandbox created: e38f8574-48ac-48f1-a0ff-d922d02b0fcb
INFO:google_adk.google.adk.plugins.plugin_manager:Plugin 'daytona_plugin' registered.
```

The DaytonaPlugin creates an isolated sandbox and registers itself with the agent.

**Step 2: Agent receives the request**

The agent receives your prompt and understands it needs to create a TypeScript `groupBy` function with proper type annotations.

**Step 3: Agent generates code and tests**

The agent writes both the implementation and test cases, then calls the `execute_code_in_daytona` tool:

```
DEBUG:google_adk.google.adk.models.google_llm:
LLM Response:
-----------------------------------------------------------
Function calls:
name: execute_code_in_daytona, args: {'code': "...", 'language': 'typescript'}
```

**Step 4: Code execution in sandbox**

```
DEBUG:daytona_adk.plugin:Before tool: execute_code_in_daytona
DEBUG:daytona_adk.tools:Executing typescript code (length: 1570 chars)
DEBUG:daytona_adk.tools:Code execution completed with exit_code: 0
DEBUG:daytona_adk.plugin:After tool: execute_code_in_daytona
```

The plugin executes the code in the isolated TypeScript environment and returns the result.

**Step 5: Agent iterates if needed**

If tests fail (exit_code != 0), the agent analyzes the error, fixes the code, and re-executes until all tests pass.

**Step 6: Agent returns verified code**

Once tests pass, the agent responds with only the working function. If you included an instruction to return tests in your prompt, the tests will also be included in the response.

**Step 7: Cleanup**

```
INFO:daytona_adk.plugin:Deleting Daytona sandbox...
INFO:daytona_adk.plugin:Daytona sandbox deleted.
INFO:google_adk.google.adk.runners:Runner closed.
```

When the context manager exits, the sandbox is automatically deleted.

#### Example Output

When the agent completes the task, you'll see output like:

````
AGENT RESPONSE:
------------------------------------------------------------
```typescript
function groupBy<T, K extends keyof any>(
  array: T[],
  keyFn: (item: T) => K
): Record<K, T[]> {
  return array.reduce((result, item) => {
    const key = keyFn(item);
    if (!result[key]) {
      result[key] = [];
    }
    result[key].push(item);
    return result;
  }, {} as Record<K, T[]>);
}
```
============================================================

App closed, sandbox cleaned up. Done!
````

The agent has already tested this code in the sandbox before returning it, so you can trust that the implementation works correctly.

#### Requesting Tests in the Response

If you want to see the tests that were executed in the sandbox, include an instruction to return them in your prompt:

```python
prompt = "Write a TypeScript function called 'groupBy' that takes an array and a key function, and groups array elements by the key. Use proper type annotations. Return the tests also in a separate code block"
```

With this prompt, the agent will return both the function and the tests:

````
```typescript
function groupBy<T, K extends keyof any>(
  array: T[],
  keyFn: (item: T) => K
): Record<K, T[]> {
  return array.reduce((result, item) => {
    const key = keyFn(item);
    if (!result[key]) {
      result[key] = [];
    }
    result[key].push(item);
    return result;
  }, {} as Record<K, T[]>);
}
```

```typescript
import { deepStrictEqual } from 'assert';

// Test case 1: Group by a property of an object
const array1 = [
  { id: 1, category: 'A' },
  { id: 2, category: 'B' },
  { id: 3, category: 'A' },
];
const result1 = groupBy(array1, (item) => item.category);
deepStrictEqual(result1, {
  A: [
    { id: 1, category: 'A' },
    { id: 3, category: 'A' },
  ],
  B: [{ id: 2, category: 'B' }],
});

// Test case 2: Group by length of strings
const array2 = ['apple', 'banana', 'cherry', 'date'];
const result2 = groupBy(array2, (item) => item.length);
deepStrictEqual(result2, {
  5: ['apple'],
  6: ['banana', 'cherry'],
  4: ['date'],
});

console.log('All tests passed!');
```
````

### 11. Complete Implementation

Here is the complete, ready-to-run example with additional output formatting for better readability:

<Tabs>
  <TabItem label="Python" icon="seti:python">
    ```python
    """Code Generator & Tester Agent Example."""

    import asyncio
    import logging

    from dotenv import load_dotenv
    from google.adk.agents import Agent
    from google.adk.apps import App
    from google.adk.runners import InMemoryRunner

    from daytona_adk import DaytonaPlugin

    load_dotenv()

    logging.basicConfig(level=logging.DEBUG)


    def extract_final_response(response: list) -> str:
        """Extract the final text response from a list of ADK events."""
        for event in reversed(response):
            text_parts = []

            if hasattr(event, "text") and event.text:
                return event.text
            if hasattr(event, "content") and event.content:
                content = event.content
                if hasattr(content, "parts") and content.parts:
                    for part in content.parts:
                        if hasattr(part, "text") and part.text:
                            text_parts.append(part.text)
                    if text_parts:
                        return "".join(text_parts)
                if hasattr(content, "text") and content.text:
                    return content.text
            if isinstance(event, dict):
                text = event.get("text") or event.get("content", {}).get("text")
                if text:
                    return text

        return ""


    AGENT_INSTRUCTION = """You are a code generator agent that writes verified, working code.
    You support Python, JavaScript, and TypeScript.

    Your workflow for every code request:
    1. Write the function
    2. Write tests for it
    3. EXECUTE the code in the sandbox to verify it works - do not skip this step
    4. If execution fails, fix and re-execute until tests pass
    5. Once verified, respond with ONLY the function (no tests)

    You must always execute code before responding. Never return untested code.
    Only include tests in your response if the user explicitly asks for them.
    """


    async def main() -> None:
        """Run the code generator agent example."""
        plugin = DaytonaPlugin(
            labels={"example": "code-generator"},
        )

        agent = Agent(
            model="gemini-2.5-pro",
            name="code_generator_agent",
            instruction=AGENT_INSTRUCTION,
            tools=plugin.get_tools(),
        )

        app = App(
            name="code_generator_app",
            root_agent=agent,
            plugins=[plugin],
        )

        async with InMemoryRunner(app=app) as runner:
            prompt = "Write a TypeScript function called 'groupBy' that takes an array and a key function, and groups array elements by the key. Use proper type annotations."

            print("\n" + "=" * 60)
            print("USER PROMPT:")
            print("=" * 60)
            print(prompt)
            print("-" * 60)

            response = await runner.run_debug(prompt)

            final_response = extract_final_response(response)
            print("\nAGENT RESPONSE:")
            print("-" * 60)
            print(final_response)
            print("=" * 60)

        print("\nApp closed, sandbox cleaned up. Done!")


    if __name__ == "__main__":
        asyncio.run(main())
    ```
  </TabItem>
</Tabs>

**Key advantages of this approach:**

- **Verified code:** Every response has been tested in a real execution environment
- **Secure execution:** Code runs in isolated Daytona sandboxes, not on your machine
- **Multi-language support:** Generate and test Python, JavaScript, or TypeScript
- **Automatic iteration:** Agent fixes issues until tests pass
- **Flexible output:** Returns only the working function by default, or includes tests if explicitly requested in the prompt

### 12. API Reference

For the complete API reference of the Daytona ADK plugin, including all available tools and configuration options, see the [daytona-adk documentation](https://github.com/daytonaio/daytona-adk-plugin#available-tools).
