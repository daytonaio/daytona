# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

import asyncio
import json
import threading
import time
import warnings
from typing import Callable, Dict, List, Optional

from daytona_api_client_async import ApiClient, BuildImage, Configuration, CreateBuildInfo
from daytona_api_client_async import CreateWorkspace as CreateSandbox
from daytona_api_client_async import ImagesApi, ImageState, ObjectStorageApi
from daytona_api_client_async import ToolboxApi as ToolboxApi
from daytona_api_client_async import VolumesApi as VolumesApi
from daytona_api_client_async import WorkspaceApi as SandboxApi
from daytona_api_client_async import WorkspaceState as SandboxState
from daytona_sdk._async.object_storage import ObjectStorage
from daytona_sdk._async.sandbox import AsyncSandbox, SandboxTargetRegion
from daytona_sdk._async.volume import AsyncVolumeService
from daytona_sdk._utils.enum import to_enum
from daytona_sdk._utils.errors import DaytonaError, intercept_errors
from daytona_sdk._utils.stream import process_streaming_response
from daytona_sdk._utils.timeout import with_timeout
from daytona_sdk.code_toolbox.sandbox_python_code_toolbox import SandboxPythonCodeToolbox
from daytona_sdk.code_toolbox.sandbox_ts_code_toolbox import SandboxTsCodeToolbox
from daytona_sdk.common.daytona import CodeLanguage, CreateSandboxParams, DaytonaConfig, Image
from deprecated import deprecated
from environs import Env

AsyncWorkspace = AsyncSandbox


class AsyncDaytona:
    """Main class for interacting with the Daytona API.

    This class provides asynchronous methods to create, manage, and interact with Daytona Sandboxes.
    It can be initialized either with explicit configuration or using environment variables.

    Attributes:
        api_key (str): API key for authentication.
        api_url (str): URL of the Daytona API.
        target (str): Default target location for Sandboxes.
        volume (AsyncVolumeService): Service for managing volumes.

    Example:
        Using environment variables:
        ```python
        daytona = AsyncDaytona()  # Uses DAYTONA_API_KEY, DAYTONA_API_URL
        sandbox = await daytona.create()
        ```

        Using explicit configuration:
        ```python
        config = DaytonaConfig(
            api_key="your-api-key",
            api_url="https://your-api.com",
            target="us"
        )
        daytona = AsyncDaytona(config)
        sandbox = await daytona.create()
        ```
    """

    def __init__(self, config: Optional[DaytonaConfig] = None):
        """Initializes Daytona instance with optional configuration.

        If no config is provided, reads from environment variables:
        - `DAYTONA_API_KEY`: Required API key for authentication
        - `DAYTONA_API_URL`: Required api URL
        - `DAYTONA_TARGET`: Optional target environment (defaults to SandboxTargetRegion.US)

        Args:
            config (Optional[DaytonaConfig]): Object containing api_key, api_url, and target.

        Raises:
            DaytonaError: If API key is not provided either through config or environment variables

        Example:
            ```python
            from daytona_sdk import Daytona, DaytonaConfig
            # Using environment variables
            daytona1 = AsyncDaytona()
            # Using explicit configuration
            config = DaytonaConfig(
                api_key="your-api-key",
                api_url="https://your-api.com",
                target="us"
            )
            daytona2 = AsyncDaytona(config)
            ```
        """

        default_api_url = "https://app.daytona.io/api"
        default_target = SandboxTargetRegion.US
        self.default_language = CodeLanguage.PYTHON

        if config is None or (
            not all([config.api_key, config.api_url, config.target])
            and not all(
                [
                    config.jwt_token,
                    config.organization_id,
                    config.api_url,
                    config.target,
                ]
            )
        ):
            # Initialize env - it automatically reads from .env and .env.local
            env = Env()
            env.read_env()  # reads .env
            # reads .env.local and overrides values
            env.read_env(".env.local", override=True)

            self.api_key = env.str("DAYTONA_API_KEY", None)
            self.jwt_token = env.str("DAYTONA_JWT_TOKEN", None)
            self.organization_id = env.str("DAYTONA_ORGANIZATION_ID", None)
            self.api_url = env.str("DAYTONA_API_URL", None) or env.str("DAYTONA_SERVER_URL", default_api_url)
            self.target = env.str("DAYTONA_TARGET", default_target)

            if env.str("DAYTONA_SERVER_URL", None) and not env.str("DAYTONA_API_URL", None):
                warnings.warn(
                    "Environment variable `DAYTONA_SERVER_URL` is deprecated and will be removed in future versions. "
                    + "Use `DAYTONA_API_URL` instead.",
                    DeprecationWarning,
                    stacklevel=2,
                )

        if config:
            if not config.api_key and config.jwt_token:
                self.api_key = None
            else:
                self.api_key = config.api_key or getattr(self, "api_key", None)
            self.jwt_token = config.jwt_token or getattr(self, "jwt_token", None)
            self.organization_id = config.organization_id or getattr(self, "organization_id", None)
            self.api_url = config.api_url or self.api_url
            self.target = config.target or self.target

        if not self.api_key and not self.jwt_token:
            raise DaytonaError("API key or JWT token is required")

        # Create API configuration without api_key
        configuration = Configuration(host=self.api_url)
        self.api_client = ApiClient(configuration)
        self.api_client.default_headers["Authorization"] = f"Bearer {self.api_key or self.jwt_token}"
        self.api_client.default_headers["X-Daytona-Source"] = "python-sdk"
        if not self.api_key:
            if not self.organization_id:
                raise DaytonaError("Organization ID is required when using JWT token")
            self.api_client.default_headers["X-Daytona-Organization-ID"] = self.organization_id

        # Initialize API clients with the api_client instance
        self.sandbox_api = SandboxApi(self.api_client)
        self.toolbox_api = ToolboxApi(self.api_client)
        self.image_api = ImagesApi(self.api_client)
        self.object_storage_api = ObjectStorageApi(self.api_client)

        # Initialize volume service
        self.volume = AsyncVolumeService(VolumesApi(self.api_client))

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
        if hasattr(self, "api_client") and self.api_client:
            await self.api_client.close()

    # unasync: delete end

    @intercept_errors(message_prefix="Failed to create sandbox: ")
    async def create(
        self,
        params: Optional[CreateSandboxParams] = None,
        timeout: Optional[float] = 60,
        *,
        on_image_build_logs: Callable[[str], None] = None,
    ) -> AsyncSandbox:
        """Creates Sandboxes with default or custom configurations. You can specify various parameters,
        including language, image, resources, environment variables, and volumes for the Sandbox.

        Args:
            params (Optional[CreateSandboxParams]): Parameters for Sandbox creation. If not provided,
                   defaults to Python language.
            timeout (Optional[float]): Timeout (in seconds) for sandbox creation. 0 means no timeout.
                Default is 60 seconds.
            on_image_build_logs (Callable[[str], None]): This callback function handles image build logs.
                It's invoked only when `params.image` is an instance of `Image` and there's no existing
                image in Daytona with the same configuration.

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
            params = CreateSandboxParams(
                language="python",
                image="debian:12.9",
                env_vars={"DEBUG": "true"},
                resources=SandboxResources(cpu=2, memory=4),
                auto_stop_interval=0,
                auto_archive_interval=60
            )
            sandbox = await daytona.create(params, 40)
            ```
        """
        # If no params provided, create default params for Python
        if params is None:
            params = CreateSandboxParams(language=self.default_language)
        if params.language is None:
            params.language = self.default_language

        effective_timeout = params.timeout if params.timeout else timeout

        return await self._create(params, effective_timeout, on_image_build_logs=on_image_build_logs)

    @with_timeout(
        error_message=lambda self, timeout: (
            f"Failed to create and start sandbox within {timeout} seconds timeout period."
        )
    )
    async def _create(
        self,
        params: Optional[CreateSandboxParams] = None,
        timeout: Optional[float] = 60,
        *,
        on_image_build_logs: Callable[[str], None] = None,
    ) -> AsyncSandbox:
        """Creates a new Sandbox and waits for it to start.

        Args:
            params (Optional[CreateSandboxParams]): Parameters for Sandbox creation. If not provided,
                   defaults to Python language.
            timeout (Optional[float]): Timeout (in seconds) for sandbox creation. 0 means no timeout.
                Default is 60 seconds.
            on_image_build_logs (Callable[[str], None]): This callback function handles image build logs.
                It's invoked only when `params.image` is an instance of `Image` and there's no existing
                image in Daytona with the same configuration.

        Returns:
            Sandbox: The created Sandbox instance.

        Raises:
            DaytonaError: If timeout, auto_stop_interval or auto_archive_interval is negative;
                If sandbox fails to start or times out
        """
        code_toolbox = self._get_code_toolbox(params)

        if timeout < 0:
            raise DaytonaError("Timeout must be a non-negative number")

        if params.auto_stop_interval is not None and params.auto_stop_interval < 0:
            raise DaytonaError("auto_stop_interval must be a non-negative integer")

        if params.auto_archive_interval is not None and params.auto_archive_interval < 0:
            raise DaytonaError("auto_archive_interval must be a non-negative integer")

        target = self.target

        # Create sandbox using dictionary
        sandbox_data = CreateSandbox(
            user=params.os_user,
            env=params.env_vars if params.env_vars else {},
            labels=params.labels,
            public=params.public,
            target=str(target) if target else None,
            auto_stop_interval=params.auto_stop_interval,
            auto_archive_interval=params.auto_archive_interval,
            volumes=params.volumes,
        )

        if params.resources:
            sandbox_data.cpu = params.resources.cpu
            sandbox_data.memory = params.resources.memory
            sandbox_data.disk = params.resources.disk
            sandbox_data.gpu = params.resources.gpu

        if isinstance(params.image, str):
            sandbox_data.image = params.image
        elif isinstance(params.image, Image):
            context_hashes = self.__process_image_context(params.image)
            sandbox_data.build_info = CreateBuildInfo(
                context_hashes=context_hashes,
                dockerfile_content=params.image.dockerfile(),
            )

        response = await self.sandbox_api.create_workspace(sandbox_data, _request_timeout=timeout or None)

        if response.state == SandboxState.PENDING_BUILD and on_image_build_logs:
            _, url, *_ = self.sandbox_api._get_build_logs_serialize(  # pylint: disable=protected-access
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
                response_ref["response"] = await self.sandbox_api.get_workspace(response_ref["response"].id)
                return response_ref["response"].state in [
                    SandboxState.STARTED,
                    SandboxState.STARTING,
                    SandboxState.ERROR,
                ]

            while response_ref["response"].state == SandboxState.PENDING_BUILD:
                time.sleep(1)
                response_ref["response"] = await self.sandbox_api.get_workspace(response_ref["response"].id)

            await process_streaming_response(
                url=url,
                headers=self.sandbox_api.api_client.default_headers,
                on_chunk=on_image_build_logs,
                should_terminate=should_terminate,
            )
            response = response_ref["response"]

        sandbox_info = AsyncSandbox.to_sandbox_info(response)
        response.info = sandbox_info

        sandbox = AsyncSandbox(
            response.id,
            response,
            self.sandbox_api,
            self.toolbox_api,
            code_toolbox,
        )

        if sandbox.instance.state != SandboxState.STARTED:
            # Wait for sandbox to start
            try:
                await sandbox.wait_for_sandbox_start()
            finally:
                # If not Daytona SaaS, we don't need to handle pulling image state
                pass

        return sandbox

    def _get_code_toolbox(self, params: Optional[CreateSandboxParams] = None):
        """Helper method to get the appropriate code toolbox based on language.

        Args:
            params (Optional[CreateSandboxParams]): Sandbox parameters. If not provided, defaults to Python toolbox.

        Returns:
            The appropriate code toolbox instance for the specified language.

        Raises:
            DaytonaError: If an unsupported language is specified.
        """
        if not params:
            return SandboxPythonCodeToolbox()

        enum_language = to_enum(CodeLanguage, params.language)
        if enum_language is None:
            raise DaytonaError(f"Unsupported language: {params.language}")
        params.language = enum_language

        match params.language:
            case CodeLanguage.JAVASCRIPT | CodeLanguage.TYPESCRIPT:
                return SandboxTsCodeToolbox()
            case CodeLanguage.PYTHON:
                return SandboxPythonCodeToolbox()
            case _:
                raise DaytonaError(f"Unsupported language: {params.language}")

    @intercept_errors(message_prefix="Failed to remove sandbox: ")
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
        return await self.sandbox_api.delete_workspace(sandbox.id, force=True, _request_timeout=timeout or None)

    remove = delete

    @deprecated(
        reason=(
            "Method is deprecated. Use `get_current_sandbox` instead. This method will be removed in a future version."
        )
    )
    async def get_current_workspace(self, workspace_id: str) -> AsyncWorkspace:
        """Gets a Sandbox by its ID.

        Args:
            workspace_id (str): The ID of the Sandbox to retrieve.

        Returns:
            Workspace: The Sandbox instance.
        """
        return await self.get_current_sandbox(workspace_id)

    @intercept_errors(message_prefix="Failed to get sandbox: ")
    async def get_current_sandbox(self, sandbox_id: str) -> AsyncSandbox:
        """Gets a Sandbox by its ID.

        Args:
            sandbox_id (str): The ID of the Sandbox to retrieve.

        Returns:
            Sandbox: The Sandbox instance.

        Raises:
            DaytonaError: If sandbox_id is not provided.

        Example:
            ```python
            sandbox = await daytona.get_current_sandbox("my-sandbox-id")
            print(sandbox.status)
            ```
        """
        if not sandbox_id:
            raise DaytonaError("sandbox_id is required")

        # Get the sandbox instance
        sandbox_instance = await self.sandbox_api.get_workspace(sandbox_id)
        sandbox_info = AsyncSandbox.to_sandbox_info(sandbox_instance)
        sandbox_instance.info = sandbox_info

        # Create and return sandbox with Python code toolbox as default
        code_toolbox = SandboxPythonCodeToolbox()
        return AsyncSandbox(
            sandbox_id,
            sandbox_instance,
            self.sandbox_api,
            self.toolbox_api,
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
            print(sandbox.info())
            ```
        """
        if sandbox_id:
            return await self.get_current_sandbox(sandbox_id)
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
        sandboxes = await self.sandbox_api.list_workspaces(labels=json.dumps(labels))

        for sandbox in sandboxes:
            sandbox_info = AsyncSandbox.to_sandbox_info(sandbox)
            sandbox.info = sandbox_info

        return [
            AsyncSandbox(
                sandbox.id,
                sandbox,
                self.sandbox_api,
                self.toolbox_api,
                self._get_code_toolbox(
                    CreateSandboxParams(
                        language=self._validate_language_label(sandbox.labels.get("code-toolbox-language"))
                    )
                ),
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

    # def resize(self, sandbox: Sandbox, resources: SandboxResources) -> None:
    #     """Resizes a sandbox.

    #     Args:
    #         sandbox: The sandbox to resize
    #         resources: The new resources to set
    #     """
    #     self.sandbox_api. (sandbox_id=sandbox.id, resources=resources)

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

    @intercept_errors(message_prefix="Failed to build image: ")
    @with_timeout(
        error_message=lambda self, timeout: (f"Failed to build image within {timeout} seconds timeout period.")
    )
    async def create_image(
        self,
        name: str,
        image: Image,
        *,
        on_logs: Callable[[str], None] = None,
        timeout: Optional[float] = 0,  # pylint: disable=unused-argument
    ) -> None:
        """Creates and registers a new image from the given Image definition.
        Args:
            name (str): The name of the image to create.
            image (Image): The Image instance.
            on_logs (Callable[[str], None]): This callback function handles image build logs.
            timeout (Optional[float]): Default is no timeout. Timeout in seconds (0 means no timeout).
        Example:
            ```python
            image = Image.debianSlim('3.12').pipInstall('numpy')
            daytona.create_image('my-image', image, on_logs=print)
            ```
        """
        context_hashes = await self.__process_image_context(image)
        built_image = await self.image_api.build_image(
            BuildImage(
                name=name,
                build_info=CreateBuildInfo(
                    context_hashes=context_hashes,
                    dockerfile_content=image.dockerfile(),
                ),
            )
        )

        terminal_states = [ImageState.ACTIVE, ImageState.ERROR]

        def start_log_streaming():
            _, url, *_ = self.image_api._get_image_build_logs_serialize(  # pylint: disable=protected-access
                id=built_image.id,
                follow=True,
                x_daytona_organization_id=None,
                _request_auth=None,
                _content_type=None,
                _headers=None,
                _host_index=None,
            )

            async def should_terminate():
                latest_image = await self.image_api.get_image(built_image.id)
                return latest_image.state in terminal_states

            asyncio.run(
                process_streaming_response(
                    url=url,
                    headers=self.image_api.api_client.default_headers,
                    on_chunk=on_logs,
                    should_terminate=should_terminate,
                )
            )

        thread_started = False
        if on_logs:
            on_logs(f"Building image {built_image.name} ({built_image.state})")
            thread = threading.Thread(target=start_log_streaming)
            if built_image.state != ImageState.BUILD_PENDING:
                thread.start()
                thread_started = True

        previous_state = built_image.state
        while built_image.state not in terminal_states:
            if on_logs and previous_state != built_image.state:
                if built_image.state != ImageState.BUILD_PENDING and not thread_started:
                    thread.start()
                    thread_started = True
                on_logs(f"Building image {built_image.name} ({built_image.state})")
                previous_state = built_image.state
            time.sleep(1)
            built_image = await self.image_api.get_image(built_image.id)

        if on_logs:
            await asyncio.to_thread(thread.join)
            if built_image.state == ImageState.ACTIVE:
                on_logs(f"Built image {built_image.name} ({built_image.state})")

        if built_image.state == ImageState.ERROR:
            raise DaytonaError(f"Failed to build image {built_image.name}, error reason: {built_image.error_reason}")

    async def __process_image_context(self, image: Image) -> List[str]:
        """Processes the image context by uploading it to object storage.
        Args:
            image (Image): The Image instance.
        Returns:
            List[str]: List of context hashes stored in object storage.
        """
        if not image._context_list:  # pylint: disable=protected-access
            return []

        push_access_creds = await self.object_storage_api.get_push_access()
        object_storage = ObjectStorage(
            push_access_creds.storage_url,
            push_access_creds.access_key,
            push_access_creds.secret,
            push_access_creds.session_token,
            push_access_creds.bucket,
        )

        context_hashes = []
        for context in image._context_list:  # pylint: disable=protected-access
            context_hash = await object_storage.upload(
                context.source_path, push_access_creds.organization_id, context.archive_path
            )
            context_hashes.append(context_hash)

        return context_hashes
