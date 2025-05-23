# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

import asyncio
import json
import threading
import time
import warnings
from dataclasses import dataclass
from enum import Enum
from typing import Annotated, Callable, Dict, List, Optional, Union

from daytona_api_client import ApiClient, BuildImage, Configuration, CreateBuildInfo
from daytona_api_client import CreateWorkspace as CreateSandbox
from daytona_api_client import (
    ImagesApi,
    ImageState,
    ObjectStorageApi,
    SessionExecuteRequest,
    SessionExecuteResponse,
    ToolboxApi,
    VolumesApi,
)
from daytona_api_client import WorkspaceApi as SandboxApi
from daytona_api_client import WorkspaceState as SandboxState
from daytona_api_client import WorkspaceVolume as VolumeMount
from daytona_sdk._utils.errors import DaytonaError, intercept_errors
from deprecated import deprecated
from environs import Env
from pydantic import BaseModel, Field, model_validator

from ._utils.enum import to_enum
from ._utils.stream import process_streaming_response
from ._utils.timeout import with_timeout
from .code_toolbox.sandbox_python_code_toolbox import SandboxPythonCodeToolbox
from .code_toolbox.sandbox_ts_code_toolbox import SandboxTsCodeToolbox
from .image import Image
from .object_storage import ObjectStorage
from .sandbox import Sandbox, SandboxTargetRegion
from .volume import VolumeService

Workspace = Sandbox


@dataclass
class CodeLanguage(Enum):
    """Programming languages supported by Daytona

    **Enum Members**:
        - `PYTHON` ("python")
        - `TYPESCRIPT` ("typescript")
        - `JAVASCRIPT` ("javascript")
    """

    PYTHON = "python"
    TYPESCRIPT = "typescript"
    JAVASCRIPT = "javascript"

    def __str__(self):
        return self.value

    def __eq__(self, other):
        if isinstance(other, str):
            return self.value == other
        return super().__eq__(other)


class DaytonaConfig(BaseModel):
    """Configuration options for initializing the Daytona client.

    Attributes:
        api_key (Optional[str]): API key for authentication with the Daytona API. If not set, it must be provided
            via the environment variable `DAYTONA_API_KEY`, or a JWT token must be provided instead.
        jwt_token (Optional[str]): JWT token for authentication with the Daytona API. If not set, it must be provided
            via the environment variable `DAYTONA_JWT_TOKEN`, or an API key must be provided instead.
        organization_id (Optional[str]): Organization ID used for JWT-based authentication. Required if a JWT token
            is provided, and must be set either here or in the environment variable `DAYTONA_ORGANIZATION_ID`.
        api_url (Optional[str]): URL of the Daytona API. Defaults to `'https://app.daytona.io/api'` if not set
            here or in the environment variable `DAYTONA_API_URL`.
        server_url (Optional[str]): Deprecated. Use `api_url` instead. This property will be removed
            in a future version.
        target (Optional[SandboxTargetRegion]): Target environment for the Sandbox. Defaults to `'us'` if not set here
            or in the environment variable `DAYTONA_TARGET`.

    Example:
        ```python
        config = DaytonaConfig(api_key="your-api-key")
        ```
        ```python
        config = DaytonaConfig(jwt_token="your-jwt-token", organization_id="your-organization-id")
        ```
    """

    api_key: Optional[str] = None
    api_url: Optional[str] = None
    server_url: Annotated[
        Optional[str],
        Field(
            default=None,
            deprecated="`server_url` is deprecated and will be removed in a future version. Use `api_url` instead.",
        ),
    ]
    target: Optional[SandboxTargetRegion] = None
    jwt_token: Optional[str] = None
    organization_id: Optional[str] = None

    @model_validator(mode="before")
    @classmethod
    def __handle_deprecated_server_url(cls, values):  # pylint: disable=unused-private-member
        if "server_url" in values and values.get("server_url"):
            warnings.warn(
                "'server_url' is deprecated and will be removed in a future version. Use 'api_url' instead.",
                DeprecationWarning,
                stacklevel=3,
            )
            if "api_url" not in values or not values["api_url"]:
                values["api_url"] = values["server_url"]
        return values


@dataclass
class SandboxResources:
    """Resources configuration for Sandbox.

    Attributes:
        cpu (Optional[int]): Number of CPU cores to allocate.
        memory (Optional[int]): Amount of memory in GB to allocate.
        disk (Optional[int]): Amount of disk space in GB to allocate.
        gpu (Optional[int]): Number of GPUs to allocate.

    Example:
        ```python
        resources = SandboxResources(
            cpu=2,
            memory=4,  # 4GB RAM
            disk=20,   # 20GB disk
            gpu=1
        )
        params = CreateSandboxParams(
            language="python",
            resources=resources
        )
        ```
    """

    cpu: Optional[int] = None
    memory: Optional[int] = None
    disk: Optional[int] = None
    gpu: Optional[int] = None


class CreateSandboxParams(BaseModel):
    """Parameters for creating a new Sandbox.

    Attributes:
        language (Optional[CodeLanguage]): Programming language for the Sandbox ("python", "javascript", "typescript").
        Defaults to "python".
        image (Optional[str]): Custom Docker image to use for the Sandbox.
        os_user (Optional[str]): OS user for the Sandbox.
        env_vars (Optional[Dict[str, str]]): Environment variables to set in the Sandbox.
        labels (Optional[Dict[str, str]]): Custom labels for the Sandbox.
        public (Optional[bool]): Whether the Sandbox should be public.
        resources (Optional[SandboxResources]): Resource configuration for the Sandbox.
        timeout (Optional[float]): Timeout in seconds for Sandbox to be created and started.
        auto_stop_interval (Optional[int]): Interval in minutes after which Sandbox will
            automatically stop if no Sandbox event occurs during that time. Default is 15 minutes.
            0 means no auto-stop.

    Example:
        ```python
        params = CreateSandboxParams(
            language="python",
            env_vars={"DEBUG": "true"},
            resources=SandboxResources(cpu=2, memory=4),
            auto_stop_interval=20
        )
        sandbox = daytona.create(params, 50)
        ```
    """

    language: Optional[CodeLanguage] = None
    image: Optional[Union[str, Image]] = None
    os_user: Optional[str] = None
    env_vars: Optional[Dict[str, str]] = None
    labels: Optional[Dict[str, str]] = None
    public: Optional[bool] = None
    resources: Optional[SandboxResources] = None
    timeout: Annotated[
        Optional[float],
        Field(
            default=None,
            deprecated=(
                "The `timeout` field is deprecated and will be removed in future versions. "
                "Use `timeout` argument in method calls instead."
            ),
        ),
    ]
    auto_stop_interval: Optional[int] = None
    volumes: Optional[List[VolumeMount]] = None

    @model_validator(mode="before")
    @classmethod
    def __handle_deprecated_timeout(cls, values):  # pylint: disable=unused-private-member
        if "timeout" in values and values.get("timeout"):
            warnings.warn(
                "The `timeout` field is deprecated and will be removed in future versions. "
                + "Use `timeout` argument in method calls instead.",
                DeprecationWarning,
                stacklevel=3,
            )
        return values


class Daytona:
    """Main class for interacting with the Daytona API.

    This class provides methods to create, manage, and interact with Daytona Sandboxes.
    It can be initialized either with explicit configuration or using environment variables.

    Attributes:
        api_key (str): API key for authentication.
        api_url (str): URL of the Daytona API.
        target (str): Default target location for Sandboxes.
        volume (VolumeService): Service for managing volumes.

    Example:
        Using environment variables:
        ```python
        daytona = Daytona()  # Uses DAYTONA_API_KEY, DAYTONA_API_URL
        ```

        Using explicit configuration:
        ```python
        config = DaytonaConfig(
            api_key="your-api-key",
            api_url="https://your-api.com",
            target="us"
        )
        daytona = Daytona(config)
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
        api_client = ApiClient(configuration)
        api_client.default_headers["Authorization"] = f"Bearer {self.api_key or self.jwt_token}"
        api_client.default_headers["X-Daytona-Source"] = "python-sdk"
        if not self.api_key:
            if not self.organization_id:
                raise DaytonaError("Organization ID is required when using JWT token")
            api_client.default_headers["X-Daytona-Organization-ID"] = self.organization_id

        # Initialize API clients with the api_client instance
        self.sandbox_api = SandboxApi(api_client)
        self.toolbox_api = ToolboxApi(api_client)
        self.image_api = ImagesApi(api_client)
        self.object_storage_api = ObjectStorageApi(api_client)

        # Initialize volume service
        self.volume = VolumeService(VolumesApi(api_client))

    @intercept_errors(message_prefix="Failed to create sandbox: ")
    def create(
        self,
        params: Optional[CreateSandboxParams] = None,
        timeout: Optional[float] = 60,
        *,
        on_image_build_logs: Callable[[str], None] = None,
    ) -> Sandbox:
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
            DaytonaError: If timeout or auto_stop_interval is negative; If sandbox fails to start or times out

        Example:
            Create a default Python Sandbox:
            ```python
            sandbox = daytona.create()
            ```

            Create a custom Sandbox:
            ```python
            params = CreateSandboxParams(
                language="python",
                image="debian:12.9",
                env_vars={"DEBUG": "true"},
                resources=SandboxResources(cpu=2, memory=4),
                auto_stop_interval=0
            )
            sandbox = daytona.create(params, 40)
            ```
        """
        # If no params provided, create default params for Python
        if params is None:
            params = CreateSandboxParams(language=self.default_language)
        if params.language is None:
            params.language = self.default_language

        effective_timeout = params.timeout if params.timeout else timeout

        return self._create(params, effective_timeout, on_image_build_logs=on_image_build_logs)

    @with_timeout(
        error_message=lambda self, timeout: (
            f"Failed to create and start sandbox within {timeout} seconds timeout period."
        )
    )
    def _create(
        self,
        params: Optional[CreateSandboxParams] = None,
        timeout: Optional[float] = 60,
        *,
        on_image_build_logs: Callable[[str], None] = None,
    ) -> Sandbox:
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
            DaytonaError: If timeout or auto_stop_interval is negative; If sandbox fails to start or times out
        """
        code_toolbox = self._get_code_toolbox(params)

        if timeout < 0:
            raise DaytonaError("Timeout must be a non-negative number")

        if params.auto_stop_interval is not None and params.auto_stop_interval < 0:
            raise DaytonaError("auto_stop_interval must be a non-negative integer")

        target = self.target

        # Create sandbox using dictionary
        sandbox_data = CreateSandbox(
            user=params.os_user,
            env=params.env_vars if params.env_vars else {},
            labels=params.labels,
            public=params.public,
            target=str(target) if target else None,
            auto_stop_interval=params.auto_stop_interval,
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

        response = self.sandbox_api.create_workspace(sandbox_data, _request_timeout=timeout or None)

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

            def should_terminate():
                response_ref["response"] = self.sandbox_api.get_workspace(response_ref["response"].id)
                return response_ref["response"].state in [
                    SandboxState.STARTED,
                    SandboxState.STARTING,
                    SandboxState.ERROR,
                ]

            while response_ref["response"].state == SandboxState.PENDING_BUILD:
                time.sleep(1)
                response_ref["response"] = self.sandbox_api.get_workspace(response_ref["response"].id)

            asyncio.run(
                process_streaming_response(
                    url=url,
                    headers=self.sandbox_api.api_client.default_headers,
                    on_chunk=on_image_build_logs,
                    should_terminate=should_terminate,
                )
            )
            response = response_ref["response"]

        sandbox_info = Sandbox.to_sandbox_info(response)
        response.info = sandbox_info

        sandbox = Sandbox(
            response.id,
            response,
            self.sandbox_api,
            self.toolbox_api,
            code_toolbox,
        )

        # Wait for sandbox to start
        if sandbox.instance.state != SandboxState.STARTED:
            try:
                sandbox.wait_for_sandbox_start(timeout)
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
    def delete(self, sandbox: Sandbox, timeout: Optional[float] = 60) -> None:
        """Deletes a Sandbox.

        Args:
            sandbox (Sandbox): The Sandbox instance to delete.
            timeout (Optional[float]): Timeout (in seconds) for sandbox deletion. 0 means no timeout.
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
        return self.sandbox_api.delete_workspace(sandbox.id, force=True, _request_timeout=timeout or None)

    remove = delete

    @deprecated(
        reason=(
            "Method is deprecated. Use `get_current_sandbox` instead. This method will be removed in a future version."
        )
    )
    def get_current_workspace(self, workspace_id: str) -> Workspace:
        """Gets a Sandbox by its ID.

        Args:
            workspace_id (str): The ID of the Sandbox to retrieve.

        Returns:
            Workspace: The Sandbox instance.
        """
        return self.get_current_sandbox(workspace_id)

    @intercept_errors(message_prefix="Failed to get sandbox: ")
    def get_current_sandbox(self, sandbox_id: str) -> Sandbox:
        """Gets a Sandbox by its ID.

        Args:
            sandbox_id (str): The ID of the Sandbox to retrieve.

        Returns:
            Sandbox: The Sandbox instance.

        Raises:
            DaytonaError: If sandbox_id is not provided.

        Example:
            ```python
            sandbox = daytona.get_current_sandbox("my-sandbox-id")
            print(sandbox.status)
            ```
        """
        if not sandbox_id:
            raise DaytonaError("sandbox_id is required")

        # Get the sandbox instance
        sandbox_instance = self.sandbox_api.get_workspace(sandbox_id)
        sandbox_info = Sandbox.to_sandbox_info(sandbox_instance)
        sandbox_instance.info = sandbox_info

        # Create and return sandbox with Python code toolbox as default
        code_toolbox = SandboxPythonCodeToolbox()
        return Sandbox(
            sandbox_id,
            sandbox_instance,
            self.sandbox_api,
            self.toolbox_api,
            code_toolbox,
        )

    @intercept_errors(message_prefix="Failed to find sandbox: ")
    def find_one(self, sandbox_id: Optional[str] = None, labels: Optional[Dict[str, str]] = None) -> Sandbox:
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
            sandbox = daytona.find_one(labels={"my-label": "my-value"})
            print(sandbox.info())
            ```
        """
        if sandbox_id:
            return self.get_current_sandbox(sandbox_id)
        sandboxes = self.list(labels)
        if len(sandboxes) == 0:
            raise DaytonaError(f"No sandbox found with labels {labels}")
        return sandboxes[0]

    @intercept_errors(message_prefix="Failed to list sandboxes: ")
    def list(self, labels: Optional[Dict[str, str]] = None) -> List[Sandbox]:
        """Lists Sandboxes filtered by labels.

        Args:
            labels (Optional[Dict[str, str]]): Labels to filter Sandboxes.

        Returns:
            List[Sandbox]: List of Sandbox instances that match the labels.

        Example:
            ```python
            sandboxes = daytona.list(labels={"my-label": "my-value"})
            for sandbox in sandboxes:
                print(f"{sandbox.id}: {sandbox.status}")
            ```
        """
        sandboxes = self.sandbox_api.list_workspaces(labels=json.dumps(labels))

        for sandbox in sandboxes:
            sandbox_info = Sandbox.to_sandbox_info(sandbox)
            sandbox.info = sandbox_info

        return [
            Sandbox(
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

    def start(self, sandbox: Sandbox, timeout: Optional[float] = 60) -> None:
        """Starts a Sandbox and waits for it to be ready.

        Args:
            sandbox (Sandbox): The Sandbox to start.
            timeout (Optional[float]): Optional timeout in seconds to wait for the Sandbox to start.
                0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative; If Sandbox fails to start or times out
        """
        sandbox.start(timeout)

    def stop(self, sandbox: Sandbox, timeout: Optional[float] = 60) -> None:
        """Stops a Sandbox and waits for it to be stopped.

        Args:
            sandbox (Sandbox): The sandbox to stop
            timeout (Optional[float]): Optional timeout (in seconds) for sandbox stop.
                0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative; If Sandbox fails to stop or times out
        """
        sandbox.stop(timeout)

    @intercept_errors(message_prefix="Failed to build image: ")
    @with_timeout(
        error_message=lambda self, timeout: (f"Failed to build image within {timeout} seconds timeout period.")
    )
    def create_image(
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
        context_hashes = self.__process_image_context(image)
        built_image = self.image_api.build_image(
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

            asyncio.run(
                process_streaming_response(
                    url=url,
                    headers=self.image_api.api_client.default_headers,
                    on_chunk=on_logs,
                    should_terminate=lambda: built_image.state in terminal_states,
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
            built_image = self.image_api.get_image(built_image.id)

        if on_logs:
            thread.join()
            if built_image.state == ImageState.ACTIVE:
                on_logs(f"Built image {built_image.name} ({built_image.state})")

        if built_image.state == ImageState.ERROR:
            raise DaytonaError(f"Failed to build image {built_image.name}, error reason: {built_image.error_reason}")

    def __process_image_context(self, image: Image) -> List[str]:
        """Processes the image context by uploading it to object storage.

        Args:
            image (Image): The Image instance.

        Returns:
            List[str]: List of context hashes stored in object storage.
        """
        if not image._context_list:  # pylint: disable=protected-access
            return []

        push_access_creds = self.object_storage_api.get_push_access()
        object_storage = ObjectStorage(
            push_access_creds.storage_url,
            push_access_creds.access_key,
            push_access_creds.secret,
            push_access_creds.session_token,
            push_access_creds.bucket,
        )

        context_hashes = []
        for context in image._context_list:  # pylint: disable=protected-access
            context_hash = object_storage.upload(
                context.source_path, push_access_creds.organization_id, context.archive_path
            )
            context_hashes.append(context_hash)

        return context_hashes


# Export these at module level
__all__ = [
    "Daytona",
    "DaytonaConfig",
    "CreateSandboxParams",
    "CodeLanguage",
    "Sandbox",
    "SessionExecuteRequest",
    "SessionExecuteResponse",
]
