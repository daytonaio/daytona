# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import functools
from typing import Any, Generic, TypeVar

from daytona_toolbox_api_client import ApiClient
from daytona_toolbox_api_client_async import ApiClient as AsyncApiClient

from .pool_tracker import AsyncPoolSaturationTracker

# TypeVar constrained to either sync or async ApiClient
ApiClientT = TypeVar("ApiClientT", ApiClient, AsyncApiClient)


class ToolboxApiClientProxy(Generic[ApiClientT]):
    """Proxy around the toolbox API client.

    Intercepts ``param_serialize`` to prepend the sandbox ID to the resource
    path and set the host to the toolbox proxy URL.  When an
    ``AsyncPoolSaturationTracker`` is provided, ``call_api`` is also wrapped
    so that every async request is tracked against the connection-pool limit.

    All other attributes are delegated to the underlying API client.
    """

    def __init__(
        self,
        api_client: ApiClientT,
        sandbox_id: str,
        toolbox_proxy_url: str,
        pool_tracker: AsyncPoolSaturationTracker | None = None,
    ):
        self._api_client: ApiClientT = api_client
        self._sandbox_id: str = sandbox_id
        self._toolbox_base_url: str = toolbox_proxy_url
        self._pool_tracker: AsyncPoolSaturationTracker | None = pool_tracker
        self._wrapped_call_api: Any | None = None

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
        attr = getattr(self._api_client, name)
        if name == "call_api" and self._pool_tracker is not None:
            if self._wrapped_call_api is None:
                self._wrapped_call_api = self._make_tracked_call_api(attr)
            return self._wrapped_call_api
        return attr

    def _make_tracked_call_api(self, original_call_api: Any) -> Any:
        assert self._pool_tracker is not None
        tracker = self._pool_tracker

        @functools.wraps(original_call_api)
        async def tracked_call_api(*args: Any, **kwargs: Any) -> Any:
            tracker.acquire()
            try:
                return await original_call_api(*args, **kwargs)
            finally:
                tracker.release()

        return tracked_call_api
