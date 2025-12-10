# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import warnings
from dataclasses import dataclass
from typing import ClassVar

from daytona_toolbox_api_client import SessionExecuteRequest as ApiSessionExecuteRequest
from daytona_toolbox_api_client import SessionExecuteResponse as ApiSessionExecuteResponse
from daytona_toolbox_api_client_async import SessionExecuteRequest as AsyncApiSessionExecuteRequest
from pydantic import BaseModel, ConfigDict, Field, model_validator

from .charts import Chart

# 3-byte multiplexing markers inserted by the shell labelers
STDOUT_PREFIX: bytes = b"\x01\x01\x01"
STDERR_PREFIX: bytes = b"\x02\x02\x02"
MAX_PREFIX_LEN: int = max(len(STDOUT_PREFIX), len(STDERR_PREFIX))


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


def demux_log(data: bytes) -> tuple[bytes, bytes]:
    """Demultiplex combined stdout/stderr log data.

    Args:
        data: Combined log bytes with STDOUT_PREFIX and STDERR_PREFIX markers

    Returns:
        Tuple of (stdout_bytes, stderr_bytes)
    """
    out_buf = bytearray()
    err_buf = bytearray()
    state = ""  # none, stdout, stderr

    while len(data) > 0:
        # Find the nearest marker (stdout or stderr)
        si = data.find(STDOUT_PREFIX)
        ei = data.find(STDERR_PREFIX)

        # Pick the closest marker index and type
        next_idx = -1
        next_marker = ""
        if si != -1 and (ei == -1 or si < ei):
            next_idx, next_marker = si, "stdout"
        elif ei != -1:
            next_idx, next_marker = ei, "stderr"

        if next_idx == -1:
            # No more markers → dump remainder into current state
            if state == "stdout":
                out_buf.extend(data)
            elif state == "stderr":
                err_buf.extend(data)
            break

        # Write everything before the marker into current state
        if state == "stdout":
            out_buf.extend(data[:next_idx])
        elif state == "stderr":
            err_buf.extend(data[:next_idx])

        # Advance past marker and switch state
        if next_marker == "stdout":
            data = data[next_idx + len(STDOUT_PREFIX) :]
            state = "stdout"
        else:
            data = data[next_idx + len(STDERR_PREFIX) :]
            state = "stderr"

    return bytes(out_buf), bytes(err_buf)
