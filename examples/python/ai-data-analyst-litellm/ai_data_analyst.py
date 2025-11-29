import re
from pathlib import Path
import base64

from daytona import CreateSandboxFromSnapshotParams, Daytona
from litellm import completion


# Helper function to extract Python code from a given string
def extract_python(text: str) -> str:
    match = re.search(r"```python([\s\S]*?)```", text)
    return match.group(1).strip() if match else ""


# Make sure you have the DAYTONA_API_KEY and OPENAI_API_KEY environment variables set
def main() -> None:
    daytona = Daytona()
    sandbox = None

    try:
        # Create a Python sandbox
        sandbox = daytona.create(CreateSandboxFromSnapshotParams(language="python"))

        base_dir = Path(__file__).resolve().parents[2]
        csv_path = "cafe_sales_data.csv"

        # Upload the CSV file to the sandbox
        sandbox.fs.upload_file(str(csv_path), "cafe_sales_data.csv")

        # Generate the system prompt with the first few rows of data for context
        with Path(csv_path).open("r", encoding="utf-8") as f:
            csv_sample = "".join(f.readlines()[:3]).strip()

        # Define the user prompt
        user_prompt = "Give the three highest revenue products for the month of January and show them as a bar chart."
        print("Prompt:", user_prompt)

        system_prompt = (
            "\nYou are a helpful assistant that analyzes data.\n"
            "You can execute Python code. Pandas and numpy are installed.\n"
            "Read cafe_sales_data.csv. The first few rows are:\n"
            f"{csv_sample}\n."
        )

        # Generate the Python code with the LLM
        print("Generating code...")
        messages = [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": user_prompt},
        ]

        llm_output = completion(
            model="gpt-5.1",
            messages=messages,
        )

        first_message = llm_output.choices[0].message
        messages.append({"role": first_message.role, "content": first_message.content})

        # Extract and execute Python code from the LLM's response
        print("Running code...")
        code = extract_python(first_message.content or "")
        exec_result = sandbox.process.code_run(code)

        messages.append(
            {
                "role": "user",
                "content": f"Code execution result:\n{exec_result.result}.",
            }
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
            model="gpt-4o",
            messages=messages,
        )

        print("Response:", summary.choices[0].message.content)

    finally:
        if sandbox is not None:
            daytona.delete(sandbox)


if __name__ == "__main__":
    main()