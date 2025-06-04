# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

from dataclasses import dataclass
from enum import Enum
from typing import Annotated, Dict, Optional

from daytona_api_client import Workspace as ApiSandbox
from daytona_api_client import WorkspaceInfo as ApiSandboxInfo
from daytona_api_client import WorkspaceState as SandboxState
from daytona_api_client_async import WorkspaceInfo as AsyncApiSandboxInfo
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
class SandboxResources:
    """Resources allocated to a Sandbox.

    Attributes:
        cpu (str): Nu, "1", "2").
        gpu (Optional[str]): Number of GPUs allocated mber of CPU cores allocated (e.g.(e.g., "1") or None if no GPU.
        memory (str): Amount of memory allocated with unit (e.g., "2Gi", "4Gi").
        disk (str): Amount of disk space allocated with unit (e.g., "10Gi", "20Gi").

    Example:
        ```python
        resources = SandboxResources(
            cpu="2",
            gpu="1",
            memory="4Gi",
            disk="20Gi"
        )
        ```
    """

    cpu: str
    memory: str
    disk: str
    gpu: Optional[str] = None


class SandboxInfo(ApiSandboxInfo, AsyncApiSandboxInfo):
    """Structured information about a Sandbox.

    Attributes:
        id (str): Unique identifier for the Sandbox.
        image (Optional[str]): Docker image used for the Sandbox.
        user (str): OS user running in the Sandbox.
        env (Dict[str, str]): Environment variables set in the Sandbox.
        labels (Dict[str, str]): Custom labels attached to the Sandbox.
        public (bool): Whether the Sandbox is publicly accessible.
        target (str): Target environment where the Sandbox runs.
        resources (SandboxResources): Resource allocations for the Sandbox.
        state (str): Current state of the Sandbox (e.g., "started", "stopped").
        error_reason (Optional[str]): Error message if Sandbox is in error state.
        snapshot_state (Optional[str]): Current state of Sandbox snapshot.
        snapshot_created_at (Optional[str]): When the snapshot was created.
        node_domain (str): Domain name of the Sandbox node.
        region (str): Region of the Sandbox node.
        class_name (str): Sandbox class.
        updated_at (str): When the Sandbox was last updated.
        last_snapshot (Optional[str]): When the last snapshot was created.
        auto_stop_interval (int): Auto-stop interval in minutes.
        auto_archive_interval (int): Auto-archive interval in minutes.

    Example:
        ```python
        sandbox = daytona.create()
        info = sandbox.info()
        print(f"Sandbox {info.id} is {info.state}")
        print(f"Resources: {info.resources.cpu} CPU, {info.resources.memory} RAM")
        ```
    """

    id: str
    name: Annotated[
        str,
        Field(
            default="",
            deprecated="The `name` field is deprecated.",
        ),
    ]
    image: Optional[str]
    user: str
    env: Dict[str, str]
    labels: Dict[str, str]
    public: bool
    target: SandboxTargetRegion
    resources: SandboxResources
    state: SandboxState
    error_reason: Optional[str]
    snapshot_state: Optional[str]
    snapshot_created_at: Optional[str]
    node_domain: str
    region: str
    class_name: str
    updated_at: str
    last_snapshot: Optional[str]
    auto_stop_interval: int
    auto_archive_interval: int
    provider_metadata: Annotated[
        Optional[str],
        Field(
            deprecated=(
                "The `provider_metadata` field is deprecated. Use `state`, `node_domain`, `region`, `class_name`,"
                " `updated_at`, `last_snapshot`, `resources`, `auto_stop_interval`, `auto_archive_interval` instead."
            )
        ),
    ]


class SandboxInstance(ApiSandbox):
    """Represents a Daytona Sandbox instance."""

    info: Optional[SandboxInfo]
