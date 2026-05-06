# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from pathlib import Path
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from daytona.common.computer_use import ScreenshotOptions, ScreenshotRegion


def _make_async_computer_use():
    from daytona._async.computer_use import AsyncComputerUse

    api_client = AsyncMock()
    return AsyncComputerUse(api_client), api_client


class TestAsyncComputerUse:
    @pytest.mark.asyncio
    async def test_init_exposes_subsystems(self):
        computer_use, _api_client = _make_async_computer_use()

        assert computer_use.mouse is not None
        assert computer_use.keyboard is not None
        assert computer_use.screenshot is not None
        assert computer_use.display is not None
        assert computer_use.recording is not None

    @pytest.mark.asyncio
    async def test_mouse_methods_delegate_to_api(self):
        computer_use, api_client = _make_async_computer_use()
        api_client.get_mouse_position.return_value = MagicMock(x=10, y=20)
        api_client.move_mouse.return_value = MagicMock(x=30, y=40)
        api_client.click.return_value = MagicMock(success=True)
        api_client.drag.return_value = MagicMock(success=True)
        api_client.scroll.return_value = MagicMock(success=False)

        assert (await computer_use.mouse.get_position()).x == 10
        assert (await computer_use.mouse.move(30, 40)).y == 40
        assert (await computer_use.mouse.click(1, 2, button="right", double=True)).success is True
        assert (await computer_use.mouse.drag(1, 2, 3, 4, button="middle")).success is True
        assert await computer_use.mouse.scroll(5, 6, "down", amount=3) is False

        assert api_client.move_mouse.call_args.args[0].x == 30
        click_request = api_client.click.call_args.args[0]
        assert click_request.button == "right"
        assert click_request.double is True

    @pytest.mark.asyncio
    async def test_keyboard_methods_delegate_to_api(self):
        computer_use, api_client = _make_async_computer_use()

        await computer_use.keyboard.type("hello", delay=50)
        await computer_use.keyboard.press("Enter")
        await computer_use.keyboard.hotkey("ctrl+shift+t")

        type_request = api_client.type_text.call_args.kwargs["request"]
        assert type_request.text == "hello"
        assert type_request.delay == 50
        press_request = api_client.press_key.call_args.kwargs["request"]
        assert press_request.modifiers == []
        hotkey_request = api_client.press_hotkey.call_args.kwargs["request"]
        assert hotkey_request.keys == "ctrl+shift+t"

    @pytest.mark.asyncio
    async def test_screenshot_methods_delegate_to_api(self):
        computer_use, api_client = _make_async_computer_use()
        api_client.take_screenshot.return_value = MagicMock(width=1920)
        api_client.take_region_screenshot.return_value = MagicMock(width=300)
        api_client.take_compressed_screenshot.return_value = MagicMock(size_bytes=123)
        api_client.take_compressed_region_screenshot.return_value = MagicMock(size_bytes=456)
        region = ScreenshotRegion(x=10, y=20, width=300, height=200)

        assert (await computer_use.screenshot.take_full_screen(show_cursor=True)).width == 1920
        assert (await computer_use.screenshot.take_region(region)).width == 300
        assert (await computer_use.screenshot.take_compressed()).size_bytes == 123
        assert (
            await computer_use.screenshot.take_compressed(
                ScreenshotOptions(show_cursor=True, fmt="jpeg", quality=90, scale=0.5)
            )
        ).size_bytes == 123
        assert (await computer_use.screenshot.take_compressed_region(region)).size_bytes == 456

        api_client.take_screenshot.assert_awaited_once_with(show_cursor=True)

    @pytest.mark.asyncio
    async def test_display_and_recording_methods_delegate_to_api(self, tmp_path: Path):
        computer_use, api_client = _make_async_computer_use()
        api_client.get_display_info.return_value = MagicMock(total_displays=1)
        api_client.get_windows.return_value = MagicMock(count=2)
        api_client.start_recording.return_value = MagicMock(id="rec-1")
        api_client.stop_recording.return_value = MagicMock(id="rec-1")
        api_client.list_recordings.return_value = MagicMock(recordings=[])
        api_client.get_recording.return_value = MagicMock(id="rec-1")

        assert (await computer_use.display.get_info()).total_displays == 1
        assert (await computer_use.display.get_windows()).count == 2
        assert (await computer_use.recording.start("label")).id == "rec-1"
        assert (await computer_use.recording.stop("rec-1")).id == "rec-1"
        assert (await computer_use.recording.list()).recordings == []
        assert (await computer_use.recording.get("rec-1")).id == "rec-1"
        await computer_use.recording.delete("rec-1")

        class FakeContent:
            async def iter_chunked(self, _chunk_size):
                for part in [b"part1", b"part2"]:
                    yield part

        class FakeResponse:
            content = FakeContent()

            async def __aenter__(self):
                return self

            async def __aexit__(self, exc_type, exc, tb):
                return False

            def raise_for_status(self):
                return None

        class FakeClient:
            async def __aenter__(self):
                return self

            async def __aexit__(self, exc_type, exc, tb):
                return False

            def request(self, method, url, headers=None, timeout=None):
                assert method == "GET"
                assert url == "https://download"
                assert headers == {"Authorization": "Bearer token"}
                return FakeResponse()

            async def close(self):
                pass

        api_client._download_recording_serialize = MagicMock(
            return_value=("GET", "https://download", {"Authorization": "Bearer token"}, None)
        )
        destination = tmp_path / "nested" / "recording.mp4"

        api_client.api_client.http_session = FakeClient()
        await computer_use.recording.download("rec-1", str(destination))

        assert destination.read_bytes() == b"part1part2"

    @pytest.mark.asyncio
    async def test_top_level_methods_delegate_to_api(self):
        computer_use, api_client = _make_async_computer_use()
        api_client.start_computer_use.return_value = MagicMock(message="started")
        api_client.stop_computer_use.return_value = MagicMock(message="stopped")
        api_client.get_computer_use_status.return_value = MagicMock(status="running")
        api_client.get_process_status.return_value = MagicMock()
        api_client.get_process_status.return_value.name = "xvfb"
        api_client.restart_process.return_value = MagicMock(message="restarted")
        api_client.get_process_logs.return_value = MagicMock(logs=["line"])
        api_client.get_process_errors.return_value = MagicMock(errors=["err"])

        assert (await computer_use.start()).message == "started"
        assert (await computer_use.stop()).message == "stopped"
        assert (await computer_use.get_status()).status == "running"
        assert (await computer_use.get_process_status("xvfb")).name == "xvfb"
        assert (await computer_use.restart_process("xvfb")).message == "restarted"
        assert (await computer_use.get_process_logs("xvfb")).logs == ["line"]
        assert (await computer_use.get_process_errors("xvfb")).errors == ["err"]

        api_client.get_process_status.assert_awaited_once_with(process_name="xvfb")
