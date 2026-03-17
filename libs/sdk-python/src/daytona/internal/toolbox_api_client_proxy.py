# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import asyncio
import inspect
import logging
import time
from typing import Any, Generic, TypeVar, cast

from daytona_toolbox_api_client import ApiClient
from daytona_toolbox_api_client_async import ApiClient as AsyncApiClient

# TypeVar constrained to either sync or async ApiClient
ApiClientT = TypeVar("ApiClientT", ApiClient, AsyncApiClient)

_logger = logging.getLogger(__name__)

_RETRYABLE_STATUS_CODES = {502, 503, 504}
_MAX_RETRIES = 3
_BASE_DELAY_SECONDS = 0.5

_RETRYABLE_EXCEPTION_TYPE_NAMES = frozenset({
    "MaxRetryError",
    "NewConnectionError",
    "ProtocolError",
    "ClientConnectionError",
    "ServerDisconnectedError",
})


def _is_retryable_exception(e: Exception) -> bool:
    if isinstance(e, ConnectionError):
        return True
    type_names = {t.__name__ for t in type(e).__mro__}
    return bool(type_names & _RETRYABLE_EXCEPTION_TYPE_NAMES)


def _retry_delay(attempt: int) -> float:
    return _BASE_DELAY_SECONDS * (2**attempt)


class ToolboxApiClientProxy(Generic[ApiClientT]):
    """Wrapper around an API client that adjusts `param_serialize` and `call_api`.

    It intercepts `param_serialize` to prepend the sandbox ID to the `resource_path` and
    set `_host` to the toolbox proxy URL, while delegating all other attributes
    and methods to the underlying API client.

    It intercepts `call_api` to add automatic retries for transient HTTP errors
    (502, 503, 504) and connection-level failures.
    """

    def __init__(self, api_client: ApiClientT, sandbox_id: str, toolbox_proxy_url: str):
        self._api_client: ApiClientT = api_client
        self._sandbox_id: str = sandbox_id
        self._toolbox_base_url: str = toolbox_proxy_url
        self._is_async: bool = inspect.iscoroutinefunction(getattr(api_client, "call_api"))

    def param_serialize(self, *args: object, **kwargs: object) -> Any:
        """Intercepts param_serialize to prepend sandbox ID to resource_path."""
        resource_path = kwargs.get("resource_path")

        if resource_path:
            resource_path = f"/{self._sandbox_id}{resource_path}"

        kwargs["resource_path"] = resource_path
        kwargs["_host"] = self._toolbox_base_url

        return self._api_client.param_serialize(*args, **kwargs)

    def call_api(self, *args: Any, **kwargs: Any) -> Any:
        """Intercepts call_api to add retry logic for transient errors."""
        if self._is_async:
            return self._call_api_with_retry_async(*args, **kwargs)
        return self._call_api_with_retry_sync(*args, **kwargs)

    def _call_api_with_retry_sync(self, *args: Any, **kwargs: Any) -> Any:
        sync_client = cast(ApiClient, self._api_client)
        last_exception: Exception | None = None
        for attempt in range(_MAX_RETRIES + 1):
            try:
                response = sync_client.call_api(*args, **kwargs)
                status: int = response.status
                if status in _RETRYABLE_STATUS_CODES and attempt < _MAX_RETRIES:
                    delay = _retry_delay(attempt)
                    _logger.warning(
                        "Toolbox API returned %d for sandbox %s, retrying in %.1fs (attempt %d/%d)",
                        status,
                        self._sandbox_id,
                        delay,
                        attempt + 1,
                        _MAX_RETRIES,
                    )
                    response.read()
                    time.sleep(delay)
                    continue
                return response
            except Exception as e:
                if attempt < _MAX_RETRIES and _is_retryable_exception(e):
                    delay = _retry_delay(attempt)
                    _logger.warning(
                        "Toolbox API connection error for sandbox %s: %s, retrying in %.1fs (attempt %d/%d)",
                        self._sandbox_id,
                        e,
                        delay,
                        attempt + 1,
                        _MAX_RETRIES,
                    )
                    time.sleep(delay)
                    last_exception = e
                    continue
                raise
        assert last_exception is not None
        raise last_exception

    async def _call_api_with_retry_async(self, *args: Any, **kwargs: Any) -> Any:
        async_client = cast(AsyncApiClient, self._api_client)
        last_exception: Exception | None = None
        for attempt in range(_MAX_RETRIES + 1):
            try:
                response = await async_client.call_api(*args, **kwargs)
                status: int = response.status
                if status in _RETRYABLE_STATUS_CODES and attempt < _MAX_RETRIES:
                    delay = _retry_delay(attempt)
                    _logger.warning(
                        "Toolbox API returned %d for sandbox %s, retrying in %.1fs (attempt %d/%d)",
                        status,
                        self._sandbox_id,
                        delay,
                        attempt + 1,
                        _MAX_RETRIES,
                    )
                    await response.read()
                    await asyncio.sleep(delay)
                    continue
                return response
            except Exception as e:
                if attempt < _MAX_RETRIES and _is_retryable_exception(e):
                    delay = _retry_delay(attempt)
                    _logger.warning(
                        "Toolbox API connection error for sandbox %s: %s, retrying in %.1fs (attempt %d/%d)",
                        self._sandbox_id,
                        e,
                        delay,
                        attempt + 1,
                        _MAX_RETRIES,
                    )
                    await asyncio.sleep(delay)
                    last_exception = e
                    continue
                raise
        assert last_exception is not None
        raise last_exception

    def __getattr__(self, name: str) -> Any:
        """Delegate all other attributes to the wrapped client."""
        return getattr(self._api_client, name)
