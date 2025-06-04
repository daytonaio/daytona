# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

import warnings
from dataclasses import dataclass
from enum import Enum
from typing import Annotated, Dict, List, Optional, Union

from daytona_sdk.common.image import Image
from daytona_sdk.common.sandbox import SandboxTargetRegion
from daytona_sdk.common.volume import VolumeMount
from pydantic import BaseModel, Field, model_validator


@dataclass
class CodeLanguage(Enum):
    """Programming languages supported by Daytona

    **Enum Members**:
        - `PYTHON` ("python")
        - `TYPESCRIPT` ("typescript")
        - `JAVASCRIPT` ("javascript")
    """

    PYTHON = "python"
    TYPESCRIPT = "typescript"
    JAVASCRIPT = "javascript"

    def __str__(self):
        return self.value

    def __eq__(self, other):
        if isinstance(other, str):
            return self.value == other
        return super().__eq__(other)


class DaytonaConfig(BaseModel):
    """Configuration options for initializing the Daytona client.

    Attributes:
        api_key (Optional[str]): API key for authentication with the Daytona API. If not set, it must be provided
            via the environment variable `DAYTONA_API_KEY`, or a JWT token must be provided instead.
        jwt_token (Optional[str]): JWT token for authentication with the Daytona API. If not set, it must be provided
            via the environment variable `DAYTONA_JWT_TOKEN`, or an API key must be provided instead.
        organization_id (Optional[str]): Organization ID used for JWT-based authentication. Required if a JWT token
            is provided, and must be set either here or in the environment variable `DAYTONA_ORGANIZATION_ID`.
        api_url (Optional[str]): URL of the Daytona API. Defaults to `'https://app.daytona.io/api'` if not set
            here or in the environment variable `DAYTONA_API_URL`.
        server_url (Optional[str]): Deprecated. Use `api_url` instead. This property will be removed
            in a future version.
        target (Optional[SandboxTargetRegion]): Target environment for the Sandbox. Defaults to `'us'` if not set here
            or in the environment variable `DAYTONA_TARGET`.

    Example:
        ```python
        config = DaytonaConfig(api_key="your-api-key")
        ```
        ```python
        config = DaytonaConfig(jwt_token="your-jwt-token", organization_id="your-organization-id")
        ```
    """

    api_key: Optional[str] = None
    api_url: Optional[str] = None
    server_url: Annotated[
        Optional[str],
        Field(
            default=None,
            deprecated="`server_url` is deprecated and will be removed in a future version. Use `api_url` instead.",
        ),
    ]
    target: Optional[SandboxTargetRegion] = None
    jwt_token: Optional[str] = None
    organization_id: Optional[str] = None

    @model_validator(mode="before")
    @classmethod
    def __handle_deprecated_server_url(cls, values):  # pylint: disable=unused-private-member
        if "server_url" in values and values.get("server_url"):
            warnings.warn(
                "'server_url' is deprecated and will be removed in a future version. Use 'api_url' instead.",
                DeprecationWarning,
                stacklevel=3,
            )
            if "api_url" not in values or not values["api_url"]:
                values["api_url"] = values["server_url"]
        return values


@dataclass
class SandboxResources:
    """Resources configuration for Sandbox.

    Attributes:
        cpu (Optional[int]): Number of CPU cores to allocate.
        memory (Optional[int]): Amount of memory in GB to allocate.
        disk (Optional[int]): Amount of disk space in GB to allocate.
        gpu (Optional[int]): Number of GPUs to allocate.

    Example:
        ```python
        resources = SandboxResources(
            cpu=2,
            memory=4,  # 4GB RAM
            disk=20,   # 20GB disk
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


class CreateSandboxParams(BaseModel):
    """Parameters for creating a new Sandbox.

    Attributes:
        language (Optional[CodeLanguage]): Programming language for the Sandbox ("python", "javascript", "typescript").
        Defaults to "python".
        image (Optional[Union[str, Image]]): Custom Docker image to use for the Sandbox. If an Image object is provided,
            the image will be dynamically built.
        os_user (Optional[str]): OS user for the Sandbox.
        env_vars (Optional[Dict[str, str]]): Environment variables to set in the Sandbox.
        labels (Optional[Dict[str, str]]): Custom labels for the Sandbox.
        public (Optional[bool]): Whether the Sandbox should be public.
        resources (Optional[SandboxResources]): Resource configuration for the Sandbox.
        timeout (Optional[float]): Timeout in seconds for Sandbox to be created and started.
        auto_stop_interval (Optional[int]): Interval in minutes after which Sandbox will
            automatically stop if no Sandbox event occurs during that time. Default is 15 minutes.
            0 means no auto-stop.
        auto_archive_interval (Optional[int]): Interval in minutes after which a continuously stopped Sandbox will
            automatically archive. Default is 7 days.
            0 means the maximum interval will be used.

    Example:
        ```python
        params = CreateSandboxParams(
            language="python",
            env_vars={"DEBUG": "true"},
            resources=SandboxResources(cpu=2, memory=4),
            auto_stop_interval=20,
            auto_archive_interval=60
        )
        sandbox = daytona.create(params, 50)
        ```
    """

    language: Optional[CodeLanguage] = None
    image: Optional[Union[str, Image]] = None
    os_user: Optional[str] = None
    env_vars: Optional[Dict[str, str]] = None
    labels: Optional[Dict[str, str]] = None
    public: Optional[bool] = None
    resources: Optional[SandboxResources] = None
    timeout: Annotated[
        Optional[float],
        Field(
            default=None,
            deprecated=(
                "The `timeout` field is deprecated and will be removed in future versions. "
                "Use `timeout` argument in method calls instead."
            ),
        ),
    ]
    auto_stop_interval: Optional[int] = None
    auto_archive_interval: Optional[int] = None
    volumes: Optional[List[VolumeMount]] = None

    @model_validator(mode="before")
    @classmethod
    def __handle_deprecated_timeout(cls, values):  # pylint: disable=unused-private-member
        if "timeout" in values and values.get("timeout"):
            warnings.warn(
                "The `timeout` field is deprecated and will be removed in future versions. "
                + "Use `timeout` argument in method calls instead.",
                DeprecationWarning,
                stacklevel=3,
            )
        return values
