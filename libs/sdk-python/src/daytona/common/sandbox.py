# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from dataclasses import dataclass

TOOLBOX_PORT = 2280


@dataclass
class Resources:
    """Resources configuration for Sandbox.

    Attributes:
        cpu (int | None): Number of CPU cores to allocate.
        memory (int | None): Amount of memory in GiB to allocate.
        disk (int | None): Amount of disk space in GiB to allocate.
        gpu (int | None): Number of GPUs to allocate.

    Example:
        ```python
        resources = Resources(
            cpu=2,
            memory=4,  # 4GiB RAM
            disk=20,   # 20GiB disk
            gpu=1
        )
        params = CreateSandboxFromImageParams(
            image=Image.debian_slim("3.12"),
            language="python",
            resources=resources
        )
        ```
    """

    cpu: int | None = None
    memory: int | None = None
    disk: int | None = None
    gpu: int | None = None
