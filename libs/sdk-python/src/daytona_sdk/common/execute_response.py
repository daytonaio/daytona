# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

from typing import Dict, List, Optional

from daytona_api_client import ExecuteResponse as ClientExecuteResponse
from pydantic import ConfigDict

from ..charts import Chart


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
