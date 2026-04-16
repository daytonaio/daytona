# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import functools
import threading
from typing import Any, Callable

from deprecated import deprecated
from pydantic import ConfigDict, PrivateAttr

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

from .._utils.errors import intercept_errors
from .._utils.otel_decorator import with_instrumentation
from .._utils.timeout import http_timeout, with_timeout
from ..common.errors import DaytonaError, DaytonaNotFoundError, DaytonaValidationError
from ..common.lsp_server import LspLanguageId, LspLanguageIdLiteral
from ..common.protocols import SandboxCodeToolbox
from ..common.sandbox import Resources
from ..internal.event_subscription_manager import SyncEventSubscriptionManager
from ..internal.toolbox_api_client_proxy import ToolboxApiClientProxy
from .code_interpreter import CodeInterpreter
from .computer_use import ComputerUse
from .filesystem import FileSystem
from .git import Git
from .lsp_server import LspServer
from .process import Process



def with_events(cls: type) -> type:
    for name in list(vars(cls)):
        if name.startswith("_"):
            continue
        method = vars(cls)[name]
        if not callable(method):
            continue

        @functools.wraps(method)
        def wrapper(self: Any, *args: Any, _m: Any = method, **kwargs: Any) -> Any:
            if getattr(self, "__pydantic_private__", None) is not None:
                self._ensure_subscribed()
            return _m(self, *args, **kwargs)

        setattr(cls, name, wrapper)
    return cls


@with_events
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
    _state_waiters: list[Callable[[SandboxState | None], None]] = PrivateAttr(default_factory=list)
    _state_waiters_lock: threading.Lock = PrivateAttr(default_factory=threading.Lock)
    _sub_id: str | None = PrivateAttr(default=None)

    # TODO: Remove model_config once everything is migrated to pydantic # pylint: disable=fixme
    model_config: ConfigDict = ConfigDict(arbitrary_types_allowed=True)

    def __init__(
        self,
        sandbox_dto: SandboxDto,
        toolbox_api: ApiClient,
        sandbox_api: SandboxApi,
        code_toolbox: SandboxCodeToolbox,
        subscription_manager: SyncEventSubscriptionManager,
    ):
        """Initialize a new Sandbox instance.

        Args:
            sandbox_dto (SandboxDto): The sandbox data from the API.
            toolbox_api (ApiClient): API client for toolbox operations.
            sandbox_api (SandboxApi): API client for Sandbox operations.
            code_toolbox (SandboxCodeToolbox): Language-specific toolbox implementation.
            subscription_manager: SyncEventSubscriptionManager for real-time updates.
        """
        super().__init__(**sandbox_dto.model_dump())
        self.__process_sandbox_dto(sandbox_dto)
        self._sandbox_api: SandboxApi = sandbox_api
        self._code_toolbox: SandboxCodeToolbox = code_toolbox
        self._subscription_manager: SyncEventSubscriptionManager = subscription_manager
        if not self.toolbox_proxy_url:
            proxy_url = self._sandbox_api.get_toolbox_proxy_url(self.id)
            self.toolbox_proxy_url = proxy_url.url
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

        self._ensure_subscribed()

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
    def stop(self, timeout: float | None = 60, force: bool = False):
        """Stops the Sandbox and waits for it to be fully stopped.

        Args:
            timeout (float | None): Maximum time to wait in seconds. 0 means no timeout. Default is 60 seconds.
            force (bool): If True, uses SIGKILL instead of SIGTERM to stop the sandbox. Default is False.

        Raises:
            DaytonaError: If timeout is negative; If sandbox fails to stop or times out

        Example:
            ```python
            sandbox = daytona.get("my-sandbox-id")
            sandbox.stop()
            print("Sandbox stopped successfully")
            ```
        """
        _ = self._sandbox_api.stop_sandbox(
            self.id,
            force=force,  # pyright: ignore[reportCallIssue]
            _request_timeout=http_timeout(timeout),
        )
        self.__refresh_data_safe()
        # This method already handles a timeout, so we don't need to pass one to internal methods
        self.wait_for_sandbox_stop(timeout=0)


    @intercept_errors(message_prefix="Failed to remove sandbox: ")
    @with_timeout()
    @with_instrumentation()
    def delete(
        self,
        timeout: float | None = 60,  # pylint: disable=unused-argument
    ) -> None:
        """Deletes the Sandbox and waits for it to reach the 'destroyed' state.

        Args:
            timeout (float | None): Timeout (in seconds) for sandbox deletion. 0 means no timeout.
                Default is 60 seconds.
        """
        sandbox = self._sandbox_api.delete_sandbox(self.id, _request_timeout=http_timeout(timeout))
        self.__process_sandbox_dto(sandbox)

        try:
            if self.state != SandboxState.DESTROYED:
                self._wait_for_state(
                    [SandboxState.DESTROYED],
                    [SandboxState.ERROR, SandboxState.BUILD_FAILED],
                    safe_refresh=True,
                )
        finally:
            self._unsubscribe_from_events()


    @intercept_errors(message_prefix="Failure during waiting for sandbox to start: ")
    @with_timeout()
    @with_instrumentation()
    def wait_for_sandbox_start(
        self,
        timeout: float | None = 60,  # pylint: disable=unused-argument # pyright: ignore[reportUnusedParameter]
    ) -> None:
        """Waits for the Sandbox to reach the 'started' state.

        Args:
            timeout (float | None): Maximum time to wait in seconds. 0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative; If Sandbox fails to start or times out;
        """
        if self.state == SandboxState.STARTED:
            return

        self._wait_for_state(
            [SandboxState.STARTED],
            [SandboxState.ERROR, SandboxState.BUILD_FAILED],
        )


    @intercept_errors(message_prefix="Failure during waiting for sandbox to stop: ")
    @with_timeout()
    @with_instrumentation()
    def wait_for_sandbox_stop(
        self,
        timeout: float | None = 60,  # pylint: disable=unused-argument # pyright: ignore[reportUnusedParameter]
    ) -> None:
        """Waits for the Sandbox to reach the 'stopped' state.
        Treats destroyed as stopped to cover ephemeral sandboxes that are automatically deleted after stopping.

        Args:
            timeout (float | None): Maximum time to wait in seconds. 0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative. If Sandbox fails to stop or times out.
        """
        if self.state in [SandboxState.STOPPED, SandboxState.DESTROYED]:
            return

        self._wait_for_state(
            [SandboxState.STOPPED, SandboxState.DESTROYED],
            [SandboxState.ERROR, SandboxState.BUILD_FAILED],
        )


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
            DaytonaValidationError: If interval is negative

        Example:
            ```python
            # Auto-stop after 1 hour
            sandbox.set_autostop_interval(60)
            # Or disable auto-stop
            sandbox.set_autostop_interval(0)
            ```
        """
        if interval < 0:
            raise DaytonaValidationError("Auto-stop interval must be a non-negative integer")

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
            DaytonaValidationError: If interval is negative

        Example:
            ```python
            # Auto-archive after 1 hour
            sandbox.set_auto_archive_interval(60)
            # Or use the maximum interval
            sandbox.set_auto_archive_interval(0)
            ```
        """
        if interval < 0:
            raise DaytonaValidationError("Auto-archive interval must be a non-negative integer")

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
        """Waits for the Sandbox resize operation to complete.

        Args:
            timeout (Optional[float]): Maximum time to wait in seconds. 0 means no timeout. Default is 60 seconds.

        Raises:
            DaytonaError: If timeout is negative. If resize operation times out.
        """
        if self.state != SandboxState.RESIZING:
            return

        error_states = [SandboxState.ERROR, SandboxState.BUILD_FAILED]
        exclude = {SandboxState.RESIZING} | set(error_states)
        target_states = [s for s in SandboxState if s not in exclude]

        self._wait_for_state(target_states, error_states)


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

    def _ensure_subscribed(self) -> None:
        with self._state_waiters_lock:
            if self._sub_id is not None:
                if self._subscription_manager.refresh(self._sub_id):
                    return
                self._sub_id = None

            self._sub_id = self._subscription_manager.subscribe(
                self.id,
                self._handle_event,
                events=["sandbox.state.updated", "sandbox.created"],
            )

    def _handle_event(self, event_name: str, data: Any) -> None:
        if not isinstance(data, dict):
            return
        raw: object = data.get("sandbox", data)  # pyright: ignore[reportUnknownVariableType]

        if event_name == "sandbox.created":
            sandbox_dto = SandboxDto.from_dict(raw)  # pyright: ignore[reportArgumentType]
            if sandbox_dto is not None:
                self.__process_sandbox_dto(sandbox_dto)
        else:
            new_state = (  # pyright: ignore[reportUnknownVariableType]
                raw.get("state") if isinstance(raw, dict) else None
            ) or data.get("newState")
            if new_state is not None:
                try:
                    self._apply_state(SandboxState(new_state))
                except ValueError:
                    pass

    def _unsubscribe_from_events(self) -> None:
        with self._state_waiters_lock:
            if self._sub_id is not None:
                self._subscription_manager.unsubscribe(self._sub_id)
                self._sub_id = None

    def _apply_state(self, new_state: SandboxState | None) -> None:
        if new_state == self.state:
            return

        self.state: SandboxState | None = new_state

        with self._state_waiters_lock:
            for waiter in list(self._state_waiters):
                waiter(new_state)

    def _wait_for_state(
        self,
        target_states: list[SandboxState],
        error_states: list[SandboxState],
        safe_refresh: bool = False,
    ) -> None:
        """Wait for sandbox to reach a target state via WebSocket events with periodic polling safety net.

        Args:
            target_states: States that indicate success.
            error_states: States that indicate failure.
            safe_refresh: If True, use safe refresh that treats 404 as destroyed (for delete operations).
        """
        self._ensure_subscribed()

        if self.state in target_states:
            return
        if self.state in error_states:
            raise DaytonaError(f"Sandbox {self.id} is in error state: {self.state}, error reason: {self.error_reason}")

        state_resolved = threading.Event()
        resolve_lock = threading.Lock()
        result_state: SandboxState | None = None

        def _waiter(state: SandboxState | None) -> None:
            nonlocal result_state
            if state is None or (state not in target_states and state not in error_states):
                return

            with resolve_lock:
                if state_resolved.is_set():
                    return

                result_state = state
                state_resolved.set()

        with self._state_waiters_lock:
            self._state_waiters.append(_waiter)
        try:
            _waiter(self.state)

            while not state_resolved.is_set():
                # Wait for event or poll interval, whichever comes first
                # (timeout is handled by outer @with_timeout decorator)
                is_set = state_resolved.wait(timeout=1)

                if is_set:
                    break

                if safe_refresh:
                    self.__refresh_data_safe()
                else:
                    self.refresh_data()

            if result_state in error_states:
                raise DaytonaError(
                    f"Sandbox {self.id} entered error state: {result_state}, error reason: {self.error_reason}"
                )
        finally:
            with self._state_waiters_lock:
                if _waiter in self._state_waiters:
                    self._state_waiters.remove(_waiter)

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
        new_proxy_url = sandbox_dto.toolbox_proxy_url
        if new_proxy_url and new_proxy_url != self.toolbox_proxy_url and hasattr(self, "_toolbox_api"):
            self._toolbox_api._toolbox_base_url = new_proxy_url
        self.toolbox_proxy_url: str = new_proxy_url
        self._apply_state(sandbox_dto.state)

    def __refresh_data_safe(self) -> None:
        """Refreshes the Sandbox data from the API, but does not throw an error if the sandbox has been deleted.
        Instead, it sets the state to destroyed.
        """
        try:
            self.refresh_data()
        except DaytonaNotFoundError:
            self._apply_state(SandboxState.DESTROYED)


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
