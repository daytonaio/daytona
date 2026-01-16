# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""RLM core module."""


# Lazy imports to avoid circular dependencies
def __getattr__(name):
    if name == "RLMAgent":
        from rlm.agent import RLMAgent

        return RLMAgent
    elif name == "AgentResult":
        from rlm.types import AgentResult

        return AgentResult
    elif name == "RLMConfig":
        from rlm.types import RLMConfig

        return RLMConfig
    raise AttributeError(f"module {__name__!r} has no attribute {name!r}")


__all__ = ["RLMAgent", "AgentResult", "RLMConfig"]
