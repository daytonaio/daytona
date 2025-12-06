# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import warnings
from enum import Enum
from typing import Annotated, Literal

from pydantic import BaseModel, Field, model_validator
from typing_extensions import override

from .image import Image
from .sandbox import Resources
from .volume import VolumeMount


class CodeLanguage(str, Enum):
    """Programming languages supported by Daytona

    **Enum Members**:
        - `PYTHON` ("python")
        - `TYPESCRIPT` ("typescript")
        - `JAVASCRIPT` ("javascript")
    """

    PYTHON = "python"
    TYPESCRIPT = "typescript"
    JAVASCRIPT = "javascript"

    @override
    def __str__(self):
        return self.value

    @override
    def __eq__(self, other: str | CodeLanguage | object) -> bool:
        if isinstance(other, str):
            return self.value == other
        if isinstance(other, CodeLanguage):
            return self is other
        return False


CodeLanguageLiteral = Literal["python", "typescript", "javascript"]


class DaytonaConfig(BaseModel):
    """Configuration options for initializing the Daytona client.

    Attributes:
        api_key (str | None): API key for authentication with the Daytona API. If not set, it must be provided
            via the environment variable `DAYTONA_API_KEY`, or a JWT token must be provided instead.
        jwt_token (str | None): JWT token for authentication with the Daytona API. If not set, it must be provided
            via the environment variable `DAYTONA_JWT_TOKEN`, or an API key must be provided instead.
        organization_id (str | None): Organization ID used for JWT-based authentication. Required if a JWT token
            is provided, and must be set either here or in the environment variable `DAYTONA_ORGANIZATION_ID`.
        api_url (str | None): URL of the Daytona API. Defaults to `'https://app.daytona.io/api'` if not set
            here or in the environment variable `DAYTONA_API_URL`.
        server_url (str | None): Deprecated. Use `api_url` instead. This property will be removed
            in a future version.
        target (str | None): Target runner location for the Sandbox. Default region for the organization is used
            if not set here or in the environment variable `DAYTONA_TARGET`.

    Example:
        ```python
        config = DaytonaConfig(api_key="your-api-key")
        ```
        ```python
        config = DaytonaConfig(jwt_token="your-jwt-token", organization_id="your-organization-id")
        ```
    """

    api_key: str | None = None
    api_url: str | None = None
    server_url: Annotated[
        str | None,
        Field(
            default=None,
            deprecated="`server_url` is deprecated and will be removed in a future version. Use `api_url` instead.",
        ),
    ] = None
    target: str | None = None
    jwt_token: str | None = None
    organization_id: str | None = None

    @model_validator(mode="before")
    @classmethod
    def _handle_deprecated_server_url(cls, values: dict[str, object]) -> dict[str, object]:
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
        name (str | None): Name of the Sandbox.
        language (CodeLanguage | CodeLanguageLiteral | None): Programming language for the Sandbox.
            Defaults to "python".
        os_user (str | None): OS user for the Sandbox.
        env_vars (dict[str, str] | None): Environment variables to set in the Sandbox.
        labels (dict[str, str] | None): Custom labels for the Sandbox.
        public (bool | None): Whether the Sandbox should be public.
        timeout (float | None): Timeout in seconds for Sandbox to be created and started.
        auto_stop_interval (int | None): Interval in minutes after which Sandbox will
            automatically stop if no Sandbox event occurs during that time. Default is 15 minutes.
            0 means no auto-stop.
        auto_archive_interval (int | None): Interval in minutes after which a continuously stopped Sandbox will
            automatically archive. Default is 7 days.
            0 means the maximum interval will be used.
        auto_delete_interval (int | None): Interval in minutes after which a continuously stopped Sandbox will
            automatically be deleted. By default, auto-delete is disabled.
            Negative value means disabled, 0 means delete immediately upon stopping.
        volumes (list[VolumeMount] | None): List of volumes mounts to attach to the Sandbox.
        network_block_all (bool | None): Whether to block all network access for the Sandbox.
        network_allow_list (str | None): Comma-separated list of allowed CIDR network addresses for the Sandbox.
        ephemeral (bool | None): Whether the Sandbox should be ephemeral.
            If True, auto_delete_interval will be set to 0.
    """

    name: str | None = None
    language: CodeLanguage | CodeLanguageLiteral | None = None
    os_user: str | None = None
    env_vars: dict[str, str] | None = None
    labels: dict[str, str] | None = None
    public: bool | None = None
    auto_stop_interval: int | None = None
    auto_archive_interval: int | None = None
    auto_delete_interval: int | None = None
    volumes: list[VolumeMount] | None = None
    network_block_all: bool | None = None
    network_allow_list: str | None = None
    ephemeral: bool | None = None

    @model_validator(mode="before")
    @classmethod
    def _handle_ephemeral_auto_delete_conflict(cls, values: dict[str, object]) -> dict[str, object]:
        if "ephemeral" in values and values.get("ephemeral"):
            if "auto_delete_interval" in values and values.get("auto_delete_interval"):
                warnings.warn(
                    """'ephemeral' and 'auto_delete_interval' cannot be used together.
                        If ephemeral is True, auto_delete_interval will be ignored and set to 0.""",
                    UserWarning,
                    stacklevel=3,
                )
            values["auto_delete_interval"] = 0
        return values


class CreateSandboxFromImageParams(CreateSandboxBaseParams):
    """Parameters for creating a new Sandbox from an image.

    Attributes:
        image (str | Image): Custom Docker image to use for the Sandbox. If an Image object is provided,
            the image will be dynamically built.
        resources (Resources | None): Resource configuration for the Sandbox. If not provided, sandbox will
            have default resources.
    """

    image: str | Image
    resources: Resources | None = None


class CreateSandboxFromSnapshotParams(CreateSandboxBaseParams):
    """Parameters for creating a new Sandbox from a snapshot.

    Attributes:
        snapshot (str | None): Name of the snapshot to use for the Sandbox.
    """

    snapshot: str | None = None
