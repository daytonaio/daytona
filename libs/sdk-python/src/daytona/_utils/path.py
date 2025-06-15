# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import os
from pathlib import PurePosixPath
from typing import Optional


def prefix_relative_path(prefix: str, path: Optional[str] = None) -> str:
    result = prefix

    if path:
        path = path.strip()
        if path == "~":
            result = prefix
        elif path.startswith("~/"):
            result = os.path.join(prefix, path[2:])
        elif PurePosixPath(path).is_absolute():
            result = path
        else:
            result = os.path.join(prefix, path)

    return result
