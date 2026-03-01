# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import threading
from typing import Any

from daytona_api_client import BuildInfo
from daytona_api_client import PaginatedSandboxes as PaginatedSandboxesDto
from daytona_api_client import PortPreviewUrl, ResizeSandbox
from daytona_api_client import Sandbox as SandboxDto
from daytona_api_client import (
    SandboxApi,
    SandboxLabels,
    SandboxState,
    SandboxVolume,
    SignedPortPreviewUrl,
    SshAccessDto,
    SshAccessValidationDto,
)
from daytona_toolbox_api_client import (
    ApiClient,
    ComputerUseApi,
    FileSystemApi,
    GitApi,
    InfoApi,
    InterpreterApi,
    LspApi,
    ProcessApi,
)
from deprecated import deprecated
from pydantic import ConfigDict, PrivateAttr

from .._utils.errors import intercept_errors
from .._utils.otel_decorator import with_instrumentation
from .._utils.timeout import http_timeout, with_timeout
from ..common.errors import DaytonaError, DaytonaNotFoundError
from ..common.lsp_server import LspLanguageId, LspLanguageIdLiteral
from ..common.protocols import SandboxCodeToolbox
from ..common.sandbox import Resources
from ..internal.event_subscriber import SyncEventSubscriber
from ..internal.toolbox_api_client_proxy import ToolboxApiClientProxy
from .code_interpreter import CodeInterpreter
from .computer_use import ComputerUse
from .filesystem import FileSystem
from .git import Git
from .lsp_server import LspServer
from .process import Process


class Sandbox(SandboxDto):
    """Represents a Daytona Sandbox.

    Attributes:
        fs (FileSystem): File system operations interface.
        git (Git): Git operations interface.
        process (Process): Process execution interface.
        computer_use (ComputerUse): Computer use operations interface for desktop automation.
        code_interpreter (CodeInterpreter): Stateful interpreter interface for executing code.
            Currently supports only Python. For other languages, use the `process.code_run` interface.
        id (str): Unique identifier for the Sandbox.
        name (str): Name of the Sandbox.
        organization_id (str): Organization ID of the Sandbox.
        snapshot (str): Daytona snapshot used to create the Sandbox.
        user (str): OS user running in the Sandbox.
        env (dict[str, str]): Environment variables set in the Sandbox.
        labels (dict[str, str]): Custom labels attached to the Sandbox.
        public (bool): Whether the Sandbox is publicly accessible.
        target (str): Target location of the runner where the Sandbox runs.
        cpu (int): Number of CPUs allocated to the Sandbox.
        gpu (int): Number of GPUs allocated to the Sandbox.
        memory (int): Amount of memory allocated to the Sandbox in GiB.
        disk (int): Amount of disk space allocated to the Sandbox in GiB.
        state (SandboxState): Current state of the Sandbox (e.g., "started", "stopped").
        error_reason (str): Error message if Sandbox is in error state.
        recoverable (bool): Whether the Sandbox error is recoverable.
        backup_state (SandboxBackupStateEnum): Current state of Sandbox backup.
        backup_created_at (str): When the backup was created.
        auto_stop_interval (int): Auto-stop interval in minutes.
        auto_archive_interval (int): Auto-archive interval in minutes.
        auto_delete_interval (int): Auto-delete interval in minutes.
        volumes (list[str]): Volumes attached to the Sandbox.
        build_info (str): Build information for the Sandbox if it was created from dynamic build.
        created_at (str): When the Sandbox was created.
        updated_at (str): When the Sandbox was last updated.
        network_block_all (bool): Whether to block all network access for the Sandbox.
        network_allow_list (str): Comma-separated list of allowed CIDR network addresses for the Sandbox.
    """

    _fs: FileSystem = PrivateAttr()
    _git: Git = PrivateAttr()
    _process: Process = PrivateAttr()
    _computer_use: ComputerUse = PrivateAttr()
    _code_interpreter: CodeInterpreter = PrivateAttr()

    # TODO: Remove model_config once everything is migrated to pydantic # pylint: disable=fixme
    model_config: ConfigDict = ConfigDict(arbitrary_types_allowed=True)

    def __init__(
        self,
        sandbox_dto: SandboxDto,
        toolbox_api: ApiClient,
        sandbox_api: SandboxApi,
        code_toolbox: SandboxCodeToolbox,
        event_subscriber: SyncEventSubscriber | None = None,
    ):
        """Initialize a new Sandbox instance.

        Args:
            sandbox_dto (SandboxDto): The sandbox data from the API.
            toolbox_api (ApiClient): API client for toolbox operations.
            sandbox_api (SandboxApi): API client for Sandbox operations.
            code_toolbox (SandboxCodeToolbox): Language-specific toolbox implementation.
            event_subscriber: Optional SyncEventSubscriber for real-time updates.
        """
        super().__init__(**sandbox_dto.model_dump())
        self.__process_sandbox_dto(sandbox_dto)
        self._sandbox_api: SandboxApi = sandbox_api
        self._code_toolbox: SandboxCodeToolbox = code_toolbox
        self._event_subscriber: SyncEventSubscriber | None = event_subscriber
        # Wrap the toolbox API client to inject the sandbox ID into the resource path
        self._toolbox_api: ToolboxApiClientProxy[ApiClient] = ToolboxApiClientProxy(
            toolbox_api, self.id, self.toolbox_proxy_url
        )

        self._fs = FileSystem(FileSystemApi(self._toolbox_api))
        self._git = Git(GitApi(self._toolbox_api))
        self._process = Process(code_toolbox, ProcessApi(self._toolbox_api))
        self._computer_use = ComputerUse(ComputerUseApi(self._toolbox_api))
        self._code_interpreter = CodeInterpreter(InterpreterApi(self._toolbox_api))
        self._info_api: InfoApi = InfoApi(self._toolbox_api)

        # Subscribe to real-time events for this sandbox
        self._subscribe_to_events()

    @property
    def fs(self) -> FileSystem:
        return self._fs

    @property
    def git(self) -> Git:
        return self._git

    @property
    def process(self) -> Process:
        return self._process

    @property
    def computer_use(self) -> ComputerUse:
        return self._computer_use

    @property
    def code_interpreter(self) -> CodeInterpreter:
        return self._code_interpreter

    @intercept_errors(message_prefix="Failed to refresh sandbox data: ")
    @with_instrumentation()
    def refresh_data(self) -> None:
        """Refreshes the Sandbox data from the API.

        Example:
            ```python
            sandbox.refresh_data()
            print(f"Sandbox {sandbox.id}:")
            print(f"State: {sandbox.state}")
            print(f"Resources: {sandbox.cpu} CPU, {sandbox.memory} GiB RAM")
            ```
        """
        instance = self._sandbox_api.get_sandbox(self.id)
        self.__process_sandbox_dto(instance)

    @intercept_errors(message_prefix="Failed to get user home directory: ")
    @with_instrumentation()
    def get_user_home_dir(self) -> str:
        """Gets the user's home directory path inside the Sandbox.

        Returns:
            str: The absolute path to the user's home directory inside the Sandbox.

        Example:
            ```python
            user_home_dir = sandbox.get_user_home_dir()
            print(f"Sandbox user home: {user_home_dir}")
            ```
        """
        response = self._info_api.get_user_home_dir()
        return response.dir

    @deprecated(
        reason=(
            "Method is deprecated. Use `get_user_home_dir` instead. This method will be removed in a future version."
        )
    )
    @with_instrumentation()
    def get_user_root_dir(self) -> str:
        return self.get_user_home_dir()

    @intercept_errors(message_prefix="Failed to get working directory path: ")
    @with_instrumentation()
    def get_work_dir(self) -> str:
        """Gets the working directory path inside the Sandbox.

        Returns:
            str: The absolute path to the Sandbox working directory. Uses the WORKDIR specified in
            the Dockerfile if present, or falling back to the user's home directory if not.

        Example:
            ```python
            work_dir = sandbox.get_work_dir()
            print(f"Sandbox working directory: {work_dir}")
            ```
        """
        response = self._info_api.get_work_dir()
        return response.dir

    @with_instrumentation()
    def create_lsp_server(self, language_id: LspLanguageId | LspLanguageIdLiteral, path_to_project: str) -> LspServer:
        """Creates a new Language Server Protocol (LSP) server instance.

        The LSP server provides language-specific features like code completion,
        diagnostics, and more.

        Args:
            language_id (LspLanguageId | LspLanguageIdLiteral): The language server type (e.g., LspLanguageId.PYTHON).
            path_to_project (str): Path to the project root directory. Relative paths are resolved
            based on the sandbox working directory.

        Returns:
            LspServer: A new LSP server instance configured for the specified language.

        Example:
            ```python
            lsp = sandbox.create_lsp_server("python", "workspace/project")
            ```
        """
        return LspServer(
            language_id,
            path_to_project,
            LspApi(self._toolbox_api),
        )

    @intercept_errors(message_prefix="Failed to set labels: ")
    @with_instrumentation()
    def set_labels(self, labels: dict[str, str]) -> dict[str, str]:
        """Sets labels for the Sandbox.

        Labels are key-value pairs that can be used to organize and identify Sandboxes.

        Args:
            labels (dict[str, str]): Dictionary of key-value pairs representing Sandbox labels.

        Returns:
            dict[str, str]: Dictionary containing the updated Sandbox labels.

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
        self.labels = (self._sandbox_api.replace_labels(self.id, SandboxLabels(labels=labels))).labels
        return self.labels

    @intercept_errors(message_prefix="Failed to start sandbox: ")
    @with_timeout()
    @with_instrumentation()
    def start(self, timeout: float | None = 60):
        """Starts the Sandbox and waits for it to be ready.

        Args:
            timeout (float | None): Maximum time to wait in seconds. 0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative. If sandbox fails to start or times out.

        Example:
            ```python
            sandbox = daytona.get("my-sandbox-id")
            sandbox.start(timeout=40)  # Wait up to 40 seconds
            print("Sandbox started successfully")
            ```
        """
        sandbox = self._sandbox_api.start_sandbox(self.id, _request_timeout=http_timeout(timeout))
        self.__process_sandbox_dto(sandbox)
        # This method already handles a timeout, so we don't need to pass one to internal methods
        self.wait_for_sandbox_start(timeout=0)

    @intercept_errors(message_prefix="Failed to recover sandbox: ")
    @with_timeout()
    def recover(self, timeout: float | None = 60):
        """Recovers the Sandbox from a recoverable error and waits for it to be ready.

        Args:
            timeout (float | None): Maximum time to wait in seconds. 0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative. If sandbox fails to recover or times out.

        Example:
            ```python
            sandbox = daytona.get("my-sandbox-id")
            sandbox.recover(timeout=40)  # Wait up to 40 seconds
            print("Sandbox recovered successfully")
            ```
        """
        sandbox = self._sandbox_api.recover_sandbox(self.id, _request_timeout=http_timeout(timeout))
        self.__process_sandbox_dto(sandbox)
        # This method already handles a timeout, so we don't need to pass one to internal methods
        self.wait_for_sandbox_start(timeout=0)

    @intercept_errors(message_prefix="Failed to stop sandbox: ")
    @with_timeout()
    @with_instrumentation()
    def stop(self, timeout: float | None = 60):
        """Stops the Sandbox and waits for it to be fully stopped.

        Args:
            timeout (float | None): Maximum time to wait in seconds. 0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative; If sandbox fails to stop or times out

        Example:
            ```python
            sandbox = daytona.get("my-sandbox-id")
            sandbox.stop()
            print("Sandbox stopped successfully")
            ```
        """
        _ = self._sandbox_api.stop_sandbox(self.id, _request_timeout=http_timeout(timeout))
        self.__refresh_data_safe()
        # This method already handles a timeout, so we don't need to pass one to internal methods
        self.wait_for_sandbox_stop(timeout=0)

    @intercept_errors(message_prefix="Failed to remove sandbox: ")
    @with_instrumentation()
    def delete(self, timeout: float | None = 60) -> None:
        """Deletes the Sandbox.

        Args:
            timeout (float | None): Timeout (in seconds) for sandbox deletion. 0 means no timeout.
                Default is 60 seconds.
        """
        _ = self._sandbox_api.delete_sandbox(self.id, _request_timeout=http_timeout(timeout))
        self.__refresh_data_safe()

    @intercept_errors(message_prefix="Failure during waiting for sandbox to start: ")
    @with_timeout()
    @with_instrumentation()
    def wait_for_sandbox_start(
        self,
        timeout: float | None = 60,  # pylint: disable=unused-argument # pyright: ignore[reportUnusedParameter]
    ) -> None:
        """Waits for the Sandbox to reach the 'started' state via WebSocket events.

        Args:
            timeout (float | None): Maximum time to wait in seconds. 0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative; If Sandbox fails to start or times out;
                If WebSocket connection is not available
        """
        if self.state == "started":
            return

        if not self._event_subscriber or not self._event_subscriber.is_connected:
            raise DaytonaError("WebSocket connection not available for sandbox event subscription")

        self._wait_for_state_via_events(["started"], ["error", "build_failed"])

    @intercept_errors(message_prefix="Failure during waiting for sandbox to stop: ")
    @with_timeout()
    @with_instrumentation()
    def wait_for_sandbox_stop(
        self,
        timeout: float | None = 60,  # pylint: disable=unused-argument # pyright: ignore[reportUnusedParameter]
    ) -> None:
        """Waits for the Sandbox to reach the 'stopped' state. Uses WebSocket events when available,
        falls back to polling otherwise.
        Treats destroyed as stopped to cover ephemeral sandboxes that are automatically deleted after stopping.

        Args:
            timeout (float | None): Maximum time to wait in seconds. 0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative. If Sandbox fails to stop or times out.
        """
        if self.state in ["stopped", "destroyed"]:
            return

        if not self._event_subscriber or not self._event_subscriber.is_connected:
            raise DaytonaError("WebSocket connection not available for sandbox event subscription")

        self._wait_for_state_via_events(["stopped", "destroyed"], ["error", "build_failed"])

    @intercept_errors(message_prefix="Failed to set auto-stop interval: ")
    @with_instrumentation()
    def set_autostop_interval(self, interval: int) -> None:
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
        if interval < 0:
            raise DaytonaError("Auto-stop interval must be a non-negative integer")

        _ = self._sandbox_api.set_autostop_interval(self.id, interval)
        self.auto_stop_interval = interval

    @intercept_errors(message_prefix="Failed to set auto-archive interval: ")
    @with_instrumentation()
    def set_auto_archive_interval(self, interval: int) -> None:
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
        if interval < 0:
            raise DaytonaError("Auto-archive interval must be a non-negative integer")

        _ = self._sandbox_api.set_auto_archive_interval(self.id, interval)
        self.auto_archive_interval = interval

    @intercept_errors(message_prefix="Failed to set auto-delete interval: ")
    @with_instrumentation()
    def set_auto_delete_interval(self, interval: int) -> None:
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
        _ = self._sandbox_api.set_auto_delete_interval(self.id, interval)
        self.auto_delete_interval = interval

    @intercept_errors(message_prefix="Failed to get preview link: ")
    @with_instrumentation()
    def get_preview_link(self, port: int) -> PortPreviewUrl:
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
        return self._sandbox_api.get_port_preview_url(self.id, port)

    @intercept_errors(message_prefix="Failed to create signed preview url: ")
    def create_signed_preview_url(self, port: int, expires_in_seconds: int | None = None) -> SignedPortPreviewUrl:
        """Creates a signed preview URL for the sandbox at the specified port.

        Args:
            port (int): The port to open the preview link on.
            expires_in_seconds (int | None): The number of seconds the signed preview
                url will be valid for. Defaults to 60 seconds.

        Returns:
            SignedPortPreviewUrl: The response object for the signed preview url.
        """
        return self._sandbox_api.get_signed_port_preview_url(self.id, port, expires_in_seconds=expires_in_seconds)

    @intercept_errors(message_prefix="Failed to expire signed preview url: ")
    def expire_signed_preview_url(self, port: int, token: str) -> None:
        """Expires a signed preview URL for the sandbox at the specified port.

        Args:
            port (int): The port to expire the signed preview url on.
            token (str): The token to expire the signed preview url on.
        """
        self._sandbox_api.expire_signed_port_preview_url(self.id, port, token)

    @intercept_errors(message_prefix="Failed to archive sandbox: ")
    @with_instrumentation()
    def archive(self) -> None:
        """Archives the sandbox, making it inactive and preserving its state. When sandboxes are
        archived, the entire filesystem state is moved to cost-effective object storage, making it
        possible to keep sandboxes available for an extended period. The tradeoff between archived
        and stopped states is that starting an archived sandbox takes more time, depending on its size.
        Sandbox must be stopped before archiving.
        """
        _ = self._sandbox_api.archive_sandbox(self.id)
        self.refresh_data()

    @intercept_errors(message_prefix="Failed to resize sandbox: ")
    @with_timeout()
    @with_instrumentation()
    def resize(self, resources: Resources, timeout: float | None = 60) -> None:
        """Resizes the Sandbox resources.

        Changes the CPU, memory, or disk allocation for the Sandbox. Hot resize (on running
        sandbox) only allows CPU/memory increases. Disk resize requires a stopped sandbox.

        Args:
            resources (Resources): New resource configuration. Only specified fields will be updated.
                - cpu: Number of CPU cores (minimum: 1). For hot resize, can only be increased.
                - memory: Memory in GiB (minimum: 1). For hot resize, can only be increased.
                - disk: Disk space in GiB (can only be increased, requires stopped sandbox).
            timeout (Optional[float]): Timeout (in seconds) for the resize operation. 0 means no timeout.
                Default is 60 seconds.

        Raises:
            DaytonaError: If hot resize constraints are violated (CPU/memory decrease on running sandbox).
            DaytonaError: If disk resize attempted on running sandbox.
            DaytonaError: If disk size decrease is attempted.
            DaytonaError: If resize operation times out.
            DaytonaError: If no resource changes are specified.

        Example:
            ```python
            # Increase CPU/memory on running sandbox (hot resize)
            sandbox.resize(Resources(cpu=4, memory=8))

            # Change disk (sandbox must be stopped)
            sandbox.stop()
            sandbox.resize(Resources(cpu=2, memory=4, disk=30))
            ```
        """
        resize_request = ResizeSandbox(
            cpu=resources.cpu,
            memory=resources.memory,
            disk=resources.disk,
        )
        sandbox = self._sandbox_api.resize_sandbox(self.id, resize_request, _request_timeout=timeout or None)
        self.__process_sandbox_dto(sandbox)
        self.wait_for_resize_complete(timeout=0)

    @intercept_errors(message_prefix="Failure during waiting for resize to complete: ")
    @with_timeout()
    @with_instrumentation()
    def wait_for_resize_complete(
        self,
        timeout: float | None = 60,  # pylint: disable=unused-argument # pyright: ignore[reportUnusedParameter]
    ) -> None:
        """Waits for the Sandbox resize operation to complete via WebSocket events.

        Args:
            timeout (Optional[float]): Maximum time to wait in seconds. 0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative. If resize operation times out.
                If WebSocket connection is not available.
        """
        if self.state != "resizing":
            return

        if not self._event_subscriber or not self._event_subscriber.is_connected:
            raise DaytonaError("WebSocket connection not available for sandbox event subscription")

        self._wait_for_state_via_events(["started", "stopped", "archived"], ["error", "build_failed"])

    @intercept_errors(message_prefix="Failed to create SSH access: ")
    @with_instrumentation()
    def create_ssh_access(self, expires_in_minutes: int | None = None) -> SshAccessDto:
        """Creates an SSH access token for the sandbox.

        Args:
            expires_in_minutes (int | None): The number of minutes the SSH access token will be valid for.
        """
        return self._sandbox_api.create_ssh_access(self.id, expires_in_minutes=expires_in_minutes)

    @intercept_errors(message_prefix="Failed to revoke SSH access: ")
    @with_instrumentation()
    def revoke_ssh_access(self, token: str) -> None:
        """Revokes an SSH access token for the sandbox.

        Args:
            token (str): The token to revoke.
        """
        _ = self._sandbox_api.revoke_ssh_access(self.id, token)

    @intercept_errors(message_prefix="Failed to validate SSH access: ")
    @with_instrumentation()
    def validate_ssh_access(self, token: str) -> SshAccessValidationDto:
        """Validates an SSH access token for the sandbox.

        Args:
            token (str): The token to validate.
        """
        return self._sandbox_api.validate_ssh_access(token)

    @intercept_errors(message_prefix="Failed to refresh sandbox activity: ")
    def refresh_activity(self) -> None:
        """Refreshes the sandbox activity to reset the timer for automated lifecycle management actions.

        This method updates the sandbox's last activity timestamp without changing its state.
        It is useful for keeping long-running sessions alive while there is still user activity.

        Example:
            ```python
            sandbox.refresh_activity()
            ```
        """
        self._sandbox_api.update_last_activity(self.id)

    def _subscribe_to_events(self) -> None:
        """Subscribe to real-time events for this sandbox, auto-updating metadata."""
        if not self._event_subscriber or not self._event_subscriber.is_connected:
            return

        def _handle_event(event_type: str, data: Any) -> None:
            sandbox_data: dict[str, Any] | None = None
            if event_type in ("state.updated", "desired-state.updated") and isinstance(data, dict):
                raw: object = data.get("sandbox")  # pyright: ignore[reportUnknownVariableType]
                if isinstance(raw, dict):
                    sandbox_data = dict(raw)  # pyright: ignore[reportUnknownArgumentType]
            elif event_type == "created" and isinstance(data, dict):
                sandbox_data = dict(data)  # pyright: ignore[reportUnknownArgumentType]

            if sandbox_data is not None:
                self._update_from_event_data(sandbox_data)

        _ = self._event_subscriber.subscribe(self.id, _handle_event)

    def _update_from_event_data(self, data: dict[str, Any]) -> None:
        """Update sandbox fields from a camelCase event payload dict."""
        field_map = {
            "id": "id",
            "name": "name",
            "organizationId": "organization_id",
            "snapshot": "snapshot",
            "user": "user",
            "env": "env",
            "labels": "labels",
            "public": "public",
            "target": "target",
            "cpu": "cpu",
            "gpu": "gpu",
            "memory": "memory",
            "disk": "disk",
            "state": "state",
            "errorReason": "error_reason",
            "recoverable": "recoverable",
            "backupState": "backup_state",
            "backupCreatedAt": "backup_created_at",
            "autoStopInterval": "auto_stop_interval",
            "autoArchiveInterval": "auto_archive_interval",
            "autoDeleteInterval": "auto_delete_interval",
            "createdAt": "created_at",
            "updatedAt": "updated_at",
            "networkBlockAll": "network_block_all",
            "networkAllowList": "network_allow_list",
        }
        for camel_key, snake_key in field_map.items():
            if camel_key in data:
                setattr(self, snake_key, data[camel_key])

    _POLL_SAFETY_INTERVAL: float = 3.0

    def _wait_for_state_via_events(
        self, target_states: list[str], error_states: list[str]
    ) -> None:
        """Wait for sandbox to reach a target state via WebSocket events with periodic polling safety net."""
        if self.state in target_states:
            return
        if self.state in error_states:
            raise DaytonaError(
                f"Sandbox {self.id} is in error state: {self.state}, error reason: {self.error_reason}"
            )

        assert self._event_subscriber is not None

        event = threading.Event()
        result: dict[str, str] = {}

        def _handler(event_type: str, data: Any) -> None:
            if event_type != "state.updated":
                return
            raw_state: object = data.get("newState") if isinstance(data, dict) else None  # pyright: ignore[reportUnknownVariableType]
            new_state: str | None = str(raw_state) if isinstance(raw_state, str) else None
            if new_state and (new_state in target_states or new_state in error_states):
                result["state"] = new_state
                event.set()

        unsubscribe = self._event_subscriber.subscribe(self.id, _handler)
        try:
            # Poll periodically as safety net for missed WebSocket events
            while not event.is_set():
                try:
                    self.refresh_data()
                except Exception:
                    pass  # Poll failed, will retry on next interval
                if self.state in target_states:
                    return
                if self.state in error_states:
                    raise DaytonaError(
                        f"Sandbox {self.id} is in error state: {self.state}, error reason: {self.error_reason}"
                    )
                # Wait for event or poll interval, whichever comes first
                # (timeout is handled by outer @with_timeout decorator)
                _ = event.wait(timeout=self._POLL_SAFETY_INTERVAL)

            final_state: str | None = result.get("state")
            if final_state in error_states:
                raise DaytonaError(
                    f"Sandbox {self.id} entered error state: {final_state}, error reason: {self.error_reason}"
                )
        finally:
            unsubscribe()

    def __process_sandbox_dto(self, sandbox_dto: SandboxDto) -> None:
        self.id: str = sandbox_dto.id
        self.name: str = sandbox_dto.name
        self.organization_id: str = sandbox_dto.organization_id
        self.snapshot: str | None = sandbox_dto.snapshot
        self.user: str = sandbox_dto.user
        self.env: dict[str, str] = sandbox_dto.env
        self.labels: dict[str, str] = sandbox_dto.labels
        self.public: bool = sandbox_dto.public
        self.target: str = sandbox_dto.target
        self.cpu: float | int = sandbox_dto.cpu
        self.gpu: float | int = sandbox_dto.gpu
        self.memory: float | int = sandbox_dto.memory
        self.disk: float | int = sandbox_dto.disk
        self.state: SandboxState | None = sandbox_dto.state
        self.error_reason: str | None = sandbox_dto.error_reason
        self.recoverable: bool | None = sandbox_dto.recoverable
        self.backup_state: str | None = sandbox_dto.backup_state
        self.backup_created_at: str | None = sandbox_dto.backup_created_at
        self.auto_stop_interval: float | int | None = sandbox_dto.auto_stop_interval
        self.auto_archive_interval: float | int | None = sandbox_dto.auto_archive_interval
        self.auto_delete_interval: float | int | None = sandbox_dto.auto_delete_interval
        self.volumes: list[SandboxVolume] | None = sandbox_dto.volumes
        self.build_info: BuildInfo | None = sandbox_dto.build_info
        self.created_at: str | None = sandbox_dto.created_at
        self.updated_at: str | None = sandbox_dto.updated_at
        self.network_block_all: bool = sandbox_dto.network_block_all
        self.network_allow_list: str | None = sandbox_dto.network_allow_list
        self.toolbox_proxy_url: str = sandbox_dto.toolbox_proxy_url

    def __refresh_data_safe(self) -> None:
        """Refreshes the Sandbox data from the API, but does not throw an error if the sandbox has been deleted.
        Instead, it sets the state to destroyed.
        """
        try:
            self.refresh_data()
        except DaytonaNotFoundError:
            self.state = SandboxState.DESTROYED


class PaginatedSandboxes(PaginatedSandboxesDto):
    """Represents a paginated list of Daytona Sandboxes.

    Attributes:
        items (list[Sandbox]): List of Sandbox instances in the current page.
        total (int): Total number of Sandboxes across all pages.
        page (int): Current page number.
        total_pages (int): Total number of pages available.
    """

    items: list[Sandbox]  # pyright: ignore[reportIncompatibleVariableOverride]

    model_config: ConfigDict = ConfigDict(arbitrary_types_allowed=True)
