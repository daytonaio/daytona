# Copyright 2026 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import os
from dataclasses import asdict, dataclass

from dotenv import load_dotenv

# pylint: disable=import-error
from daytona import CreateSandboxFromSnapshotParams, Daytona
from tealtiger import TealOpenAI

MODEL = "gpt-4o-mini"


@dataclass
class ExecutionReceipt:
    governance_provider: str
    allowed: bool
    reason: str
    sandbox_id: str | None
    exit_code: int | None
    result_preview: str


def generate_governed_code(prompt: str) -> str:
    client = TealOpenAI(
        api_key=os.environ["OPENAI_API_KEY"],
        guardrails={
            "secret_detection": True,
            "prompt_injection": True,
        },
        budget={
            "max_cost_per_session": 2.00,
        },
    )

    response = client.chat.completions.create(
        model=MODEL,
        messages=[
            {
                "role": "system",
                "content": (
                    "Generate a short, self-contained Python script. "
                    "Return only executable Python code, with no markdown fences."
                ),
            },
            {"role": "user", "content": prompt},
        ],
    )

    return response.choices[0].message.content or ""


def run_in_daytona(code: str) -> ExecutionReceipt:
    daytona = Daytona()
    sandbox = None

    try:
        sandbox = daytona.create(CreateSandboxFromSnapshotParams(language="python"))
        result = sandbox.process.code_run(code)
        return ExecutionReceipt(
            governance_provider="tealtiger",
            allowed=True,
            reason="Governed generation succeeded before Daytona execution",
            sandbox_id=sandbox.id,
            exit_code=result.exit_code,
            result_preview=result.result[:500],
        )
    finally:
        if sandbox is not None:
            daytona.delete(sandbox)


def print_receipt(receipt: ExecutionReceipt) -> None:
    print("\nExecution receipt")
    for key, value in asdict(receipt).items():
        print(f"{key}: {value}")


def main() -> None:
    load_dotenv()

    prompt = "Write Python code that calculates the first 10 Fibonacci numbers and prints them."
    print("Generating governed code...")
    try:
        code = generate_governed_code(prompt)
    except Exception as exc:
        print_receipt(
            ExecutionReceipt(
                governance_provider="tealtiger",
                allowed=False,
                reason=f"Governance blocked code before Daytona execution: {exc}",
                sandbox_id=None,
                exit_code=None,
                result_preview="",
            )
        )
        return

    print("\nApproved code")
    print("-" * 60)
    print(code)
    print("-" * 60)

    print("\nRunning approved code in Daytona...")
    receipt = run_in_daytona(code)
    print_receipt(receipt)


if __name__ == "__main__":
    main()
