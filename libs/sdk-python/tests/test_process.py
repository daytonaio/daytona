from __future__ import annotations

import json
from unittest.mock import AsyncMock, MagicMock

import pytest

from daytona.common.errors import DaytonaError
from daytona.common.process import CodeRunParams, ExecutionArtifacts


class TestProcessParseOutput:
    def test_plain_output(self):
        from daytona._sync.process import Process

        lines = ["hello", "world"]
        artifacts = Process._parse_output(lines)
        assert artifacts.stdout == "hello\nworld"
        assert artifacts.charts == []

    def test_empty_output(self):
        from daytona._sync.process import Process

        artifacts = Process._parse_output([])
        assert artifacts.stdout == ""

    def test_artifact_lines_filtered(self):
        from daytona._sync.process import Process

        chart_data = {
            "type": "chart",
            "value": {"type": "line", "title": "Test", "elements": []},
        }
        lines = [
            "regular output",
            f"dtn_artifact_k39fd2:{json.dumps(chart_data)}",
            "more output",
        ]
        artifacts = Process._parse_output(lines)
        assert artifacts.stdout == "regular output\nmore output"
        assert len(artifacts.charts) == 1
        assert artifacts.charts[0].title == "Test"


class TestSyncProcessExec:
    def _make_process(self):
        from daytona._sync.process import Process

        mock_toolbox = MagicMock()
        mock_api = MagicMock()
        return Process(mock_toolbox, mock_api), mock_toolbox, mock_api

    def test_exec_simple_command(self):
        proc, toolbox, api = self._make_process()
        mock_response = MagicMock()
        mock_response.result = "Hello, World!"
        mock_response.exit_code = 0
        mock_response.additional_properties = {}
        api.execute_command.return_value = mock_response
        result = proc.exec("echo 'Hello, World!'")
        assert result.exit_code == 0
        assert result.result == "Hello, World!"

    def test_exec_with_cwd(self):
        proc, toolbox, api = self._make_process()
        mock_response = MagicMock()
        mock_response.result = "file.txt"
        mock_response.exit_code = 0
        mock_response.additional_properties = {}
        api.execute_command.return_value = mock_response
        result = proc.exec("ls", cwd="/workspace")
        assert result.exit_code == 0
        call_args = api.execute_command.call_args
        assert call_args.kwargs["request"].cwd == "/workspace"

    def test_exec_with_env(self):
        proc, toolbox, api = self._make_process()
        mock_response = MagicMock()
        mock_response.result = "value"
        mock_response.exit_code = 0
        mock_response.additional_properties = {}
        api.execute_command.return_value = mock_response
        result = proc.exec("echo $MY_VAR", env={"MY_VAR": "value"})
        assert result.exit_code == 0
        call_args = api.execute_command.call_args
        assert "export MY_VAR" in call_args.kwargs["request"].command

    def test_exec_with_invalid_env_key_raises(self):
        proc, toolbox, api = self._make_process()
        with pytest.raises(DaytonaError, match="Invalid environment variable name"):
            proc.exec("echo test", env={"invalid-key": "value"})

    def test_code_run(self):
        proc, toolbox, api = self._make_process()
        toolbox.get_run_command.return_value = 'python3 -c "print(42)"'
        mock_response = MagicMock()
        mock_response.result = "42"
        mock_response.exit_code = 0
        mock_response.additional_properties = {}
        api.execute_command.return_value = mock_response
        result = proc.code_run("print(42)")
        assert result.exit_code == 0
        toolbox.get_run_command.assert_called_once()

    def test_code_run_with_params(self):
        proc, toolbox, api = self._make_process()
        toolbox.get_run_command.return_value = 'python3 -c "print(42)"'
        mock_response = MagicMock()
        mock_response.result = "42"
        mock_response.exit_code = 0
        mock_response.additional_properties = {}
        api.execute_command.return_value = mock_response
        params = CodeRunParams(argv=["--verbose"], env={"DEBUG": "1"})
        result = proc.code_run("print(42)", params=params)
        assert result.exit_code == 0


class TestSyncProcessSessions:
    def _make_process(self):
        from daytona._sync.process import Process

        mock_toolbox = MagicMock()
        mock_api = MagicMock()
        return Process(mock_toolbox, mock_api), mock_api

    def test_create_session(self):
        proc, api = self._make_process()
        api.create_session.return_value = None
        proc.create_session("my-session")
        api.create_session.assert_called_once()

    def test_get_session(self):
        proc, api = self._make_process()
        mock_session = MagicMock()
        mock_session.session_id = "my-session"
        api.get_session.return_value = mock_session
        result = proc.get_session("my-session")
        assert result.session_id == "my-session"

    def test_list_sessions(self):
        proc, api = self._make_process()
        api.list_sessions.return_value = [MagicMock()]
        result = proc.list_sessions()
        assert len(result) == 1

    def test_delete_session(self):
        proc, api = self._make_process()
        api.delete_session.return_value = None
        proc.delete_session("my-session")
        api.delete_session.assert_called_once_with(session_id="my-session")


class TestAsyncProcessExec:
    def _make_process(self):
        from daytona._async.process import AsyncProcess

        mock_toolbox = MagicMock()
        mock_api = AsyncMock()
        return AsyncProcess(mock_toolbox, mock_api), mock_toolbox, mock_api

    @pytest.mark.asyncio
    async def test_exec_simple(self):
        proc, toolbox, api = self._make_process()
        mock_response = MagicMock()
        mock_response.result = "output"
        mock_response.exit_code = 0
        mock_response.additional_properties = {}
        api.execute_command.return_value = mock_response
        result = await proc.exec("echo hello")
        assert result.exit_code == 0
        assert result.result == "output"

    @pytest.mark.asyncio
    async def test_code_run(self):
        proc, toolbox, api = self._make_process()
        toolbox.get_run_command.return_value = 'python3 -c "print(1)"'
        mock_response = MagicMock()
        mock_response.result = "1"
        mock_response.exit_code = 0
        mock_response.additional_properties = {}
        api.execute_command.return_value = mock_response
        result = await proc.code_run("print(1)")
        assert result.exit_code == 0

    @pytest.mark.asyncio
    async def test_exec_with_invalid_env_key_raises(self):
        proc, toolbox, api = self._make_process()
        with pytest.raises(DaytonaError, match="Invalid environment variable name"):
            await proc.exec("echo test", env={"bad-key": "value"})
