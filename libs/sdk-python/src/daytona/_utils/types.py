# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from typing import Any

from typing_extensions import TypeGuard

from ..common.protocols import HasBody, HasStatus


def has_body(obj: Any) -> TypeGuard[HasBody]:
    return hasattr(obj, "body") and obj.body is not None


def has_status(obj: object) -> TypeGuard[HasStatus]:
    return hasattr(obj, "status")
