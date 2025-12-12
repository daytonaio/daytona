# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import base64
import re
from pathlib import Path

from litellm import completion  # pylint: disable=import-error

# pylint: disable=import-error
from daytona import CreateSandboxFromSnapshotParams, Daytona

CODING_MODEL = "anthropic/claude-sonnet-4-0"
SUMMARY_MODEL = "anthropic/claude-haiku-4-5"


# Helper function to extract Python code from a given string
def extract_python(text: str) -> str:
    match = re.search(r"```python([\s\S]*?)```", text)
    return match.group(1).strip() if match else ""


# Make sure you have the DAYTONA_API_KEY environment variable set
def main() -> None:
    daytona = Daytona()
    sandbox = None

    try:
        # Create a Python sandbox
        sandbox = daytona.create(CreateSandboxFromSnapshotParams(language="python"))

        csv_path = "cafe_sales_data.csv"
        sandbox_csv_path = csv_path

        # Upload the CSV file to the sandbox
        sandbox.fs.upload_file(str(csv_path), sandbox_csv_path)

        # Generate the system prompt with the first few rows of data for context
        with Path(csv_path).open("r", encoding="utf-8") as f:
            csv_sample = "".join(f.readlines()[:3]).strip()

        # Define the user prompt
        user_prompt = "Give the three highest revenue products for the month of January and show them as a bar chart."
        print("Prompt:", user_prompt)

        system_prompt = (
            "\nYou are a helpful assistant that analyzes data.\n"
            "To run Python code in a sandbox, output a single block of code.\n"
            f"The sandbox:\n - has pandas and numpy installed.\n - contains {sandbox_csv_path}."
            "Plot any charts that you create."
            f"The first few rows of {sandbox_csv_path} are:\n"
            f"{csv_sample}\n"
            "After seeing the results of the code, answer the user's query."
        )

        # Generate the Python code with the LLM
        print("Generating code...")
        messages = [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": user_prompt},
        ]

        # LiteLLM supports a variety of model providers
        # Make sure to have the right environment variables set
        llm_output = completion(
            model=CODING_MODEL,
            messages=messages,
        )

        first_message = llm_output.choices[0].message
        print("LLM output:", first_message)
        messages.append({"role": first_message.role, "content": first_message.content})

        # Extract and execute Python code from the LLM's response
        print("Running code...")
        code = extract_python(first_message.content or "")
        exec_result = sandbox.process.code_run(code)

        messages.append(
            {
                "role": "assistant",
                "content": f"Code execution result:\n{exec_result.result}.",
            },
        )

        artifacts = getattr(exec_result, "artifacts", None)
        charts = getattr(artifacts, "charts", None) if artifacts is not None else None
        if charts:
            for index, chart in enumerate(charts):
                png_data = chart.get("png") if isinstance(chart, dict) else getattr(chart, "png", None)
                if png_data:
                    filename = f"chart-{index}.png"
                    Path(filename).write_bytes(base64.b64decode(png_data))
                    print(f"✓ Chart saved to {filename}")

        # Generate the final response with the LLM
        summary = completion(
            model=SUMMARY_MODEL,
            messages=messages,
        )

        print("Response:", summary.choices[0].message.content)

    finally:
        if sandbox is not None:
            daytona.delete(sandbox)


if __name__ == "__main__":
    main()
