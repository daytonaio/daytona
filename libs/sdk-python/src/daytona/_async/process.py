# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import base64
import json
import re
from typing import Awaitable, Callable, Dict, List, Optional, Union

import websockets
from daytona_api_client_async import (
    Command,
    CreateSessionRequest,
    ExecuteRequest,
    PortPreviewUrl,
    PtyCreateRequest,
    PtyResizeRequest,
    PtySessionInfo,
    Session,
    ToolboxApi,
)
from websockets.asyncio.client import connect

from .._utils.errors import intercept_errors
from .._utils.stream import std_demux_stream
from ..code_toolbox.sandbox_python_code_toolbox import SandboxPythonCodeToolbox
from ..common.charts import parse_chart
from ..common.process import (
    CodeRunParams,
    ExecuteResponse,
    ExecutionArtifacts,
    SessionCommandLogsResponse,
    SessionExecuteRequest,
    SessionExecuteResponse,
    demux_log,
    parse_session_command_logs,
)
from ..common.pty import PtySize
from ..handle.async_pty_handle import AsyncPtyHandle


class AsyncProcess:
    """Handles process and code execution within a Sandbox."""

    def __init__(
        self,
        sandbox_id: str,
        code_toolbox: SandboxPythonCodeToolbox,
        toolbox_api: ToolboxApi,
        get_preview_link: Callable[[int], Awaitable[PortPreviewUrl]],
    ):
        """Initialize a new Process instance.

        Args:
            sandbox_id (str): The ID of the Sandbox.
            code_toolbox (SandboxPythonCodeToolbox): Language-specific code execution toolbox.
            toolbox_api (ToolboxApi): API client for Sandbox operations.
        """
        self._sandbox_id = sandbox_id
        self._code_toolbox = code_toolbox
        self._toolbox_api = toolbox_api
        self._get_preview_link = get_preview_link

    @staticmethod
    def _parse_output(lines: List[str]) -> Optional[ExecutionArtifacts]:
        """
        Parse the output of a command to extract ExecutionArtifacts.

        Args:
            lines: A list of lines of output from a command

        Returns:
            ExecutionArtifacts: The artifacts from the command execution
        """
        artifacts = ExecutionArtifacts("", [])
        for line in lines:
            if not line.startswith("dtn_artifact_k39fd2:"):
                artifacts.stdout += line
                artifacts.stdout += "\n"
            else:
                # Remove the prefix and parse JSON
                json_str = line.replace("dtn_artifact_k39fd2:", "", 1).strip()
                data = json.loads(json_str)
                data_type = data.pop("type")

                # Check if this is chart data
                if data_type == "chart":
                    chart_data = data.get("value", {})
                    artifacts.charts.append(parse_chart(**chart_data))

        return artifacts

    @intercept_errors(message_prefix="Failed to execute command: ")
    async def exec(
        self,
        command: str,
        cwd: Optional[str] = None,
        env: Optional[Dict[str, str]] = None,
        timeout: Optional[int] = None,
    ) -> ExecuteResponse:
        """Execute a shell command in the Sandbox.

        Args:
            command (str): Shell command to execute.
            cwd (Optional[str]): Working directory for command execution. If not
                specified, uses the sandbox working directory.
            env (Optional[Dict[str, str]]): Environment variables to set for the command.
            timeout (Optional[int]): Maximum time in seconds to wait for the command
                to complete. 0 means wait indefinitely.

        Returns:
            ExecuteResponse: Command execution results containing:
                - exit_code: The command's exit status
                - result: Standard output from the command
                - artifacts: ExecutionArtifacts object containing `stdout` (same as result)
                and `charts` (matplotlib charts metadata)

        Example:
            ```python
            # Simple command
            response = await sandbox.process.exec("echo 'Hello'")
            print(response.artifacts.stdout)  # Prints: Hello

            # Command with working directory
            result = await sandbox.process.exec("ls", cwd="workspace/src")

            # Command with timeout
            result = await sandbox.process.exec("sleep 10", timeout=5)
            ```
        """
        base64_user_cmd = base64.b64encode(command.encode()).decode()
        command = f"echo '{base64_user_cmd}' | base64 -d | sh"

        if env and len(env.items()) > 0:
            safe_env_exports = (
                ";".join(
                    [
                        f"export {key}=$(echo '{base64.b64encode(value.encode()).decode()}' | base64 -d)"
                        for key, value in env.items()
                    ]
                )
                + ";"
            )
            command = f"{safe_env_exports} {command}"

        command = f'sh -c "{command}"'
        execute_request = ExecuteRequest(command=command, cwd=cwd, timeout=timeout)

        response = await self._toolbox_api.execute_command(sandbox_id=self._sandbox_id, execute_request=execute_request)

        # Post-process the output to extract ExecutionArtifacts
        artifacts = AsyncProcess._parse_output(response.result.splitlines())

        # Create new response with processed output and charts
        # TODO: Remove model_construct once everything is migrated to pydantic # pylint: disable=fixme
        return ExecuteResponse.model_construct(
            exit_code=response.exit_code,
            result=artifacts.stdout,
            artifacts=artifacts,
            additional_properties=response.additional_properties,
        )

    async def code_run(
        self,
        code: str,
        params: Optional[CodeRunParams] = None,
        timeout: Optional[int] = None,
    ) -> ExecuteResponse:
        """Executes code in the Sandbox using the appropriate language runtime.

        Args:
            code (str): Code to execute.
            params (Optional[CodeRunParams]): Parameters for code execution.
            timeout (Optional[int]): Maximum time in seconds to wait for the code
                to complete. 0 means wait indefinitely.

        Returns:
            ExecuteResponse: Code execution result containing:
                - exit_code: The execution's exit status
                - result: Standard output from the code
                - artifacts: ExecutionArtifacts object containing `stdout` (same as result)
                and `charts` (matplotlib charts metadata)

        Example:
            ```python
            # Run Python code
            response = await sandbox.process.code_run('''
                x = 10
                y = 20
                print(f"Sum: {x + y}")
            ''')
            print(response.artifacts.stdout)  # Prints: Sum: 30
            ```

            Matplotlib charts are automatically detected and returned in the `charts` field
            of the `ExecutionArtifacts` object.
            ```python
            code = '''
            import matplotlib.pyplot as plt
            import numpy as np

            x = np.linspace(0, 10, 30)
            y = np.sin(x)

            plt.figure(figsize=(8, 5))
            plt.plot(x, y, 'b-', linewidth=2)
            plt.title('Line Chart')
            plt.xlabel('X-axis (seconds)')
            plt.ylabel('Y-axis (amplitude)')
            plt.grid(True)
            plt.show()
            '''

            response = await sandbox.process.code_run(code)
            chart = response.artifacts.charts[0]

            print(f"Type: {chart.type}")
            print(f"Title: {chart.title}")
            if chart.type == ChartType.LINE and isinstance(chart, LineChart):
                print(f"X Label: {chart.x_label}")
                print(f"Y Label: {chart.y_label}")
                print(f"X Ticks: {chart.x_ticks}")
                print(f"X Tick Labels: {chart.x_tick_labels}")
                print(f"X Scale: {chart.x_scale}")
                print(f"Y Ticks: {chart.y_ticks}")
                print(f"Y Tick Labels: {chart.y_tick_labels}")
                print(f"Y Scale: {chart.y_scale}")
                print("Elements:")
                for element in chart.elements:
                    print(f"Label: {element.label}")
                    print(f"Points: {element.points}")
            ```
        """
        command = self._code_toolbox.get_run_command(code, params)
        return await self.exec(command, env=params.env if params else None, timeout=timeout)

    @intercept_errors(message_prefix="Failed to create session: ")
    async def create_session(self, session_id: str) -> None:
        """Creates a new long-running background session in the Sandbox.

        Sessions are background processes that maintain state between commands, making them ideal for
        scenarios requiring multiple related commands or persistent environment setup. You can run
        long-running commands and monitor process status.

        Args:
            session_id (str): Unique identifier for the new session.

        Example:
            ```python
            # Create a new session
            session_id = "my-session"
            await sandbox.process.create_session(session_id)
            session = await sandbox.process.get_session(session_id)
            # Do work...
            await sandbox.process.delete_session(session_id)
            ```
        """
        request = CreateSessionRequest(sessionId=session_id)
        await self._toolbox_api.create_session(self._sandbox_id, create_session_request=request)

    @intercept_errors(message_prefix="Failed to get session: ")
    async def get_session(self, session_id: str) -> Session:
        """Gets a session in the Sandbox.

        Args:
            session_id (str): Unique identifier of the session to retrieve.

        Returns:
            Session: Session information including:
                - session_id: The session's unique identifier
                - commands: List of commands executed in the session

        Example:
            ```python
            session = await sandbox.process.get_session("my-session")
            for cmd in session.commands:
                print(f"Command: {cmd.command}")
            ```
        """
        return await self._toolbox_api.get_session(self._sandbox_id, session_id=session_id)

    @intercept_errors(message_prefix="Failed to get session command: ")
    async def get_session_command(self, session_id: str, command_id: str) -> Command:
        """Gets information about a specific command executed in a session.

        Args:
            session_id (str): Unique identifier of the session.
            command_id (str): Unique identifier of the command.

        Returns:
            Command: Command information including:
                - id: The command's unique identifier
                - command: The executed command string
                - exit_code: Command's exit status (if completed)

        Example:
            ```python
            cmd = await sandbox.process.get_session_command("my-session", "cmd-123")
            if cmd.exit_code == 0:
                print(f"Command {cmd.command} completed successfully")
            ```
        """
        return await self._toolbox_api.get_session_command(
            self._sandbox_id, session_id=session_id, command_id=command_id
        )

    @intercept_errors(message_prefix="Failed to execute session command: ")
    async def execute_session_command(
        self,
        session_id: str,
        req: SessionExecuteRequest,
        timeout: Optional[int] = None,
    ) -> SessionExecuteResponse:
        """Executes a command in the session.

        Args:
            session_id (str): Unique identifier of the session to use.
            req (SessionExecuteRequest): Command execution request containing:
                - command: The command to execute
                - run_async: Whether to execute asynchronously

        Returns:
            SessionExecuteResponse: Command execution results containing:
                - cmd_id: Unique identifier for the executed command
                - output: Combined command output (stdout and stderr) (if synchronous execution)
                - stdout: Standard output from the command
                - stderr: Standard error from the command
                - exit_code: Command exit status (if synchronous execution)

        Example:
            ```python
            # Execute commands in sequence, maintaining state
            session_id = "my-session"

            # Change directory
            req = SessionExecuteRequest(command="cd /workspace")
            await sandbox.process.execute_session_command(session_id, req)

            # Create a file
            req = SessionExecuteRequest(command="echo 'Hello' > test.txt")
            await sandbox.process.execute_session_command(session_id, req)

            # Read the file
            req = SessionExecuteRequest(command="cat test.txt")
            result = await sandbox.process.execute_session_command(session_id, req)
            print(f"Command stdout: {result.stdout}")
            print(f"Command stderr: {result.stderr}")
            ```
        """
        response = await self._toolbox_api.execute_session_command(
            self._sandbox_id,
            session_id=session_id,
            session_execute_request=req,
            _request_timeout=timeout or None,
        )

        stdout, stderr = demux_log(response.output.encode("utf-8", "ignore") if response.output else b"")

        return SessionExecuteResponse.model_construct(
            cmd_id=response.cmd_id,
            output=response.output,
            stdout=stdout.decode("utf-8", "ignore"),
            stderr=stderr.decode("utf-8", "ignore"),
            exit_code=response.exit_code,
            additional_properties=response.additional_properties,
        )

    @intercept_errors(message_prefix="Failed to get session command logs: ")
    async def get_session_command_logs(self, session_id: str, command_id: str) -> SessionCommandLogsResponse:
        """Get the logs for a command executed in a session.

        Args:
            session_id (str): Unique identifier of the session.
            command_id (str): Unique identifier of the command.

        Returns:
            SessionCommandLogsResponse: Command logs including:
                - output: Combined command output (stdout and stderr)
                - stdout: Standard output from the command
                - stderr: Standard error from the command

        Example:
            ```python
            logs = await sandbox.process.get_session_command_logs(
                "my-session",
                "cmd-123"
            )
            print(f"Command stdout: {logs.stdout}")
            print(f"Command stderr: {logs.stderr}")
            ```
        """
        response = await self._toolbox_api.get_session_command_logs_without_preload_content(
            self._sandbox_id, session_id=session_id, command_id=command_id
        )

        # unasync: delete start
        response.data = await response.content.read()
        # unasync: delete end

        return parse_session_command_logs(response.data)

    # unasync: preserve start
    @intercept_errors(message_prefix="Failed to get session command logs: ")
    async def get_session_command_logs_async(
        self, session_id: str, command_id: str, on_stdout: Callable[[str], None], on_stderr: Callable[[str], None]
    ) -> None:
        """Asynchronously retrieves and processes the logs for a command executed in a session as they become available.

        Args:
            session_id (str): Unique identifier of the session.
            command_id (str): Unique identifier of the command.
            on_stdout (Callable[[str], None]): Callback function to handle stdout log chunks as they arrive.
            on_stderr (Callable[[str], None]): Callback function to handle stderr log chunks as they arrive.

        Example:
            ```python
            await sandbox.process.get_session_command_logs_async(
                "my-session",
                "cmd-123",
                lambda log: print(f"[STDOUT]: {log}"),
                lambda log: print(f"[STDERR]: {log}"),
            )
            ```
        """
        _, url, headers, *_ = self._toolbox_api._get_session_command_logs_serialize(  # pylint: disable=protected-access
            sandbox_id=self._sandbox_id,
            session_id=session_id,
            command_id=command_id,
            x_daytona_organization_id=None,
            follow=True,
            _request_auth=None,
            _content_type=None,
            _headers=None,
            _host_index=None,
        )
        # unasync: preserve end

        preview_link = await self._get_preview_link(2280)
        url = re.sub(r"^http", "ws", preview_link.url) + url[url.index("/process") :]

        # unasync: preserve start
        async with websockets.connect(
            url,
            additional_headers={
                **headers,
                "X-Daytona-Preview-Token": preview_link.token,
            },
        ) as ws:
            await std_demux_stream(ws, on_stdout, on_stderr)

    # unasync: preserve end

    @intercept_errors(message_prefix="Failed to list sessions: ")
    async def list_sessions(self) -> List[Session]:
        """Lists all sessions in the Sandbox.

        Returns:
            List[Session]: List of all sessions in the Sandbox.

        Example:
            ```python
            sessions = await sandbox.process.list_sessions()
            for session in sessions:
                print(f"Session {session.session_id}:")
                print(f"  Commands: {len(session.commands)}")
            ```
        """
        return await self._toolbox_api.list_sessions(self._sandbox_id)

    @intercept_errors(message_prefix="Failed to delete session: ")
    async def delete_session(self, session_id: str) -> None:
        """Terminates and removes a session from the Sandbox, cleaning up any resources
        associated with it.

        Args:
            session_id (str): Unique identifier of the session to delete.

        Example:
            ```python
            # Create and use a session
            await sandbox.process.create_session("temp-session")
            # ... use the session ...

            # Clean up when done
            await sandbox.process.delete_session("temp-session")
            ```
        """
        await self._toolbox_api.delete_session(self._sandbox_id, session_id=session_id)

    @intercept_errors(message_prefix="Failed to create PTY session: ")
    async def create_pty_session(
        self,
        id: str,
        # unasync: delete start
        on_data: Union[Callable[[bytes], None], Callable[[bytes], Awaitable[None]]] = None,
        # unasync: delete end
        cwd: Optional[str] = None,
        envs: Optional[Dict[str, str]] = None,
        pty_size: Optional[PtySize] = None,
    ) -> AsyncPtyHandle:
        """Creates a new PTY (pseudo-terminal) session in the Sandbox.

        Creates an interactive terminal session that can execute commands and handle user input.
        The PTY session behaves like a real terminal, supporting features like command history.

        Args:
            id: Unique identifier for the PTY session. Must be unique within the Sandbox.
            cwd: Working directory for the PTY session. Defaults to the sandbox's working directory.
            env: Environment variables to set in the PTY session. These will be merged with
                the Sandbox's default environment variables.
            pty_size: Terminal size configuration. Defaults to 80x24 if not specified.

        Returns:
            AsyncPtyHandle: Handle for managing the created PTY session. Use this to send input,
                           receive output, resize the terminal, and manage the session lifecycle.

        Raises:
            DaytonaError: If the PTY session creation fails or the session ID is already in use.
        """
        response = await self._toolbox_api.create_pty_session(
            self._sandbox_id,
            pty_create_request=PtyCreateRequest(
                id=id,
                cwd=cwd,
                envs=envs,
                cols=pty_size.cols if pty_size else None,
                rows=pty_size.rows if pty_size else None,
                lazy_start=True,
            ),
        )

        return await self.connect_pty_session(
            response.session_id,
            # unasync: delete start
            on_data,
            # unasync: delete end
        )

    @intercept_errors(message_prefix="Failed to connect PTY session: ")
    async def connect_pty_session(
        self,
        session_id: str,
        # unasync: delete start
        on_data: Union[Callable[[bytes], None], Callable[[bytes], Awaitable[None]]],
        # unasync: delete end
    ) -> AsyncPtyHandle:
        """Connects to an existing PTY session in the Sandbox.

        Establishes a WebSocket connection to an existing PTY session, allowing you to
        interact with a previously created terminal session.

        Args:
            session_id: Unique identifier of the PTY session to connect to.

        Returns:
            AsyncPtyHandle: Handle for managing the connected PTY session.

        Raises:
            DaytonaError: If the PTY session doesn't exist or connection fails.
        """
        _, url, headers, *_ = self._toolbox_api._connect_pty_session_serialize(  # pylint: disable=protected-access
            sandbox_id=self._sandbox_id,
            session_id=session_id,
            x_daytona_organization_id=None,
            _request_auth=None,
            _content_type=None,
            _headers=None,
            _host_index=None,
        )

        preview_link = await self._get_preview_link(2280)
        url = re.sub(r"^http", "ws", preview_link.url) + url[url.index("/process") :]

        ws = await connect(
            url,
            additional_headers={
                **headers,
                "X-Daytona-Preview-Token": preview_link.token,
            },
        )

        # Create resize and kill handlers
        async def resize_handler(pty_size: PtySize) -> PtySessionInfo:
            return await self.resize_pty_session(session_id, pty_size)

        async def kill_handler() -> None:
            await self.kill_pty_session(session_id)

        handle = AsyncPtyHandle(
            ws,
            # unasync: delete start
            on_data,
            # unasync: delete end
            session_id=session_id,
            handle_resize=resize_handler,
            handle_kill=kill_handler,
        )
        await handle.wait_for_connection()
        return handle

    @intercept_errors(message_prefix="Failed to list PTY sessions: ")
    async def list_pty_sessions(self) -> List[PtySessionInfo]:
        """Lists all PTY sessions in the Sandbox.

        Retrieves information about all PTY sessions in this Sandbox.

        Returns:
            List[PtySessionInfo]: List of PTY session information objects containing
                                details about each session's state, creation time, and configuration.

        Example:
            ```python
            # List all PTY sessions
            sessions = await sandbox.process.list_pty_sessions()

            for session in sessions:
                print(f"Session ID: {session.id}")
                print(f"Active: {session.active}")
                print(f"Created: {session.created_at}")
            ```
        """
        return await self._toolbox_api.list_pty_sessions(self._sandbox_id)

    @intercept_errors(message_prefix="Failed to get PTY session info: ")
    async def get_pty_session_info(self, session_id: str) -> PtySessionInfo:
        """Gets detailed information about a specific PTY session.

        Retrieves comprehensive information about a PTY session including its current state,
        configuration, and metadata.

        Args:
            session_id: Unique identifier of the PTY session to retrieve information for.

        Returns:
            PtySessionInfo: Detailed information about the PTY session including ID, state,
                           creation time, working directory, environment variables, and more.

        Raises:
            DaytonaError: If the PTY session doesn't exist.

        Example:
            ```python
            # Get details about a specific PTY session
            session_info = await sandbox.process.get_pty_session_info("my-session")

            print(f"Session ID: {session_info.id}")
            print(f"Active: {session_info.active}")
            print(f"Working Directory: {session_info.cwd}")
            print(f"Terminal Size: {session_info.cols}x{session_info.rows}")
            ```
        """
        return await self._toolbox_api.get_pty_session(self._sandbox_id, session_id=session_id)

    @intercept_errors(message_prefix="Failed to kill PTY session: ")
    async def kill_pty_session(self, session_id: str) -> None:
        """Kills a PTY session and terminates its associated process.

        Forcefully terminates the PTY session and cleans up all associated resources.
        This will close any active connections and kill the underlying shell process.
        This operation is irreversible. Any unsaved work in the terminal session will be lost.

        Args:
            session_id: Unique identifier of the PTY session to kill.

        Raises:
            DaytonaError: If the PTY session doesn't exist or cannot be killed.

        Example:
            ```python
            # Kill a specific PTY session
            await sandbox.process.kill_pty_session("my-session")

            # Verify the session no longer exists
            pty_sessions = await sandbox.process.list_pty_sessions()
            for pty_session in pty_sessions:
                print(f"PTY session: {pty_session.id}")
            ```
        """
        await self._toolbox_api.delete_pty_session(self._sandbox_id, session_id=session_id)

    @intercept_errors(message_prefix="Failed to resize PTY session: ")
    async def resize_pty_session(self, session_id: str, pty_size: PtySize) -> PtySessionInfo:
        """Resizes a PTY session's terminal dimensions.

        Changes the terminal size of an active PTY session. This is useful when the
        client terminal is resized or when you need to adjust the display for different
        output requirements.

        Args:
            session_id: Unique identifier of the PTY session to resize.
            pty_size: New terminal dimensions containing the desired columns and rows.

        Returns:
            PtySessionInfo: Updated session information reflecting the new terminal size.

        Raises:
            DaytonaError: If the PTY session doesn't exist or resize operation fails.

        Example:
            ```python
            from daytona.common.pty import PtySize

            # Resize a PTY session to a larger terminal
            new_size = PtySize(rows=40, cols=150)
            updated_info = await sandbox.process.resize_pty_session("my-session", new_size)

            print(f"Terminal resized to {updated_info.cols}x{updated_info.rows}")

            # You can also use the AsyncPtyHandle's resize method
            await pty_handle.resize(new_size)
            ```
        """
        return await self._toolbox_api.resize_pty_session(
            self._sandbox_id,
            session_id=session_id,
            pty_resize_request=PtyResizeRequest(cols=pty_size.cols, rows=pty_size.rows),
        )
