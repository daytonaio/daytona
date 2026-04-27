# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import json
from types import SimpleNamespace
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from daytona.common.errors import DaytonaConnectionError, DaytonaTimeoutError


def _make_async_interpreter():
    from daytona._async.code_interpreter import AsyncCodeInterpreter

    api_client = MagicMock()
    api_client.create_interpreter_context = AsyncMock()
    api_client.list_interpreter_contexts = AsyncMock()
    api_client.delete_interpreter_context = AsyncMock()
    api_client._execute_interpreter_code_serialize.return_value = (
        "GET",
        "https://toolbox.example/interpreter",
        {"Authorization": "Bearer token"},
        None,
    )
    return AsyncCodeInterpreter(api_client), api_client


def _close_event(code: int | None = None, reason: str | None = None, *, use_sent: bool = False):
    payload = SimpleNamespace(code=code, reason=reason)
    return SimpleNamespace(rcvd=None if use_sent else payload, sent=payload if use_sent else None)


class TestAsyncCodeInterpreterRunCode:
    @pytest.mark.asyncio
    async def test_run_code_streams_messages_and_invokes_callbacks(self):
        interpreter, _api_client = _make_async_interpreter()
        stdout_messages: list[str] = []
        stderr_messages: list[str] = []
        error_names: list[str] = []

        class FakeConnectionClosed(Exception):
            pass

        class FakeConnectionClosedOK(FakeConnectionClosed):
            pass

        websocket = AsyncMock()
        websocket.recv.side_effect = [
            json.dumps({"type": "stdout", "text": "hello "}),
            json.dumps({"type": "stderr", "text": "warn"}),
            json.dumps({"type": "error", "name": "ValueError", "value": "bad", "traceback": "tb"}),
            FakeConnectionClosedOK(),
        ]
        websocket_cm = AsyncMock()
        websocket_cm.__aenter__.return_value = websocket
        websocket_cm.__aexit__.return_value = False

        with (
            patch("daytona._async.code_interpreter.connect", return_value=websocket_cm),
            patch("daytona._async.code_interpreter.ConnectionClosed", FakeConnectionClosed),
            patch("daytona._async.code_interpreter.ConnectionClosedOK", FakeConnectionClosedOK),
        ):
            result = await interpreter.run_code(
                "print('hi')",
                on_stdout=lambda msg: stdout_messages.append(msg.output),
                on_stderr=lambda msg: stderr_messages.append(msg.output),
                on_error=lambda err: error_names.append(err.name),
            )

        sent_payload = json.loads(websocket.send.call_args.args[0])
        assert sent_payload == {"code": "print('hi')"}
        assert result.stdout == "hello "
        assert result.stderr == "warn"
        assert result.error is not None
        assert result.error.value == "bad"
        assert stdout_messages == ["hello "]
        assert stderr_messages == ["warn"]
        assert error_names == ["ValueError"]

    @pytest.mark.asyncio
    async def test_run_code_includes_context_envs_and_timeout(self):
        interpreter, _api_client = _make_async_interpreter()

        class FakeConnectionClosed(Exception):
            pass

        class FakeConnectionClosedOK(FakeConnectionClosed):
            pass

        websocket = AsyncMock()
        websocket.recv.side_effect = [FakeConnectionClosedOK()]
        websocket_cm = AsyncMock()
        websocket_cm.__aenter__.return_value = websocket
        websocket_cm.__aexit__.return_value = False
        context = SimpleNamespace(id="ctx-1")

        with (
            patch("daytona._async.code_interpreter.connect", return_value=websocket_cm),
            patch("daytona._async.code_interpreter.ConnectionClosed", FakeConnectionClosed),
            patch("daytona._async.code_interpreter.ConnectionClosedOK", FakeConnectionClosedOK),
        ):
            await interpreter.run_code(
                "print('hi')",
                context=context,
                envs={"DEBUG": "1"},
                timeout=15,
            )

        assert json.loads(websocket.send.call_args.args[0]) == {
            "code": "print('hi')",
            "contextId": "ctx-1",
            "envs": {"DEBUG": "1"},
            "timeout": 15,
        }

    def test_raise_from_ws_close_timeout(self):
        interpreter, _api_client = _make_async_interpreter()

        with pytest.raises(DaytonaTimeoutError, match="Execution timed out"):
            interpreter._raise_from_ws_close(_close_event(4008, "timed out"))

    def test_raise_from_ws_close_connection_error(self):
        interpreter, _api_client = _make_async_interpreter()

        with pytest.raises(DaytonaConnectionError, match=r"closed unexpectedly \(close code 4999\)"):
            interpreter._raise_from_ws_close(_close_event(4999, "closed unexpectedly"))


class TestAsyncCodeInterpreterContextManagement:
    @pytest.mark.asyncio
    async def test_create_context_passes_cwd(self):
        interpreter, api_client = _make_async_interpreter()
        api_client.create_interpreter_context.return_value = SimpleNamespace(id="ctx-1")

        context = await interpreter.create_context(cwd="/workspace")

        assert context.id == "ctx-1"
        request = api_client.create_interpreter_context.call_args.kwargs["request"]
        assert request.cwd == "/workspace"

    @pytest.mark.asyncio
    async def test_list_contexts_returns_empty_list_when_contexts_missing(self):
        interpreter, api_client = _make_async_interpreter()
        api_client.list_interpreter_contexts.return_value = SimpleNamespace(contexts=None)

        assert await interpreter.list_contexts() == []

    @pytest.mark.asyncio
    async def test_list_contexts_returns_contexts(self):
        interpreter, api_client = _make_async_interpreter()
        contexts = [SimpleNamespace(id="ctx-1")]
        api_client.list_interpreter_contexts.return_value = SimpleNamespace(contexts=contexts)

        assert await interpreter.list_contexts() == contexts

    @pytest.mark.asyncio
    async def test_delete_context_uses_context_id(self):
        interpreter, api_client = _make_async_interpreter()

        await interpreter.delete_context(SimpleNamespace(id="ctx-123"))

        api_client.delete_interpreter_context.assert_awaited_once_with(id="ctx-123")
