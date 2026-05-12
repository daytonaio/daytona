# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest
from httpx_ws import WebSocketDisconnect
from wsproto.events import BytesMessage, CloseConnection, TextMessage

from daytona.common.errors import DaytonaConnectionError, DaytonaError, DaytonaTimeoutError
from daytona.common.pty import PtySize


def _text(data: str) -> TextMessage:
    return TextMessage(data=data, frame_finished=True, message_finished=True)


def _binary(data: bytes) -> BytesMessage:
    return BytesMessage(data=data, frame_finished=True, message_finished=True)


class TestPtyHandle:
    def _make_handle(self):
        from daytona.handle.pty_handle import PtyHandle

        ws = MagicMock()
        return PtyHandle(ws, "session-1"), ws

    def test_properties_and_connection_state(self):
        handle, _ws = self._make_handle()

        assert handle.session_id == "session-1"
        assert handle.exit_code is None
        assert handle.error is None
        assert handle.is_connected() is True

    def test_wait_for_connection_returns_when_already_established(self):
        handle, ws = self._make_handle()
        handle._connection_established = True

        handle.wait_for_connection()

        ws.receive.assert_not_called()

    def test_wait_for_connection_raises_when_websocket_missing(self):
        handle, _ws = self._make_handle()
        handle._ws = None

        with pytest.raises(DaytonaConnectionError, match="not available"):
            handle.wait_for_connection()

    def test_wait_for_connection_processes_connected_control_message(self):
        handle, ws = self._make_handle()
        ws.receive.side_effect = [_text('{"type":"control","status":"connected"}')]

        handle.wait_for_connection(timeout=0.2)

        assert handle._connection_established is True

    def test_wait_for_connection_raises_when_error_message_received(self):
        handle, ws = self._make_handle()
        ws.receive.side_effect = [_text('{"type":"control","status":"error","error":"boom"}')]

        with pytest.raises(DaytonaConnectionError, match="boom"):
            handle.wait_for_connection(timeout=0.2)

    def test_wait_for_connection_times_out(self, monkeypatch):
        handle, ws = self._make_handle()
        ws.receive.side_effect = TimeoutError()
        times = iter([0.0, 0.2])
        monkeypatch.setattr("daytona.handle.pty_handle.time.time", lambda: next(times))

        with pytest.raises(DaytonaTimeoutError, match="connection timeout"):
            handle.wait_for_connection(timeout=0.1)

    def test_send_input_encodes_string(self):
        handle, ws = self._make_handle()

        handle.send_input("echo hi\n")

        ws.send_bytes.assert_called_once_with(b"echo hi\n")

    def test_send_input_passes_bytes_through(self):
        handle, ws = self._make_handle()

        handle.send_input(b"raw")

        ws.send_bytes.assert_called_once_with(b"raw")

    def test_send_input_raises_when_disconnected(self):
        handle, _ws = self._make_handle()
        handle._connected = False

        with pytest.raises(DaytonaConnectionError, match="not connected"):
            handle.send_input("hi")

    def test_resize_raises_without_handler(self):
        handle, _ws = self._make_handle()

        with pytest.raises(DaytonaError, match="Resize handler not available"):
            handle.resize(PtySize(rows=24, cols=80))

    def test_resize_wraps_handler_errors(self):
        handle, _ws = self._make_handle()
        handle._handle_resize = MagicMock(side_effect=RuntimeError("bad resize"))

        with pytest.raises(DaytonaError, match="Failed to resize PTY: bad resize"):
            handle.resize(PtySize(rows=24, cols=80))

    def test_kill_raises_without_handler(self):
        handle, _ws = self._make_handle()

        with pytest.raises(DaytonaError, match="Kill handler not available"):
            handle.kill()

    def test_kill_wraps_handler_errors(self):
        handle, _ws = self._make_handle()
        handle._handle_kill = MagicMock(side_effect=RuntimeError("bad kill"))

        with pytest.raises(DaytonaError, match="Failed to kill PTY: bad kill"):
            handle.kill()

    def test_iterator_yields_text_and_bytes_but_skips_control_messages(self):
        handle, ws = self._make_handle()
        ws.receive.side_effect = [
            _text('{"type":"control","status":"connected"}'),
            _text("hello"),
            _binary(b"world"),
            CloseConnection(code=1000, reason=None),
        ]

        assert list(handle) == [b"hello", b"world"]
        assert handle.exit_code == 0

    def test_wait_forwards_data_to_callback(self):
        handle, ws = self._make_handle()
        ws.receive.side_effect = [
            _text("hello"),
            _binary(b"world"),
            CloseConnection(code=1000, reason=None),
        ]
        received: list[bytes] = []

        result = handle.wait(on_data=received.append)

        assert received == [b"hello", b"world"]
        assert result.exit_code == 0
        assert result.error is None

    def test_iterator_terminates_on_websocket_disconnect(self):
        handle, ws = self._make_handle()
        ws.receive.side_effect = [
            _text("hello"),
            WebSocketDisconnect(code=1000, reason='{"exitCode":0}'),
        ]

        assert list(handle) == [b"hello"]
        assert handle.exit_code == 0

    def test_disconnect_closes_websocket_and_marks_disconnected(self):
        handle, ws = self._make_handle()

        handle.disconnect()

        ws.close.assert_called_once()
        assert handle._ws is None
        assert handle.is_connected() is False

    def test_disconnect_uses_context_manager_when_provided(self):
        from daytona.handle.pty_handle import PtyHandle

        ws = MagicMock()
        ws_cm = MagicMock()
        ws_cm.__exit__ = MagicMock(return_value=False)
        handle = PtyHandle(ws, "session-1", ws_context_manager=ws_cm)

        handle.disconnect()

        ws_cm.__exit__.assert_called_once_with(None, None, None)
        ws.close.assert_not_called()
        assert handle._ws is None

    def test_handle_close_parses_structured_exit_payload(self):
        handle, _ws = self._make_handle()

        handle._handle_close(1000, '{"exitCode":2,"exitReason":"failed","error":"boom"}')

        assert handle.exit_code == 2
        assert handle.error == "boom"

    def test_handle_close_defaults_exit_code_zero_for_normal_close(self):
        handle, _ws = self._make_handle()

        handle._handle_close(1000, "plain reason")

        assert handle.exit_code == 0


class TestAsyncPtyHandle:
    async def _build_handle(self, monkeypatch, ws: AsyncMock | None = None):
        from daytona.handle.async_pty_handle import AsyncPtyHandle

        async def noop(self):
            return None

        monkeypatch.setattr(AsyncPtyHandle, "_handle_websocket", noop)
        ws = ws or AsyncMock()
        ws.close_code = None
        ws.closed = False
        return AsyncPtyHandle(ws, session_id="session-1"), ws

    @pytest.mark.asyncio
    async def test_properties_and_connection_state(self, monkeypatch):
        handle, _ws = await self._build_handle(monkeypatch)

        assert handle.session_id == "session-1"
        assert handle.exit_code is None
        assert handle.error is None
        assert handle.is_connected() is False
        await handle.disconnect()

    @pytest.mark.asyncio
    async def test_wait_for_connection_returns_when_established(self, monkeypatch):
        handle, _ws = await self._build_handle(monkeypatch)
        handle._connection_established = True

        await handle.wait_for_connection()
        await handle.disconnect()

    @pytest.mark.asyncio
    async def test_wait_for_connection_raises_on_error(self, monkeypatch):
        handle, _ws = await self._build_handle(monkeypatch)
        handle._error = "boom"

        with pytest.raises(DaytonaConnectionError, match="boom"):
            await handle.wait_for_connection()
        await handle.disconnect()

    @pytest.mark.asyncio
    async def test_send_input_encodes_string(self, monkeypatch):
        handle, ws = await self._build_handle(monkeypatch)
        handle._connected = True

        await handle.send_input("echo hi\n")

        ws.send_bytes.assert_awaited_once_with(b"echo hi\n")
        await handle.disconnect()

    @pytest.mark.asyncio
    async def test_send_input_raises_when_disconnected(self, monkeypatch):
        handle, _ws = await self._build_handle(monkeypatch)

        with pytest.raises(DaytonaConnectionError, match="not connected"):
            await handle.send_input("hi")
        await handle.disconnect()

    @pytest.mark.asyncio
    async def test_wait_returns_result_from_background_task(self, monkeypatch):
        handle, _ws = await self._build_handle(monkeypatch)
        handle._exit_code = 3
        handle._error = "bad"

        result = await handle.wait()

        assert result.exit_code == 3
        assert result.error == "bad"
        await handle.disconnect()

    @pytest.mark.asyncio
    async def test_resize_and_kill_require_handlers(self, monkeypatch):
        handle, _ws = await self._build_handle(monkeypatch)

        with pytest.raises(DaytonaError, match="Resize handler not available"):
            await handle.resize(PtySize(rows=24, cols=80))
        with pytest.raises(DaytonaError, match="Kill handler not available"):
            await handle.kill()
        await handle.disconnect()

    @pytest.mark.asyncio
    async def test_handle_message_routes_text_and_binary_to_callback(self, monkeypatch):
        received: list[bytes] = []
        from daytona.handle.async_pty_handle import AsyncPtyHandle

        async def noop(self):
            return None

        monkeypatch.setattr(AsyncPtyHandle, "_handle_websocket", noop)
        ws = AsyncMock()
        ws.close_code = None
        handle = AsyncPtyHandle(ws, on_data=received.append, session_id="session-1")

        await handle._handle_message('{"type":"control","status":"connected"}')
        await handle._handle_message("hello")
        await handle._handle_message(b"world")

        assert handle._connection_established is True
        assert received == [b"hello", b"world"]
        await handle.disconnect()

    @pytest.mark.asyncio
    async def test_call_data_handler_supports_async_callbacks(self, monkeypatch):
        received: list[bytes] = []

        async def on_data(data: bytes):
            received.append(data)

        handle, _ws = await self._build_handle(monkeypatch)
        handle._on_data = on_data

        await handle._call_data_handler(b"chunk")

        assert received == [b"chunk"]
        await handle.disconnect()

    @pytest.mark.asyncio
    async def test_disconnect_closes_websocket_when_open(self, monkeypatch):
        handle, ws = await self._build_handle(monkeypatch)
        handle._connected = True

        await handle.disconnect()

        ws.close.assert_awaited_once()

    @pytest.mark.asyncio
    async def test_handle_close_parses_structured_exit_payload(self, monkeypatch):
        handle, _ws = await self._build_handle(monkeypatch)

        await handle._handle_close(1000, '{"exitCode":5,"exitReason":"failed","error":"boom"}')

        assert handle.exit_code == 5
        assert handle.error == "boom"
        await handle.disconnect()

    @pytest.mark.asyncio
    async def test_handle_close_defaults_exit_code_zero_for_normal_close(self, monkeypatch):
        handle, _ws = await self._build_handle(monkeypatch)

        await handle._handle_close(1000, "plain reason")

        assert handle.exit_code == 0
        await handle.disconnect()
