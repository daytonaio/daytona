# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import importlib
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from daytona_api_client import SandboxState
    from daytona_toolbox_api_client import SessionExecuteResponse

    from ._async.computer_use import AsyncComputerUse, AsyncDisplay, AsyncKeyboard, AsyncMouse, AsyncScreenshot
    from ._async.daytona import AsyncDaytona
    from ._async.sandbox import AsyncPaginatedSandboxes, AsyncSandbox
    from ._sync.daytona import Daytona
    from ._sync.sandbox import Sandbox
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
    from .common.code_interpreter import ExecutionError, ExecutionResult, OutputMessage
    from .common.computer_use import ScreenshotOptions, ScreenshotRegion
    from .common.daytona import (
        CodeLanguage,
        CreateSandboxBaseParams,
        CreateSandboxFromImageParams,
        CreateSandboxFromSnapshotParams,
        DaytonaConfig,
    )
    from .common.errors import (
        DaytonaAuthenticationError,
        DaytonaAuthorizationError,
        DaytonaConflictError,
        DaytonaConnectionError,
        DaytonaError,
        DaytonaNotFoundError,
        DaytonaRateLimitError,
        DaytonaTimeoutError,
        DaytonaValidationError,
        DownloadAbortedError,
        UploadAbortedError,
    )
    from .common.filesystem import (
        DownloadFileOptions,
        FileDownloadErrorDetails,
        FileDownloadRequest,
        FileDownloadResponse,
        FileUpload,
        TransferProgressCallback,
        UploadFileOptions,
    )
    from .common.image import Image
    from .common.lsp_server import LspCompletionPosition, LspLanguageId
    from .common.process import CodeRunParams, ExecuteResponse, ExecutionArtifacts, OutputHandler, SessionExecuteRequest
    from .common.pty import PtySize
    from .common.sandbox import Resources
    from .common.snapshot import CreateSnapshotParams
    from .common.volume import VolumeMount

__all__ = [
    "Daytona",
    "DaytonaConfig",
    "CodeLanguage",
    "SessionExecuteRequest",
    "SessionExecuteResponse",
    "DaytonaError",
    "DaytonaNotFoundError",
    "DaytonaRateLimitError",
    "LspLanguageId",
    "CodeRunParams",
    "Sandbox",
    "Resources",
    "SandboxState",
    "ChartType",
    "Chart",
    "LineChart",
    "ScatterChart",
    "BarChart",
    "PieChart",
    "BoxAndWhiskerChart",
    "CompositeChart",
    "FileDownloadRequest",
    "FileDownloadResponse",
    "FileDownloadErrorDetails",
    "FileUpload",
    "VolumeMount",
    "AsyncDaytona",
    "AsyncSandbox",
    "AsyncPaginatedSandboxes",
    "AsyncComputerUse",
    "AsyncMouse",
    "AsyncKeyboard",
    "AsyncScreenshot",
    "AsyncDisplay",
    "ScreenshotRegion",
    "ScreenshotOptions",
    "Image",
    "CreateSandboxBaseParams",
    "CreateSandboxFromImageParams",
    "CreateSandboxFromSnapshotParams",
    "CreateSnapshotParams",
    "PtySize",
    "LspCompletionPosition",
    "ExecutionArtifacts",
    "ExecuteResponse",
    "ExecutionResult",
    "ExecutionError",
    "OutputMessage",
    "OutputHandler",
    "DaytonaTimeoutError",
    "DaytonaAuthenticationError",
    "DaytonaAuthorizationError",
    "DaytonaConflictError",
    "DaytonaValidationError",
    "DaytonaConnectionError",
    "DownloadAbortedError",
    "DownloadFileOptions",
    "TransferProgressCallback",
    "UploadAbortedError",
    "UploadFileOptions",
]

# Mapping of symbol name -> (absolute module path, attribute name) for external packages
_EXTERNAL_IMPORTS: dict[str, tuple[str, str]] = {
    "SandboxState": ("daytona_api_client.models.sandbox_state", "SandboxState"),
    "SessionExecuteResponse": (
        "daytona_toolbox_api_client.models.session_execute_response",
        "SessionExecuteResponse",
    ),
}

# Mapping of symbol name -> relative submodule path (within the daytona package)
_DYNAMIC_IMPORTS: dict[str, str] = {
    # _sync
    "Daytona": "_sync.daytona",
    "Sandbox": "_sync.sandbox",
    # _async
    "AsyncDaytona": "_async.daytona",
    "AsyncSandbox": "_async.sandbox",
    "AsyncPaginatedSandboxes": "_async.sandbox",
    "AsyncComputerUse": "_async.computer_use",
    "AsyncMouse": "_async.computer_use",
    "AsyncKeyboard": "_async.computer_use",
    "AsyncScreenshot": "_async.computer_use",
    "AsyncDisplay": "_async.computer_use",
    # common.charts
    "BarChart": "common.charts",
    "BoxAndWhiskerChart": "common.charts",
    "Chart": "common.charts",
    "ChartType": "common.charts",
    "CompositeChart": "common.charts",
    "LineChart": "common.charts",
    "PieChart": "common.charts",
    "ScatterChart": "common.charts",
    # common.code_interpreter
    "ExecutionError": "common.code_interpreter",
    "ExecutionResult": "common.code_interpreter",
    "OutputMessage": "common.code_interpreter",
    # common.computer_use
    "ScreenshotOptions": "common.computer_use",
    "ScreenshotRegion": "common.computer_use",
    # common.daytona
    "CodeLanguage": "common.daytona",
    "CreateSandboxBaseParams": "common.daytona",
    "CreateSandboxFromImageParams": "common.daytona",
    "CreateSandboxFromSnapshotParams": "common.daytona",
    "DaytonaConfig": "common.daytona",
    # common.errors
    "DaytonaError": "common.errors",
    "DaytonaNotFoundError": "common.errors",
    "DaytonaRateLimitError": "common.errors",
    "DaytonaTimeoutError": "common.errors",
    "DaytonaAuthenticationError": "common.errors",
    "DaytonaAuthorizationError": "common.errors",
    "DaytonaConflictError": "common.errors",
    "DaytonaValidationError": "common.errors",
    "DaytonaConnectionError": "common.errors",
    "DownloadAbortedError": "common.errors",
    "UploadAbortedError": "common.errors",
    # common.filesystem
    "FileDownloadErrorDetails": "common.filesystem",
    "FileDownloadRequest": "common.filesystem",
    "FileDownloadResponse": "common.filesystem",
    "FileUpload": "common.filesystem",
    "DownloadFileOptions": "common.filesystem",
    "TransferProgressCallback": "common.filesystem",
    "UploadFileOptions": "common.filesystem",
    # common.image
    "Image": "common.image",
    # common.lsp_server
    "LspCompletionPosition": "common.lsp_server",
    "LspLanguageId": "common.lsp_server",
    # common.process
    "CodeRunParams": "common.process",
    "ExecuteResponse": "common.process",
    "ExecutionArtifacts": "common.process",
    "OutputHandler": "common.process",
    "SessionExecuteRequest": "common.process",
    # common.pty
    "PtySize": "common.pty",
    # common.sandbox
    "Resources": "common.sandbox",
    # common.snapshot
    "CreateSnapshotParams": "common.snapshot",
    # common.volume
    "VolumeMount": "common.volume",
}


def __getattr__(attr_name: str) -> object:
    # Check external package imports first
    external = _EXTERNAL_IMPORTS.get(attr_name)
    if external is not None:
        module_path, name = external
        mod = importlib.import_module(module_path)
        value = getattr(mod, name)
        globals()[attr_name] = value
        return value

    # Check internal submodule imports
    submodule = _DYNAMIC_IMPORTS.get(attr_name)
    if submodule is not None:
        mod = importlib.import_module(f".{submodule}", __name__)
        value = getattr(mod, attr_name)
        globals()[attr_name] = value
        return value

    raise AttributeError(f"module {__name__!r} has no attribute {attr_name!r}")


def __dir__() -> list[str]:
    return list(__all__)
