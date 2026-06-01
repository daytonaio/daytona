# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""Plan-and-execute LangGraph agent that ETLs GitHub data into a Daytona-sandboxed SQLite."""

import re
from typing import TypedDict

from dotenv import load_dotenv  # pylint: disable=import-error
from langchain_anthropic import ChatAnthropic  # pylint: disable=import-error
from langchain_core.messages import HumanMessage, SystemMessage  # pylint: disable=import-error
from langgraph.graph import END, START, StateGraph  # pylint: disable=import-error
from pydantic import BaseModel, Field  # pylint: disable=import-error

from daytona import Daytona, Sandbox

load_dotenv()

USER_REQUEST = """Profile the maintenance health of the public GitHub repository `langchain-ai/langgraph`.

Fetch the 100 most recently updated issues from
  https://api.github.com/repos/langchain-ai/langgraph/issues?state=all&per_page=100&sort=updated
and the 100 most recently updated pull requests from
  https://api.github.com/repos/langchain-ai/langgraph/pulls?state=all&per_page=100&sort=updated
(public GitHub REST API, no authentication required). Load them into a SQLite database in the sandbox.

Then answer these three questions, printing the SQL query before each result:

1. PR merge rate among CLOSED PRs in the dataset (% merged vs closed-without-merge).
2. Top 5 PR authors by total PR count, with each author's personal merge rate.
3. The single most-commented currently-OPEN issue: title, comment count, author login, and created_at.
"""

PLAN_SYSTEM_PROMPT = """You are the planner stage of a plan-and-execute data agent.

Produce an ordered list of 4-8 atomic plan steps. Each step is one natural-language sentence describing
a single chunk of Python code that the executor stage will then write and run in a persistent Daytona sandbox.

Rules:
- Sandbox state PERSISTS across steps. Imports, variables, and files from step N are visible in step N+1.
- Step 1 should establish any package installs or imports.
- Each step is one coherent action. Group tightly-coupled work that shares variables (fetch + filter, or
  create-schema + insert-data) into a SINGLE step so the executor doesn't have to guess prior variable names
  across step boundaries. Keep loosely-coupled work in separate steps.
- PRESERVE any specific URLs, endpoints, file paths, table names, or identifiers from the user's request
  VERBATIM inside the plan step that uses them. Do not paraphrase URLs.
- Do NOT write code in the plan. Describe what each step does.
"""

EXECUTE_SYSTEM_PROMPT = """You are the executor stage of a plan-and-execute data agent.

You receive the user's original request, the full plan, and the index of the current step. You must output
ONLY Python code that accomplishes the current step. The code runs in a persistent Daytona sandbox; prior
steps' variables, imports, and files are still in scope. Always `print()` the relevant output so later
stages can see the results.

Rules:
- Use the EXACT URLs / endpoints / file paths from the user's original request. Do not invent or paraphrase.
- CRITICAL: Before referencing any variable from a prior step, scan the prior code shown below and use
  EXACTLY the variable name that the prior step assigned. Never invent variable names. If you cannot find
  the variable you need in the prior code, re-derive it from scratch within your current step.
- Output format: a single ```python fenced block, nothing else. No prose.
- If a previous attempt failed, you will see the error and the failing code. Diagnose the root cause and
  produce a materially different fix. Do not repeat the failing approach. If the error is a NameError,
  the missing variable was never defined in the shown prior code; re-derive it from raw data.
"""

SUMMARIZE_SYSTEM_PROMPT = """You are the summarizer stage of a plan-and-execute data agent.

You will be shown the user's original request and the stdout from each successfully executed plan step.
Produce a clear, factual answer in 1-3 short paragraphs. Cite specific numbers from the stdout. Do not
hallucinate values that are not present in the stdout. If a step failed, say so plainly.
"""


class Plan(BaseModel):
    """Structured-output schema for the planner LLM call."""

    steps: list[str] = Field(
        description="Atomic plan steps the executor will implement, in order.",
        min_length=1,
        max_length=10,
    )


class AgentState(TypedDict):
    """Graph state. Carries the sandbox handle, plan, progress, and per-step outputs + code."""

    sandbox: Sandbox | None
    user_request: str
    plan: list[str]
    step_idx: int
    attempts: int
    max_attempts: int
    last_error: str | None
    last_code: str | None
    step_outputs: list[str]
    step_codes: list[str]
    final_answer: str


CODE_FENCE_RE = re.compile(r"```(?:python)?\n(.*?)```", re.DOTALL)


def extract_code(text: str) -> str:
    """Pull the first ```python fenced block out of an LLM response; fall back to the raw text."""
    match = CODE_FENCE_RE.search(text)
    return match.group(1).strip() if match else text.strip()


def build_graph(model: ChatAnthropic):
    """Wire the 6-node plan-and-execute state graph."""
    plan_llm = model.with_structured_output(Plan)

    def provision(state: AgentState) -> dict:
        print("\n[provision] creating Daytona sandbox...")
        sandbox = Daytona().create()
        print(f"[provision] sandbox ready (id={sandbox.id})")
        return {"sandbox": sandbox}

    def plan_node(state: AgentState) -> dict:
        print("\n[plan] asking the LLM for a multi-step plan...")
        result = plan_llm.invoke(
            [
                SystemMessage(content=PLAN_SYSTEM_PROMPT),
                HumanMessage(content=state["user_request"]),
            ]
        )
        assert isinstance(result, Plan)
        print(f"[plan] {len(result.steps)} step(s):")
        for i, step in enumerate(result.steps, 1):
            print(f"  {i}. {step}")
        return {"plan": result.steps}

    def execute(state: AgentState) -> dict:
        idx = state["step_idx"]
        step_text = state["plan"][idx]
        attempt = state["attempts"] + 1
        max_attempts = state["max_attempts"]
        print(f"\n[execute] step {idx + 1}/{len(state['plan'])} attempt {attempt}/{max_attempts}: {step_text}")

        plan_listing = "\n".join(
            f"  {i + 1}. {s}{' <-- CURRENT' if i == idx else ''}" for i, s in enumerate(state["plan"])
        )
        prompt_parts = [
            f"Original user request:\n{state['user_request']}",
            f"Full plan:\n{plan_listing}",
            f"Current step ({idx + 1} of {len(state['plan'])}): {step_text}",
        ]
        if state["step_codes"]:
            prompt_parts.append("Code already executed in this sandbox (variables and imports still in scope):")
            for i, prior in enumerate(state["step_codes"], 1):
                prompt_parts.append(f"--- step {i} code ---\n{prior}")
        if state["step_outputs"]:
            prompt_parts.append("Stdout from those prior steps:")
            for i, output in enumerate(state["step_outputs"], 1):
                prompt_parts.append(f"--- step {i} stdout ---\n{output[:1500]}")
        if state["last_error"] and state["last_code"]:
            prompt_parts.append(f"--- previous attempt error ---\n{state['last_error'][:1500]}")
            prompt_parts.append(f"--- previous failing code ---\n{state['last_code']}")
            prompt_parts.append("Diagnose and write a corrected implementation.")

        response = model.invoke(
            [
                SystemMessage(content=EXECUTE_SYSTEM_PROMPT),
                HumanMessage(content="\n\n".join(prompt_parts)),
            ]
        )
        content = response.content if isinstance(response.content, str) else str(response.content)
        code = extract_code(content)
        print(f"[execute] generated {len(code)} chars of code, running in sandbox...")

        sandbox = state["sandbox"]
        assert sandbox is not None, "sandbox missing from state"
        result = sandbox.code_interpreter.run_code(code, timeout=180)
        stdout = result.stdout or ""

        if result.error is not None:
            err = f"{result.error.name}: {result.error.value}\n{result.error.traceback}".strip()
            print(f"[execute] step FAILED:\n{err[:400]}")
            return {"last_error": err, "last_code": code}

        snippet = stdout if len(stdout) <= 400 else stdout[:400] + f"...({len(stdout) - 400} more chars)"
        print(f"[execute] step OK. stdout:\n{snippet}")
        return {
            "last_error": None,
            "last_code": code,
            "step_outputs": state["step_outputs"] + [stdout],
            "step_codes": state["step_codes"] + [code],
        }

    def check(state: AgentState) -> dict:
        """Deterministic: advance on success, increment attempts on failure."""
        if state["last_error"]:
            return {"attempts": state["attempts"] + 1}
        return {"step_idx": state["step_idx"] + 1, "attempts": 0, "last_error": None, "last_code": None}

    def route_from_check(state: AgentState) -> str:
        if state["last_error"]:
            if state["attempts"] >= state["max_attempts"]:
                print(f"[check] step failed after {state['max_attempts']} attempts; giving up and summarizing")
                return "summarize"
            print(f"[check] retrying (attempt {state['attempts'] + 1}/{state['max_attempts']})")
            return "execute"
        if state["step_idx"] >= len(state["plan"]):
            print("[check] all plan steps complete; summarizing")
            return "summarize"
        print(f"[check] step {state['step_idx']} done; advancing to step {state['step_idx'] + 1}")
        return "execute"

    def summarize(state: AgentState) -> dict:
        print("\n[summarize] asking the LLM for a final answer...")
        parts = [f"Original request:\n{state['user_request']}", "Outputs from executed plan steps:"]
        for i, output in enumerate(state["step_outputs"], 1):
            parts.append(f"--- step {i} stdout ---\n{output}")
        if state["last_error"]:
            parts.append(f"NOTE: the agent gave up before finishing. Last error:\n{state['last_error']}")
        response = model.invoke(
            [SystemMessage(content=SUMMARIZE_SYSTEM_PROMPT), HumanMessage(content="\n\n".join(parts))]
        )
        content = response.content if isinstance(response.content, str) else str(response.content)
        return {"final_answer": content}

    def cleanup(state: AgentState) -> dict:
        sandbox = state.get("sandbox")
        if sandbox is not None:
            print(f"\n[cleanup] deleting sandbox {sandbox.id}...")
            sandbox.delete()
            print("[cleanup] done")
        return {"sandbox": None}

    graph = StateGraph(AgentState)
    graph.add_node("provision", provision)
    graph.add_node("plan", plan_node)
    graph.add_node("execute", execute)
    graph.add_node("check", check)
    graph.add_node("summarize", summarize)
    graph.add_node("cleanup", cleanup)

    graph.add_edge(START, "provision")
    graph.add_edge("provision", "plan")
    graph.add_edge("plan", "execute")
    graph.add_edge("execute", "check")
    graph.add_conditional_edges("check", route_from_check, {"execute": "execute", "summarize": "summarize"})
    graph.add_edge("summarize", "cleanup")
    graph.add_edge("cleanup", END)
    return graph.compile()


def main() -> None:
    model = ChatAnthropic(
        model_name="claude-opus-4-6",
        temperature=0,
        timeout=120,
        max_retries=2,
        stop=None,
    )
    app = build_graph(model)

    print("=" * 60)
    print("USER REQUEST")
    print("=" * 60)
    print(USER_REQUEST)

    initial_state: AgentState = {
        "sandbox": None,
        "user_request": USER_REQUEST,
        "plan": [],
        "step_idx": 0,
        "attempts": 0,
        "max_attempts": 3,
        "last_error": None,
        "last_code": None,
        "step_outputs": [],
        "step_codes": [],
        "final_answer": "",
    }

    # Stream the graph so we always have the latest state snapshot. If anything
    # in the run raises before the `cleanup` node executes, the finally block
    # will still see the live sandbox in `final_state` and delete it.
    final_state: AgentState | None = None
    try:
        for chunk in app.stream(initial_state, config={"recursion_limit": 50}, stream_mode="values"):
            final_state = chunk  # type: ignore[assignment]
    finally:
        if final_state is not None:
            sandbox = final_state.get("sandbox")
            if sandbox is not None:
                try:
                    sandbox.delete()
                    print("\n[main] defensive cleanup: deleted orphaned sandbox")
                except Exception as e:  # noqa: BLE001
                    print(f"\n[main] defensive cleanup failed: {e}")

    print("\n" + "=" * 60)
    print("FINAL ANSWER")
    print("=" * 60)
    print(final_state["final_answer"] if final_state else "(run did not produce a final state)")


if __name__ == "__main__":
    main()
