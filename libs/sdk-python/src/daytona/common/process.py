# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import re
import warnings
from collections.abc import Awaitable, Callable
from dataclasses import dataclass
from typing import ClassVar, TypeVar, Union

from pydantic import BaseModel, ConfigDict, Field, model_validator

from daytona_toolbox_api_client import SessionExecuteRequest as ApiSessionExecuteRequest
from daytona_toolbox_api_client import SessionExecuteResponse as ApiSessionExecuteResponse
from daytona_toolbox_api_client_async import SessionExecuteRequest as AsyncApiSessionExecuteRequest

from .charts import Chart

# 3-byte multiplexing markers inserted by the shell labelers
STDOUT_PREFIX: bytes = b"\x01\x01\x01"
STDERR_PREFIX: bytes = b"\x02\x02\x02"
MAX_PREFIX_LEN: int = max(len(STDOUT_PREFIX), len(STDERR_PREFIX))

_VALID_ENV_KEY_REGEX = re.compile(r"^[A-Za-z_][A-Za-z0-9_]*$")


@dataclass
class CodeRunParams:
    """Parameters for code execution.

    Attributes:
        argv (list[str] | None): Command line arguments
        env (dict[str, str] | None): Environment variables
    """

    argv: list[str] | None = None
    env: dict[str, str] | None = None


class SessionExecuteRequest(ApiSessionExecuteRequest, AsyncApiSessionExecuteRequest):
    """Contains the request for executing a command in a session.

    Attributes:
        command (str): The command to execute.
        run_async (bool | None): Whether to execute the command asynchronously.
        var_async (bool | None): Deprecated. Use `run_async` instead.
        suppress_input_echo (bool | None): Whether to suppress input echo. Default is `False`.
    """

    @model_validator(mode="before")
    @classmethod
    def _handle_deprecated_var_async(cls, values: dict[str, object]):
        if "var_async" in values and values.get("var_async"):
            warnings.warn(
                "'var_async' is deprecated and will be removed in a future version. Use 'run_async' instead.",
                DeprecationWarning,
                stacklevel=3,
            )
            if "run_async" not in values or not values["run_async"]:
                values["run_async"] = values.pop("var_async")
        return values


@dataclass
class ExecutionArtifacts:
    """Artifacts from the command execution.

    Attributes:
        stdout (str): Standard output from the command, same as `result` in `ExecuteResponse`
        charts (list[Chart] | None): List of chart metadata from matplotlib
    """

    stdout: str = ""
    charts: list[Chart] | None = None


class ExecuteResponse(BaseModel):
    """Response from the command execution.

    Attributes:
        exit_code (int): The exit code from the command execution
        result (str): The output from the command execution
        artifacts (ExecutionArtifacts | None): Artifacts from the command execution
    """

    exit_code: int
    result: str
    artifacts: ExecutionArtifacts | None = None
    additional_properties: dict[str, object] = Field(default_factory=dict)

    model_config: ClassVar[ConfigDict] = ConfigDict(arbitrary_types_allowed=True, extra="allow")


class SessionExecuteResponse(ApiSessionExecuteResponse):
    """Response from the session command execution.

    Attributes:
        cmd_id (str): The ID of the executed command
        stdout (str | None): The stdout from the command execution
        stderr (str | None): The stderr from the command execution
        output (str): The output from the command execution
        exit_code (int): The exit code from the command execution
    """

    cmd_id: str
    stdout: str | None = None
    stderr: str | None = None
    output: str | None = None
    exit_code: int | None = None
    additional_properties: dict[str, object] = Field(default_factory=dict)

    model_config: ClassVar[ConfigDict] = ConfigDict(arbitrary_types_allowed=True)


@dataclass
class SessionCommandLogsResponse:
    """Response from the command logs.

    Attributes:
        output (str | None): The combined output from the command
        stdout (str | None): The stdout from the command
        stderr (str | None): The stderr from the command
    """

    output: str | None = None
    stdout: str | None = None
    stderr: str | None = None


def parse_session_command_logs(data: bytes) -> SessionCommandLogsResponse:
    """Parse combined stdout/stderr output into separate streams.

    Args:
        data: Combined log bytes with STDOUT_PREFIX and STDERR_PREFIX markers

    Returns:
        SessionCommandLogsResponse with separated stdout and stderr
    """
    stdout_bytes, stderr_bytes = demux_log(data)

    # Convert bytes to strings, ignoring potential encoding issues
    stdout_str = stdout_bytes.decode("utf-8", "ignore")
    stderr_str = stderr_bytes.decode("utf-8", "ignore")

    # For backwards compatibility, logs field contains the original combined data
    output_str = data.decode("utf-8", "ignore")

    return SessionCommandLogsResponse(output=output_str, stdout=stdout_str, stderr=stderr_str)


_MARKER_RE = re.compile(re.escape(STDOUT_PREFIX) + b"|" + re.escape(STDERR_PREFIX))


def demux_log(data: bytes) -> tuple[bytes, bytes]:
    """Demultiplex combined stdout/stderr log data.

    Args:
        data: Combined log bytes with STDOUT_PREFIX and STDERR_PREFIX markers

    Returns:
        Tuple of (stdout_bytes, stderr_bytes)
    """
    out_parts: list[bytes] = []
    err_parts: list[bytes] = []
    state = ""
    pos = 0

    for match in _MARKER_RE.finditer(data):
        start = match.start()
        if pos < start:
            chunk = data[pos:start]
            if state == "stdout":
                out_parts.append(chunk)
            elif state == "stderr":
                err_parts.append(chunk)

        state = "stdout" if match.group() == STDOUT_PREFIX else "stderr"
        pos = match.end()

    if pos < len(data):
        tail = data[pos:]
        if state == "stdout":
            out_parts.append(tail)
        elif state == "stderr":
            err_parts.append(tail)

    return b"".join(out_parts), b"".join(err_parts)


# Type aliases for callbacks
T = TypeVar("T")
OutputHandler = Union[
    Callable[[T], None],
    Callable[[T], Awaitable[None]],
]
"""Callback type that accepts both sync and async handlers.

Blocking synchronous operations inside handlers may cause WebSocket disconnections.
"""
