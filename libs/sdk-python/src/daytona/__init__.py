# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import importlib
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from daytona_api_client import SandboxListSortDirection, SandboxListSortField, SandboxState
    from daytona_api_client_async import GpuType, SandboxClass
    from daytona_toolbox_api_client import SessionExecuteResponse

    from ._async.computer_use import AsyncComputerUse, AsyncDisplay, AsyncKeyboard, AsyncMouse, AsyncScreenshot
    from ._async.daytona import AsyncDaytona
    from ._async.sandbox import AsyncSandbox
    from ._async.session import AsyncSessionService
    from ._sync.daytona import Daytona
    from ._sync.sandbox import Sandbox
    from ._sync.session import SessionService
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
    )
    from .common.filesystem import (
        CancelEvent,
        DownloadProgress,
        FileDownloadErrorDetails,
        FileDownloadRequest,
        FileDownloadResponse,
        FileUpload,
        UploadProgress,
    )
    from .common.image import Image
    from .common.lsp_server import LspCompletionPosition, LspLanguageId
    from .common.process import CodeRunParams, ExecuteResponse, ExecutionArtifacts, OutputHandler, SessionExecuteRequest
    from .common.pty import PtySize
    from .common.sandbox import ListSandboxesQuery, Resources
    from .common.session import (
        SessionAccess,
        SessionDisplay,
        SessionExecutionError,
        SessionExpiredError,
        SessionInvalidatedError,
        SessionRef,
        SessionRunOptions,
        SessionRunResult,
    )
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
    "GpuType",
    "SandboxClass",
    "SandboxState",
    "SandboxListSortField",
    "SandboxListSortDirection",
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
    "DownloadProgress",
    "UploadProgress",
    "CancelEvent",
    "FileUpload",
    "VolumeMount",
    "ListSandboxesQuery",
    "AsyncDaytona",
    "AsyncSandbox",
    "AsyncComputerUse",
    "AsyncMouse",
    "AsyncKeyboard",
    "AsyncScreenshot",
    "AsyncDisplay",
    "ScreenshotRegion",
    "ScreenshotOptions",
    "Image",
    "SessionAccess",
    "SessionExpiredError",
    "SessionInvalidatedError",
    "SessionRef",
    "SessionDisplay",
    "SessionExecutionError",
    "SessionRunOptions",
    "SessionRunResult",
    "SessionService",
    "AsyncSessionService",
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
]

# Mapping of symbol name -> (absolute module path, attribute name) for external packages
_EXTERNAL_IMPORTS: dict[str, tuple[str, str]] = {
    "GpuType": ("daytona_api_client_async.models.gpu_type", "GpuType"),
    "SandboxClass": ("daytona_api_client_async.models.sandbox_class", "SandboxClass"),
    "SandboxState": ("daytona_api_client.models.sandbox_state", "SandboxState"),
    "SandboxListSortField": ("daytona_api_client.models.sandbox_list_sort_field", "SandboxListSortField"),
    "SandboxListSortDirection": (
        "daytona_api_client.models.sandbox_list_sort_direction",
        "SandboxListSortDirection",
    ),
    "SessionExecuteResponse": ("daytona_toolbox_api_client.models.session_execute_response", "SessionExecuteResponse"),
}

# Mapping of symbol name -> relative submodule path (within the daytona package)
_DYNAMIC_IMPORTS: dict[str, str] = {
    # _sync
    "Daytona": "_sync.daytona",
    "Sandbox": "_sync.sandbox",
    # _async
    "AsyncDaytona": "_async.daytona",
    "AsyncSandbox": "_async.sandbox",
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
    # common.filesystem
    "FileDownloadErrorDetails": "common.filesystem",
    "DownloadProgress": "common.filesystem",
    "UploadProgress": "common.filesystem",
    "CancelEvent": "common.filesystem",
    "FileDownloadRequest": "common.filesystem",
    "FileDownloadResponse": "common.filesystem",
    "FileUpload": "common.filesystem",
    # common.image
    "Image": "common.image",
    # common.session
    "SessionAccess": "common.session",
    "SessionExpiredError": "common.session",
    "SessionInvalidatedError": "common.session",
    "SessionRef": "common.session",
    "SessionDisplay": "common.session",
    "SessionExecutionError": "common.session",
    "SessionRunOptions": "common.session",
    "SessionRunResult": "common.session",
    # _sync.session
    "SessionService": "_sync.session",
    # _async.session
    "AsyncSessionService": "_async.session",
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
    "ListSandboxesQuery": "common.sandbox",
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
