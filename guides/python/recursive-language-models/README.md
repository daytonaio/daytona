# Run Recursive Language Models on Daytona

## Overview

This guide demonstrates how to implement a recursive language model (RLM) agent system built on Daytona sandboxes, based on the approach pioneered in [Recursive Language Models](https://arxiv.org/abs/2512.24601) (Zhang, Kraska, Khattab). Unlike traditional single-agent approaches, agents can spawn sub-agents recursively, each in its own isolated sandbox with a fresh clone of the target repository.

The system enables tree-structured problem decomposition: a root agent can delegate subtasks to child agents, which can spawn their own children, creating a hierarchy of specialized workers collaborating on complex software engineering tasks.

## Features

- **Sandboxed code execution:** Each agent runs in an isolated Daytona sandbox with a fresh repository clone
- **Recursive agent spawning:** Agents spawn sub-agents via `rlm_query()`, each with their own sandbox
- **Parallel sub-agent execution:** `rlm_query_batched()` spawns multiple sub-agents concurrently using thread pools
- **Budget management:** Global sandbox limit (default: 25) shared across the entire agent tree
- **LLM agnostic:** LiteLLM integration enables any provider (OpenRouter, OpenAI, Anthropic, etc.)
- **Git patch output:** Root agents produce git patches as their final output
- **Interactive viewer:** Web-based D3.js visualization of agent execution trees

## Requirements

- **Python:** Version 3.10 or higher

## Environment Variables

- `DAYTONA_API_KEY`: Required for access to Daytona sandboxes. Get it from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)
- `LLM_API_KEY`: Required for your LLM provider via [LiteLLM](https://docs.litellm.ai/) (OpenRouter, OpenAI, Anthropic, etc.)

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

3. Set your API keys in `.env` (copy from `.env.example`):

```bash
cp .env.example .env
# Edit .env with your DAYTONA_API_KEY and LLM_API_KEY
```

4. Run the agent:

```bash
python run.py https://github.com/user/repo --prompt "Fix the bug in auth.py"
```

### CLI Options

- `repo` - GitHub repository URL (required, positional)
- `-p, --prompt` - Task prompt for the agent (required)
- `-b, --branch` - Branch name (optional)
- `--commit` - Specific commit SHA (optional)
- `-c, --config` - Path to YAML configuration file (default: `config.yaml`)
- `-o, --output` - Output file for the patch (default: stdout)
- `--verbose / --quiet` - Enable verbose output (default: verbose)

## Configuration

The script has several configurable parameters in `config.yaml`:

### Sandbox Settings

- `max_sandboxes`: Maximum total sandboxes that can be created over the entire run (default: 50)
- `global_timeout`: Total timeout in seconds for the entire run (default: 1800 = 30 minutes)

### Model Settings

- `model.name`: The LLM model to use in LiteLLM format (default: `openrouter/google/gemini-3-flash-preview`)

### Agent Settings

- `max_iterations`: Maximum iterations per agent before timeout (default: 50)
- `result_truncation_limit`: Maximum characters for sub-agent results (default: 20000)

## How It Works

The system runs a recursive agent architecture where each agent operates in its own sandbox.

1. **Initialization:** Load config and create root agent (depth=0) with a Daytona sandbox containing a fresh clone of the target repository
2. **Iteration loop:** The agent runs an iteration loop: LLM call → extract Python code blocks → execute in REPL
3. **Sub-agent spawning:** When code calls `rlm_query(task)`, a new sub-agent is created with its own sandbox
4. **Recursive delegation:** Sub-agents can spawn their own sub-agents (unlimited depth)
5. **Result propagation:** Results flow back up the tree as sub-agents complete
6. **Completion:** Root agent calls `FINAL()` or times out, producing a git patch of all changes
7. **Cleanup:** Sandboxes are deleted and results are logged to JSON

## Agent Hierarchy Example

```
Root Agent (depth=0)
├── Sub-Agent A (depth=1)
│   ├── Sub-Agent A1 (depth=2)
│   └── Sub-Agent A2 (depth=2)
└── Sub-Agent B (depth=1)
    ├── Sub-Agent B1 (depth=2)
    └── Sub-Agent B2 (depth=2)
```

## Key Functions Available in Agent Code

| Function | Description |
| -------- | ----------- |
| `rlm_query(task)` | Spawn a single sub-agent with the given task, returns result string |
| `rlm_query_batched(tasks)` | Spawn multiple sub-agents in parallel, returns list of result strings |
| `FINAL(answer)` | Submit final result (root agent: triggers git patch extraction) |
| `FINAL_VAR(var_name)` | Submit the value of a variable as the result |
| `edit_file(path, old, new)` | Edit a file with syntax validation |

Variables and imports persist between iterations within the same agent.

## Viewer

Start a local server and open the viewer to visualize agent execution:

```bash
python -m http.server 8000
# Open http://localhost:8000/viewer/
```

The viewer provides:

- Interactive tree visualization of the agent hierarchy
- Iteration details with code and output for each agent

## Output

After running, results are saved in the `results/` directory:

- `{run_id}.detail.json`: Full agent tree with all iterations, code blocks, and outputs
- `index.json`: Index of all runs for the viewer
- Final output: Git patch printed to stdout (or saved to `-o` file)

## License

See the main project LICENSE file for details.

## References

- [Recursive Language Models](https://arxiv.org/abs/2512.24601) - Zhang, Kraska, Khattab
- [LiteLLM](https://docs.litellm.ai/)
