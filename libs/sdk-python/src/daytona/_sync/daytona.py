# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import asyncio
import json
import time
import warnings
import weakref
from copy import deepcopy
from importlib.metadata import version
from typing import Callable, cast, overload

import httpx
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.aiohttp_client import AioHttpClientInstrumentor
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.semconv.attributes import service_attributes

from daytona_api_client import (
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
from daytona_api_client import VolumesApi as VolumesApi
from daytona_toolbox_api_client import ApiClient as ToolboxApiClient

from .._utils.enum import to_enum
from .._utils.env import DaytonaEnvReader
from .._utils.errors import intercept_errors
from .._utils.otel_decorator import with_instrumentation
from .._utils.stream import process_streaming_response
from .._utils.timeout import http_timeout, with_timeout
from ..common.daytona import (
    CODE_TOOLBOX_LANGUAGE_LABEL,
    CodeLanguage,
    CreateSandboxFromImageParams,
    CreateSandboxFromSnapshotParams,
    DaytonaConfig,
)
from ..common.errors import DaytonaAuthenticationError, DaytonaValidationError
from ..common.image import Image
from ..internal.http_client import build_sync_http_client
from ..internal.urllib3_retry import RemoteDisconnectedRetry
from .sandbox import PaginatedSandboxes, Sandbox
from .snapshot import SnapshotService
from .volume import VolumeService


class Daytona:
    """Main class for interacting with the Daytona API.

    This class provides methods to create, manage, and interact with Daytona Sandboxes.
    It can be initialized either with explicit configuration or using environment variables.

    Attributes:
        volume (VolumeService): Service for managing volumes.
        snapshot (SnapshotService): Service for managing snapshots.

    Example:
        Using environment variables:
        ```python
        daytona = Daytona()  # Uses DAYTONA_API_KEY, DAYTONA_API_URL
        sandbox = daytona.create()
        ```

        Using explicit configuration:
        ```python
        config = DaytonaConfig(
            api_key="your-api-key",
            api_url="https://your-api.com",
            target="us"
        )
        daytona = Daytona(config)
        sandbox = daytona.create()
        ```

        Using OpenTelemetry tracing:
        ```python
        config = DaytonaConfig(
            api_key="your-api-key",
            experimental={"otelEnabled": True}  # Enable OpenTelemetry tracing through experimental config
        )
        async with Daytona(config) as daytona:
            sandbox = daytona.create()
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
            daytona1 = Daytona()

            # Using explicit configuration
            config = DaytonaConfig(
                api_key="your-api-key",
                api_url="https://your-api.com",
                target="us"
            )
            daytona2 = Daytona(config)

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

        env_reader: DaytonaEnvReader | None = None

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
            env_reader = DaytonaEnvReader()
            self._api_key = self._api_key or (env_reader.get("DAYTONA_API_KEY") if not self._jwt_token else None)
            self._jwt_token = self._jwt_token or env_reader.get("DAYTONA_JWT_TOKEN")
            self._organization_id = self._organization_id or env_reader.get("DAYTONA_ORGANIZATION_ID")
            api_url = api_url or env_reader.get("DAYTONA_API_URL") or env_reader.get("DAYTONA_SERVER_URL")
            self._target = self._target or env_reader.get("DAYTONA_TARGET")

            if env_reader.get("DAYTONA_SERVER_URL") and not env_reader.get("DAYTONA_API_URL"):
                warnings.warn(
                    "Environment variable `DAYTONA_SERVER_URL` is deprecated and will be removed in future versions. "
                    + "Use `DAYTONA_API_URL` instead.",
                    DeprecationWarning,
                    stacklevel=2,
                )

        self._api_url = api_url or default_api_url

        if not self._api_key and not self._jwt_token:
            msg = (
                "Authentication credentials not found."
                + " Set DAYTONA_API_KEY, or both DAYTONA_JWT_TOKEN and DAYTONA_ORGANIZATION_ID."
                + " These can also be provided via DaytonaConfig."
            )
            raise DaytonaAuthenticationError(msg)

        # Create API configuration without api_key
        configuration = Configuration(host=self._api_url)
        # When None, keep urllib3 default (cpu_count * 5) — unlike aiohttp,
        # urllib3 treats None as maxsize=1 which would hurt performance.
        pool_size = config.connection_pool_maxsize if config else 250
        if pool_size is not None:
            configuration.connection_pool_maxsize = pool_size
        self._api_client: ApiClient = ApiClient(configuration)
        self._api_client.default_headers["Authorization"] = f"Bearer {self._api_key or self._jwt_token}"
        self._api_client.default_headers["X-Daytona-Source"] = "sdk-python"

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
        self._api_client.user_agent = f"sdk-python/{sdk_version}"

        if not self._api_key:
            if not self._organization_id:
                raise DaytonaAuthenticationError(
                    "DAYTONA_ORGANIZATION_ID is required when authenticating with DAYTONA_JWT_TOKEN."
                    + " It can also be provided via DaytonaConfig."
                )
            self._api_client.default_headers["X-Daytona-Organization-ID"] = self._organization_id

        # Shared pooled client for file-transfer paths; honors connection_pool_maxsize.
        self._http_client: httpx.Client = build_sync_http_client(pool_size)
        # Close the pool deterministically when the Daytona instance is dereferenced.
        # The finalizer captures only the httpx.Client (not self), so it doesn't keep
        # the Daytona alive — it just runs httpx.Client.close() on GC.
        _ = weakref.finalize(self, self._http_client.close)

        # Initialize API clients with the api_client instance
        self._sandbox_api: SandboxApi = SandboxApi(self._api_client)
        self._object_storage_api: ObjectStorageApi = ObjectStorageApi(self._api_client)
        self._config_api: ConfigApi = ConfigApi(self._api_client)
        self._toolbox_api_client: ToolboxApiClient = self._clone_api_client_to_toolbox_api_client()

        # Initialize services
        self.volume: VolumeService = VolumeService(VolumesApi(self._api_client))
        self.snapshot: SnapshotService = SnapshotService(
            SnapshotsApi(self._api_client), self._object_storage_api, self._target
        )

        # Initialize OpenTelemetry if enabled
        env = env_reader or DaytonaEnvReader()
        otel_enabled = (
            (config and config.otel_enabled)
            or (config and config._experimental and config._experimental.get("otelEnabled"))
            or env.get("DAYTONA_OTEL_ENABLED") == "true"
            or env.get("DAYTONA_EXPERIMENTAL_OTEL_ENABLED") == "true"
        )
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

    @overload
    def create(
        self,
        params: CreateSandboxFromSnapshotParams | None = None,
        *,
        timeout: float = 60,
    ) -> Sandbox:
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
            sandbox = daytona.create()
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
            sandbox = daytona.create(params, timeout=40)
            ```
        """

    @overload
    def create(
        self,
        params: CreateSandboxFromImageParams | None = None,
        *,
        timeout: float = 60,
        on_snapshot_create_logs: Callable[[str], None] | None = None,
    ) -> Sandbox:
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
            sandbox = daytona.create(CreateSandboxFromImageParams(image="debian:12.9"))
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
            sandbox = daytona.create(
                params,
                timeout=40,
                on_snapshot_create_logs=lambda chunk: print(chunk, end=""),
            )
            ```
        """

    @intercept_errors(message_prefix="Failed to create sandbox: ")
    @with_timeout()
    @with_instrumentation()
    def create(
        self,
        params: CreateSandboxFromSnapshotParams | CreateSandboxFromImageParams | None = None,
        *,
        timeout: float = 60,
        on_snapshot_create_logs: Callable[[str], None] | None = None,
    ) -> Sandbox:
        # If no params provided, create default params for Python
        if not params:
            params = CreateSandboxFromSnapshotParams(language=self.default_language)
        elif not params.language:
            params.language = self.default_language

        return self._create(params, timeout=timeout, on_snapshot_create_logs=on_snapshot_create_logs)

    def _create(
        self,
        params: CreateSandboxFromSnapshotParams | CreateSandboxFromImageParams,
        *,
        timeout: float = 60,
        on_snapshot_create_logs: Callable[[str], None] | None = None,
    ) -> Sandbox:
        validated_language = self._validate_language_label(str(params.language) if params.language else None)
        params.labels = params.labels or {}
        params.labels[CODE_TOOLBOX_LANGUAGE_LABEL] = validated_language.value

        if timeout and timeout < 0:
            raise DaytonaValidationError("Timeout must be a non-negative number")

        if params.auto_stop_interval is not None and params.auto_stop_interval < 0:
            raise DaytonaValidationError("auto_stop_interval must be a non-negative integer")

        if params.auto_archive_interval is not None and params.auto_archive_interval < 0:
            raise DaytonaValidationError("auto_archive_interval must be a non-negative integer")

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
                context_hashes = SnapshotService.process_image_context(self._object_storage_api, params.image)
                sandbox_data.build_info = CreateBuildInfo(
                    context_hashes=context_hashes,
                    dockerfile_content=params.image.dockerfile(),
                )

            if params.resources:
                sandbox_data.cpu = params.resources.cpu
                sandbox_data.memory = params.resources.memory
                sandbox_data.disk = params.resources.disk
                sandbox_data.gpu = params.resources.gpu

        response = self._sandbox_api.create_sandbox(sandbox_data, _request_timeout=http_timeout(timeout))

        if response.state == SandboxState.PENDING_BUILD and on_snapshot_create_logs:
            build_logs_url = (self._sandbox_api.get_build_logs_url(response.id)).url

            response_ref = {"response": response}

            def should_terminate():
                response_ref["response"] = self._sandbox_api.get_sandbox(response_ref["response"].id)
                return response_ref["response"].state in [
                    SandboxState.STARTED,
                    SandboxState.STARTING,
                    SandboxState.ERROR,
                    SandboxState.BUILD_FAILED,
                ]

            while response_ref["response"].state == SandboxState.PENDING_BUILD:
                time.sleep(1)
                response_ref["response"] = self._sandbox_api.get_sandbox(response_ref["response"].id)

            asyncio.run(
                process_streaming_response(
                    url=build_logs_url + "?follow=true",
                    headers=cast(dict[str, str], self._sandbox_api.api_client.default_headers),
                    on_chunk=lambda chunk: on_snapshot_create_logs(chunk.rstrip()),
                    should_terminate=should_terminate,
                )
            )
            response = response_ref["response"]

        sandbox = Sandbox(
            response,
            self._toolbox_api_client,
            self._sandbox_api,
            validated_language.value,
            http_client=self._http_client,
        )

        if sandbox.state != SandboxState.STARTED:
            # Wait for sandbox to start. This method already handles a timeout,
            # so we don't need to pass one to internal methods.
            sandbox.wait_for_sandbox_start(timeout=0)

        return sandbox

    @with_instrumentation()
    def delete(self, sandbox: Sandbox, timeout: float = 60) -> None:
        """Deletes a Sandbox.

        Args:
            sandbox (Sandbox): The Sandbox instance to delete.
            timeout (float): Timeout (in seconds) for sandbox deletion. 0 means no timeout.
                Default is 60 seconds.

        Raises:
            DaytonaError: If sandbox fails to delete or times out

        Example:
            ```python
            sandbox = daytona.create()
            # ... use sandbox ...
            daytona.delete(sandbox)  # Clean up when done
            ```
        """
        return sandbox.delete(timeout)

    @intercept_errors(message_prefix="Failed to get sandbox: ")
    @with_instrumentation()
    def get(self, sandbox_id_or_name: str) -> Sandbox:
        """Gets a Sandbox by its ID or name.

        Args:
            sandbox_id_or_name (str): The ID or name of the Sandbox to retrieve.

        Returns:
            Sandbox: The Sandbox instance.

        Raises:
            DaytonaError: If sandbox_id_or_name is not provided.

        Example:
            ```python
            sandbox = daytona.get("my-sandbox-id-or-name")
            print(sandbox.state)
            ```
        """
        if not sandbox_id_or_name:
            raise DaytonaValidationError("sandbox_id_or_name is required")

        # Get the sandbox instance
        sandbox_instance = self._sandbox_api.get_sandbox(sandbox_id_or_name)
        language = self._validate_language_label(sandbox_instance.labels.get(CODE_TOOLBOX_LANGUAGE_LABEL)).value
        return Sandbox(
            sandbox_instance,
            self._toolbox_api_client,
            self._sandbox_api,
            language,
            http_client=self._http_client,
        )

    @intercept_errors(message_prefix="Failed to list sandboxes: ")
    @with_instrumentation()
    def list(
        self, labels: dict[str, str] | None = None, page: int | None = None, limit: int | None = None
    ) -> PaginatedSandboxes:
        """Returns paginated list of Sandboxes filtered by labels.

        Args:
            labels (dict[str, str] | None): Labels to filter Sandboxes.
            page (int | None): Page number for pagination (starting from 1).
            limit (int | None): Maximum number of items per page.

        Returns:
            PaginatedSandboxes: Paginated list of Sandbox instances that match the labels.

        Example:
            ```python
            result = daytona.list(labels={"my-label": "my-value"}, page=2, limit=10)
            for sandbox in result.items:
                print(f"{sandbox.id}: {sandbox.state}")
            ```
        """
        if page is not None and page < 1:
            raise DaytonaValidationError("page must be a positive integer")

        if limit is not None and limit < 1:
            raise DaytonaValidationError("limit must be a positive integer")

        response = self._sandbox_api.list_sandboxes_paginated(labels=json.dumps(labels), page=page, limit=limit)

        items: list[Sandbox] = []
        for sandbox in response.items:
            language = self._validate_language_label(sandbox.labels.get(CODE_TOOLBOX_LANGUAGE_LABEL)).value
            items.append(
                Sandbox(
                    sandbox,
                    self._toolbox_api_client,
                    self._sandbox_api,
                    language,
                    http_client=self._http_client,
                )
            )

        return PaginatedSandboxes(
            items=items,
            total=response.total,
            page=response.page,
            total_pages=response.total_pages,
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
            raise DaytonaValidationError(f"Invalid {CODE_TOOLBOX_LANGUAGE_LABEL}: {language}")
        return enum_language

    @with_instrumentation()
    def start(self, sandbox: Sandbox, timeout: float = 60) -> None:
        """Starts a Sandbox and waits for it to be ready.

        Args:
            sandbox (Sandbox): The Sandbox to start.
            timeout (float): Optional timeout in seconds to wait for the Sandbox to start.
                0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative; If Sandbox fails to start or times out
        """
        sandbox.start(timeout)

    @with_instrumentation()
    def stop(self, sandbox: Sandbox, timeout: float = 60) -> None:
        """Stops a Sandbox and waits for it to be stopped.

        Args:
            sandbox (Sandbox): The sandbox to stop
            timeout (float): Optional timeout (in seconds) for sandbox stop.
                0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative; If Sandbox fails to stop or times out
        """
        sandbox.stop(timeout)

    def _clone_api_client_to_toolbox_api_client(self) -> ToolboxApiClient:
        """Creates the toolbox API client from the main API client with empty host.

        Returns:
            ToolboxApiClient: The toolbox API client.
        """
        assert isinstance(self._api_client.configuration, Configuration)
        config = deepcopy(self._api_client.configuration)
        config.host = ""
        # Retry only on RemoteDisconnected (stale pool connections).
        # The daemon may close idle connections; urllib3 would normally not retry
        # POST, causing RemoteDisconnected to propagate. Using a targeted subclass
        # (instead of urllib3.Retry with allowed_methods=None) avoids also retrying
        # IncompleteRead, where the server already started processing and sending a
        # response — retrying that would execute the operation a second time.
        config.retries = RemoteDisconnectedRetry(total=3, raise_on_status=False)
        toolbox_api_client = ToolboxApiClient(config)
        toolbox_api_client.default_headers = deepcopy(cast(dict[str, str], self._api_client.default_headers))

        return toolbox_api_client
