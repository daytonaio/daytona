# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import os
from contextlib import contextmanager
from copy import deepcopy


@contextmanager
def isolated_env(temp_env=None):
    """Temporarily replaces os.environ with a controlled copy."""
    old_env = deepcopy(os.environ)
    os.environ.clear()
    if temp_env:
        os.environ.update(temp_env)
    try:
        yield
    finally:
        os.environ.clear()
        os.environ.update(old_env)
