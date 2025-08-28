# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from dataclasses import dataclass
from typing import Optional


@dataclass
class Resources:
    """Resources configuration for Sandbox.

    Attributes:
        cpu (Optional[int]): Number of CPU cores to allocate.
        memory (Optional[int]): Amount of memory in GiB to allocate.
        disk (Optional[int]): Amount of disk space in GiB to allocate.
        gpu (Optional[int]): Number of GPUs to allocate.

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

    cpu: Optional[int] = None
    memory: Optional[int] = None
    disk: Optional[int] = None
    gpu: Optional[int] = None
