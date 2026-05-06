# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import functools
from typing import Any, Generic, cast

import aiohttp
from typing_extensions import override

from .api_client_proxy import ApiClientT
from .pool_tracker import AsyncPoolSaturationTracker


class ToolboxApiClientProxy(Generic[ApiClientT]):
    """Proxy around the toolbox API client.

    Intercepts ``param_serialize`` to prepend the sandbox ID to the resource path
    and set the host to the toolbox proxy URL. When an ``AsyncPoolSaturationTracker``
    is provided, ``call_api`` is wrapped to track each async request against the
    connection-pool limit.

    ``__getattr__`` and ``__setattr__`` make the proxy transparent: anything not
    in the proxy's own ``__dict__`` (snapshotted at the end of ``__init__``)
    delegates to the underlying api_client.
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

        # Snapshot proxy-owned attribute names so __setattr__ can tell them apart from
        # writes meant to delegate to the underlying api_client. Beats a hand-edited
        # whitelist that drifts out of sync with __init__.
        object.__setattr__(self, "_proxy_own_attrs", frozenset(self.__dict__))

    @property
    def http_session(self) -> aiohttp.ClientSession:
        """Shared ``aiohttp.ClientSession`` reachable through the underlying ApiClient.

        Only meaningful for the async variant. ``SharedAiohttpSession`` populates
        ``rest_client.pool_manager`` on the first REST call, so any path going through
        ``daytona.create/get/list`` is safe.

        Raises:
            RuntimeError: if no REST request has fired yet — only reachable when a
                ``Sandbox`` is built outside the standard entrypoints.
        """
        # ApiClientT covers all four generated variants, so pyright widens
        # pool_manager to a urllib3 / aiohttp union; this property is async-only,
        # so the cast keeps the typed return honest.
        session = self._api_client.rest_client.pool_manager
        if session is None:
            raise RuntimeError(
                "Shared aiohttp session not initialized; this happens when toolbox"
                + " APIs are called before any main API call. Use AsyncDaytona.create/get/list"
                + " to obtain a Sandbox first."
            )
        return cast(aiohttp.ClientSession, session)

    def param_serialize(self, *args: object, **kwargs: object) -> Any:
        """Intercepts param_serialize to prepend sandbox ID to resource_path."""
        resource_path = kwargs.get("resource_path")

        if resource_path:
            resource_path = f"/{self._sandbox_id}{resource_path}"

        kwargs["resource_path"] = resource_path
        kwargs["_host"] = self._toolbox_base_url

        return self._api_client.param_serialize(*args, **kwargs)

    def __getattr__(self, name: str) -> Any:
        attr = getattr(self._api_client, name)
        if name == "call_api" and self._pool_tracker is not None:
            if self._wrapped_call_api is None:
                self._wrapped_call_api = self._make_tracked_call_api(attr)
            return self._wrapped_call_api
        return attr

    @override
    def __setattr__(self, name: str, value: Any) -> None:
        # Pre-snapshot (during __init__): all writes land on the proxy itself.
        # Post-snapshot: only names captured in __init__ stay on the proxy; everything
        # else delegates so callers can do e.g. ``api_client.user_agent = ...``.
        own = self.__dict__.get("_proxy_own_attrs")
        if own is None or name in own:
            object.__setattr__(self, name, value)
        else:
            setattr(self._api_client, name, value)

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
