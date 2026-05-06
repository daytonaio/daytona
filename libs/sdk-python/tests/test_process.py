# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest

from daytona.common.process import CodeRunParams
from daytona_toolbox_api_client import Chart as GeneratedChart


class TestSyncProcessExec:
    def _make_process(self):
        from daytona._sync.process import Process

        mock_api = MagicMock()
        return Process("python", mock_api, http_client=MagicMock()), mock_api

    def test_exec_simple_command(self):
        proc, api = self._make_process()
        api.execute_command.return_value = MagicMock(result="Hello, World!", exit_code=0, additional_properties={})
        result = proc.exec("echo 'Hello, World!'")
        assert result.exit_code == 0
        assert result.result == "Hello, World!"

    def test_exec_with_cwd_and_env(self):
        proc, api = self._make_process()
        api.execute_command.return_value = MagicMock(result="value", exit_code=0, additional_properties={})
        proc.exec("echo $MY_VAR", cwd="/workspace", env={"MY_VAR": "value"})
        request = api.execute_command.call_args.kwargs["request"]
        assert request.command == "echo $MY_VAR"
        assert request.cwd == "/workspace"
        assert request.envs == {"MY_VAR": "value"}

    def test_exec_falls_back_to_additional_properties_code(self):
        proc, api = self._make_process()
        api.execute_command.return_value = MagicMock(result="oops", exit_code=None, additional_properties={"code": 42})
        result = proc.exec("false")
        assert result.exit_code == 42

    def test_code_run_uses_language_and_params(self):
        proc, api = self._make_process()
        api.code_run.return_value = MagicMock(
            result="42\n",
            exit_code=0,
            artifacts=None,
            additional_properties={},
        )
        result = proc.code_run("print(42)", params=CodeRunParams(argv=["--flag"], env={"DEBUG": "1"}), timeout=5)
        request = api.code_run.call_args.kwargs["request"]
        assert request.language == "python"
        assert request.argv == ["--flag"]
        assert request.envs == {"DEBUG": "1"}
        assert request.timeout == 5
        assert result.result == "42\n"

    def test_code_run_parses_charts(self):
        proc, api = self._make_process()
        api.code_run.return_value = MagicMock(
            result="chart output",
            exit_code=0,
            artifacts=MagicMock(charts=[GeneratedChart(type="line", title="Line", elements=[])]),
            additional_properties={},
        )
        result = proc.code_run("print('chart')")
        assert result.artifacts is not None
        assert result.artifacts.charts is not None
        assert len(result.artifacts.charts) == 1
        assert result.artifacts.charts[0].title == "Line"


class TestSyncProcessSessions:
    def _make_process(self):
        from daytona._sync.process import Process

        mock_api = MagicMock()
        return Process("python", mock_api, http_client=MagicMock()), mock_api

    def test_create_session(self):
        proc, api = self._make_process()
        proc.create_session("my-session")
        request = api.create_session.call_args.kwargs["request"]
        assert request.session_id == "my-session"

    def test_get_session(self):
        proc, api = self._make_process()
        api.get_session.return_value = MagicMock(session_id="my-session")
        assert proc.get_session("my-session").session_id == "my-session"

    def test_execute_session_command(self):
        proc, api = self._make_process()
        api.session_execute_command.return_value = MagicMock(
            cmd_id="cmd-1",
            output="all",
            stdout="out",
            stderr="err",
            exit_code=0,
            additional_properties={},
        )
        result = proc.execute_session_command("my-session", req=MagicMock())
        assert result.cmd_id == "cmd-1"
        assert result.stdout == "out"
        assert result.stderr == "err"

    def test_get_session_command_logs(self):
        proc, api = self._make_process()
        api.get_session_command_logs.return_value = MagicMock(output="all", stdout="out", stderr="err")
        result = proc.get_session_command_logs("my-session", "cmd-1")
        assert result.output == "all"
        assert result.stdout == "out"
        assert result.stderr == "err"

    def test_delete_session(self):
        proc, api = self._make_process()
        proc.delete_session("my-session")
        api.delete_session.assert_called_once_with(session_id="my-session")

    def test_get_entrypoint_session(self):
        proc, api = self._make_process()
        api.get_entrypoint_session.return_value = MagicMock(session_id="entrypoint")

        assert proc.get_entrypoint_session().session_id == "entrypoint"

    def test_get_session_command(self):
        proc, api = self._make_process()
        api.get_session_command.return_value = MagicMock(id="cmd-1")

        assert proc.get_session_command("my-session", "cmd-1").id == "cmd-1"

    def test_get_entrypoint_logs(self):
        proc, api = self._make_process()
        api.get_entrypoint_logs.return_value = MagicMock(output="all", stdout="out", stderr="err")

        result = proc.get_entrypoint_logs()

        assert result.output == "all"
        assert result.stdout == "out"
        assert result.stderr == "err"

    def test_send_session_command_input_and_list_sessions(self):
        proc, api = self._make_process()
        api.list_sessions.return_value = [MagicMock(session_id="one")]

        proc.send_session_command_input("my-session", "cmd-1", "hello")
        sessions = proc.list_sessions()

        assert sessions[0].session_id == "one"
        send_request = api.send_input.call_args.kwargs["request"]
        assert send_request.data == "hello"


class TestAsyncProcessExec:
    def _make_process(self):
        from daytona._async.process import AsyncProcess

        mock_api = AsyncMock()
        return AsyncProcess("python", mock_api), mock_api

    @pytest.mark.asyncio
    async def test_exec_simple(self):
        proc, api = self._make_process()
        api.execute_command.return_value = MagicMock(result="output", exit_code=0, additional_properties={})
        result = await proc.exec("echo hello")
        assert result.exit_code == 0
        assert result.result == "output"

    @pytest.mark.asyncio
    async def test_exec_with_env(self):
        proc, api = self._make_process()
        api.execute_command.return_value = MagicMock(result="value", exit_code=0, additional_properties={})
        await proc.exec("echo $MY_VAR", env={"MY_VAR": "value"})
        request = api.execute_command.call_args.kwargs["request"]
        assert request.envs == {"MY_VAR": "value"}

    @pytest.mark.asyncio
    async def test_code_run(self):
        proc, api = self._make_process()
        api.code_run.return_value = MagicMock(result="1", exit_code=0, artifacts=None, additional_properties={})
        result = await proc.code_run("print(1)")
        request = api.code_run.call_args.kwargs["request"]
        assert request.language == "python"
        assert result.exit_code == 0

    @pytest.mark.asyncio
    async def test_create_and_delete_session(self):
        proc, api = self._make_process()
        await proc.create_session("my-session")
        request = api.create_session.call_args.kwargs["request"]
        assert request.session_id == "my-session"
        await proc.delete_session("my-session")
        api.delete_session.assert_called_once_with(session_id="my-session")

    @pytest.mark.asyncio
    async def test_execute_session_command(self):
        proc, api = self._make_process()
        api.session_execute_command.return_value = MagicMock(
            cmd_id="cmd-1",
            output="all",
            stdout="out",
            stderr="err",
            exit_code=0,
            additional_properties={},
        )
        result = await proc.execute_session_command("my-session", req=MagicMock())
        assert result.cmd_id == "cmd-1"
        assert result.output == "all"

    @pytest.mark.asyncio
    async def test_get_entrypoint_and_session_metadata(self):
        proc, api = self._make_process()
        api.get_entrypoint_session.return_value = MagicMock(session_id="entrypoint")
        api.get_session_command.return_value = MagicMock(id="cmd-1")
        api.get_entrypoint_logs.return_value = MagicMock(output="all", stdout="out", stderr="err")
        api.list_sessions.return_value = [MagicMock(session_id="one")]

        assert (await proc.get_entrypoint_session()).session_id == "entrypoint"
        assert (await proc.get_session_command("my-session", "cmd-1")).id == "cmd-1"
        assert (await proc.get_entrypoint_logs()).stdout == "out"
        assert (await proc.list_sessions())[0].session_id == "one"

    @pytest.mark.asyncio
    async def test_send_session_command_input(self):
        proc, api = self._make_process()

        await proc.send_session_command_input("my-session", "cmd-1", "hello")

        send_request = api.send_input.call_args.kwargs["request"]
        assert send_request.data == "hello"
