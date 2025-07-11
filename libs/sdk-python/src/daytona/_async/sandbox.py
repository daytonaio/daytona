# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import asyncio
from typing import Dict, Optional

from daytona_api_client_async import PortPreviewUrl
from daytona_api_client_async import Sandbox as SandboxDto
from daytona_api_client_async import SandboxApi, ToolboxApi
from pydantic import ConfigDict, PrivateAttr

from .._utils.errors import intercept_errors
from .._utils.path import prefix_relative_path
from .._utils.timeout import with_timeout
from ..common.errors import DaytonaError
from ..common.protocols import SandboxCodeToolbox
from .computer_use import AsyncComputerUse
from .filesystem import AsyncFileSystem
from .git import AsyncGit
from .lsp_server import AsyncLspServer, LspLanguageId
from .process import AsyncProcess


class AsyncSandbox(SandboxDto):
    """Represents a Daytona Sandbox.

    Attributes:
        fs (AsyncFileSystem): File system operations interface.
        git (AsyncGit): Git operations interface.
        process (AsyncProcess): Process execution interface.
        computer_use (AsyncComputerUse): Computer use operations interface for desktop automation.
        id (str): Unique identifier for the Sandbox.
        organization_id (str): Organization ID of the Sandbox.
        snapshot (str): Daytona snapshot used to create the Sandbox.
        user (str): OS user running in the Sandbox.
        env (Dict[str, str]): Environment variables set in the Sandbox.
        labels (Dict[str, str]): Custom labels attached to the Sandbox.
        public (bool): Whether the Sandbox is publicly accessible.
        target (str): Target location of the runner where the Sandbox runs.
        cpu (int): Number of CPUs allocated to the Sandbox.
        gpu (int): Number of GPUs allocated to the Sandbox.
        memory (int): Amount of memory allocated to the Sandbox in GiB.
        disk (int): Amount of disk space allocated to the Sandbox in GiB.
        state (SandboxState): Current state of the Sandbox (e.g., "started", "stopped").
        error_reason (str): Error message if Sandbox is in error state.
        backup_state (SandboxBackupStateEnum): Current state of Sandbox backup.
        backup_created_at (str): When the backup was created.
        auto_stop_interval (int): Auto-stop interval in minutes.
        auto_archive_interval (int): Auto-archive interval in minutes.
        auto_delete_interval (int): Auto-delete interval in minutes.
        runner_domain (str): Domain name of the Sandbox runner.
        volumes (List[str]): Volumes attached to the Sandbox.
        build_info (str): Build information for the Sandbox if it was created from dynamic build.
        created_at (str): When the Sandbox was created.
        updated_at (str): When the Sandbox was last updated.
    """

    _fs: AsyncFileSystem = PrivateAttr()
    _git: AsyncGit = PrivateAttr()
    _process: AsyncProcess = PrivateAttr()
    _computer_use: AsyncComputerUse = PrivateAttr()

    # TODO: Remove model_config once everything is migrated to pydantic # pylint: disable=fixme
    model_config = ConfigDict(arbitrary_types_allowed=True)

    def __init__(
        self,
        sandbox_dto: SandboxDto,
        sandbox_api: SandboxApi,
        toolbox_api: ToolboxApi,
        code_toolbox: SandboxCodeToolbox,
    ):
        """Initialize a new Sandbox instance.

        Args:
            id (str): Unique identifier for the Sandbox.
            instance (SandboxInstance): The underlying Sandbox instance.
            sandbox_api (SandboxApi): API client for Sandbox operations.
            toolbox_api (ToolboxApi): API client for toolbox operations.
            code_toolbox (SandboxCodeToolbox): Language-specific toolbox implementation.
        """
        super().__init__(**sandbox_dto.model_dump())
        self.__process_sandbox_dto(sandbox_dto)
        self._sandbox_api = sandbox_api
        self._toolbox_api = toolbox_api
        self._code_toolbox = code_toolbox
        self._root_dir = ""

        self._fs = AsyncFileSystem(self.id, toolbox_api, self.__get_root_dir)
        self._git = AsyncGit(self.id, toolbox_api, self.__get_root_dir)
        self._process = AsyncProcess(self.id, code_toolbox, toolbox_api, self.__get_root_dir)
        self._computer_use = AsyncComputerUse(self.id, toolbox_api)

    @property
    def fs(self) -> AsyncFileSystem:
        return self._fs

    @property
    def git(self) -> AsyncGit:
        return self._git

    @property
    def process(self) -> AsyncProcess:
        return self._process

    @property
    def computer_use(self) -> AsyncComputerUse:
        return self._computer_use

    async def refresh_data(self) -> None:
        """Refreshes the Sandbox data from the API.

        Example:
            ```python
            await sandbox.refresh_data()
            print(f"Sandbox {sandbox.id}:")
            print(f"State: {sandbox.state}")
            print(f"Resources: {sandbox.cpu} CPU, {sandbox.memory} GiB RAM")
            ```
        """
        instance = await self._sandbox_api.get_sandbox(self.id)
        self.__process_sandbox_dto(instance)

    @intercept_errors(message_prefix="Failed to get sandbox root directory: ")
    async def get_user_root_dir(self) -> str:
        """Gets the root directory path for the logged in user inside the Sandbox.

        Returns:
            str: The absolute path to the Sandbox root directory for the logged in user.

        Example:
            ```python
            root_dir = await sandbox.get_user_root_dir()
            print(f"Sandbox root: {root_dir}")
            ```
        """
        response = await self._toolbox_api.get_project_dir(self.id)
        return response.dir

    def create_lsp_server(self, language_id: LspLanguageId, path_to_project: str) -> AsyncLspServer:
        """Creates a new Language Server Protocol (LSP) server instance.

        The LSP server provides language-specific features like code completion,
        diagnostics, and more.

        Args:
            language_id (LspLanguageId): The language server type (e.g., LspLanguageId.PYTHON).
            path_to_project (str): Path to the project root directory. Relative paths are resolved based on the user's
            root directory.

        Returns:
            LspServer: A new LSP server instance configured for the specified language.

        Example:
            ```python
            lsp = sandbox.create_lsp_server("python", "workspace/project")
            ```
        """
        return AsyncLspServer(
            language_id,
            prefix_relative_path(self._root_dir, path_to_project),
            self._toolbox_api,
            self.id,
        )

    @intercept_errors(message_prefix="Failed to set labels: ")
    async def set_labels(self, labels: Dict[str, str]) -> Dict[str, str]:
        """Sets labels for the Sandbox.

        Labels are key-value pairs that can be used to organize and identify Sandboxes.

        Args:
            labels (Dict[str, str]): Dictionary of key-value pairs representing Sandbox labels.

        Returns:
            Dict[str, str]: Dictionary containing the updated Sandbox labels.

        Example:
            ```python
            new_labels = sandbox.set_labels({
                "project": "my-project",
                "environment": "development",
                "team": "backend"
            })
            print(f"Updated labels: {new_labels}")
            ```
        """
        # Convert all values to strings and create the expected labels structure
        string_labels = {k: str(v).lower() if isinstance(v, bool) else str(v) for k, v in labels.items()}
        labels_payload = {"labels": string_labels}
        self.labels = (await self._sandbox_api.replace_labels(self.id, labels_payload)).labels
        return self.labels

    @intercept_errors(message_prefix="Failed to start sandbox: ")
    @with_timeout(
        error_message=lambda self, timeout: (
            f"Sandbox {self.id} failed to start within the {timeout} seconds timeout period"
        )
    )
    async def start(self, timeout: Optional[float] = 60):
        """Starts the Sandbox and waits for it to be ready.

        Args:
            timeout (Optional[float]): Maximum time to wait in seconds. 0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative. If sandbox fails to start or times out.

        Example:
            ```python
            sandbox = daytona.get_current_sandbox("my-sandbox")
            sandbox.start(timeout=40)  # Wait up to 40 seconds
            print("Sandbox started successfully")
            ```
        """
        sandbox = await self._sandbox_api.start_sandbox(self.id, _request_timeout=timeout or None)
        self.__process_sandbox_dto(sandbox)
        await self.wait_for_sandbox_start()

    @intercept_errors(message_prefix="Failed to stop sandbox: ")
    @with_timeout(
        error_message=lambda self, timeout: (
            f"Sandbox {self.id} failed to stop within the {timeout} seconds timeout period"
        )
    )
    async def stop(self, timeout: Optional[float] = 60):
        """Stops the Sandbox and waits for it to be fully stopped.

        Args:
            timeout (Optional[float]): Maximum time to wait in seconds. 0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative; If sandbox fails to stop or times out

        Example:
            ```python
            sandbox = daytona.get_current_sandbox("my-sandbox")
            sandbox.stop()
            print("Sandbox stopped successfully")
            ```
        """
        await self._sandbox_api.stop_sandbox(self.id, _request_timeout=timeout or None)
        await self.refresh_data()
        await self.wait_for_sandbox_stop()

    @intercept_errors(message_prefix="Failed to remove sandbox: ")
    async def delete(self, timeout: Optional[float] = 60) -> None:
        """Deletes the Sandbox.

        Args:
            timeout (Optional[float]): Timeout (in seconds) for sandbox deletion. 0 means no timeout.
                Default is 60 seconds.
        """
        await self._sandbox_api.delete_sandbox(self.id, force=True, _request_timeout=timeout or None)
        await self.refresh_data()

    @intercept_errors(message_prefix="Failure during waiting for sandbox to start: ")
    @with_timeout(
        error_message=lambda self, timeout: (
            f"Sandbox {self.id} failed to become ready within the {timeout} seconds timeout period"
        )
    )
    async def wait_for_sandbox_start(
        self,
        timeout: Optional[float] = 60,  # pylint: disable=unused-argument
    ) -> None:
        """Waits for the Sandbox to reach the 'started' state. Polls the Sandbox status until it
        reaches the 'started' state, encounters an error or times out.

        Args:
            timeout (Optional[float]): Maximum time to wait in seconds. 0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative; If Sandbox fails to start or times out
        """
        while self.state != "started":
            await self.refresh_data()

            if self.state == "started":
                return

            if self.state in ["error", "build_failed"]:
                err_msg = (
                    f"Sandbox {self.id} failed to start with state: {self.state}, error reason: {self.error_reason}"
                )
                raise DaytonaError(err_msg)

            await asyncio.sleep(0.1)  # Wait 100ms between checks

    @intercept_errors(message_prefix="Failure during waiting for sandbox to stop: ")
    @with_timeout(
        error_message=lambda self, timeout: (
            f"Sandbox {self.id} failed to become stopped within the {timeout} seconds timeout period"
        )
    )
    async def wait_for_sandbox_stop(
        self,
        timeout: Optional[float] = 60,  # pylint: disable=unused-argument
    ) -> None:
        """Waits for the Sandbox to reach the 'stopped' state. Polls the Sandbox status until it
        reaches the 'stopped' state, encounters an error or times out. It will wait up to 60 seconds
        for the Sandbox to stop.

        Args:
            timeout (Optional[float]): Maximum time to wait in seconds. 0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative. If Sandbox fails to stop or times out.
        """
        while self.state != "stopped":
            try:
                await self.refresh_data()

                if self.state in ["error", "build_failed"]:
                    err_msg = (
                        f"Sandbox {self.id} failed to stop with status: {self.state}, error reason: {self.error_reason}"
                    )
                    raise DaytonaError(err_msg)
            except Exception as e:
                # If there's a validation error, continue waiting
                if "validation error" not in str(e):
                    raise e

            await asyncio.sleep(0.1)  # Wait 100ms between checks

    @intercept_errors(message_prefix="Failed to set auto-stop interval: ")
    async def set_autostop_interval(self, interval: int) -> None:
        """Sets the auto-stop interval for the Sandbox.

        The Sandbox will automatically stop after being idle (no new events) for the specified interval.
        Events include any state changes or interactions with the Sandbox through the SDK.
        Interactions using Sandbox Previews are not included.

        Args:
            interval (int): Number of minutes of inactivity before auto-stopping.
                Set to 0 to disable auto-stop. Defaults to 15.

        Raises:
            DaytonaError: If interval is negative

        Example:
            ```python
            # Auto-stop after 1 hour
            sandbox.set_autostop_interval(60)
            # Or disable auto-stop
            sandbox.set_autostop_interval(0)
            ```
        """
        if not isinstance(interval, int) or interval < 0:
            raise DaytonaError("Auto-stop interval must be a non-negative integer")

        await self._sandbox_api.set_autostop_interval(self.id, interval)
        self.auto_stop_interval = interval

    @intercept_errors(message_prefix="Failed to set auto-archive interval: ")
    async def set_auto_archive_interval(self, interval: int) -> None:
        """Sets the auto-archive interval for the Sandbox.

        The Sandbox will automatically archive after being continuously stopped for the specified interval.

        Args:
            interval (int): Number of minutes after which a continuously stopped Sandbox will be auto-archived.
                Set to 0 for the maximum interval. Default is 7 days.

        Raises:
            DaytonaError: If interval is negative

        Example:
            ```python
            # Auto-archive after 1 hour
            sandbox.set_auto_archive_interval(60)
            # Or use the maximum interval
            sandbox.set_auto_archive_interval(0)
            ```
        """
        if not isinstance(interval, int) or interval < 0:
            raise DaytonaError("Auto-archive interval must be a non-negative integer")
        await self._sandbox_api.set_auto_archive_interval(self.id, interval)
        self.auto_archive_interval = interval

    @intercept_errors(message_prefix="Failed to set auto-delete interval: ")
    async def set_auto_delete_interval(self, interval: int) -> None:
        """Sets the auto-delete interval for the Sandbox.

        The Sandbox will automatically delete after being continuously stopped for the specified interval.

        Args:
            interval (int): Number of minutes after which a continuously stopped Sandbox will be auto-deleted.
                Set to negative value to disable auto-delete. Set to 0 to delete immediately upon stopping.
                By default, auto-delete is disabled.

        Example:
            ```python
            # Auto-delete after 1 hour
            sandbox.set_auto_delete_interval(60)
            # Or delete immediately upon stopping
            sandbox.set_auto_delete_interval(0)
            # Or disable auto-delete
            sandbox.set_auto_delete_interval(-1)
            ```
        """
        await self._sandbox_api.set_auto_delete_interval(self.id, interval)
        self.auto_delete_interval = interval

    @intercept_errors(message_prefix="Failed to get preview link: ")
    async def get_preview_link(self, port: int) -> PortPreviewUrl:
        """Retrieves the preview link for the sandbox at the specified port. If the port is closed,
        it will be opened automatically. For private sandboxes, a token is included to grant access
        to the URL.

        Args:
            port (int): The port to open the preview link on.

        Returns:
            PortPreviewUrl: The response object for the preview link, which includes the `url`
            and the `token` (to access private sandboxes).

        Example:
            ```python
            preview_link = sandbox.get_preview_link(3000)
            print(f"Preview URL: {preview_link.url}")
            print(f"Token: {preview_link.token}")
            ```
        """
        return await self._sandbox_api.get_port_preview_url(self.id, port)

    @intercept_errors(message_prefix="Failed to archive sandbox: ")
    async def archive(self) -> None:
        """Archives the sandbox, making it inactive and preserving its state. When sandboxes are
        archived, the entire filesystem state is moved to cost-effective object storage, making it
        possible to keep sandboxes available for an extended period. The tradeoff between archived
        and stopped states is that starting an archived sandbox takes more time, depending on its size.
        Sandbox must be stopped before archiving.
        """
        await self._sandbox_api.archive_sandbox(self.id)
        await self.refresh_data()

    async def __get_root_dir(self) -> str:
        if not self._root_dir:
            self._root_dir = await self.get_user_root_dir()
        return self._root_dir

    def __process_sandbox_dto(self, sandbox_dto: SandboxDto) -> None:
        self.id = sandbox_dto.id
        self.organization_id = sandbox_dto.organization_id
        self.snapshot = sandbox_dto.snapshot
        self.user = sandbox_dto.user
        self.env = sandbox_dto.env
        self.labels = sandbox_dto.labels
        self.public = sandbox_dto.public
        self.target = sandbox_dto.target
        self.cpu = sandbox_dto.cpu
        self.gpu = sandbox_dto.gpu
        self.memory = sandbox_dto.memory
        self.disk = sandbox_dto.disk
        self.state = sandbox_dto.state
        self.error_reason = sandbox_dto.error_reason
        self.backup_state = sandbox_dto.backup_state
        self.backup_created_at = sandbox_dto.backup_created_at
        self.auto_stop_interval = sandbox_dto.auto_stop_interval
        self.auto_archive_interval = sandbox_dto.auto_archive_interval
        self.auto_delete_interval = sandbox_dto.auto_delete_interval
        self.runner_domain = sandbox_dto.runner_domain
        self.volumes = sandbox_dto.volumes
        self.build_info = sandbox_dto.build_info
        self.created_at = sandbox_dto.created_at
        self.updated_at = sandbox_dto.updated_at
