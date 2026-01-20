# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from typing import Protocol

from ..common.process import CodeRunParams


class SandboxCodeToolbox(Protocol):
    def get_run_command(self, code: str, params: CodeRunParams | None = None) -> str:
        ...


class HasBody(Protocol):
    body: object


class HasStatus(Protocol):
    status: int
