"""Core type definitions for deeper-rlm."""

from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

import yaml


@dataclass
class ModelConfig:
    """LLM model configuration."""

    name: str  # LiteLLM format, e.g. "openrouter/google/gemini-3-flash-preview"


@dataclass
class RLMConfig:
    """RLM execution configuration."""

    max_sandboxes: int
    max_iterations: int
    global_timeout: int
    result_truncation_limit: int


@dataclass
class Config:
    """Complete configuration."""

    model: ModelConfig
    rlm: RLMConfig

    @classmethod
    def from_yaml(cls, path: str | Path) -> "Config":
        """Load configuration from a YAML file."""
        with open(path) as f:
            data: dict[str, Any] = yaml.safe_load(f)

        model = data["model"]
        rlm = data["rlm"]

        return cls(
            model=ModelConfig(name=model["name"]),
            rlm=RLMConfig(
                max_sandboxes=rlm["max_sandboxes"],
                max_iterations=rlm["max_iterations"],
                global_timeout=rlm["global_timeout"],
                result_truncation_limit=rlm["result_truncation_limit"],
            ),
        )


@dataclass
class UsageStats:
    """Token usage statistics."""

    input_tokens: int = 0
    output_tokens: int = 0
    cost: float = 0.0

    def __add__(self, other: "UsageStats") -> "UsageStats":
        return UsageStats(
            input_tokens=self.input_tokens + other.input_tokens,
            output_tokens=self.output_tokens + other.output_tokens,
            cost=self.cost + other.cost,
        )


@dataclass
class CodeBlockResult:
    """Result of executing a code block."""

    code: str
    stdout: str
    stderr: str
    execution_time: float
    error: str | None = None


@dataclass
class Iteration:
    """A single iteration of agent execution."""

    iteration: int
    prompt: str | list[dict[str, str]]
    raw_response: str
    parsed_code_blocks: list[CodeBlockResult] = field(default_factory=list)
    spawned_agents: list["AgentResult"] = field(default_factory=list)


@dataclass
class AgentResult:
    """Result from an RLM agent execution."""

    agent_id: str
    depth: int
    sandbox_id: str
    task: str | None = None  # None for root agent

    iterations: list[Iteration] = field(default_factory=list)
    spawned_agents: list["AgentResult"] = field(default_factory=list)

    result: str | None = None  # The FINAL() answer
    result_truncated: bool = False
    usage: UsageStats = field(default_factory=UsageStats)
    execution_time: float = 0.0
    error: str | None = None
