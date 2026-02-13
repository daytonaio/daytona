---
title: Build deep Recursive Language Models
description: Implement recursive language model agents where each agent runs in its own isolated Daytona sandbox.
---

import { TabItem, Tabs } from '@astrojs/starlight/components'

This guide demonstrates how to build a recursive language model (RLM) agent system that uses Daytona sandboxes, based on the approach pioneered in [Recursive Language Models](https://arxiv.org/abs/2512.24601) (Zhang, Kraska, Khattab) and further explored by [Prime Intellect](https://www.primeintellect.ai/blog/rlm).

While the original paper and Prime Intellect's implementation focus on single-level recursion (depth=1), this guide extends the concept to **unlimited recursion depth** — agents can spawn sub-agents, which can spawn their own sub-agents, and so on. Each agent runs in its own isolated Daytona sandbox with a fresh clone of the target repository.

---

### 1. Workflow Overview

The system implements a recursive agent architecture where agents can delegate subtasks to child agents:

1. **Initialize**: Root agent receives a task and gets a Daytona sandbox with a fresh repository clone
2. **Iterate**: Agent runs a loop: LLM call → extract Python code → execute in REPL
3. **Delegate**: Code can call `rlm_query()` to spawn sub-agents, each with their own sandbox
4. **Aggregate**: Sub-agents return results; parent synthesizes findings and optionally runs more code
5. **Complete**: Root agent receives all sub-agent results, produces a git patch; all sandboxes are cleaned up

```
Root Agent (depth=0)
├── Sub-Agent A (depth=1)
│   ├── Sub-Agent A1 (depth=2)
│   └── Sub-Agent A2 (depth=2)
└── Sub-Agent B (depth=1)
    ├── Sub-Agent B1 (depth=2)
    └── Sub-Agent B2 (depth=2)
```

Each agent runs in its own isolated Daytona sandbox with a fresh repository clone, enabling parallel exploration.

### 2. Setup

#### Clone the Repository

Clone the [Daytona repository](https://github.com/daytonaio/daytona.git) and navigate to the example directory:

```bash
git clone https://github.com/daytonaio/daytona.git
cd daytona/guides/python/recursive-language-models
```

#### Create Virtual Environment

```bash
python3.10 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

#### Install Dependencies

```bash
pip install -e .
```

This installs:
- `daytona` - Daytona SDK for sandbox management
- `litellm` - Unified LLM interface for any provider
- `typer` - CLI framework
- `pyyaml` - Configuration parsing

#### Configure Environment

Get your Daytona API key from the [Daytona Dashboard](https://app.daytona.io/dashboard/keys) and create a `.env` file:

```bash
DAYTONA_API_KEY=your_daytona_api_key
LLM_API_KEY=your_llm_api_key
```

The `LLM_API_KEY` is used via [LiteLLM](https://docs.litellm.ai/), supporting OpenRouter, OpenAI, Anthropic, and other providers.

### 3. Running an Agent

With setup complete, let's run an agent. Here's an example that investigates TODO comments in scikit-learn:

```bash
python run.py https://github.com/scikit-learn/scikit-learn \
  -p "Investigate TODO comments across this repository. Spawn sub-agents to explore different modules. Find the easiest TODO and fix it."
```

This spawns a root agent that explores the codebase, delegates to sub-agents for parallel investigation, and produces a git patch fixing the easiest TODO it finds. We'll walk through the results and trace the execution in detail later, but first, let's look at how the code works.

#### CLI Options

| Option | Description |
|--------|-------------|
| `repo` | GitHub repository URL (required) |
| `-p, --prompt` | Task prompt for the agent (required) |
| `-b, --branch` | Branch name (optional) |
| `--commit` | Specific commit SHA (optional) |
| `-c, --config` | Path to config file (default: `config.yaml`) |
| `-o, --output` | Output file for patch (default: stdout) |

### 4. Understanding the Code

Let's walk through the key components of the agent system.

#### Agent Execution Loop

Each agent runs an iteration loop that calls the LLM, extracts code blocks, and executes them. The core loop in `agent.py`:

<Tabs>
  <TabItem label="Python" icon="seti:python">
    ```python
    def _run_loop(self) -> None:
        """Run the main iteration loop."""
        system_prompt = build_system_prompt(depth=self.depth)
        messages = [{"role": "system", "content": system_prompt}]
        execution_result = None

        for iteration in range(self.config.rlm.max_iterations):
            # Check global timeout
            if self._is_timeout():
                break

            # Build user prompt with previous execution result
            user_prompt = build_user_prompt(iteration, execution_result)
            messages.append({"role": "user", "content": user_prompt})

            # Get model completion
            response = self.client.completion(messages)
            messages.append({"role": "assistant", "content": response})

            # Execute code blocks in REPL
            repl_result = self.repl.execute_response(response)

            # Check for final answer
            if repl_result.final_answer is not None:
                self._result = repl_result.final_answer
                break

            # Format result for next iteration
            execution_result = format_execution_result(...)
    ```
  </TabItem>
</Tabs>

Each iteration:
1. Builds a prompt with context from previous execution
2. Gets an LLM completion
3. Extracts and executes Python code blocks
4. Checks if the agent called `FINAL()` to submit results
5. Formats the output for the next iteration

#### Sub-Agent Spawning

When agent code calls `rlm_query()`, a new sub-agent is created with its own sandbox:

<Tabs>
  <TabItem label="Python" icon="seti:python">
    ```python
    def _handle_rlm_query(self, task: str) -> str:
        """Spawn a sub-agent for a specific task."""
        # Check sandbox budget
        if not self.sandbox_manager.budget.can_acquire():
            return "Error: sandbox budget exhausted"

        # Create sub-agent at depth + 1
        sub_agent = RLMAgent(
            client=self.client,
            sandbox_manager=self.sandbox_manager,
            config=self.config,
            depth=self.depth + 1,
            task=task,
            # ... other params
        )

        # Run sub-agent (blocking)
        result = sub_agent.run()

        # Return result, truncated if necessary
        return result.result or "No result"
    ```
  </TabItem>
</Tabs>

For parallel spawning, `rlm_query_batched()` uses a thread pool:

<Tabs>
  <TabItem label="Python" icon="seti:python">
    ```python
    def _handle_rlm_query_batched(self, tasks: list[str]) -> list[str]:
        """Spawn multiple sub-agents in parallel."""
        results = [""] * len(tasks)

        with ThreadPoolExecutor(max_workers=10) as executor:
            future_to_idx = {
                executor.submit(self._handle_rlm_query, task): i
                for i, task in enumerate(tasks)
            }
            for future in as_completed(future_to_idx):
                idx = future_to_idx[future]
                results[idx] = future.result()

        return results
    ```
  </TabItem>
</Tabs>

#### Agent Code Interface

Inside the REPL, agents have access to these functions:

| Function | Description |
|----------|-------------|
| `rlm_query(task)` | Spawn a single sub-agent, returns result string |
| `rlm_query_batched(tasks)` | Spawn multiple sub-agents in parallel |
| `FINAL(answer)` | Submit final result (root: triggers patch extraction) |
| `FINAL_VAR(var_name)` | Submit the value of a variable as result |
| `edit_file(path, old, new)` | Edit a file with syntax validation |

Example spawning pattern used by agents:

<Tabs>
  <TabItem label="Python" icon="seti:python">
    ```python
    # Spawn multiple sub-agents to explore different modules
    results = rlm_query_batched([
        "Search for TODO comments in sklearn/linear_model/ and assess difficulty",
        "Search for TODO comments in sklearn/ensemble/ and assess difficulty",
        "Search for TODO comments in sklearn/tree/ and assess difficulty",
    ])

    for i, result in enumerate(results):
        print(f"=== Sub-agent {i+1} findings ===")
        print(result)
    ```
  </TabItem>
</Tabs>

### 5. Example Walkthrough

Let's trace what happens when we run an agent on a popular machine learning library, scikit-learn:

```bash
python run.py https://github.com/scikit-learn/scikit-learn \
  -p "Investigate TODO comments across this repository. Spawn sub-agents to explore different modules under sklearn/ in parallel. For each TODO found, assess how difficult it would be to fix (easy/medium/hard). After gathering results, pick the easiest TODO and fix it."
```

Note that there are about 400 lines in scikit-learn that contain the substring "# TODO".

**Step 1: Root agent explores and spawns depth-1 sub-agents**

The root agent (depth=0) examines the repository structure, identifies all sklearn modules, and spawns 25 sub-agents in parallel:

<Tabs>
  <TabItem label="Python" icon="seti:python">
    ```python
    # Define the subdirectories to investigate
    subdirs = [
        "cluster", "compose", "covariance", "cross_decomposition", "datasets",
        "decomposition", "ensemble", "feature_extraction", "feature_selection",
        "gaussian_process", "impute", "inspection", "linear_model", "manifold",
        "metrics", "mixture", "model_selection", "neighbors", "neural_network",
        "preprocessing", "semi_supervised", "svm", "tree", "utils"
    ]

    # Create queries for sub-agents
    queries = [
        f"Search for 'TODO' comments in 'sklearn/{subdir}/'. For each TODO found, provide: "
        f"1. The file path and line number. 2. The content of the TODO. 3. An assessment "
        f"of how difficult it would be to fix (easy/medium/hard) with a brief justification."
        for subdir in subdirs
    ]

    results = rlm_query_batched(queries)
    ```
  </TabItem>
</Tabs>

Each of these 25 sub-agents gets its own Daytona sandbox with a fresh clone of scikit-learn.

**Step 2: Depth-1 agents spawn depth-2 agents**

Some depth-1 agents decide their module is too large and spawn their own sub-agents. For example, the `sklearn/metrics/` agent spawned 3 depth-2 agents:

<Tabs>
  <TabItem label="Python" icon="seti:python">
    ```python
    # Inside the sklearn/metrics/ agent (depth=1)
    # To efficiently handle the large number of TODOs, spawn sub-agents for sub-directories

    tasks = [
        "Identify and assess TODOs in 'sklearn/metrics/cluster/'. Provide file, line, content, and difficulty.",
        "Identify and assess TODOs in 'sklearn/metrics/tests/'. Provide file, line, content, and difficulty.",
        "Identify and assess TODOs in 'sklearn/metrics/_plot/' and its 'tests' sub-directory."
    ]

    results = rlm_query_batched(tasks)
    ```
  </TabItem>
</Tabs>

**Step 3: Results propagate back**

Each sub-agent returns findings via `FINAL()`. Results flow back up:
- Depth-2 → Depth-1: Detailed analysis of specific subdirectories
- Depth-1 → Root: Module-level summaries with difficulty ratings

**Step 4: Root agent synthesizes and acts**

The root agent reviews all findings, identifies the easiest TODO, and makes the fix.

**Step 5: Git patch produced**

<Tabs>
  <TabItem label="Python" icon="seti:python">
    ```python
    import subprocess
    subprocess.run(['git', 'add', '-A'], cwd='/workspace')
    result = subprocess.run(['git', 'diff', '--cached', 'HEAD'],
                            capture_output=True, text=True, cwd='/workspace')
    FINAL(result.stdout)
    ```
  </TabItem>
</Tabs>

#### Results

- Execution time: **316 seconds** (~5.3 minutes)
- Agents spawned: **40** (25 at depth 1, 15 at depth 2)

**Generated patch:**
```diff
diff --git a/sklearn/utils/_array_api.py b/sklearn/utils/_array_api.py
--- a/sklearn/utils/_array_api.py
+++ b/sklearn/utils/_array_api.py
@@ -19,8 +19,7 @@ from sklearn.externals.array_api_compat import numpy as np_compat
 from sklearn.utils._dataframe import is_df_or_series
 from sklearn.utils.fixes import parse_version

-# TODO: complete __all__
-__all__ = ["xpx"]  # we import xpx here just to re-export it, need this to appease ruff
+__all__ = ['device', 'get_namespace', 'get_namespace_and_device', 'indexing_dtype', 'move_to', 'size', 'supported_float_dtypes', 'xpx', 'yield_namespace_device_dtype_combinations', 'yield_namespaces']
```

The agent found the easiest TODO (`# TODO: complete __all__` in `sklearn/utils/_array_api.py`) and completed the `__all__` list with all public symbols from the module.

### 6. Configuration

Configure the agent in `config.yaml`:

<Tabs>
  <TabItem label="YAML" icon="seti:yaml">
    ```yaml
    # Model configuration - using LiteLLM format
    model:
      name: "openrouter/google/gemini-3-flash-preview"

    # RLM configuration
    rlm:
      max_sandboxes: 50
      max_iterations: 50
      global_timeout: 3600
      result_truncation_limit: 10000
    ```
  </TabItem>
</Tabs>

| Parameter | Default | Description |
|-----------|---------|-------------|
| `model.name` | `openrouter/google/gemini-3-flash-preview` | LLM model in LiteLLM format |
| `rlm.max_sandboxes` | 50 | Maximum total sandboxes across entire rollout |
| `rlm.max_iterations` | 50 | Maximum iterations per agent |
| `rlm.global_timeout` | 3600 | Total timeout in seconds |
| `rlm.result_truncation_limit` | 10000 | Max chars in sub-agent results |

:::tip[Scaling Tips]
- Increase `max_sandboxes` for tasks requiring more parallel exploration
- The sandbox budget tracks total sandboxes created over the lifetime of the rollout
- Sub-agent sandboxes are deleted immediately after completion
:::

### 7. Viewing Results

Results are saved to the `results/` directory as JSON files. Use the built-in viewer:

```bash
python -m http.server 8000
# Open http://localhost:8000/viewer/
```

The viewer provides:
- Interactive tree visualization of the agent hierarchy
- Iteration details with code and output for each agent
- Statistics: agent count, max depth, total iterations

### 8. Conclusion

Current language models aren't specifically trained to leverage recursive delegation, so RLMs don't necessarily outperform single-agent approaches on benchmarks yet. However, the architecture demonstrates compelling properties for complex tasks.

In our scikit-learn example, 40 agents ran in parallel across the agent tree, each with its own isolated sandbox, completing the entire run in just over 5 minutes. This level of parallelism, where each agent can freely modify files, run tests, and explore without affecting others, would be difficult to achieve without per-agent sandboxes.

**Key advantages of this approach:**

- **Recursive decomposition**: Complex tasks naturally break into sub-tasks handled by specialized agents
- **Isolated execution**: Each agent gets a fresh sandbox, preventing interference
- **Parallel exploration**: `rlm_query_batched()` enables concurrent investigation