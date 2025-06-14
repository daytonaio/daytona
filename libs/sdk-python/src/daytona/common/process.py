# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import warnings
from dataclasses import dataclass
from typing import Dict, List, Optional

from daytona_api_client import ExecuteResponse as ClientExecuteResponse
from daytona_api_client import SessionExecuteRequest as ApiSessionExecuteRequest
from daytona_api_client_async import SessionExecuteRequest as AsyncApiSessionExecuteRequest
from pydantic import ConfigDict, model_validator

from .charts import Chart


@dataclass
class CodeRunParams:
    """Parameters for code execution.

    Attributes:
        argv (Optional[List[str]]): Command line arguments
        env (Optional[Dict[str, str]]): Environment variables
    """

    argv: Optional[List[str]] = None
    env: Optional[Dict[str, str]] = None


class SessionExecuteRequest(ApiSessionExecuteRequest, AsyncApiSessionExecuteRequest):
    """Contains the request for executing a command in a session.

    Attributes:
        command (str): The command to execute.
        run_async (Optional[bool]): Whether to execute the command asynchronously.
        var_async (Optional[bool]): Deprecated. Use `run_async` instead.
    """

    @model_validator(mode="before")
    @classmethod
    def __handle_deprecated_var_async(cls, values):  # pylint: disable=unused-private-member
        if "var_async" in values and values.get("var_async"):
            warnings.warn(
                "'var_async' is deprecated and will be removed in a future version. Use 'run_async' instead.",
                DeprecationWarning,
                stacklevel=3,
            )
            if "run_async" not in values or not values["run_async"]:
                values["run_async"] = values.pop("var_async")
        return values


class ExecutionArtifacts:
    """Artifacts from the command execution.

    Attributes:
        stdout (str): Standard output from the command, same as `result` in `ExecuteResponse`
        charts (Optional[List[Chart]]): List of chart metadata from matplotlib
    """

    stdout: str
    charts: Optional[List[Chart]] = None

    def __init__(self, stdout: str = "", charts: Optional[List[Chart]] = None):
        self.stdout = stdout
        self.charts = charts


class ExecuteResponse(ClientExecuteResponse):
    """Response from the command execution.

    Attributes:
        exit_code (int): The exit code from the command execution
        result (str): The output from the command execution
        artifacts (Optional[ExecutionArtifacts]): Artifacts from the command execution
    """

    artifacts: Optional[ExecutionArtifacts] = None

    # TODO: Remove model_config once everything is migrated to pydantic # pylint: disable=fixme
    model_config = ConfigDict(arbitrary_types_allowed=True)

    # pylint: disable=super-init-not-called
    def __init__(
        self,
        exit_code: int,
        result: str,
        artifacts: Optional[ExecutionArtifacts] = None,
        additional_properties: Dict = None,
    ):
        self.exit_code = exit_code
        self.result = result
        self.additional_properties = additional_properties or {}
        self.artifacts = artifacts
