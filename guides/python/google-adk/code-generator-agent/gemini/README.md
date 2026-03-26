# Code Generator Agent Example (Google ADK + Daytona)

## Overview

This example demonstrates how to build a [Google ADK](https://google.github.io/adk-docs/) agent that generates and verifies code using [Daytona](https://daytona.io) sandboxes. The agent uses the `DaytonaPlugin` to execute code in an isolated environment, enabling automated code generation workflows with built-in testing.

In this example, the agent is tasked with writing a TypeScript `groupBy` function that takes an array and a key function, then groups array elements by the key. The agent generates the implementation, creates test cases, executes them in the sandbox, and iterates until all tests pass before returning the verified code.

## Features

- **Secure sandbox execution:** All code runs in isolated Daytona sandboxes
- **Multi-language support:** Generate code in Python, JavaScript, or TypeScript
- **Automatic testing:** Agent creates and runs tests to verify implementations
- **Iterative refinement:** Agent fixes code until tests pass before responding
- **Natural language interface:** Describe your function in plain English

## Requirements

- **Python:** Version 3.10 or higher is required

> [!TIP]
> It's recommended to use a virtual environment (`venv` or `poetry`) to isolate project dependencies.

## Environment Variables

To run this example, you need to set the following environment variables:

- `DAYTONA_API_KEY`: Required for access to Daytona sandboxes. Get it from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)
- `GOOGLE_API_KEY`: Required for Gemini model access. Get it from [Google AI Studio](https://aistudio.google.com/apikey)

See the `.env.example` file for the exact structure. Copy `.env.example` to `.env` and fill in your API keys before running.

## Getting Started

Before proceeding, complete the following steps:

1. Ensure Python 3.10 or higher is installed
2. Copy `.env.example` to `.env` and add your API keys

### Setup and Run

1. Create and activate a virtual environment:

   ```bash
   python3.10 -m venv venv
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   ```

2. Install dependencies:

   ```bash
   pip install -U google-adk daytona-adk python-dotenv
   ```

3. Run the example:

   ```bash
   python main.py
   ```

## Configuration

- **Agent Instruction:** The `AGENT_INSTRUCTION` constant in `main.py` defines the agent's behavior. You can customize this to change how the agent approaches code generation and testing.

## How It Works

When you run the example, the agent follows this workflow:

1. **Receive Request:** The agent receives your natural language description of a function
2. **Generate Code:** Agent writes the function implementation in the specified language
3. **Create Tests:** Agent generates test cases to verify the implementation
4. **Execute in Sandbox:** Code and tests run in an isolated Daytona sandbox
5. **Iterate:** If tests fail, agent fixes the code and re-executes until tests pass
6. **Return Result:** Once verified, the agent returns only the working function code
7. **Cleanup:** Sandbox resources are automatically cleaned up

## Example Output

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

## API Reference

For the complete API reference, see the [daytona-adk documentation](https://github.com/daytonaio/daytona-adk-plugin#available-tools).

## License

See the main project LICENSE file for details.

## References

- [Google ADK Documentation](https://google.github.io/adk-docs/)
- [Daytona ADK Plugin](https://pypi.org/project/daytona-adk/)
- [Daytona](https://daytona.io)
