# Run DSPy RLMs on Daytona

## Overview

DSPy's [RLM (Recursive Language Model)](https://dspy.ai/) module gives an LLM a Python REPL. The LLM writes code, executes it, sees the output, and repeats — building up state across iterations until it calls `SUBMIT()` with a final answer. Within that code, the LLM can also call `llm_query()` to invoke sub-LLM reasoning (the "recursive" part), mixing procedural computation with natural-language understanding.

`DaytonaInterpreter` is a `CodeInterpreter` backend that runs all of this inside a Daytona cloud sandbox, so LLM-generated code never executes on the host.

## Features

- **Sandboxed REPL:** LLM-generated code runs in an isolated Daytona sandbox
- **Persistent state:** Variables and imports survive across iterations within a session
- **Sub-LLM calls:** `llm_query()` and `llm_query_batched()` are bridged into the sandbox, letting generated code invoke nested LLM reasoning
- **Custom tools:** Host-side Python functions (database queries, APIs, etc.) can be passed into the sandbox through a broker server
- **Typed SUBMIT:** Output fields can have explicit types so the LLM knows the expected schema
- **Context manager:** `DaytonaInterpreter` supports `with` for automatic cleanup

## Requirements

- **Python:** Version 3.10 or higher

## Environment Variables

- `DAYTONA_API_KEY`: Required for access to Daytona sandboxes. Get it from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)
- `OPENROUTER_API_KEY`, `OPENAI_API_KEY`, or `ANTHROPIC_API_KEY`: Required for your LLM provider (depending on which model you use)

## Getting Started

### Setup and Run

1. Create and activate a virtual environment:

```bash
python3.10 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

2. Install dependencies:

```bash
pip install -e .
```

To also run `demo.py` (which plots results with matplotlib), install with the demo extra:

```bash
pip install -e ".[demo]"
```

3. Set your API keys in `.env` (copy from `.env.example`):

```bash
cp .env.example .env
# Edit .env with your DAYTONA_API_KEY and LLM provider key
```

4. Run the example:

```bash
python demo.py
```

## How It Works

### The REPL loop

Each RLM call runs an iterative loop:

1. RLM prompts the LLM with the task inputs and the REPL history so far
2. The LLM responds with reasoning and a Python code snippet
3. The code executes in the Daytona sandbox
4. The output (stdout, errors, or a final result) is appended to the REPL history
5. Steps 1–4 repeat until the code calls `SUBMIT()` or the iteration limit is reached

State persists across iterations — variables, imports, and function definitions all carry over. This lets the LLM explore data incrementally, inspect intermediate results with `print()`, and refine its approach before committing a final answer.

```
         ┌──────────────────────────────────┐
         │            DSPy RLM              │
         │                                  │
         │  Prompt LLM (inputs + history)   │
         │         │                        │
         │         ▼                        │
         │  LLM writes Python code          │
         │         │                        │
         │         ▼                        │
         │  Execute in sandbox ─────────────┼──▶ Daytona Sandbox
         │         │                        │    (persistent REPL)
         │         ▼                        │
         │  Append output to history        │
         │         │                        │
         │         ▼                        │
         │  SUBMIT() called? ──no──▶ loop   │
         │         │                        │
         │        yes                       │
         │         ▼                        │
         │  Return final answer             │
         └──────────────────────────────────┘
```

### Sub-LLM calls

The generated code has access to two built-in functions for invoking an LLM from within the REPL:

- `llm_query(prompt)` — send a single prompt, get a string back
- `llm_query_batched(prompts)` — send multiple prompts concurrently

This is what makes RLM "recursive": the LLM can write code that delegates semantic work to another LLM call, then processes the result with Python. For example:

```python
texts = [page1, page2, page3]
summaries = llm_query_batched([f"Summarize: {t}" for t in texts])
combined = "\n".join(summaries)
SUBMIT(answer=combined)
```

These functions execute on the host (they need LLM API access). `DaytonaInterpreter` bridges them into the sandbox through the broker, the same mechanism used for custom tools.

### Custom tools

You can pass host-side functions into the sandbox via the `tools` dict. The interpreter bridges them using a broker server that runs inside the sandbox:

1. A Flask server starts inside the sandbox on port 3000
2. For each tool, a wrapper function is injected into the sandbox that calls the broker over HTTP
3. The host polls the broker for pending requests, executes the function, and posts the result back

From the LLM's perspective, these look like regular Python functions.

## Usage Examples

### Basic: Reasoning with RLM

```python
import dspy
from dotenv import load_dotenv
from daytona_interpreter import DaytonaInterpreter

load_dotenv()

# Configure the LLM
lm = dspy.LM("openrouter/google/gemini-3-flash-preview")
dspy.configure(lm=lm)

# Create an RLM with the Daytona interpreter
interpreter = DaytonaInterpreter()

rlm = dspy.RLM(
    signature="question -> answer: str",
    interpreter=interpreter,
    verbose=True,
)

result = rlm(question="What is the sum of the first 10 prime numbers?")
print(result.answer)

interpreter.shutdown()
```

### With Custom Tools

Pass host-side functions into the sandbox so the LLM's generated code can call them:

```python
import json
import dspy
from dotenv import load_dotenv
from daytona_interpreter import DaytonaInterpreter

load_dotenv()

lm = dspy.LM("openrouter/google/gemini-3-flash-preview")
dspy.configure(lm=lm)

# Define tools that run on the host
def search_knowledge_base(query: str) -> str:
    """Search a knowledge base and return relevant results."""
    # Replace with your actual search logic
    return json.dumps({"results": [f"Result for: {query}"]})

# Pass tools to the interpreter
interpreter = DaytonaInterpreter(tools={"search_knowledge_base": search_knowledge_base})

rlm = dspy.RLM(
    signature="question -> answer: str",
    interpreter=interpreter,
    verbose=True,
)

result = rlm(question="Search for information about Python generators and summarize it.")
print(result.answer)

interpreter.shutdown()
```

Inside the sandbox, the LLM can call `search_knowledge_base(...)` like a regular function. The call is routed to the host through the broker. See [Custom tools](#custom-tools) for how this works.

## Key Concepts

### SUBMIT

`SUBMIT()` ends the REPL loop and returns a final answer. Its arguments match the output fields of the DSPy signature:

```python
SUBMIT(answer="The sum is 129")
```

If the signature has typed output fields, `SUBMIT` gets a typed signature in the sandbox so the LLM knows the expected schema:

```python
# Automatically generated:
def SUBMIT(answer: str):
    ...
```

If the LLM never calls `SUBMIT()` within the iteration limit, RLM falls back to extracting an answer from the REPL history.

### Broker

The broker is a small Flask server inside the sandbox that bridges function calls between the sandbox and the host. It handles both RLM's built-in `llm_query` / `llm_query_batched` and any custom tools you provide. It starts automatically when tools are present.

## License

See the main project LICENSE file for details.

## References

- [DSPy](https://dspy.ai/) — The framework for programming with foundation models
- [Daytona](https://www.daytona.io/) — Secure cloud development environments
- [Recursive Language Models](https://arxiv.org/abs/2512.24601) — Zhang, Kraska, Khattab
