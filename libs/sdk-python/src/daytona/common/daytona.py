# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import warnings
from dataclasses import dataclass
from enum import Enum
from typing import Annotated, Dict, List, Optional, Union

from pydantic import BaseModel, Field, model_validator

from .image import Image
from .sandbox import Resources
from .volume import VolumeMount


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
        target (Optional[str]): Target runner location for the Sandbox. Defaults to `'us'` if not set here
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
    ] = None
    target: Optional[str] = None
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


class CreateSandboxBaseParams(BaseModel):
    """Base parameters for creating a new Sandbox.

    Attributes:
        language (Optional[CodeLanguage]): Programming language for the Sandbox. Defaults to "python".
        os_user (Optional[str]): OS user for the Sandbox.
        env_vars (Optional[Dict[str, str]]): Environment variables to set in the Sandbox.
        labels (Optional[Dict[str, str]]): Custom labels for the Sandbox.
        public (Optional[bool]): Whether the Sandbox should be public.
        timeout (Optional[float]): Timeout in seconds for Sandbox to be created and started.
        auto_stop_interval (Optional[int]): Interval in minutes after which Sandbox will
            automatically stop if no Sandbox event occurs during that time. Default is 15 minutes.
            0 means no auto-stop.
        auto_archive_interval (Optional[int]): Interval in minutes after which a continuously stopped Sandbox will
            automatically archive. Default is 7 days.
            0 means the maximum interval will be used.
        auto_delete_interval (Optional[int]): Interval in minutes after which a continuously stopped Sandbox will
            automatically be deleted. By default, auto-delete is disabled.
            Negative value means disabled, 0 means delete immediately upon stopping.
        volumes (Optional[List[VolumeMount]]): List of volumes mounts to attach to the Sandbox.
    """

    language: Optional[CodeLanguage] = None
    os_user: Optional[str] = None
    env_vars: Optional[Dict[str, str]] = None
    labels: Optional[Dict[str, str]] = None
    public: Optional[bool] = None
    auto_stop_interval: Optional[int] = None
    auto_archive_interval: Optional[int] = None
    auto_delete_interval: Optional[int] = None
    volumes: Optional[List[VolumeMount]] = None


class CreateSandboxFromImageParams(CreateSandboxBaseParams):
    """Parameters for creating a new Sandbox from an image.

    Attributes:
        image (Union[str, Image]): Custom Docker image to use for the Sandbox. If an Image object is provided,
            the image will be dynamically built.
        resources (Optional[Resources]): Resource configuration for the Sandbox. If not provided, sandbox will
            have default resources.
    """

    image: Union[str, Image]
    resources: Optional[Resources] = None


class CreateSandboxFromSnapshotParams(CreateSandboxBaseParams):
    """Parameters for creating a new Sandbox from a snapshot.

    Attributes:
        snapshot (Optional[str]): Name of the snapshot to use for the Sandbox.
    """

    snapshot: Optional[str] = None
