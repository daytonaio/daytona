# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from dataclasses import dataclass
from datetime import datetime

from daytona_api_client import SandboxListSortDirection, SandboxListSortField, SandboxState
from daytona_api_client_async import GpuType

TOOLBOX_PORT = 2280


@dataclass
class Resources:
    """Resources configuration for Sandbox.

    Attributes:
        cpu (int | None): Number of CPU cores to allocate.
        memory (int | None): Amount of memory in GiB to allocate.
        disk (int | None): Amount of disk space in GiB to allocate.
        gpu (int | None): Number of GPUs to allocate.
        gpu_type (GpuType | list[GpuType] | None): Preferred GPU type for the Sandbox.

    Example:
        ```python
        resources = Resources(
            cpu=2,
            memory=4,  # 4GiB RAM
            disk=20,   # 20GiB disk
            gpu=1,
            gpu_type=GpuType.H100,
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
    gpu_type: GpuType | list[GpuType] | None = None


@dataclass
class ListSandboxesQuery:
    """Query parameters for filtering and sorting when listing Sandboxes.

    Attributes:
        limit: Per-page fetch size. Does NOT limit the total number of
            Sandboxes returned.
        id: Filter by ID prefix (case-insensitive).
        name: Filter by name prefix (case-insensitive).
        labels: Filter by labels.
        states: Filter by states.
        snapshots: Filter by snapshot names.
        targets: Filter by targets.
        min_cpu: Filter by minimum CPU.
        max_cpu: Filter by maximum CPU.
        min_memory_gib: Filter by minimum memory in GiB.
        max_memory_gib: Filter by maximum memory in GiB.
        min_disk_gib: Filter by minimum disk space in GiB.
        max_disk_gib: Filter by maximum disk space in GiB.
        is_public: Filter by public status.
        is_recoverable: Filter by recoverable status.
        created_at_after (datetime): Include sandboxes created after this timestamp.
        created_at_before (datetime): Include sandboxes created before this timestamp.
        last_activity_after (datetime): Include sandboxes with last activity after this timestamp.
        last_activity_before (datetime): Include sandboxes with last activity before this timestamp.
        sort: Field to sort by.
        order: Sort direction.
    """

    limit: int | None = None
    id: str | None = None
    name: str | None = None
    labels: dict[str, str] | None = None
    states: list[SandboxState] | None = None
    snapshots: list[str] | None = None
    targets: list[str] | None = None
    min_cpu: int | None = None
    max_cpu: int | None = None
    min_memory_gib: int | None = None
    max_memory_gib: int | None = None
    min_disk_gib: int | None = None
    max_disk_gib: int | None = None
    is_public: bool | None = None
    is_recoverable: bool | None = None
    created_at_after: datetime | None = None
    created_at_before: datetime | None = None
    last_activity_after: datetime | None = None
    last_activity_before: datetime | None = None
    sort: SandboxListSortField | None = None
    order: SandboxListSortDirection | None = None
