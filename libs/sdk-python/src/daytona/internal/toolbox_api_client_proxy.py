# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from typing import Callable, Generic, TypeVar

from daytona_toolbox_api_client import ApiClient
from daytona_toolbox_api_client_async import ApiClient as AsyncApiClient
from typing_extensions import Awaitable

# TypeVar constrained to either sync or async ApiClient
ApiClientT = TypeVar("ApiClientT", ApiClient, AsyncApiClient)


class _ToolboxApiClientProxy(Generic[ApiClientT]):
    """Wrapper around an API client that adjusts `param_serialize` method.

    It intercepts `param_serialize` to prepend the sandbox ID to the `resource_path` and
    set `_host` to the toolbox base URL, while delegating all other attributes
    and methods to the underlying API client.
    """

    def __init__(self, api_client: ApiClientT, sandbox_id: str):
        self._api_client: ApiClientT = api_client
        self._sandbox_id = sandbox_id
        self._toolbox_base_url = None

    def param_serialize(self, *args, **kwargs):
        """Intercepts param_serialize to prepend sandbox ID to resource_path."""
        resource_path = kwargs.get("resource_path")

        if resource_path:
            resource_path = f"/{self._sandbox_id}{resource_path}"

        kwargs["resource_path"] = resource_path
        kwargs["_host"] = self._toolbox_base_url

        return self._api_client.param_serialize(*args, **kwargs)

    def __getattr__(self, name):
        """Delegate all other attributes to the wrapped client."""
        return getattr(self._api_client, name)


class AsyncToolboxApiClientProxyLazyBaseUrl(_ToolboxApiClientProxy[AsyncApiClient]):
    """Wrapper around an async API client that adjusts `call_api` method.

    It intercepts `call_api` to prepend the toolbox base URL to the `url` if it is not already set.
    While delegating all other attributes and methods to the underlying async API client.
    """

    def __init__(self, api_client: AsyncApiClient, sandbox_id: str, get_toolbox_base_url: Callable[[], Awaitable[str]]):
        super().__init__(api_client, sandbox_id)
        self._get_toolbox_base_url = get_toolbox_base_url

    async def call_api(self, *args, **kwargs):
        url = str(args[1])

        if url.startswith("/"):
            await self.load_toolbox_base_url()
            url = self._toolbox_base_url + url
            args = (args[0], url, *args[2:])

        return await self._api_client.call_api(*args, **kwargs)

    async def load_toolbox_base_url(self):
        if self._toolbox_base_url is None:
            self._toolbox_base_url = await self._get_toolbox_base_url()


class ToolboxApiClientProxyLazyBaseUrl(_ToolboxApiClientProxy[ApiClient]):
    """Wrapper around a sync API client that adjusts `call_api` method.

    It intercepts `call_api` to prepend the toolbox base URL to the `url` if it is not already set.
    While delegating all other attributes and methods to the underlying sync API client.
    """

    def __init__(self, api_client: ApiClient, sandbox_id: str, get_toolbox_base_url: Callable[[], str]):
        super().__init__(api_client, sandbox_id)
        self._get_toolbox_base_url = get_toolbox_base_url

    def call_api(self, *args, **kwargs):
        url = str(args[1])

        if url.startswith("/"):
            self.load_toolbox_base_url()
            url = self._toolbox_base_url + url
            args = (args[0], url, *args[2:])

        return self._api_client.call_api(*args, **kwargs)

    def load_toolbox_base_url(self):
        if self._toolbox_base_url is None:
            self._toolbox_base_url = self._get_toolbox_base_url()
