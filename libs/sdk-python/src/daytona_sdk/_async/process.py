# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

import base64
import json
from typing import Awaitable, Callable, Dict, List, Optional

from daytona_api_client_async import (
    Command,
    CreateSessionRequest,
    ExecuteRequest,
    Session,
    SessionExecuteResponse,
    ToolboxApi,
)
from daytona_sdk._utils.errors import intercept_errors
from daytona_sdk._utils.stream import process_streaming_response
from daytona_sdk.code_toolbox.sandbox_python_code_toolbox import SandboxPythonCodeToolbox
from daytona_sdk.common.charts import parse_chart
from daytona_sdk.common.process import CodeRunParams, ExecuteResponse, ExecutionArtifacts, SessionExecuteRequest
from daytona_sdk.common.protocols import SandboxInstance


class AsyncProcess:
    """Handles process and code execution within a Sandbox.

    Attributes:
        code_toolbox (SandboxPythonCodeToolbox): Language-specific code execution toolbox.
        toolbox_api (ToolboxApi): API client for Sandbox operations.
        instance (SandboxInstance): The Sandbox instance this process belongs to.
    """

    def __init__(
        self,
        code_toolbox: SandboxPythonCodeToolbox,
        toolbox_api: ToolboxApi,
        instance: SandboxInstance,
        get_root_dir: Callable[[], Awaitable[str]],
    ):
        """Initialize a new Process instance.

        Args:
            code_toolbox (SandboxPythonCodeToolbox): Language-specific code execution toolbox.
            toolbox_api (ToolboxApi): API client for Sandbox operations.
            instance (SandboxInstance): The Sandbox instance this process belongs to.
            get_root_dir (Callable[[], str]): A function to get the default root directory of the Sandbox.
        """
        self.code_toolbox = code_toolbox
        self.toolbox_api = toolbox_api
        self.instance = instance
        self._get_root_dir = get_root_dir

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
                specified, uses the Sandbox root directory. Default is the user's root directory.
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
        execute_request = ExecuteRequest(command=command, cwd=cwd or await self._get_root_dir(), timeout=timeout)

        response = await self.toolbox_api.execute_command(
            workspace_id=self.instance.id, execute_request=execute_request
        )

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
                    print(f"\n\tLabel: {element.label}")
                    print(f"\tPoints: {element.points}")
            ```
        """
        command = self.code_toolbox.get_run_command(code, params)
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
        await self.toolbox_api.create_session(self.instance.id, create_session_request=request)

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
        return await self.toolbox_api.get_session(self.instance.id, session_id=session_id)

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
        return await self.toolbox_api.get_session_command(
            self.instance.id, session_id=session_id, command_id=command_id
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
                - output: Command output (if synchronous execution)
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
            print(result.output)  # Prints: Hello
            ```
        """
        return await self.toolbox_api.execute_session_command(
            self.instance.id,
            session_id=session_id,
            session_execute_request=req,
            _request_timeout=timeout or None,
        )

    @intercept_errors(message_prefix="Failed to get session command logs: ")
    async def get_session_command_logs(self, session_id: str, command_id: str) -> str:
        """Get the logs for a command executed in a session. Retrieves the complete output
        (stdout and stderr) from a command executed in a session.

        Args:
            session_id (str): Unique identifier of the session.
            command_id (str): Unique identifier of the command.

        Returns:
            str: Complete command output including both stdout and stderr.

        Example:
            ```python
            logs = await sandbox.process.get_session_command_logs(
                "my-session",
                "cmd-123"
            )
            print(f"Command output: {logs}")
            ```
        """
        return await self.toolbox_api.get_session_command_logs(
            self.instance.id, session_id=session_id, command_id=command_id
        )

    # unasync: preserve start
    @intercept_errors(message_prefix="Failed to get session command logs: ")
    async def get_session_command_logs_async(
        self, session_id: str, command_id: str, on_logs: Callable[[str], None]
    ) -> None:
        """Asynchronously retrieves and processes the logs for a command executed in a session as they become available.

        Args:
            session_id (str): Unique identifier of the session.
            command_id (str): Unique identifier of the command.
            on_logs (Callable[[str], None]): Callback function to handle log chunks as they arrive.

        Example:
            ```python
            await sandbox.process.get_session_command_logs_async(
                "my-session",
                "cmd-123",
                lambda chunk: print(f"Log chunk: {chunk}")
            )
            ```
        """
        _, url, *_ = self.toolbox_api._get_session_command_logs_serialize(  # pylint: disable=protected-access
            workspace_id=self.instance.id,
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
        async def should_terminate():
            return (await self.get_session_command(session_id, command_id)).exit_code is not None

        # unasync: preserve start

        await process_streaming_response(
            url=url,
            headers=self.toolbox_api.api_client.default_headers,
            on_chunk=on_logs,
            should_terminate=should_terminate,
        )

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
        return await self.toolbox_api.list_sessions(self.instance.id)

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
        await self.toolbox_api.delete_session(self.instance.id, session_id=session_id)
