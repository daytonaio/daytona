# AG2 Bug Fixer Agent

## Overview

This example demonstrates how to build an AI-powered bug fixer using two [AG2](https://ag2.ai/) agents and [Daytona](https://daytona.io) sandboxes. A `bug_fixer` LLM agent analyzes broken code and proposes fixes, while a `code_executor` agent runs each fix attempt inside an isolated Daytona sandbox. If execution fails, the bug fixer sees the full error output and retries — looping until the code passes or the turn limit is reached.

The executor supports Python, JavaScript, TypeScript, and Bash — inferring the language automatically from the code block returned by the LLM.

## Features

- **Execution-verified fixes:** Every proposed fix is actually run in a sandbox — the agent only terminates when the code passes, not just when it looks correct
- **Secure execution:** Fix attempts run in isolated Daytona sandboxes, not on your machine
- **Multi-language support:** Python, JavaScript, TypeScript, and Bash — language is inferred automatically from the LLM's fenced code block
- **Iterative refinement:** If a fix fails, the agent sees the full error output and retries automatically
- **Automatic cleanup:** The sandbox is deleted as soon as `fix_bug` returns, regardless of outcome

## Requirements

- **Python:** Version 3.10 or higher

## Environment Variables

- `DAYTONA_API_KEY`: Required for access to Daytona sandboxes. Get it from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)
- `OPENAI_API_KEY`: Required for GPT-4o-mini access. Get it from [OpenAI Platform](https://platform.openai.com/api-keys)

## Getting Started

1. Create and activate a virtual environment:

```bash
python3.10 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

2. Install dependencies:

```bash
pip install "ag2[daytona,openai]" python-dotenv
```

3. Set your API keys in `.env`:

```bash
DAYTONA_API_KEY=your_daytona_api_key
OPENAI_API_KEY=your_openai_api_key
```

4. Run the example:

```bash
python main.py
```

## How It Works

1. `code_executor` sends the broken code to `bug_fixer` as the opening message
2. `bug_fixer` (GPT-4o-mini) analyzes it and replies with a fix wrapped in a fenced code block
3. AG2 detects the code block, runs it in a Daytona sandbox via `DaytonaCodeExecutor`, and sends the result back
4. If execution fails, `bug_fixer` sees the error output and proposes a new fix
5. The loop continues until the code passes or `max_turns` is reached
6. The sandbox is automatically deleted when `fix_bug` returns

## Examples

The script runs three examples — one per supported language:

- **Python** — postfix expression evaluator with swapped operands for `-` and `/`
- **JavaScript** — run-length encoder with wrong concatenation order in two places
- **TypeScript** — Kadane's max subarray algorithm using `Math.min` instead of `Math.max`

## Example Output

````
============================================================
Example 1: Python — Postfix Expression Evaluator Bug
============================================================
code_executor (to bug_fixer):

Fix this broken code:
...

>>>>>>>> USING AUTO REPLY...
bug_fixer (to code_executor):

```python
def eval_postfix(expression):
    ...
    elif token == '-':
        stack.append(a - b)  # Fixed order of operands for subtraction
    elif token == '/':
        stack.append(a // b)  # Fixed order of operands for division
    ...
```

>>>>>>>> USING AUTO REPLY...
>>>>>>>> EXECUTING CODE BLOCK (inferred language is python)...
code_executor (to bug_fixer):

exitcode: 0 (execution succeeded)
Code output: All postfix tests passed!

>>>>>>>> USING AUTO REPLY...
bug_fixer (to code_executor):

TERMINATE
````

## API Reference

For the complete API reference, see the [DaytonaCodeExecutor documentation](https://docs.ag2.ai/latest/docs/api-reference/autogen/coding/DaytonaCodeExecutor).

## License

See the main project LICENSE file for details.

## References

- [AG2](https://ag2.ai/)
- [AG2 DaytonaCodeExecutor](https://docs.ag2.ai/latest/docs/api-reference/autogen/coding/DaytonaCodeExecutor)
- [Daytona](https://daytona.io)
