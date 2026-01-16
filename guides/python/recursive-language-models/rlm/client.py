# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""LiteLLM-based unified client for all LLM providers."""

import logging
import time
from abc import ABC, abstractmethod
from typing import Any

import litellm

from rlm.types import UsageStats

logger = logging.getLogger(__name__)


class BaseLMClient(ABC):
    """Base class for all language model clients."""

    def __init__(self, model_name: str, **kwargs):
        self.model_name = model_name
        self.kwargs = kwargs
        self._last_usage = UsageStats()

    @abstractmethod
    def completion(self, prompt: str | list[dict[str, Any]]) -> str:
        """Generate a completion for the given prompt."""
        raise NotImplementedError

    @property
    def last_usage(self) -> UsageStats:
        """Get usage stats from the last completion call."""
        return self._last_usage


class LiteLLMClient(BaseLMClient):
    """
    Unified LLM client using LiteLLM.

    Supports all providers via LiteLLM's model naming convention:
    - OpenRouter: "openrouter/google/gemini-3-flash-preview"
    - OpenAI: "gpt-4o"
    - Anthropic: "claude-3-opus-20240229"
    """

    def __init__(
        self,
        model_name: str,
        api_key: str | None = None,
        **kwargs,
    ):
        super().__init__(model_name=model_name, **kwargs)
        self.api_key = api_key

    def completion(self, prompt: str | list[dict[str, Any]], max_retries: int = 5) -> str:
        messages = self._prepare_messages(prompt)

        for attempt in range(max_retries):
            response = litellm.completion(
                model=self.model_name,
                messages=messages,
                api_key=self.api_key,
            )
            self._track_usage(response)

            if response.choices and response.choices[0].message.content:
                return response.choices[0].message.content

            # Empty response - retry
            if attempt < max_retries - 1:
                logger.warning(f"Empty API response, retrying ({attempt + 1}/{max_retries})")
                time.sleep(1 * (attempt + 1))  # Backoff

        raise ValueError("API returned empty response after retries")

    def _prepare_messages(self, prompt: str | list[dict[str, Any]]) -> list[dict[str, Any]]:
        """Prepare messages for API."""
        if isinstance(prompt, str):
            return [{"role": "user", "content": prompt}]
        elif isinstance(prompt, list):
            return prompt
        else:
            raise ValueError(f"Invalid prompt type: {type(prompt)}")

    def _track_usage(self, response):
        """Track token usage from response."""
        usage = getattr(response, "usage", None)
        if usage is None:
            self._last_usage = UsageStats()
            return

        input_tokens = getattr(usage, "prompt_tokens", 0) or 0
        output_tokens = getattr(usage, "completion_tokens", 0) or 0

        # Use LiteLLM's cost calculation
        try:
            cost = litellm.completion_cost(response)
        except Exception:
            cost = 0

        self._last_usage = UsageStats(
            input_tokens=input_tokens,
            output_tokens=output_tokens,
            cost=cost,
        )


def create_client(model_name: str, api_key: str) -> LiteLLMClient:
    """Create an LLM client using LiteLLM."""
    return LiteLLMClient(model_name=model_name, api_key=api_key)
