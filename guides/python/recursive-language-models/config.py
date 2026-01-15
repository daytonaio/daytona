"""Configuration loading and management."""

from pathlib import Path
from typing import Any

import yaml

from rlm.types import (
    Config,
    ModelConfig,
    RLMConfig,
)


def load_config(config_path: str | Path) -> Config:
    """Load configuration from a YAML file."""
    with open(config_path) as f:
        data = yaml.safe_load(f)

    return parse_config(data)


def parse_config(data: dict[str, Any]) -> Config:
    """Parse configuration dictionary into Config object."""
    model = data["model"]
    rlm = data["rlm"]

    return Config(
        model=ModelConfig(name=model["name"]),
        rlm=RLMConfig(
            max_sandboxes=rlm["max_sandboxes"],
            max_iterations=rlm["max_iterations"],
            global_timeout=rlm["global_timeout"],
            result_truncation_limit=rlm["result_truncation_limit"],
        ),
    )


