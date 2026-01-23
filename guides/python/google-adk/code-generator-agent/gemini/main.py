"""Code Generator & Tester Agent Example."""

import asyncio
import logging

from daytona_adk import DaytonaPlugin  # pylint: disable=import-error
from dotenv import load_dotenv
from google.adk.agents import Agent  # pylint: disable=import-error
from google.adk.apps import App  # pylint: disable=import-error
from google.adk.runners import InMemoryRunner  # pylint: disable=import-error

load_dotenv()

logging.basicConfig(level=logging.DEBUG)


def extract_final_response(response: list) -> str:
    """Extract the final text response from a list of ADK events."""
    for event in reversed(response):
        text_parts = []

        if hasattr(event, "text") and event.text:
            return event.text
        if hasattr(event, "content") and event.content:
            content = event.content
            if hasattr(content, "parts") and content.parts:
                for part in content.parts:
                    if hasattr(part, "text") and part.text:
                        text_parts.append(part.text)
                if text_parts:
                    return "".join(text_parts)
            if hasattr(content, "text") and content.text:
                return content.text
        if isinstance(event, dict):
            text = event.get("text") or event.get("content", {}).get("text")
            if text:
                return text

    return ""


AGENT_INSTRUCTION = """You are a code generator agent that writes verified, working code.
You support Python, JavaScript, and TypeScript.

Your workflow for every code request:
1. Write the function
2. Write tests for it
3. EXECUTE the code in the sandbox to verify it works - do not skip this step
4. If execution fails, fix and re-execute until tests pass
5. Once verified, respond with ONLY the function (no tests)

You must always execute code before responding. Never return untested code.
Only include tests in your response if the user explicitly asks for them.
"""


async def main() -> None:
    """Run the code generator agent example."""
    plugin = DaytonaPlugin(
        labels={"example": "code-generator"},
    )

    agent = Agent(
        model="gemini-2.5-pro",
        name="code_generator_agent",
        instruction=AGENT_INSTRUCTION,
        tools=plugin.get_tools(),
    )

    app = App(
        name="code_generator_app",
        root_agent=agent,
        plugins=[plugin],
    )

    async with InMemoryRunner(app=app) as runner:
        prompt = (
            "Write a TypeScript function called 'groupBy' that takes an array "
            "and a key function, and groups array elements by the key. "
            "Use proper type annotations."
        )

        print("\n" + "=" * 60)
        print("USER PROMPT:")
        print("=" * 60)
        print(prompt)
        print("-" * 60)

        response = await runner.run_debug(prompt)

        final_response = extract_final_response(response)
        print("\nAGENT RESPONSE:")
        print("-" * 60)
        print(final_response)
        print("=" * 60)

    print("\nApp closed, sandbox cleaned up. Done!")


if __name__ == "__main__":
    asyncio.run(main())
