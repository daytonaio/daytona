# LangGraph Plan-and-Execute Data Agent (LangGraph + Daytona)

## Overview

This example demonstrates how to build a [LangGraph](https://langchain-ai.github.io/langgraph/) **plan-and-execute** data agent that runs an end-to-end ETL + analytical-SQL workflow inside a [Daytona](https://daytona.io) sandbox. The LLM plans the work as an ordered list of atomic steps, writes Python for each step, and the sandbox executes it. The graph is hand-wired as a six-node state machine (`provision -> plan -> execute -> check -> summarize -> cleanup`). The `check` node sits between `execute` and the next stage and inspects the result: on success it routes forward (advance to the next plan step, or to `summarize` when the plan is complete), on failure it routes back to `execute` with the failing code as context so the LLM can retry. Every node and edge in the agent's control flow is explicit and inspectable.

In this example, the agent profiles the maintenance health of the public [`langchain-ai/langgraph`](https://github.com/langchain-ai/langgraph) GitHub repository: extract issues + pull requests from the public GitHub REST API, transform and normalize them, load them into a SQLite database in the sandbox, run three analytical SQL queries, and summarize findings. The LangGraph state machine handles planning, sequential execution, per-step retries on failure, and cleanup.

## Features

- **Six-node plan-and-execute state machine:** Explicit nodes for sandbox provisioning, planning, code execution, retry checking, summarization, and cleanup
- **LLM-emitted structured plan:** The planner uses Pydantic structured output to emit an ordered `list[str]` of atomic plan steps the executor implements one at a time
- **Per-step retry with bounded attempts:** A deterministic `check` node tracks `(attempts, max_attempts, last_error, last_code)` and routes a failing step back to `execute` with the failure context, up to `max_attempts` times before giving up
- **Persistent interpreter context across the entire run:** Variables, imports, and files set up in one plan step are still in scope in the next (the agent uses the SDK's `code_interpreter` API, which keeps a shared interpreter context alive across calls)
- **Custom graph state beyond messages:** Demonstrates `TypedDict` state with task-specific fields (`plan`, `step_idx`, `step_outputs`, `step_codes`, `last_error`, ...), the canonical LangGraph pattern for non-chat workflows
- **Guaranteed cleanup:** The `cleanup` node always runs after `summarize`, deleting the sandbox regardless of whether the agent succeeded or gave up

## Requirements

- **Python:** Version 3.10 or higher

> [!TIP]
> It's recommended to use a virtual environment (`venv` or `poetry`) to isolate project dependencies.

## Environment Variables

- `DAYTONA_API_KEY`: Required for access to Daytona sandboxes. Get it from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)
- `ANTHROPIC_API_KEY`: Required for Claude model access. Get it from [Anthropic Console](https://console.anthropic.com/)

Copy `.env.example` to `.env` and fill in your API keys before running.

## Getting Started

1. Create and activate a virtual environment:

   ```bash
   python3.10 -m venv venv
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   ```

2. Install dependencies:

   ```bash
   pip install -U langgraph langchain-core langchain-anthropic daytona pydantic python-dotenv
   ```

3. Set your API keys in `.env`:

   ```bash
   DAYTONA_API_KEY=your_daytona_api_key
   ANTHROPIC_API_KEY=your_anthropic_api_key
   ```

4. Run the example:

   ```bash
   python main.py
   ```

## How It Works

1. **`provision`**: `Daytona().create()` provisions a fresh sandbox and stores it in graph state
2. **`plan`**: the LLM (called via `with_structured_output(Plan)`) emits an ordered `list[str]` of plan steps into `state["plan"]`
3. **`execute`**: given the original user request, the full plan, the current step index, all prior steps' code and stdout, and any previous-attempt error, the LLM emits Python for the current step; the graph runs it in the sandbox via `sandbox.code_interpreter.run_code(code)`, which executes inside the sandbox's persistent shared interpreter context so imports and variables from earlier steps remain in scope
4. **`check`**: deterministic. On success it advances `step_idx` and resets `attempts`. On failure it increments `attempts`. Then a conditional edge routes back to `execute` (retry or advance) or forward to `summarize` (done or gave up after `max_attempts`)
5. **`summarize`**: the LLM reads every successful step's stdout and produces a factual natural-language answer to the user's request
6. **`cleanup`**: `sandbox.delete()` runs unconditionally after `summarize`

## Example Output

The agent typically emits a 6-step plan, executes each step in the persistent interpreter context, runs three analytical SQL queries, and summarizes. Because variables and imports survive across steps, an `import requests` in step 1 is still in scope when step 2 calls `requests.get(...)`, so well-formed code rarely needs the retry path. The canonical run below completes all six steps on the first attempt:

```
[provision] creating Daytona sandbox...
[provision] sandbox ready (id=b9cf758d-9b93-4117-96b3-9a406c86b1b8)

[plan] asking the LLM for a multi-step plan...
[plan] 6 step(s):
  1. Install needed packages and import requests, sqlite3, json
  2. Fetch the 100 most recently updated issues and PRs from the
     langchain-ai/langgraph /issues and /pulls endpoints; store JSON responses
  3. Create a SQLite database (langgraph.db) with two tables, define schemas,
     and insert the fetched data
  4. SQL: PR merge rate among closed PRs
  5. SQL: top 5 PR authors by count with personal merge rates
  6. SQL: most-commented currently-open issue

[execute] step 1/6 attempt 1/3: Install needed packages ...
[execute] step OK.
[check] step 1 done; advancing to step 2

[execute] step 2/6 attempt 1/3: Fetch the issues and PRs ...
[execute] step OK. stdout: Fetched 100 issues, 100 PRs
[check] step 2 done; advancing to step 3

[execute] step 3/6 attempt 1/3: Create a SQLite database, insert data ...
[execute] step OK. stdout: Created tables issues, pull_requests; inserted 40 issues, 100 PRs
[check] step 3 done; advancing to step 4

... (steps 4-6 succeed) ...

[summarize] asking the LLM for a final answer...
[cleanup] deleting sandbox ...
[cleanup] done

============================================================
FINAL ANSWER
============================================================
PR Merge Rate: 96/100 closed PRs merged = 96.0%.

Top 5 PR authors (total PRs, personal merge rate):
  nfcampos       40   100.00%
  hinthornw      18   100.00%
  hwchase17      15    93.33%
  rlancemartin   10   100.00%
  baskaryan       4   100.00%

Most-commented open issue:
  "Long tool calls (~180s+) silently re-executed from checkpoint on LangGraph Cloud"
  25 comments, opened by MarioAlessandroNapoli on 2026-04-05.
```

The retry path still exists for genuine code failures (syntax errors, runtime exceptions, the LLM hallucinating a non-existent method, a bad SQL query). When a step fails, the `execute` node serializes `result.error` into `state["last_error"]`; `check` sees the populated error, increments `attempts`, and `route_from_check` sends control back to `execute` with the failing code and traceback included in the next prompt, up to `max_attempts` times.

## License

See the main project LICENSE file for details.

## References

- [LangGraph](https://langchain-ai.github.io/langgraph/)
- [Daytona](https://daytona.io)
