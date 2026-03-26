# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from typing import Any, Generic, TypeVar

from daytona_toolbox_api_client import ApiClient
from daytona_toolbox_api_client_async import ApiClient as AsyncApiClient

# TypeVar constrained to either sync or async ApiClient
ApiClientT = TypeVar("ApiClientT", ApiClient, AsyncApiClient)


class ToolboxApiClientProxy(Generic[ApiClientT]):
    """Wrapper around an API client that adjusts `param_serialize` method.

    It intercepts `param_serialize` to prepend the sandbox ID to the `resource_path` and
    set `_host` to the toolbox proxy URL, while delegating all other attributes
    and methods to the underlying API client.
    """

    def __init__(self, api_client: ApiClientT, sandbox_id: str, toolbox_proxy_url: str):
        self._api_client: ApiClientT = api_client
        self._sandbox_id: str = sandbox_id
        self._toolbox_base_url: str = toolbox_proxy_url

    def param_serialize(self, *args: object, **kwargs: object) -> Any:
        """Intercepts param_serialize to prepend sandbox ID to resource_path."""
        resource_path = kwargs.get("resource_path")

        if resource_path:
            resource_path = f"/{self._sandbox_id}{resource_path}"

        kwargs["resource_path"] = resource_path
        kwargs["_host"] = self._toolbox_base_url

        return self._api_client.param_serialize(*args, **kwargs)

    def __getattr__(self, name: str) -> Any:
        """Delegate all other attributes to the wrapped client."""
        return getattr(self._api_client, name)
