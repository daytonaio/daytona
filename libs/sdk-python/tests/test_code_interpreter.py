# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import json
from types import SimpleNamespace
from unittest.mock import MagicMock, patch

import pytest
from httpx_ws import WebSocketDisconnect
from wsproto.events import CloseConnection, TextMessage

from daytona.common.errors import DaytonaConnectionError, DaytonaTimeoutError


def _make_interpreter():
    from daytona._sync.code_interpreter import CodeInterpreter

    api_client = MagicMock()
    api_client._execute_interpreter_code_serialize.return_value = (
        "GET",
        "https://toolbox.example/interpreter",
        {"Authorization": "Bearer token"},
        None,
    )
    return CodeInterpreter(api_client, http_client=MagicMock()), api_client


def _text(payload: dict) -> TextMessage:
    return TextMessage(data=json.dumps(payload), frame_finished=True, message_finished=True)


def _make_ws_cm(events):
    websocket = MagicMock()
    websocket.receive.side_effect = events
    websocket_cm = MagicMock()
    websocket_cm.__enter__.return_value = websocket
    websocket_cm.__exit__.return_value = False
    return websocket, websocket_cm


class TestCodeInterpreterRunCode:
    def test_run_code_streams_messages_and_invokes_callbacks(self):
        interpreter, _api_client = _make_interpreter()
        stdout_messages: list[str] = []
        stderr_messages: list[str] = []
        error_names: list[str] = []

        ws, ws_cm = _make_ws_cm(
            [
                _text({"type": "stdout", "text": "hello "}),
                _text({"type": "stderr", "text": "warn"}),
                _text({"type": "error", "name": "ValueError", "value": "bad", "traceback": "tb"}),
                _text({"type": "stdout", "text": "world"}),
                CloseConnection(code=1000, reason=None),
            ]
        )

        with patch("daytona._sync.code_interpreter.httpx_ws.connect_ws", return_value=ws_cm):
            result = interpreter.run_code(
                "print('hi')",
                on_stdout=lambda msg: stdout_messages.append(msg.output),
                on_stderr=lambda msg: stderr_messages.append(msg.output),
                on_error=lambda err: error_names.append(err.name),
            )

        sent_payload = json.loads(ws.send_text.call_args.args[0])
        assert sent_payload == {"code": "print('hi')"}
        assert result.stdout == "hello world"
        assert result.stderr == "warn"
        assert result.error is not None
        assert result.error.name == "ValueError"
        assert stdout_messages == ["hello ", "world"]
        assert stderr_messages == ["warn"]
        assert error_names == ["ValueError"]

    def test_run_code_includes_context_envs_and_timeout(self):
        interpreter, _api_client = _make_interpreter()

        ws, ws_cm = _make_ws_cm([CloseConnection(code=1000, reason=None)])
        context = SimpleNamespace(id="ctx-1")

        with patch("daytona._sync.code_interpreter.httpx_ws.connect_ws", return_value=ws_cm):
            interpreter.run_code(
                "print('hi')",
                context=context,
                envs={"DEBUG": "1"},
                timeout=15,
            )

        assert json.loads(ws.send_text.call_args.args[0]) == {
            "code": "print('hi')",
            "contextId": "ctx-1",
            "envs": {"DEBUG": "1"},
            "timeout": 15,
        }

    def test_run_code_ignores_unknown_chunk_types(self):
        interpreter, _api_client = _make_interpreter()

        _ws, ws_cm = _make_ws_cm(
            [
                _text({"type": "unknown", "text": "ignored"}),
                CloseConnection(code=1000, reason=None),
            ]
        )

        with patch("daytona._sync.code_interpreter.httpx_ws.connect_ws", return_value=ws_cm):
            result = interpreter.run_code("print('hi')")

        assert result.stdout == ""
        assert result.stderr == ""
        assert result.error is None

    def test_run_code_raises_timeout_when_disconnect_has_code_4008(self):
        interpreter, _api_client = _make_interpreter()

        _ws, ws_cm = _make_ws_cm([WebSocketDisconnect(code=4008, reason="timed out")])

        with patch("daytona._sync.code_interpreter.httpx_ws.connect_ws", return_value=ws_cm):
            with pytest.raises(DaytonaTimeoutError, match="Execution timed out"):
                interpreter.run_code("print('hi')")

    def test_run_code_raises_connection_error_on_unexpected_disconnect(self):
        interpreter, _api_client = _make_interpreter()

        _ws, ws_cm = _make_ws_cm([WebSocketDisconnect(code=4001, reason="socket ended")])

        with patch("daytona._sync.code_interpreter.httpx_ws.connect_ws", return_value=ws_cm):
            with pytest.raises(DaytonaConnectionError, match=r"socket ended \(close code 4001\)"):
                interpreter.run_code("print('hi')")

    def test_run_code_raises_connection_error_via_close_frame(self):
        interpreter, _api_client = _make_interpreter()

        _ws, ws_cm = _make_ws_cm([CloseConnection(code=4100, reason="writer closed")])

        with patch("daytona._sync.code_interpreter.httpx_ws.connect_ws", return_value=ws_cm):
            with pytest.raises(DaytonaConnectionError, match=r"writer closed \(close code 4100\)"):
                interpreter.run_code("print('hi')")

    def test_maybe_raise_from_close_normal_close_does_not_raise(self):
        interpreter, _api_client = _make_interpreter()

        interpreter._maybe_raise_from_close(1000, None)
        interpreter._maybe_raise_from_close(None, None)


class TestCodeInterpreterContextManagement:
    def test_create_context_passes_cwd(self):
        interpreter, api_client = _make_interpreter()
        api_client.create_interpreter_context.return_value = SimpleNamespace(id="ctx-1")

        context = interpreter.create_context(cwd="/workspace")

        assert context.id == "ctx-1"
        request = api_client.create_interpreter_context.call_args.kwargs["request"]
        assert request.cwd == "/workspace"

    def test_create_context_without_cwd(self):
        interpreter, api_client = _make_interpreter()
        api_client.create_interpreter_context.return_value = SimpleNamespace(id="ctx-1")

        interpreter.create_context()

        request = api_client.create_interpreter_context.call_args.kwargs["request"]
        assert request.cwd is None

    def test_list_contexts_returns_contexts(self):
        interpreter, api_client = _make_interpreter()
        contexts = [SimpleNamespace(id="ctx-1"), SimpleNamespace(id="ctx-2")]
        api_client.list_interpreter_contexts.return_value = SimpleNamespace(contexts=contexts)

        assert interpreter.list_contexts() == contexts

    def test_list_contexts_returns_empty_list_when_contexts_missing(self):
        interpreter, api_client = _make_interpreter()
        api_client.list_interpreter_contexts.return_value = SimpleNamespace(contexts=None)

        assert interpreter.list_contexts() == []

    def test_delete_context_uses_context_id(self):
        interpreter, api_client = _make_interpreter()

        interpreter.delete_context(SimpleNamespace(id="ctx-123"))

        api_client.delete_interpreter_context.assert_called_once_with(id="ctx-123")
