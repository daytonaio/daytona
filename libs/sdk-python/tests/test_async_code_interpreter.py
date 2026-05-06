# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import json
from types import SimpleNamespace
from unittest.mock import AsyncMock, MagicMock

import aiohttp
import pytest

from daytona.common.errors import DaytonaConnectionError, DaytonaTimeoutError


def _make_async_interpreter(http_session=None):
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
    if http_session is not None:
        api_client.api_client.http_session = http_session
    return AsyncCodeInterpreter(api_client), api_client


def _text_msg(payload: str) -> SimpleNamespace:
    return SimpleNamespace(type=aiohttp.WSMsgType.TEXT, data=payload, extra=None)


def _closed_msg() -> SimpleNamespace:
    return SimpleNamespace(type=aiohttp.WSMsgType.CLOSED, data=None, extra=None)


def _close_msg(code: int, reason: str | None = None) -> SimpleNamespace:
    return SimpleNamespace(type=aiohttp.WSMsgType.CLOSE, data=code, extra=reason)


def _make_session_with_ws(ws: AsyncMock) -> AsyncMock:
    session = AsyncMock(spec=aiohttp.ClientSession)
    session.ws_connect = AsyncMock(return_value=ws)
    return session


def _make_ws() -> AsyncMock:
    ws = AsyncMock()
    ws.closed = False
    ws.close_code = 1000
    return ws


class TestAsyncCodeInterpreterRunCode:
    @pytest.mark.asyncio
    async def test_run_code_streams_messages_and_invokes_callbacks(self):
        ws = _make_ws()
        ws.receive.side_effect = [
            _text_msg(json.dumps({"type": "stdout", "text": "hello "})),
            _text_msg(json.dumps({"type": "stderr", "text": "warn"})),
            _text_msg(json.dumps({"type": "error", "name": "ValueError", "value": "bad", "traceback": "tb"})),
            _closed_msg(),
        ]
        session = _make_session_with_ws(ws)

        interpreter, _api_client = _make_async_interpreter(http_session=session)
        stdout_messages: list[str] = []
        stderr_messages: list[str] = []
        error_names: list[str] = []

        result = await interpreter.run_code(
            "print('hi')",
            on_stdout=lambda msg: stdout_messages.append(msg.output),
            on_stderr=lambda msg: stderr_messages.append(msg.output),
            on_error=lambda err: error_names.append(err.name),
        )

        sent_payload = json.loads(ws.send_str.call_args.args[0])
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
        ws = _make_ws()
        ws.receive.side_effect = [_closed_msg()]
        session = _make_session_with_ws(ws)

        interpreter, _api_client = _make_async_interpreter(http_session=session)
        context = SimpleNamespace(id="ctx-1")

        await interpreter.run_code(
            "print('hi')",
            context=context,
            envs={"DEBUG": "1"},
            timeout=15,
        )

        assert json.loads(ws.send_str.call_args.args[0]) == {
            "code": "print('hi')",
            "contextId": "ctx-1",
            "envs": {"DEBUG": "1"},
            "timeout": 15,
        }

    @pytest.mark.asyncio
    async def test_run_code_raises_timeout_on_close_code_4008(self):
        ws = _make_ws()
        ws.receive.side_effect = [_close_msg(4008, "timed out")]
        session = _make_session_with_ws(ws)

        interpreter, _api_client = _make_async_interpreter(http_session=session)

        with pytest.raises(DaytonaTimeoutError, match="Execution timed out"):
            await interpreter.run_code("print('hi')")

    @pytest.mark.asyncio
    async def test_run_code_raises_connection_error_on_unexpected_close(self):
        ws = _make_ws()
        ws.receive.side_effect = [_close_msg(4999, "closed unexpectedly")]
        session = _make_session_with_ws(ws)

        interpreter, _api_client = _make_async_interpreter(http_session=session)

        with pytest.raises(DaytonaConnectionError, match=r"closed unexpectedly \(close code 4999\)"):
            await interpreter.run_code("print('hi')")

    def test_maybe_raise_from_ws_close_timeout(self):
        interpreter, _api_client = _make_async_interpreter()

        with pytest.raises(DaytonaTimeoutError, match="Execution timed out"):
            interpreter._maybe_raise_from_ws_close(4008, "timed out")

    def test_maybe_raise_from_ws_close_connection_error(self):
        interpreter, _api_client = _make_async_interpreter()

        with pytest.raises(DaytonaConnectionError, match=r"closed unexpectedly \(close code 4999\)"):
            interpreter._maybe_raise_from_ws_close(4999, "closed unexpectedly")

    def test_maybe_raise_from_ws_close_normal_close_does_not_raise(self):
        interpreter, _api_client = _make_async_interpreter()

        interpreter._maybe_raise_from_ws_close(1000, None)
        interpreter._maybe_raise_from_ws_close(None, None)


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
