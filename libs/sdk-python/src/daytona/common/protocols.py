# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from typing import Protocol


class HasBody(Protocol):
    body: object


class HasStatus(Protocol):
    status: int
