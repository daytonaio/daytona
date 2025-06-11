# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

from dataclasses import dataclass
from enum import Enum
from typing import Annotated, Dict, Optional

from daytona_api_client import Sandbox as ApiSandbox
from daytona_api_client import SandboxInfo as ApiSandboxInfo
from daytona_api_client import SandboxState
from daytona_api_client_async import SandboxInfo as AsyncApiSandboxInfo
from pydantic import Field


@dataclass
class SandboxTargetRegion(str, Enum):
    """Target regions for Sandboxes

    **Enum Members**:
        - `EU` ("eu")
        - `US` ("us")
        - `ASIA` ("asia")
    """

    EU = "eu"
    US = "us"
    ASIA = "asia"

    def __str__(self):
        return self.value

    def __eq__(self, other):
        if isinstance(other, str):
            return self.value == other
        return super().__eq__(other)


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
        params = CreateSandboxParams(
            language="python",
            resources=resources
        )
        ```
    """

    cpu: Optional[int] = None
    memory: Optional[int] = None
    disk: Optional[int] = None
    gpu: Optional[int] = None


class SandboxInfo(ApiSandboxInfo, AsyncApiSandboxInfo):
    """Structured information about a Sandbox.

    Attributes:
        id (str): Unique identifier for the Sandbox.
        snapshot (Optional[str]): Daytona snapshot used to create the Sandbox.
        user (str): OS user running in the Sandbox.
        env (Dict[str, str]): Environment variables set in the Sandbox.
        labels (Dict[str, str]): Custom labels attached to the Sandbox.
        public (bool): Whether the Sandbox is publicly accessible.
        target (str): Target environment where the Sandbox runs.
        resources (Resources): Resource allocations for the Sandbox.
        state (str): Current state of the Sandbox (e.g., "started", "stopped").
        error_reason (Optional[str]): Error message if Sandbox is in error state.
        backup_state (Optional[str]): Current state of Sandbox backup.
        backup_created_at (Optional[str]): When the backup was created.
        node_domain (str): Domain name of the Sandbox node.
        region (str): Region of the Sandbox node.
        class_name (str): Sandbox class.
        updated_at (str): When the Sandbox was last updated.
        last_backup (Optional[str]): When the last backup was created.
        auto_stop_interval (int): Auto-stop interval in minutes.
        auto_archive_interval (int): Auto-archive interval in minutes.
    """

    id: str
    name: Annotated[
        str,
        Field(
            default="",
            deprecated="The `name` field is deprecated.",
        ),
    ]
    snapshot: Optional[str]
    user: str
    env: Dict[str, str]
    labels: Dict[str, str]
    public: bool
    target: SandboxTargetRegion
    resources: Resources
    state: SandboxState
    error_reason: Optional[str]
    backup_state: Optional[str]
    backup_created_at: Optional[str]
    node_domain: str
    region: str
    class_name: str
    updated_at: str
    last_backup: Optional[str]
    auto_stop_interval: int
    auto_archive_interval: int
    provider_metadata: Annotated[
        Optional[str],
        Field(
            deprecated=(
                "The `provider_metadata` field is deprecated. Use `state`, `node_domain`, `region`, `class_name`,"
                " `updated_at`, `last_backup`, `resources`, `auto_stop_interval`, `auto_archive_interval` instead."
            )
        ),
    ]


class SandboxInstance(ApiSandbox):
    """Represents a Daytona Sandbox instance."""

    info: Optional[SandboxInfo]
