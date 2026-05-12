# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from pathlib import Path
from unittest.mock import MagicMock, patch

from daytona.common.computer_use import ScreenshotOptions, ScreenshotRegion


def _make_computer_use():
    from daytona._sync.computer_use import ComputerUse

    api_client = MagicMock()
    return ComputerUse(api_client, http_client=MagicMock()), api_client


class TestComputerUse:
    def test_init_exposes_subsystems(self):
        computer_use, _api_client = _make_computer_use()

        assert computer_use.mouse is not None
        assert computer_use.keyboard is not None
        assert computer_use.screenshot is not None
        assert computer_use.display is not None
        assert computer_use.recording is not None
        assert computer_use.accessibility is not None

    def test_mouse_methods_delegate_to_api(self):
        computer_use, api_client = _make_computer_use()
        api_client.get_mouse_position.return_value = MagicMock(x=10, y=20)
        api_client.move_mouse.return_value = MagicMock(x=30, y=40)
        api_client.click.return_value = MagicMock(success=True)
        api_client.drag.return_value = MagicMock(success=True)
        api_client.scroll.return_value = MagicMock(success=True)

        assert computer_use.mouse.get_position().x == 10
        assert computer_use.mouse.move(30, 40).y == 40
        assert computer_use.mouse.click(1, 2, button="right", double=True).success is True
        assert computer_use.mouse.drag(1, 2, 3, 4, button="middle").success is True
        assert computer_use.mouse.scroll(5, 6, "down", amount=3) is True

        assert api_client.move_mouse.call_args.args[0].x == 30
        click_request = api_client.click.call_args.args[0]
        assert click_request.button == "right"
        assert click_request.double is True
        drag_request = api_client.drag.call_args.kwargs["request"]
        assert drag_request.end_x == 3
        scroll_request = api_client.scroll.call_args.kwargs["request"]
        assert scroll_request.amount == 3

    def test_mouse_scroll_returns_false_when_api_reports_failure(self):
        computer_use, api_client = _make_computer_use()
        api_client.scroll.return_value = MagicMock(success=False)

        assert computer_use.mouse.scroll(1, 2, "up") is False

    def test_keyboard_methods_delegate_to_api(self):
        computer_use, api_client = _make_computer_use()

        computer_use.keyboard.type("hello", delay=50)
        computer_use.keyboard.press("Enter")
        computer_use.keyboard.press("c", modifiers=["ctrl"])
        computer_use.keyboard.hotkey("ctrl+shift+t")

        type_request = api_client.type_text.call_args.kwargs["request"]
        assert type_request.text == "hello"
        assert type_request.delay == 50
        first_press = api_client.press_key.call_args_list[0].kwargs["request"]
        second_press = api_client.press_key.call_args_list[1].kwargs["request"]
        assert first_press.modifiers == []
        assert second_press.modifiers == ["ctrl"]
        hotkey_request = api_client.press_hotkey.call_args.kwargs["request"]
        assert hotkey_request.keys == "ctrl+shift+t"

    def test_screenshot_methods_delegate_to_api(self):
        computer_use, api_client = _make_computer_use()
        api_client.take_screenshot.return_value = MagicMock(width=1920)
        api_client.take_region_screenshot.return_value = MagicMock(width=300)
        api_client.take_compressed_screenshot.return_value = MagicMock(size_bytes=123)
        api_client.take_compressed_region_screenshot.return_value = MagicMock(size_bytes=456)
        region = ScreenshotRegion(x=10, y=20, width=300, height=200)

        assert computer_use.screenshot.take_full_screen(show_cursor=True).width == 1920
        assert computer_use.screenshot.take_region(region).width == 300
        assert computer_use.screenshot.take_compressed().size_bytes == 123
        assert (
            computer_use.screenshot.take_compressed(
                ScreenshotOptions(show_cursor=True, fmt="jpeg", quality=90, scale=0.5)
            ).size_bytes
            == 123
        )
        assert computer_use.screenshot.take_compressed_region(region).size_bytes == 456

        api_client.take_screenshot.assert_called_once_with(show_cursor=True)
        api_client.take_region_screenshot.assert_called_once_with(height=200, width=300, y=20, x=10, show_cursor=False)
        compressed_kwargs = api_client.take_compressed_screenshot.call_args_list[1].kwargs
        assert compressed_kwargs == {"scale": 0.5, "quality": 90, "format": "jpeg", "show_cursor": True}
        compressed_region_kwargs = api_client.take_compressed_region_screenshot.call_args.kwargs
        assert compressed_region_kwargs["width"] == 300

    def test_display_methods_delegate_to_api(self):
        computer_use, api_client = _make_computer_use()
        api_client.get_display_info.return_value = MagicMock(total_displays=1)
        api_client.get_windows.return_value = MagicMock(count=2)

        assert computer_use.display.get_info().total_displays == 1
        assert computer_use.display.get_windows().count == 2

    def test_recording_service_methods_delegate_to_api(self, tmp_path: Path):
        computer_use, api_client = _make_computer_use()
        api_client.start_recording.return_value = MagicMock(id="rec-1")
        api_client.stop_recording.return_value = MagicMock(id="rec-1")
        api_client.list_recordings.return_value = MagicMock(recordings=[])
        api_client.get_recording.return_value = MagicMock(id="rec-1")

        assert computer_use.recording.start("label").id == "rec-1"
        assert computer_use.recording.stop("rec-1").id == "rec-1"
        assert computer_use.recording.list().recordings == []
        assert computer_use.recording.get("rec-1").id == "rec-1"
        computer_use.recording.delete("rec-1")

        assert api_client.start_recording.call_args.kwargs["request"].label == "label"
        assert api_client.stop_recording.call_args.kwargs["request"].id == "rec-1"
        api_client.delete_recording.assert_called_once_with(id="rec-1")

        stream_response = MagicMock()
        stream_response.__enter__.return_value = stream_response
        stream_response.__exit__.return_value = False
        stream_response.iter_bytes.return_value = iter([b"part1", b"part2"])
        client = MagicMock()
        client.__enter__.return_value = client
        client.__exit__.return_value = False
        client.stream.return_value = stream_response
        api_client._download_recording_serialize.return_value = (
            "GET",
            "https://download",
            {"Authorization": "Bearer token"},
            None,
        )
        destination = tmp_path / "nested" / "recording.mp4"

        computer_use.recording._http_client = client
        computer_use.recording.download("rec-1", str(destination))

        assert destination.read_bytes() == b"part1part2"
        from daytona.internal.http_client import request_timeout

        client.stream.assert_called_once_with(
            "GET", "https://download", headers={"Authorization": "Bearer token"}, timeout=request_timeout(30 * 60)
        )

    def test_top_level_methods_delegate_to_api(self):
        computer_use, api_client = _make_computer_use()
        api_client.start_computer_use.return_value = MagicMock(message="started")
        api_client.stop_computer_use.return_value = MagicMock(message="stopped")
        api_client.get_computer_use_status.return_value = MagicMock(status="running")
        api_client.get_process_status.return_value = MagicMock()
        api_client.get_process_status.return_value.name = "xvfb"
        api_client.restart_process.return_value = MagicMock(message="restarted")
        api_client.get_process_logs.return_value = MagicMock(logs=["line"])
        api_client.get_process_errors.return_value = MagicMock(errors=["err"])

        assert computer_use.start().message == "started"
        assert computer_use.stop().message == "stopped"
        assert computer_use.get_status().status == "running"
        assert computer_use.get_process_status("xvfb").name == "xvfb"
        assert computer_use.restart_process("xvfb").message == "restarted"
        assert computer_use.get_process_logs("xvfb").logs == ["line"]
        assert computer_use.get_process_errors("xvfb").errors == ["err"]

        api_client.get_process_status.assert_called_once_with(process_name="xvfb")
        api_client.restart_process.assert_called_once_with(process_name="xvfb")

    def test_accessibility_methods_delegate_to_api(self):
        computer_use, api_client = _make_computer_use()
        api_client.get_accessibility_tree.return_value = MagicMock(root=MagicMock(id="root"))
        api_client.find_accessibility_nodes.return_value = MagicMock(matches=[MagicMock(id="node-1")])

        assert computer_use.accessibility.get_tree().root.id == "root"
        assert computer_use.accessibility.get_tree(scope="pid", pid=123, max_depth=0).root.id == "root"
        assert (
            computer_use.accessibility.find_nodes(
                scope="all",
                role="button",
                name="Submit",
                name_match="exact",
                states=["visible"],
                limit=0,
            )
            .matches[0]
            .id
            == "node-1"
        )
        computer_use.accessibility.focus_node("node-1")
        computer_use.accessibility.invoke_node("node-1")
        computer_use.accessibility.invoke_node("node-2", action="click")
        computer_use.accessibility.set_node_value("node-3", "hello")

        api_client.get_accessibility_tree.assert_any_call(scope=None, pid=None, max_depth=None)
        api_client.get_accessibility_tree.assert_any_call(scope="pid", pid=123, max_depth=0)
        find_request = api_client.find_accessibility_nodes.call_args.kwargs["request"]
        assert find_request.to_dict() == {
            "scope": "all",
            "role": "button",
            "name": "Submit",
            "nameMatch": "exact",
            "states": ["visible"],
            "limit": 0,
        }
        focus_request = api_client.focus_accessibility_node.call_args.kwargs["request"]
        assert focus_request.id == "node-1"
        first_invoke = api_client.invoke_accessibility_node.call_args_list[0].kwargs["request"]
        second_invoke = api_client.invoke_accessibility_node.call_args_list[1].kwargs["request"]
        assert first_invoke.to_dict() == {"id": "node-1"}
        assert second_invoke.to_dict() == {"action": "click", "id": "node-2"}
        value_request = api_client.set_accessibility_node_value.call_args.kwargs["request"]
        assert value_request.to_dict() == {"id": "node-3", "value": "hello"}
