# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import os
from collections.abc import Iterator, Mapping, MutableMapping
from contextlib import contextmanager
from copy import deepcopy


@contextmanager
def isolated_env(temp_env: Mapping[str, str] | None = None) -> Iterator[None]:
    """Temporarily replaces os.environ with a controlled copy."""
    old_env: MutableMapping[str, str] = deepcopy(os.environ)
    os.environ.clear()
    if temp_env:
        os.environ.update(temp_env)
    try:
        yield
    finally:
        os.environ.clear()
        os.environ.update(old_env)
