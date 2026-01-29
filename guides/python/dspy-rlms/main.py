"""Example: Run DSPy RLM with Daytona sandboxed code execution."""

import dspy
from daytona_interpreter import DaytonaInterpreter
from dotenv import load_dotenv

load_dotenv()

# Configure the LLM — pick one:
lm = dspy.LM("openrouter/google/gemini-3-flash-preview")
# lm = dspy.LM("openai/gpt-4o-mini")
# lm = dspy.LM("anthropic/claude-sonnet-4-20250514")
dspy.configure(lm=lm)

# Create an RLM backed by a Daytona sandbox
interpreter = DaytonaInterpreter()

rlm = dspy.RLM(
    signature="question -> answer: str",
    interpreter=interpreter,
    verbose=True,
)

try:
    result = rlm(question="What is the sum of the first 10 prime numbers?")
    print(f"\nAnswer: {result.answer}")
finally:
    interpreter.shutdown()
