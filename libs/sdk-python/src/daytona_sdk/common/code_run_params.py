# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

from dataclasses import dataclass
from typing import Dict, List, Optional


@dataclass
class CodeRunParams:
    """Parameters for code execution.

    Attributes:
        argv (Optional[List[str]]): Command line arguments
        env (Optional[Dict[str, str]]): Environment variables
    """

    argv: Optional[List[str]] = None
    env: Optional[Dict[str, str]] = None
