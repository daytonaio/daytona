# TealTiger Governed Sandbox Example

## Overview

This example demonstrates a pre-execution governance pattern for AI-generated code. TealTiger evaluates the LLM output before the code reaches `sandbox.process.code_run()`, while Daytona provides the isolated runtime for code that passes the governance checks.

Use this pattern when an agent generates code dynamically and you want a policy decision before the code can enter the sandbox filesystem, process logs, or network path.

## Features

- **Pre-execution governance:** TealTiger guardrails run before Daytona execution
- **Sandbox isolation:** Approved code runs inside a Daytona sandbox, not on your machine
- **Audit-friendly logging:** The example prints a receipt that links the governance step to the sandbox result
- **Guaranteed cleanup:** The sandbox is deleted even when code generation or execution fails

## Requirements

- **Python:** Version 3.10 or higher
- `DAYTONA_API_KEY`: Required for access to Daytona sandboxes. Get it from [Daytona Dashboard](https://app.daytona.io/dashboard/keys)
- `OPENAI_API_KEY`: Required for the TealTiger OpenAI wrapper

## Getting Started

1. Create and activate a virtual environment:

   ```bash
   python3.10 -m venv venv
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   ```

2. Install dependencies:

   ```bash
   pip install -e .
   ```

3. Copy `.env.example` to `.env` and fill in your keys:

   ```bash
   DAYTONA_API_KEY=dtn_***
   OPENAI_API_KEY=sk-***
   ```

4. Run the example:

   ```bash
   python governed_sandbox.py
   ```

## How It Works

1. `generate_governed_code()` calls `TealOpenAI` with secret and prompt-injection guardrails enabled.
2. TealTiger evaluates the generated code before the application passes it to Daytona.
3. If governance raises an error, the script records a blocked receipt and never creates a Daytona sandbox.
4. `run_in_daytona()` creates a Python sandbox and calls `sandbox.process.code_run(code)` only after the governed generation step succeeds.
5. The script prints an execution receipt containing the governance provider, decision reason, sandbox ID, exit code, and whether execution was allowed.
6. The sandbox is deleted in a `finally` block.

## License

See the main project LICENSE file for details.

## References

- [TealTiger](https://github.com/agentguard-ai/tealtiger)
- [Daytona Documentation](https://www.daytona.io/docs)
