# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import asyncio
import json
import warnings
from importlib.metadata import version
from typing import Callable, Dict, List, Optional, Union, overload

from daytona_api_client_async import (
    ApiClient,
    Configuration,
    CreateBuildInfo,
    CreateSandbox,
    ObjectStorageApi,
    SandboxApi,
    SandboxState,
    SnapshotsApi,
)
from daytona_api_client_async import ToolboxApi as ToolboxApi
from daytona_api_client_async import VolumesApi as VolumesApi
from environs import Env

from .._utils.enum import to_enum
from .._utils.errors import DaytonaError, intercept_errors
from .._utils.stream import process_streaming_response
from .._utils.timeout import with_timeout
from ..code_toolbox.sandbox_python_code_toolbox import SandboxPythonCodeToolbox
from ..code_toolbox.sandbox_ts_code_toolbox import SandboxTsCodeToolbox
from ..common.daytona import (
    CodeLanguage,
    CreateSandboxFromImageParams,
    CreateSandboxFromSnapshotParams,
    DaytonaConfig,
    Image,
)
from .sandbox import AsyncSandbox
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
    """

    _api_key: Optional[str] = None
    _jwt_token: Optional[str] = None
    _organization_id: Optional[str] = None
    _api_url: str
    _target: Optional[str] = None

    def __init__(self, config: Optional[DaytonaConfig] = None):
        """Initializes Daytona instance with optional configuration.

        If no config is provided, reads from environment variables:
        - `DAYTONA_API_KEY`: Required API key for authentication
        - `DAYTONA_API_URL`: Required api URL
        - `DAYTONA_TARGET`: Optional target environment (defaults to 'us')

        Args:
            config (Optional[DaytonaConfig]): Object containing api_key, api_url, and target.

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
        self.default_language = CodeLanguage.PYTHON
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
            env.read_env()
            env.read_env(".env", override=True)
            env.read_env(".env.local", override=True)

            self._api_key = self._api_key or (env.str("DAYTONA_API_KEY", None) if not self._jwt_token else None)
            self._jwt_token = self._jwt_token or env.str("DAYTONA_JWT_TOKEN", None)
            self._organization_id = self._organization_id or env.str("DAYTONA_ORGANIZATION_ID", None)
            api_url = api_url or env.str("DAYTONA_API_URL", None) or env.str("DAYTONA_SERVER_URL", default_api_url)
            self._target = self._target or env.str("DAYTONA_TARGET", None)

            if env.str("DAYTONA_SERVER_URL", None) and not env.str("DAYTONA_API_URL", None):
                warnings.warn(
                    "Environment variable `DAYTONA_SERVER_URL` is deprecated and will be removed in future versions. "
                    + "Use `DAYTONA_API_URL` instead.",
                    DeprecationWarning,
                    stacklevel=2,
                )

        self._api_url = api_url

        if not self._api_key and not self._jwt_token:
            raise DaytonaError("API key or JWT token is required")

        # Create API configuration without api_key
        configuration = Configuration(host=self._api_url)
        self._api_client = ApiClient(configuration)
        self._api_client.default_headers["Authorization"] = f"Bearer {self._api_key or self._jwt_token}"
        self._api_client.default_headers["X-Daytona-Source"] = "python-sdk"

        # Get SDK version dynamically
        try:
            sdk_version = None
            for pkg_name in ["daytona_sdk", "daytona"]:
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
        self._sandbox_api = SandboxApi(self._api_client)
        self._toolbox_api = ToolboxApi(self._api_client)
        self._object_storage_api = ObjectStorageApi(self._api_client)

        # Initialize services
        self.volume = AsyncVolumeService(VolumesApi(self._api_client))
        self.snapshot = AsyncSnapshotService(SnapshotsApi(self._api_client), self._object_storage_api)

    # unasync: delete start
    async def __aenter__(self):
        """Async context manager entry."""
        return self

    async def __aexit__(self, exc_type, exc_value, traceback):
        """Async context manager exit - ensures proper cleanup."""
        await self.close()

    async def close(self):
        """Close the HTTP session and clean up resources.

        This method should be called when you're done using the AsyncDaytona instance
        to properly close the underlying HTTP session and avoid resource leaks.

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
        if hasattr(self, "_api_client") and self._api_client:
            await self._api_client.close()

    # unasync: delete end

    @overload
    async def create(
        self,
        params: Optional[CreateSandboxFromSnapshotParams] = None,
        *,
        timeout: Optional[float] = 60,
    ) -> AsyncSandbox:
        """Creates Sandboxes from specified or default snapshot. You can specify various parameters,
        including language, image, environment variables, and volumes.

        Args:
            params (Optional[CreateSandboxFromSnapshotParams]): Parameters for Sandbox creation. If not provided,
                   defaults to default Daytona snapshot and Python language.
            timeout (Optional[float]): Timeout (in seconds) for sandbox creation. 0 means no timeout.
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
        params: Optional[CreateSandboxFromImageParams] = None,
        *,
        timeout: Optional[float] = 60,
        on_snapshot_create_logs: Callable[[str], None] = None,
    ) -> AsyncSandbox:
        """Creates Sandboxes from specified image available on some registry or declarative Daytona Image.
        You can specify various parameters, including resources, language, image, environment variables,
        and volumes. Daytona creates snapshot from provided image and uses it to create Sandbox.

        Args:
            params (Optional[CreateSandboxFromImageParams]): Parameters for Sandbox creation from image.
            timeout (Optional[float]): Timeout (in seconds) for sandbox creation. 0 means no timeout.
                Default is 60 seconds.
            on_snapshot_create_logs (Callable[[str], None]): This callback function handles snapshot creation logs.

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
    async def create(
        self,
        params: Optional[Union[CreateSandboxFromSnapshotParams, CreateSandboxFromImageParams]] = None,
        *,
        timeout: Optional[float] = 60,
        on_snapshot_create_logs: Callable[[str], None] = None,
    ) -> AsyncSandbox:
        # If no params provided, create default params for Python
        if params is None:
            params = CreateSandboxFromSnapshotParams(language=self.default_language)
        elif params.language is None:
            params.language = self.default_language

        return await self._create(params, timeout=timeout, on_snapshot_create_logs=on_snapshot_create_logs)

    @with_timeout(
        error_message=lambda self, timeout: (
            f"Failed to create and start sandbox within {timeout} seconds timeout period."
        )
    )
    async def _create(
        self,
        params: Optional[Union[CreateSandboxFromSnapshotParams, CreateSandboxFromImageParams]] = None,
        *,
        timeout: Optional[float] = 60,
        on_snapshot_create_logs: Callable[[str], None] = None,
    ) -> AsyncSandbox:
        code_toolbox = self._get_code_toolbox(params.language)

        if timeout < 0:
            raise DaytonaError("Timeout must be a non-negative number")

        if params.auto_stop_interval is not None and params.auto_stop_interval < 0:
            raise DaytonaError("auto_stop_interval must be a non-negative integer")

        if params.auto_archive_interval is not None and params.auto_archive_interval < 0:
            raise DaytonaError("auto_archive_interval must be a non-negative integer")

        target = self._target

        # Create sandbox using dictionary
        sandbox_data = CreateSandbox(
            user=params.os_user,
            env=params.env_vars if params.env_vars else {},
            labels=params.labels,
            public=params.public,
            target=str(target) if target else None,
            auto_stop_interval=params.auto_stop_interval,
            auto_archive_interval=params.auto_archive_interval,
            auto_delete_interval=params.auto_delete_interval,
            volumes=params.volumes,
        )

        if getattr(params, "snapshot", None):
            sandbox_data.snapshot = params.snapshot

        if getattr(params, "image", None):
            if isinstance(params.image, str):
                sandbox_data.build_info = CreateBuildInfo(
                    dockerfile_content=Image.base(params.image).dockerfile(),
                )
            elif isinstance(params.image, Image):
                context_hashes = await AsyncSnapshotService.process_image_context(
                    self._object_storage_api, params.image
                )
                sandbox_data.build_info = CreateBuildInfo(
                    context_hashes=context_hashes,
                    dockerfile_content=params.image.dockerfile(),
                )

        if getattr(params, "resources", None):
            sandbox_data.cpu = params.resources.cpu
            sandbox_data.memory = params.resources.memory
            sandbox_data.disk = params.resources.disk
            sandbox_data.gpu = params.resources.gpu

        response = await self._sandbox_api.create_sandbox(sandbox_data, _request_timeout=timeout or None)

        if response.state == SandboxState.PENDING_BUILD and on_snapshot_create_logs:
            _, url, *_ = self._sandbox_api._get_build_logs_serialize(  # pylint: disable=protected-access
                response.id,
                follow=True,
                x_daytona_organization_id=None,
                _request_auth=None,
                _content_type=None,
                _headers=None,
                _host_index=None,
            )

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
                url=url,
                headers=self._sandbox_api.api_client.default_headers,
                on_chunk=lambda chunk: on_snapshot_create_logs(chunk.rstrip()),
                should_terminate=should_terminate,
            )
            response = response_ref["response"]

        sandbox = AsyncSandbox(
            response,
            self._sandbox_api,
            self._toolbox_api,
            code_toolbox,
        )

        if sandbox.state != SandboxState.STARTED:
            # Wait for sandbox to start
            try:
                await sandbox.wait_for_sandbox_start()
            finally:
                # If not Daytona SaaS, we don't need to handle pulling image state
                pass

        return sandbox

    def _get_code_toolbox(self, language: Optional[CodeLanguage] = None):
        """Helper method to get the appropriate code toolbox based on language.

        Args:
            language (Optional[CodeLanguage]): Language of the code toolbox. If not provided, defaults to Python.

        Returns:
            The appropriate code toolbox instance for the specified language.

        Raises:
            DaytonaError: If an unsupported language is specified.
        """
        if not language:
            return SandboxPythonCodeToolbox()

        enum_language = to_enum(CodeLanguage, language)
        if enum_language is None:
            raise DaytonaError(f"Unsupported language: {language}")
        language = enum_language

        match language:
            case CodeLanguage.JAVASCRIPT | CodeLanguage.TYPESCRIPT:
                return SandboxTsCodeToolbox()
            case CodeLanguage.PYTHON:
                return SandboxPythonCodeToolbox()
            case _:
                raise DaytonaError(f"Unsupported language: {language}")

    async def delete(self, sandbox: AsyncSandbox, timeout: Optional[float] = 60) -> None:
        """Deletes a Sandbox.

        Args:
            sandbox (Sandbox): The Sandbox instance to delete.
            timeout (Optional[float]): Timeout (in seconds) for sandbox deletion. 0 means no timeout.
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
        return await sandbox.delete(timeout)

    @intercept_errors(message_prefix="Failed to get sandbox: ")
    async def get(self, sandbox_id: str) -> AsyncSandbox:
        """Gets a Sandbox by its ID.

        Args:
            sandbox_id (str): The ID of the Sandbox to retrieve.

        Returns:
            Sandbox: The Sandbox instance.

        Raises:
            DaytonaError: If sandbox_id is not provided.

        Example:
            ```python
            sandbox = await daytona.get("my-sandbox-id")
            print(sandbox.status)
            ```
        """
        if not sandbox_id:
            raise DaytonaError("sandbox_id is required")

        # Get the sandbox instance
        sandbox_instance = await self._sandbox_api.get_sandbox(sandbox_id)

        # Create and return sandbox with Python code toolbox as default
        code_toolbox = SandboxPythonCodeToolbox()
        return AsyncSandbox(
            sandbox_instance,
            self._sandbox_api,
            self._toolbox_api,
            code_toolbox,
        )

    @intercept_errors(message_prefix="Failed to find sandbox: ")
    async def find_one(self, sandbox_id: Optional[str] = None, labels: Optional[Dict[str, str]] = None) -> AsyncSandbox:
        """Finds a Sandbox by its ID or labels.

        Args:
            sandbox_id (Optional[str]): The ID of the Sandbox to retrieve.
            labels (Optional[Dict[str, str]]): Labels to filter Sandboxes.

        Returns:
            Sandbox: First Sandbox that matches the ID or labels.

        Raises:
            DaytonaError: If no Sandbox is found.

        Example:
            ```python
            sandbox = await daytona.find_one(labels={"my-label": "my-value"})
            print(f"Sandbox ID: {sandbox.id} State: {sandbox.state}")
            ```
        """
        if sandbox_id:
            return await self.get(sandbox_id)
        sandboxes = await self.list(labels)
        if len(sandboxes) == 0:
            raise DaytonaError(f"No sandbox found with labels {labels}")
        return sandboxes[0]

    @intercept_errors(message_prefix="Failed to list sandboxes: ")
    async def list(self, labels: Optional[Dict[str, str]] = None) -> List[AsyncSandbox]:
        """Lists Sandboxes filtered by labels.

        Args:
            labels (Optional[Dict[str, str]]): Labels to filter Sandboxes.

        Returns:
            List[Sandbox]: List of Sandbox instances that match the labels.

        Example:
            ```python
            sandboxes = await daytona.list(labels={"my-label": "my-value"})
            for sandbox in sandboxes:
                print(f"{sandbox.id}: {sandbox.status}")
            ```
        """
        sandboxes = await self._sandbox_api.list_sandboxes(labels=json.dumps(labels))

        return [
            AsyncSandbox(
                sandbox,
                self._sandbox_api,
                self._toolbox_api,
                self._get_code_toolbox(self._validate_language_label(sandbox.labels.get("code-toolbox-language"))),
            )
            for sandbox in sandboxes
        ]

    def _validate_language_label(self, language: Optional[str]) -> CodeLanguage:
        """Validates and normalizes the language label.

        Args:
            language (Optional[str]): The language label to validate.

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

    async def start(self, sandbox: AsyncSandbox, timeout: Optional[float] = 60) -> None:
        """Starts a Sandbox and waits for it to be ready.

        Args:
            sandbox (Sandbox): The Sandbox to start.
            timeout (Optional[float]): Optional timeout in seconds to wait for the Sandbox to start.
                0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative; If Sandbox fails to start or times out
        """
        await sandbox.start(timeout)

    async def stop(self, sandbox: AsyncSandbox, timeout: Optional[float] = 60) -> None:
        """Stops a Sandbox and waits for it to be stopped.

        Args:
            sandbox (Sandbox): The sandbox to stop
            timeout (Optional[float]): Optional timeout (in seconds) for sandbox stop.
                0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative; If Sandbox fails to stop or times out
        """
        await sandbox.stop(timeout)
