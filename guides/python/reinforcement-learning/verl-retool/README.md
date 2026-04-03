# veRL ReTool + Daytona Benchmark

## Overview

This directory contains the benchmark script used by Daytona's veRL guide.
`benchmark_tool_backends.py` runs the ReTool Daytona backend from a local
veRL checkout with the `recipe` submodule initialized.

## Requirements

- A local veRL checkout with the `recipe` submodule initialized
- A Python environment where veRL is already installed
- `DAYTONA_API_KEY` exported in your shell

## Quick Start

From your veRL environment:

```bash
cd /path/to/daytona/guides/python/reinforcement-learning/verl-retool
pip install -e .
```

Run the benchmark:

```bash
python benchmark_tool_backends.py \
  --backend daytona \
  --verl-root /absolute/path/to/verl \
  --concurrency 1 4 8 16 32 64 128
```

The script runs `simple_stdout`, `cpu_bound_stdout`, and `runtime_error`,
and writes `summary.json` and `results.csv` under
`outputs/daytona/<timestamp>/`.
