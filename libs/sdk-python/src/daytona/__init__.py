# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

from daytona_api_client import SandboxState, SessionExecuteResponse

from ._async.daytona import AsyncDaytona
from ._async.sandbox import AsyncSandbox
from ._sync.daytona import Daytona
from ._sync.sandbox import Sandbox

# Create deprecated aliases with proper warnings
from ._utils.deprecation import deprecated_alias
from .common.charts import (
    BarChart,
    BoxAndWhiskerChart,
    Chart,
    ChartType,
    CompositeChart,
    LineChart,
    PieChart,
    ScatterChart,
)
from .common.daytona import (
    CodeLanguage,
    CreateSandboxBaseParams,
    CreateSandboxFromImageParams,
    CreateSandboxFromSnapshotParams,
    DaytonaConfig,
)
from .common.errors import DaytonaError
from .common.filesystem import FileUpload
from .common.image import Image
from .common.lsp_server import LspLanguageId
from .common.process import CodeRunParams, SessionExecuteRequest
from .common.sandbox import Resources, SandboxTargetRegion
from .common.snapshot import CreateSnapshotParams
from .common.volume import VolumeMount

Workspace = deprecated_alias("Workspace", "Sandbox")(Sandbox)
WorkspaceTargetRegion = deprecated_alias("WorkspaceTargetRegion", "SandboxTargetRegion")(SandboxTargetRegion)
WorkspaceResources = deprecated_alias("WorkspaceResources", "SandboxResources")(Resources)
WorkspaceState = deprecated_alias("WorkspaceState", "SandboxState")(SandboxState)

__all__ = [
    "Daytona",
    "DaytonaConfig",
    "CodeLanguage",
    "SessionExecuteRequest",
    "SessionExecuteResponse",
    "DaytonaError",
    "LspLanguageId",
    "WorkspaceTargetRegion",
    "CodeRunParams",
    "Sandbox",
    "SandboxTargetRegion",
    "Resources",
    "SandboxState",
    "Workspace",
    "WorkspaceTargetRegion",
    "WorkspaceResources",
    "WorkspaceState",
    "ChartType",
    "Chart",
    "LineChart",
    "ScatterChart",
    "BarChart",
    "PieChart",
    "BoxAndWhiskerChart",
    "CompositeChart",
    "FileUpload",
    "VolumeMount",
    "AsyncDaytona",
    "AsyncSandbox",
    "Image",
    "CreateSandboxBaseParams",
    "CreateSandboxFromImageParams",
    "CreateSandboxFromSnapshotParams",
    "CreateSnapshotParams",
]
