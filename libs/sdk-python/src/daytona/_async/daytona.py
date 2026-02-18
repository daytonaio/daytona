# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0
from __future__ import annotations

import asyncio
import json
import os
import warnings
from copy import deepcopy
from importlib.metadata import version
from types import TracebackType
from typing import Callable, cast, overload

from daytona_api_client_async import (
    ApiClient,
    ConfigApi,
    Configuration,
    CreateBuildInfo,
    CreateSandbox,
    ObjectStorageApi,
    SandboxApi,
    SandboxState,
    SandboxVolume,
    SnapshotsApi,
)
from daytona_api_client_async import VolumesApi as VolumesApi
from daytona_toolbox_api_client_async import ApiClient as ToolboxApiClient
from deprecated import deprecated
from environs import Env
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.aiohttp_client import AioHttpClientInstrumentor
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.semconv.attributes import service_attributes

from .._utils.enum import to_enum
from .._utils.errors import intercept_errors
from .._utils.otel_decorator import with_instrumentation
from .._utils.stream import process_streaming_response
from .._utils.timeout import http_timeout, with_timeout
from ..code_toolbox.sandbox_js_code_toolbox import SandboxJsCodeToolbox
from ..code_toolbox.sandbox_python_code_toolbox import SandboxPythonCodeToolbox
from ..code_toolbox.sandbox_ts_code_toolbox import SandboxTsCodeToolbox
from ..common.daytona import (
    CodeLanguage,
    CodeLanguageLiteral,
    CreateSandboxFromImageParams,
    CreateSandboxFromSnapshotParams,
    DaytonaConfig,
    ListSandboxesParams,
)
from ..common.errors import DaytonaError
from ..common.image import Image
from ..common.protocols import SandboxCodeToolbox
from .sandbox import AsyncCursorPaginatedSandboxes, AsyncPaginatedSandboxes, AsyncSandbox
from .snapshot import AsyncSnapshotService
from .volume import AsyncVolumeService


class AsyncDaytona:
    """Main class for interacting with the Daytona API.

    This class provides asynchronous methods to create, manage, and interact with Daytona Sandboxes.
    It can be initialized either with explicit configuration or using environment variables.

    Attributes:
        volume (AsyncVolumeService): Service for managing volumes.
        snapshot (AsyncSnapshotService): Service for managing snapshots.

    Example:
        Using environment variables:
        ```python
        async with AsyncDaytona() as daytona:  # Uses DAYTONA_API_KEY, DAYTONA_API_URL
            sandbox = await daytona.create()
        ```

        Using explicit configuration:
        ```python
        config = DaytonaConfig(
            api_key="your-api-key",
            api_url="https://your-api.com",
            target="us"
        )
        try:
            daytona = AsyncDaytona(config)
            sandbox = await daytona.create()
        finally:
            await daytona.close()
        ```

        Using OpenTelemetry tracing:
        ```python
        config = DaytonaConfig(
            api_key="your-api-key",
            experimental={"otelEnabled": True}
        )
        async with AsyncDaytona(config) as daytona:
            sandbox = await daytona.create()
            # All SDK operations will be traced
        # OpenTelemetry traces are flushed on close
        ```
    """

    _api_key: str | None = None
    _jwt_token: str | None = None
    _organization_id: str | None = None
    _api_url: str
    _target: str | None = None
    _tracer_provider: TracerProvider | None = None

    def __init__(self, config: DaytonaConfig | None = None):
        """Initializes Daytona instance with optional configuration.

        If no config is provided, reads from environment variables:
        - `DAYTONA_API_KEY`: Required API key for authentication
        - `DAYTONA_API_URL`: Required api URL
        - `DAYTONA_TARGET`: Optional target environment (if not provided, default region for the organization is used)

        Args:
            config (DaytonaConfig | None): Object containing api_key, api_url, and target.

        Raises:
            DaytonaError: If API key is not provided either through config or environment variables

        Example:
            ```python
            from daytona import Daytona, DaytonaConfig
            # Using environment variables
            daytona1 = AsyncDaytona()
            await daytona1.close()
            # Using explicit configuration
            config = DaytonaConfig(
                api_key="your-api-key",
                api_url="https://your-api.com",
                target="us"
            )
            daytona2 = AsyncDaytona(config)
            await daytona2.close()
            ```
        """

        default_api_url = "https://app.daytona.io/api"
        self.default_language: CodeLanguage = CodeLanguage.PYTHON
        api_url = None

        if config:
            self._api_key = None if (not config.api_key and config.jwt_token) else config.api_key
            self._jwt_token = config.jwt_token
            self._organization_id = config.organization_id
            api_url = config.api_url or config.server_url
            self._target = config.target

        if config is None or (
            not all([self._api_key, api_url, self._target])
            and not all(
                [
                    self._jwt_token,
                    self._organization_id,
                    api_url,
                    self._target,
                ]
            )
        ):
            # Initialize env - it automatically reads from .env and .env.local
            env = Env()
            _ = env.read_env()
            _ = env.read_env(".env", override=True)
            _ = env.read_env(".env.local", override=True)

            self._api_key = self._api_key or (env.str("DAYTONA_API_KEY", None) if not self._jwt_token else None)
            self._jwt_token = self._jwt_token or env.str("DAYTONA_JWT_TOKEN", None)
            self._organization_id = self._organization_id or env.str("DAYTONA_ORGANIZATION_ID", None)
            api_url = api_url or env.str("DAYTONA_API_URL", None) or env.str("DAYTONA_SERVER_URL", None)
            self._target = self._target or env.str("DAYTONA_TARGET", None)

            if env.str("DAYTONA_SERVER_URL", None) and not env.str("DAYTONA_API_URL", None):
                warnings.warn(
                    "Environment variable `DAYTONA_SERVER_URL` is deprecated and will be removed in future versions. "
                    + "Use `DAYTONA_API_URL` instead.",
                    DeprecationWarning,
                    stacklevel=2,
                )

        self._api_url = api_url or default_api_url

        if not self._api_key and not self._jwt_token:
            raise DaytonaError("API key or JWT token is required")

        # Create API configuration without api_key
        configuration = Configuration(host=self._api_url)
        self._api_client: ApiClient = ApiClient(configuration)
        self._api_client.default_headers["Authorization"] = f"Bearer {self._api_key or self._jwt_token}"
        self._api_client.default_headers["X-Daytona-Source"] = "python-sdk"

        # Get SDK version dynamically
        try:
            sdk_version = None
            for pkg_name in ["daytona", "daytona_sdk"]:
                try:
                    sdk_version = version(pkg_name)
                    break
                except Exception:
                    continue

            if sdk_version is None:
                raise Exception("Neither package found")

        except Exception:
            # Fallback version if neither package metadata is available
            sdk_version = "unknown"
        self._api_client.default_headers["X-Daytona-SDK-Version"] = sdk_version

        if not self._api_key:
            if not self._organization_id:
                raise DaytonaError("Organization ID is required when using JWT token")
            self._api_client.default_headers["X-Daytona-Organization-ID"] = self._organization_id

        # Initialize API clients with the api_client instance
        self._sandbox_api: SandboxApi = SandboxApi(self._api_client)
        self._object_storage_api: ObjectStorageApi = ObjectStorageApi(self._api_client)
        self._config_api: ConfigApi = ConfigApi(self._api_client)
        # Toolbox proxy cache per region
        self._proxy_toolbox_url_tasks: dict[str, asyncio.Task[str]] = {}
        self._proxy_toolbox_url_lock: asyncio.Lock = asyncio.Lock()
        self._toolbox_api_client: ToolboxApiClient = self._clone_api_client_to_toolbox_api_client()

        # Initialize services
        self.volume: AsyncVolumeService = AsyncVolumeService(VolumesApi(self._api_client))
        self.snapshot: AsyncSnapshotService = AsyncSnapshotService(
            SnapshotsApi(self._api_client),
            self._object_storage_api,
            self._target,
        )

        # Initialize OpenTelemetry if enabled
        otel_enabled = (config and config._experimental and config._experimental.get("otelEnabled")) or os.environ.get(
            "DAYTONA_EXPERIMENTAL_OTEL_ENABLED"
        ) == "true"
        if otel_enabled:
            self._init_otel(sdk_version)

    def _init_otel(self, sdk_version: str):
        """Initialize OpenTelemetry tracing.

        Args:
            sdk_version: The SDK version to include in resource attributes
        """
        # Create resource with SDK version
        resource = Resource.create(
            {
                service_attributes.SERVICE_VERSION: sdk_version,
                service_attributes.SERVICE_NAME: "daytona-python-sdk",
            }
        )

        # Create and configure tracer provider
        self._tracer_provider = TracerProvider(resource=resource)

        otlp_exporter = OTLPSpanExporter()
        self._tracer_provider.add_span_processor(BatchSpanProcessor(otlp_exporter))

        AioHttpClientInstrumentor().instrument()

        # Set the global tracer provider
        trace.set_tracer_provider(self._tracer_provider)

    async def __aenter__(self):
        """Async context manager entry."""
        return self

    async def __aexit__(
        self,
        exc_type: type[BaseException] | None = None,
        exc_value: BaseException | None = None,
        traceback: TracebackType | None = None,
    ):
        """Async context manager exit - ensures proper cleanup."""
        await self.close()

    async def close(self):
        """Close the HTTP session and clean up resources.

        This method should be called when you're done using the AsyncDaytona instance
        to properly close the underlying HTTP sessions and avoid resource leaks.

        Example:
            ```python
            daytona = AsyncDaytona()
            try:
                sandbox = await daytona.create()
                # ... use sandbox ...
            finally:
                await daytona.close()
            ```

            Or better yet, use as async context manager:
            ```python
            async with AsyncDaytona() as daytona:
                sandbox = await daytona.create()
                # ... use sandbox ...
            # Automatically closed
            ```
        """

        # Shutdown OpenTelemetry if it was initialized
        if self._tracer_provider is not None:
            self._tracer_provider.shutdown()

        # Close the main API client
        if hasattr(self, "_api_client") and self._api_client:
            await self._api_client.close()

        # Close the toolbox API client
        if hasattr(self, "_toolbox_api_client") and self._toolbox_api_client:
            await self._toolbox_api_client.close()

    @overload
    async def create(
        self,
        params: CreateSandboxFromSnapshotParams | None = None,
        *,
        timeout: float = 60,
    ) -> AsyncSandbox:
        """Creates Sandboxes from specified or default snapshot. You can specify various parameters,
        including language, image, environment variables, and volumes.

        Args:
            params (CreateSandboxFromSnapshotParams | None): Parameters for Sandbox creation. If not provided,
                   defaults to default Daytona snapshot and Python language.
            timeout (float): Timeout (in seconds) for sandbox creation. 0 means no timeout.
                Default is 60 seconds.

        Returns:
            Sandbox: The created Sandbox instance.

        Raises:
            DaytonaError: If timeout, auto_stop_interval or auto_archive_interval is negative;
                If sandbox fails to start or times out

        Example:
            Create a default Python Sandbox:
            ```python
            sandbox = await daytona.create()
            ```

            Create a custom Sandbox:
            ```python
            params = CreateSandboxFromSnapshotParams(
                language="python",
                snapshot="my-snapshot-id",
                env_vars={"DEBUG": "true"},
                auto_stop_interval=0,
                auto_archive_interval=60,
                auto_delete_interval=120
            )
            sandbox = await daytona.create(params, timeout=40)
            ```
        """

    @overload
    async def create(
        self,
        params: CreateSandboxFromImageParams | None = None,
        *,
        timeout: float = 60,
        on_snapshot_create_logs: Callable[[str], None] | None = None,
    ) -> AsyncSandbox:
        """Creates Sandboxes from specified image available on some registry or declarative Daytona Image.
        You can specify various parameters, including resources, language, image, environment variables,
        and volumes. Daytona creates snapshot from provided image and uses it to create Sandbox.

        Args:
            params (CreateSandboxFromImageParams | None): Parameters for Sandbox creation from image.
            timeout (float): Timeout (in seconds) for sandbox creation. 0 means no timeout.
                Default is 60 seconds.
            on_snapshot_create_logs (Callable[[str], None] | None): This callback function
                handles snapshot creation logs.

        Returns:
            Sandbox: The created Sandbox instance.

        Raises:
            DaytonaError: If timeout, auto_stop_interval or auto_archive_interval is negative;
                If sandbox fails to start or times out

        Example:
            Create a default Python Sandbox from image:
            ```python
            sandbox = await daytona.create(CreateSandboxFromImageParams(image="debian:12.9"))
            ```

            Create a custom Sandbox from declarative Image definition:
            ```python
            declarative_image = (
                Image.base("alpine:3.18")
                .pipInstall(["numpy", "pandas"])
                .env({"MY_ENV_VAR": "My Environment Variable"})
            )
            params = CreateSandboxFromImageParams(
                language="python",
                image=declarative_image,
                env_vars={"DEBUG": "true"},
                resources=Resources(cpu=2, memory=4),
                auto_stop_interval=0,
                auto_archive_interval=60,
                auto_delete_interval=120
            )
            sandbox = await daytona.create(
                params,
                timeout=40,
                on_snapshot_create_logs=lambda chunk: print(chunk, end=""),
            )
            ```
        """

    @intercept_errors(message_prefix="Failed to create sandbox: ")
    @with_timeout()
    @with_instrumentation()
    async def create(
        self,
        params: CreateSandboxFromSnapshotParams | CreateSandboxFromImageParams | None = None,
        *,
        timeout: float = 60,
        on_snapshot_create_logs: Callable[[str], None] | None = None,
    ) -> AsyncSandbox:
        # If no params provided, create default params for Python
        if not params:
            params = CreateSandboxFromSnapshotParams(language=self.default_language)
        elif not params.language:
            params.language = self.default_language

        return await self._create(params, timeout=timeout, on_snapshot_create_logs=on_snapshot_create_logs)

    async def _create(
        self,
        params: CreateSandboxFromSnapshotParams | CreateSandboxFromImageParams,
        *,
        timeout: float = 60,
        on_snapshot_create_logs: Callable[[str], None] | None = None,
    ) -> AsyncSandbox:
        code_toolbox = self._get_code_toolbox(params.language)

        if timeout and timeout < 0:
            raise DaytonaError("Timeout must be a non-negative number")

        if params.auto_stop_interval is not None and params.auto_stop_interval < 0:
            raise DaytonaError("auto_stop_interval must be a non-negative integer")

        if params.auto_archive_interval is not None and params.auto_archive_interval < 0:
            raise DaytonaError("auto_archive_interval must be a non-negative integer")

        target = self._target

        volumes = []
        if params.volumes:
            volumes = [
                SandboxVolume(volume_id=volume.volume_id, mount_path=volume.mount_path, subpath=volume.subpath)
                for volume in params.volumes
            ]

        # Create sandbox using dictionary
        sandbox_data = CreateSandbox(
            name=params.name,
            user=params.os_user,
            env=params.env_vars if params.env_vars else {},
            labels=params.labels,
            public=params.public,
            target=str(target) if target else None,
            auto_stop_interval=params.auto_stop_interval,
            auto_archive_interval=params.auto_archive_interval,
            auto_delete_interval=params.auto_delete_interval,
            volumes=volumes,
            network_block_all=params.network_block_all,
            network_allow_list=params.network_allow_list,
        )

        if isinstance(params, CreateSandboxFromSnapshotParams) and params.snapshot:
            sandbox_data.snapshot = params.snapshot

        if isinstance(params, CreateSandboxFromImageParams) and params.image:
            if isinstance(params.image, str):
                sandbox_data.build_info = CreateBuildInfo(
                    dockerfile_content=Image.base(params.image).dockerfile(),
                )
            else:
                context_hashes = await AsyncSnapshotService.process_image_context(
                    self._object_storage_api, params.image
                )
                sandbox_data.build_info = CreateBuildInfo(
                    context_hashes=context_hashes,
                    dockerfile_content=params.image.dockerfile(),
                )

            if params.resources:
                sandbox_data.cpu = params.resources.cpu
                sandbox_data.memory = params.resources.memory
                sandbox_data.disk = params.resources.disk
                sandbox_data.gpu = params.resources.gpu

        response = await self._sandbox_api.create_sandbox(sandbox_data, _request_timeout=http_timeout(timeout))

        if response.state == SandboxState.PENDING_BUILD and on_snapshot_create_logs:
            build_logs_url = (await self._sandbox_api.get_build_logs_url(response.id)).url

            response_ref = {"response": response}

            async def should_terminate():
                response_ref["response"] = await self._sandbox_api.get_sandbox(response_ref["response"].id)
                return response_ref["response"].state in [
                    SandboxState.STARTED,
                    SandboxState.STARTING,
                    SandboxState.ERROR,
                    SandboxState.BUILD_FAILED,
                ]

            while response_ref["response"].state == SandboxState.PENDING_BUILD:
                await asyncio.sleep(1)
                response_ref["response"] = await self._sandbox_api.get_sandbox(response_ref["response"].id)

            await process_streaming_response(
                url=build_logs_url + "?follow=true",
                headers=cast(dict[str, str], self._sandbox_api.api_client.default_headers),
                on_chunk=lambda chunk: on_snapshot_create_logs(chunk.rstrip()),
                should_terminate=should_terminate,
            )
            response = response_ref["response"]

        sandbox = AsyncSandbox(
            response,
            self._toolbox_api_client,
            self._sandbox_api,
            code_toolbox,
            self._get_proxy_toolbox_url,
        )

        if sandbox.state != SandboxState.STARTED:
            # Wait for sandbox to start. This method already handles a timeout,
            # so we don't need to pass one to internal methods.
            await sandbox.wait_for_sandbox_start(timeout=0)

        return sandbox

    def _get_code_toolbox(self, language: CodeLanguage | CodeLanguageLiteral | None = None) -> SandboxCodeToolbox:
        """Helper method to get the appropriate code toolbox based on language.

        Args:
            language (CodeLanguage | None): Language of the code toolbox. If not provided, defaults to Python.

        Returns:
            SandboxCodeToolbox: The appropriate code toolbox instance for the specified language.

        Raises:
            DaytonaError: If an unsupported language is specified.
        """
        if not language:
            return SandboxPythonCodeToolbox()

        enum_language = to_enum(CodeLanguage, language)
        if enum_language is None:
            raise DaytonaError(f"Unsupported language: {language}")
        language = enum_language

        toolboxes = {
            CodeLanguage.JAVASCRIPT.value: SandboxJsCodeToolbox,
            CodeLanguage.TYPESCRIPT.value: SandboxTsCodeToolbox,
            CodeLanguage.PYTHON.value: SandboxPythonCodeToolbox,
        }

        try:
            return toolboxes[language.value]()
        except KeyError as e:
            raise DaytonaError(f"Unsupported language: {language}") from e

    @with_instrumentation()
    async def delete(self, sandbox: AsyncSandbox, timeout: float = 60) -> None:
        """Deletes a Sandbox.

        Args:
            sandbox (Sandbox): The Sandbox instance to delete.
            timeout (float): Timeout (in seconds) for sandbox deletion. 0 means no timeout.
                Default is 60 seconds.

        Raises:
            DaytonaError: If sandbox fails to delete or times out

        Example:
            ```python
            sandbox = await daytona.create()
            # ... use sandbox ...
            await daytona.delete(sandbox)  # Clean up when done
            ```
        """
        _ = await sandbox.delete(timeout)

    @intercept_errors(message_prefix="Failed to get sandbox: ")
    @with_instrumentation()
    async def get(self, sandbox_id_or_name: str) -> AsyncSandbox:
        """Gets a Sandbox by its ID or name.

        Args:
            sandbox_id_or_name (str): The ID or name of the Sandbox to retrieve.

        Returns:
            Sandbox: The Sandbox instance.

        Raises:
            DaytonaError: If sandbox_id_or_name is not provided.

        Example:
            ```python
            sandbox = await daytona.get("my-sandbox-id-or-name")
            print(sandbox.state)
            ```
        """
        if not sandbox_id_or_name:
            raise DaytonaError("sandbox_id_or_name is required")

        # Get the sandbox instance
        sandbox_instance = await self._sandbox_api.get_sandbox(sandbox_id_or_name)

        # Create and return sandbox with Python code toolbox as default
        code_toolbox = SandboxPythonCodeToolbox()
        return AsyncSandbox(
            sandbox_instance,
            self._toolbox_api_client,
            self._sandbox_api,
            code_toolbox,
            self._get_proxy_toolbox_url,
        )

    @deprecated(
        reason=(
            "Use `get` for name lookup. "
            "This method relies on a deprecated API endpoint that will be removed on April 1, 2026."
        )
    )
    @intercept_errors(message_prefix="Failed to find sandbox: ")
    @with_instrumentation()
    async def find_one(
        self, sandbox_id_or_name: str | None = None, labels: dict[str, str] | None = None
    ) -> AsyncSandbox:
        """Finds a Sandbox by its ID or name or labels.

        .. deprecated::
            Use :meth:`get` for name lookup.
            This method relies on a deprecated API endpoint that will be removed on April 1, 2026.
            After that date, this method will be removed.

        Args:
            sandbox_id_or_name (str | None): The ID or name of the Sandbox to retrieve.
            labels (dict[str, str] | None): Labels to filter Sandboxes.

        Returns:
            Sandbox: First Sandbox that matches the ID or name or labels.

        Raises:
            DaytonaError: If no Sandbox is found.

        Example:
            ```python
            sandbox = await daytona.find_one(labels={"my-label": "my-value"})
            print(f"Sandbox ID: {sandbox.id} State: {sandbox.state}")
            ```
        """
        if sandbox_id_or_name:
            return await self.get(sandbox_id_or_name)

        sandboxes = await self.list(labels, page=1, limit=1)
        if len(sandboxes.items) == 0:
            raise DaytonaError(f"No sandbox found with labels {labels}")
        return sandboxes.items[0]

    @overload
    async def list(
        self,
        params: ListSandboxesParams,
    ) -> AsyncCursorPaginatedSandboxes:
        """Returns a paginated list of Sandboxes with optional state filtering.

        Uses cursor-based pagination, ordered newest first.

        Args:
            params.cursor (str | None): Pagination cursor from a previous response. Omit to start from the beginning.
            params.limit (int | None): Maximum number of items per page.
            params.states (list[str] | None): List of states to filter by.

        Returns:
            AsyncCursorPaginatedSandboxes: Cursor-paginated list of Sandbox instances.

        Example:
            ```python
            # First page
            page1 = await daytona.list(ListSandboxesParams(limit=10))
            for sandbox in page1.items:
                print(f"{sandbox.id}: {sandbox.state}")

            # Next page
            if page1.next_cursor:
                page2 = await daytona.list(ListSandboxesParams(cursor=page1.next_cursor, limit=10))

            # Filter by state
            running = await daytona.list(ListSandboxesParams(limit=10, states=[SandboxState.STARTED]))
            ```
        """

    @overload
    async def list(
        self,
        labels: dict[str, str] | None = None,
        page: int | None = None,
        limit: int | None = None,
    ) -> AsyncPaginatedSandboxes:
        """Returns paginated list of Sandboxes filtered by labels.

        .. deprecated::
            Use the cursor-based overload instead. This overload uses offset-based pagination
            against a deprecated API endpoint that will be removed on April 1, 2026.

        Args:
            labels (dict[str, str] | None): Labels to filter Sandboxes.
            page (int | None): Page number for pagination (starting from 1).
            limit (int | None): Maximum number of items per page.

        Returns:
            AsyncPaginatedSandboxes: Paginated list of Sandbox instances that match the labels.

        Example:
            ```python
            result = await daytona.list(labels={"my-label": "my-value"}, page=2, limit=10)
            for sandbox in result.items:
                print(f"{sandbox.id}: {sandbox.state}")
            ```
        """

    @intercept_errors(message_prefix="Failed to list sandboxes: ")
    @with_instrumentation()
    async def list(  # pyright: ignore[reportInconsistentOverload]
        self,
        params_or_labels: ListSandboxesParams | dict[str, str] | None = None,
        page: int | None = None,
        limit: int | None = None,
    ) -> AsyncPaginatedSandboxes | AsyncCursorPaginatedSandboxes:
        if not isinstance(params_or_labels, ListSandboxesParams):
            labels = params_or_labels

            if page is not None and page < 1:
                raise DaytonaError("page must be a positive integer")

            if limit is not None and limit < 1:
                raise DaytonaError("limit must be a positive integer")

            response = await self._sandbox_api.list_sandboxes_paginated_deprecated(
                labels=json.dumps(labels), page=page, limit=limit
            )

            return AsyncPaginatedSandboxes(
                items=[
                    AsyncSandbox(
                        sandbox,
                        self._toolbox_api_client,
                        self._sandbox_api,
                        self._get_code_toolbox(
                            self._validate_language_label(sandbox.labels.get("code-toolbox-language"))
                        ),
                        self._get_proxy_toolbox_url,
                    )
                    for sandbox in response.items
                ],
                total=response.total,
                page=response.page,
                total_pages=response.total_pages,
            )

        params = params_or_labels
        response = await self._sandbox_api.list_sandboxes(
            cursor=params.cursor,
            limit=params.limit,
            states=list(params.states) if params.states else None,
        )

        return AsyncCursorPaginatedSandboxes(
            items=[
                AsyncSandbox(
                    sandbox,
                    self._toolbox_api_client,
                    self._sandbox_api,
                    self._get_code_toolbox(self._validate_language_label(sandbox.labels.get("code-toolbox-language"))),
                    self._get_proxy_toolbox_url,
                )
                for sandbox in response.items
            ],
            next_cursor=response.next_cursor,
        )

    def _validate_language_label(self, language: str | None = None) -> CodeLanguage:
        """Validates and normalizes the language label.

        Args:
            language (str | None): The language label to validate.

        Returns:
            CodeLanguage: The validated language, defaults to "python" if None

        Raises:
            DaytonaError: If the language is not supported.
        """
        if not language:
            return CodeLanguage.PYTHON

        enum_language = to_enum(CodeLanguage, language)
        if enum_language is None:
            raise DaytonaError(f"Invalid code-toolbox-language: {language}")
        return enum_language

    @with_instrumentation()
    async def start(self, sandbox: AsyncSandbox, timeout: float = 60) -> None:
        """Starts a Sandbox and waits for it to be ready.

        Args:
            sandbox (Sandbox): The Sandbox to start.
            timeout (float): Optional timeout in seconds to wait for the Sandbox to start.
                0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative; If Sandbox fails to start or times out
        """
        await sandbox.start(timeout)

    @with_instrumentation()
    async def stop(self, sandbox: AsyncSandbox, timeout: float = 60) -> None:
        """Stops a Sandbox and waits for it to be stopped.

        Args:
            sandbox (Sandbox): The sandbox to stop
            timeout (float): Optional timeout (in seconds) for sandbox stop.
                0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative; If Sandbox fails to stop or times out
        """
        await sandbox.stop(timeout)

    def _clone_api_client_to_toolbox_api_client(self) -> ToolboxApiClient:
        """Creates the toolbox API client from the main API client with empty host.

        Returns:
            ToolboxApiClient: The toolbox API client.
        """
        assert isinstance(self._api_client.configuration, Configuration)
        config = deepcopy(self._api_client.configuration)
        config.host = ""
        toolbox_api_client = ToolboxApiClient(config)
        toolbox_api_client.default_headers = deepcopy(cast(dict[str, str], self._api_client.default_headers))

        return toolbox_api_client

    async def _get_proxy_toolbox_url(self, sandbox_id: str, region_id: str) -> str:
        if self._proxy_toolbox_url_tasks.get(region_id) is not None:
            return await self._proxy_toolbox_url_tasks[region_id]

        async with self._proxy_toolbox_url_lock:
            # Double-check: another coroutine might have created the task
            if self._proxy_toolbox_url_tasks.get(region_id) is None:

                async def _fetch():
                    response = await self._sandbox_api.get_toolbox_proxy_url(sandbox_id)
                    return response.url

                self._proxy_toolbox_url_tasks[region_id] = asyncio.create_task(_fetch())

        # All coroutines that made it here can now await the same task in parallel
        return await self._proxy_toolbox_url_tasks[region_id]
