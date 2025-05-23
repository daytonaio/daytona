# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

from daytona_api_client import WorkspaceState as SandboxState

# Create deprecated aliases with proper warnings
from ._utils.deprecation import deprecated_alias
from .charts import BarChart, BoxAndWhiskerChart, Chart, ChartType, CompositeChart, LineChart, PieChart, ScatterChart
from .common.code_run_params import CodeRunParams
from .daytona import (
    CodeLanguage,
    CreateSandboxParams,
    Daytona,
    DaytonaConfig,
    DaytonaError,
    SandboxResources,
    SessionExecuteResponse,
    VolumeMount,
)
from .filesystem import FileUpload
from .image import Image
from .lsp_server import LspLanguageId
from .process import SessionExecuteRequest
from .sandbox import Sandbox, SandboxState, SandboxTargetRegion

CreateWorkspaceParams = deprecated_alias("CreateWorkspaceParams", "CreateSandboxParams")(CreateSandboxParams)
Workspace = deprecated_alias("Workspace", "Sandbox")(Sandbox)
WorkspaceTargetRegion = deprecated_alias("WorkspaceTargetRegion", "SandboxTargetRegion")(SandboxTargetRegion)
WorkspaceResources = deprecated_alias("WorkspaceResources", "SandboxResources")(SandboxResources)
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
    "CreateSandboxParams",
    "Sandbox",
    "SandboxTargetRegion",
    "SandboxResources",
    "SandboxState",
    "CreateWorkspaceParams",
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
    "Image",
]
