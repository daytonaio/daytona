# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import os
from collections.abc import Callable

import httpx
from daytona_toolbox_api_client import (
    ComputerUseApi,
    ComputerUseStartResponse,
    ComputerUseStatusResponse,
    ComputerUseStopResponse,
    DisplayInfoResponse,
    KeyboardHotkeyRequest,
    KeyboardPressRequest,
    KeyboardTypeRequest,
    ListRecordingsResponse,
    MouseClickRequest,
    MouseClickResponse,
    MouseDragRequest,
    MouseDragResponse,
    MouseMoveRequest,
    MousePositionResponse,
    MouseScrollRequest,
    ProcessErrorsResponse,
    ProcessLogsResponse,
    ProcessRestartResponse,
    ProcessStatusResponse,
    Recording,
    ScreenshotResponse,
    StartRecordingRequest,
    StopRecordingRequest,
    WindowsResponse,
)

from .._utils.errors import intercept_errors
from .._utils.otel_decorator import with_instrumentation
from ..common.computer_use import ScreenshotOptions, ScreenshotRegion


class Mouse:
    """Mouse operations for computer use functionality."""

    def __init__(self, api_client: ComputerUseApi):
        self._api_client: ComputerUseApi = api_client

    @intercept_errors(message_prefix="Failed to get mouse position: ")
    @with_instrumentation()
    def get_position(self) -> MousePositionResponse:
        """Gets the current mouse cursor position.

        Returns:
            MousePositionResponse: Current mouse position with x and y coordinates.

        Example:
            ```python
            position = sandbox.computer_use.mouse.get_position()
            print(f"Mouse is at: {position.x}, {position.y}")
            ```
        """
        response = self._api_client.get_mouse_position()
        return response

    @intercept_errors(message_prefix="Failed to move mouse: ")
    @with_instrumentation()
    def move(self, x: int, y: int) -> MousePositionResponse:
        """Moves the mouse cursor to the specified coordinates.

        Args:
            x (int): The x coordinate to move to.
            y (int): The y coordinate to move to.

        Returns:
            MousePositionResponse: Position after move.

        Example:
            ```python
            result = sandbox.computer_use.mouse.move(100, 200)
            print(f"Mouse moved to: {result.x}, {result.y}")
            ```
        """
        request = MouseMoveRequest(x=x, y=y)
        response = self._api_client.move_mouse(request)
        return response

    @intercept_errors(message_prefix="Failed to click mouse: ")
    @with_instrumentation()
    def click(self, x: int, y: int, button: str = "left", double: bool = False) -> MouseClickResponse:
        """Clicks the mouse at the specified coordinates.

        Args:
            x (int): The x coordinate to click at.
            y (int): The y coordinate to click at.
            button (str): The mouse button to click ('left', 'right', 'middle').
            double (bool): Whether to perform a double-click.

        Returns:
            MouseClickResponse: Click operation result.

        Example:
            ```python
            # Single left click
            result = sandbox.computer_use.mouse.click(100, 200)

            # Double click
            double_click = sandbox.computer_use.mouse.click(100, 200, "left", True)

            # Right click
            right_click = sandbox.computer_use.mouse.click(100, 200, "right")
            ```
        """
        request = MouseClickRequest(x=x, y=y, button=button, double=double)
        response = self._api_client.click(request)
        return response

    @intercept_errors(message_prefix="Failed to drag mouse: ")
    @with_instrumentation()
    def drag(self, start_x: int, start_y: int, end_x: int, end_y: int, button: str = "left") -> MouseDragResponse:
        """Drags the mouse from start coordinates to end coordinates.

        Args:
            start_x (int): The starting x coordinate.
            start_y (int): The starting y coordinate.
            end_x (int): The ending x coordinate.
            end_y (int): The ending y coordinate.
            button (str): The mouse button to use for dragging.

        Returns:
            MouseDragResponse: Drag operation result.

        Example:
            ```python
            result = sandbox.computer_use.mouse.drag(50, 50, 150, 150)
            print(f"Dragged from {result.from_x},{result.from_y} to {result.to_x},{result.to_y}")
            ```
        """
        request = MouseDragRequest(start_x=start_x, start_y=start_y, end_x=end_x, end_y=end_y, button=button)
        response = self._api_client.drag(request=request)
        return response

    @intercept_errors(message_prefix="Failed to scroll mouse: ")
    @with_instrumentation()
    def scroll(self, x: int, y: int, direction: str, amount: int = 1) -> bool:
        """Scrolls the mouse wheel at the specified coordinates.

        Args:
            x (int): The x coordinate to scroll at.
            y (int): The y coordinate to scroll at.
            direction (str): The direction to scroll ('up' or 'down').
            amount (int): The amount to scroll.

        Returns:
            bool: Whether the scroll operation was successful.

        Example:
            ```python
            # Scroll up
            scroll_up = sandbox.computer_use.mouse.scroll(100, 200, "up", 3)

            # Scroll down
            scroll_down = sandbox.computer_use.mouse.scroll(100, 200, "down", 5)
            ```
        """
        request = MouseScrollRequest(x=x, y=y, direction=direction, amount=amount)
        response = self._api_client.scroll(request=request)
        return response.success is True


class Keyboard:
    """Keyboard operations for computer use functionality."""

    def __init__(self, api_client: ComputerUseApi):
        self._api_client: ComputerUseApi = api_client

    @intercept_errors(message_prefix="Failed to type text: ")
    @with_instrumentation()
    def type(self, text: str, delay: int | None = None) -> None:
        """Types the specified text.

        Args:
            text (str): The text to type.
            delay (int): Delay between characters in milliseconds.

        Raises:
            DaytonaError: If the type operation fails.

        Example:
            ```python
            try:
                sandbox.computer_use.keyboard.type("Hello, World!")
                print(f"Operation success")
            except Exception as e:
                print(f"Operation failed: {e}")

            # With delay between characters
            try:
                sandbox.computer_use.keyboard.type("Slow typing", 100)
                print(f"Operation success")
            except Exception as e:
                print(f"Operation failed: {e}")
            ```
        """
        request = KeyboardTypeRequest(text=text, delay=delay)
        _ = self._api_client.type_text(request=request)

    @intercept_errors(message_prefix="Failed to press key: ")
    @with_instrumentation()
    def press(self, key: str, modifiers: list[str] | None = None) -> None:
        """Presses a key with optional modifiers.

        Args:
            key (str): The key to press (e.g., 'Enter', 'Escape', 'Tab', 'a', 'A').
            modifiers (list[str]): Modifier keys ('ctrl', 'alt', 'meta', 'shift').

        Raises:
            DaytonaError: If the press operation fails.

        Example:
            ```python
            # Press Enter
            try:
                sandbox.computer_use.keyboard.press("Return")
                print(f"Operation success")
            except Exception as e:
                print(f"Operation failed: {e}")

            # Press Ctrl+C
            try:
                sandbox.computer_use.keyboard.press("c", ["ctrl"])
                print(f"Operation success")

            # Press Ctrl+Shift+T
            try:
                sandbox.computer_use.keyboard.press("t", ["ctrl", "shift"])
                print(f"Operation success")
            except Exception as e:
                print(f"Operation failed: {e}")
            ```
        """
        request = KeyboardPressRequest(key=key, modifiers=modifiers or [])
        _ = self._api_client.press_key(request=request)

    @intercept_errors(message_prefix="Failed to press hotkey: ")
    @with_instrumentation()
    def hotkey(self, keys: str) -> None:
        """Presses a hotkey combination.

        Args:
            keys (str): The hotkey combination (e.g., 'ctrl+c', 'alt+tab', 'cmd+shift+t').

        Raises:
            DaytonaError: If the hotkey operation fails.

        Example:
            ```python
            # Copy
            try:
                sandbox.computer_use.keyboard.hotkey("ctrl+c")
                print(f"Operation success")
            except Exception as e:
                print(f"Operation failed: {e}")

            # Paste
            try:
                sandbox.computer_use.keyboard.hotkey("ctrl+v")
                print(f"Operation success")
            except Exception as e:
                print(f"Operation failed: {e}")

            # Alt+Tab
            try:
                sandbox.computer_use.keyboard.hotkey("alt+tab")
                print(f"Operation success")
            except Exception as e:
                print(f"Operation failed: {e}")
            ```
        """
        request = KeyboardHotkeyRequest(keys=keys)
        _ = self._api_client.press_hotkey(request=request)


class Screenshot:
    """Screenshot operations for computer use functionality."""

    def __init__(self, api_client: ComputerUseApi):
        self._api_client: ComputerUseApi = api_client

    @intercept_errors(message_prefix="Failed to take screenshot: ")
    @with_instrumentation()
    def take_full_screen(self, show_cursor: bool = False) -> ScreenshotResponse:
        """Takes a screenshot of the entire screen.

        Args:
            show_cursor (bool): Whether to show the cursor in the screenshot.

        Returns:
            ScreenshotResponse: Screenshot data with base64 encoded image.

        Example:
            ```python
            screenshot = sandbox.computer_use.screenshot.take_full_screen()
            print(f"Screenshot size: {screenshot.width}x{screenshot.height}")

            # With cursor visible
            with_cursor = sandbox.computer_use.screenshot.take_full_screen(True)
            ```
        """
        response = self._api_client.take_screenshot(show_cursor=show_cursor)
        return response

    @intercept_errors(message_prefix="Failed to take region screenshot: ")
    @with_instrumentation()
    def take_region(self, region: ScreenshotRegion, show_cursor: bool = False) -> ScreenshotResponse:
        """Takes a screenshot of a specific region.

        Args:
            region (ScreenshotRegion): The region to capture.
            show_cursor (bool): Whether to show the cursor in the screenshot.

        Returns:
            ScreenshotResponse: Screenshot data with base64 encoded image.

        Example:
            ```python
            region = ScreenshotRegion(x=100, y=100, width=300, height=200)
            screenshot = sandbox.computer_use.screenshot.take_region(region)
            print(f"Captured region: {screenshot.region.width}x{screenshot.region.height}")
            ```
        """
        response = self._api_client.take_region_screenshot(
            height=region.height, width=region.width, y=region.y, x=region.x, show_cursor=show_cursor
        )
        return response

    @intercept_errors(message_prefix="Failed to take compressed screenshot: ")
    @with_instrumentation()
    def take_compressed(self, options: ScreenshotOptions | None = None) -> ScreenshotResponse:
        """Takes a compressed screenshot of the entire screen.

        Args:
            options (ScreenshotOptions | None): Compression and display options.

        Returns:
            ScreenshotResponse: Compressed screenshot data.

        Example:
            ```python
            # Default compression
            screenshot = sandbox.computer_use.screenshot.take_compressed()

            # High quality JPEG
            jpeg = sandbox.computer_use.screenshot.take_compressed(
                ScreenshotOptions(format="jpeg", quality=95, show_cursor=True)
            )

            # Scaled down PNG
            scaled = sandbox.computer_use.screenshot.take_compressed(
                ScreenshotOptions(format="png", scale=0.5)
            )
            ```
        """
        if options is None:
            options = ScreenshotOptions()

        response = self._api_client.take_compressed_screenshot(
            scale=options.scale,
            quality=options.quality,
            format=options.fmt,
            show_cursor=options.show_cursor,
        )
        return response

    @intercept_errors(message_prefix="Failed to take compressed region screenshot: ")
    @with_instrumentation()
    def take_compressed_region(
        self, region: ScreenshotRegion, options: ScreenshotOptions | None = None
    ) -> ScreenshotResponse:
        """Takes a compressed screenshot of a specific region.

        Args:
            region (ScreenshotRegion): The region to capture.
            options (ScreenshotOptions | None): Compression and display options.

        Returns:
            ScreenshotResponse: Compressed screenshot data.

        Example:
            ```python
            region = ScreenshotRegion(x=0, y=0, width=800, height=600)
            screenshot = sandbox.computer_use.screenshot.take_compressed_region(
                region,
                ScreenshotOptions(format="webp", quality=80, show_cursor=True)
            )
            print(f"Compressed size: {screenshot.size_bytes} bytes")
            ```
        """
        if options is None:
            options = ScreenshotOptions()

        response = self._api_client.take_compressed_region_screenshot(
            height=region.height,
            width=region.width,
            y=region.y,
            x=region.x,
            scale=options.scale,
            quality=options.quality,
            format=options.fmt,
            show_cursor=options.show_cursor,
        )
        return response


class Display:
    """Display operations for computer use functionality."""

    def __init__(self, api_client: ComputerUseApi):
        self._api_client: ComputerUseApi = api_client

    @intercept_errors(message_prefix="Failed to get display info: ")
    @with_instrumentation()
    def get_info(self) -> DisplayInfoResponse:
        """Gets information about the displays.

        Returns:
            DisplayInfoResponse: Display information including primary display and all available displays.

        Example:
            ```python
            info = sandbox.computer_use.display.get_info()
            print(f"Primary display: {info.primary_display.width}x{info.primary_display.height}")
            print(f"Total displays: {info.total_displays}")
            for i, display in enumerate(info.displays):
                print(f"Display {i}: {display.width}x{display.height} at {display.x},{display.y}")
            ```
        """
        response = self._api_client.get_display_info()
        return response

    @intercept_errors(message_prefix="Failed to get windows: ")
    @with_instrumentation()
    def get_windows(self) -> WindowsResponse:
        """Gets the list of open windows.

        Returns:
            WindowsResponse: List of open windows with their IDs and titles.

        Example:
            ```python
            windows = sandbox.computer_use.display.get_windows()
            print(f"Found {windows.count} open windows:")
            for window in windows.windows:
                print(f"- {window.title} (ID: {window.id})")
            ```
        """
        response = self._api_client.get_windows()
        return response


class RecordingService:
    """Recording operations for computer use functionality."""

    def __init__(
        self,
        api_client: ComputerUseApi,
        ensure_toolbox_url: Callable[[], None],
    ):
        self._api_client: ComputerUseApi = api_client
        self._ensure_toolbox_url: Callable[[], None] = ensure_toolbox_url

    @intercept_errors(message_prefix="Failed to start recording: ")
    @with_instrumentation()
    def start(self, label: str | None = None) -> Recording:
        """Starts a new screen recording session.

        Args:
            label (str | None): Optional custom label for the recording.

        Returns:
            Recording: Recording start response.

        Example:
            ```python
            # Start a recording with a label
            recording = sandbox.computer_use.recording.start("my-test-recording")
            print(f"Recording started: {recording.id}")
            print(f"File: {recording.file_path}")
            ```
        """
        request = StartRecordingRequest(label=label)
        return self._api_client.start_recording(request=request)

    @intercept_errors(message_prefix="Failed to stop recording: ")
    @with_instrumentation()
    def stop(self, recording_id: str) -> Recording:
        """Stops an active screen recording session.

        Args:
            recording_id (str): The ID of the recording to stop.

        Returns:
            Recording: Recording stop response.

        Example:
            ```python
            result = sandbox.computer_use.recording.stop(recording.id)
            print(f"Recording stopped: {result.duration_seconds} seconds")
            print(f"Saved to: {result.file_path}")
            ```
        """
        request = StopRecordingRequest(id=recording_id)
        return self._api_client.stop_recording(request=request)

    @intercept_errors(message_prefix="Failed to list recordings: ")
    @with_instrumentation()
    def list(self) -> ListRecordingsResponse:
        """Lists all recordings (active and completed).

        Returns:
            ListRecordingsResponse: List of all recordings.

        Example:
            ```python
            recordings = sandbox.computer_use.recording.list()
            print(f"Found {len(recordings.recordings)} recordings")
            for rec in recordings.recordings:
                print(f"- {rec.file_name}: {rec.status}")
            ```
        """
        return self._api_client.list_recordings()

    @intercept_errors(message_prefix="Failed to get recording: ")
    @with_instrumentation()
    def get(self, recording_id: str) -> Recording:
        """Gets details of a specific recording by ID.

        Args:
            recording_id (str): The ID of the recording to retrieve.

        Returns:
            Recording: Recording details.

        Example:
            ```python
            recording = sandbox.computer_use.recording.get(recording_id)
            print(f"Recording: {recording.file_name}")
            print(f"Status: {recording.status}")
            print(f"Duration: {recording.duration_seconds} seconds")
            ```
        """
        return self._api_client.get_recording(id=recording_id)

    @intercept_errors(message_prefix="Failed to delete recording: ")
    @with_instrumentation()
    def delete(self, recording_id: str) -> None:
        """Deletes a recording by ID.

        Args:
            recording_id (str): The ID of the recording to delete.

        Example:
            ```python
            sandbox.computer_use.recording.delete(recording_id)
            print("Recording deleted")
            ```
        """
        self._api_client.delete_recording(id=recording_id)

    @intercept_errors(message_prefix="Failed to download recording: ")
    @with_instrumentation()
    def download(self, recording_id: str, local_path: str) -> None:
        """Downloads a recording file from the Sandbox and saves it to a local file.

        The file is streamed directly to disk without loading the entire content into memory.

        Args:
            recording_id (str): The ID of the recording to download.
            local_path (str): Path to save the recording file locally.

        Example:
            ```python
            # Download recording to file
            sandbox.computer_use.recording.download(recording_id, "local_recording.mp4")
            print("Recording downloaded")
            ```
        """
        # Ensure the toolbox URL is loaded before making the request
        self._ensure_toolbox_url()

        # Serialize the request to get the URL and headers
        method, url, headers, *_ = self._api_client._download_recording_serialize(
            id=recording_id,
            _request_auth=None,
            _content_type=None,
            _headers=None,
            _host_index=None,
        )

        # Create parent directory if it doesn't exist
        parent_dir = os.path.dirname(os.path.abspath(local_path))
        if parent_dir:
            os.makedirs(parent_dir, exist_ok=True)

        # Stream the download directly to file
        with httpx.Client(timeout=30 * 60) as client:
            with client.stream(method, url, headers=headers) as response:
                _ = response.raise_for_status()

                with open(local_path, "wb") as f:
                    for chunk in response.iter_bytes(64 * 1024):
                        _ = f.write(chunk)


class ComputerUse:
    """Computer Use functionality for interacting with the desktop environment.

    Provides access to mouse, keyboard, screenshot, display, and recording operations
    for automating desktop interactions within a sandbox.

    Attributes:
        mouse (Mouse): Mouse operations interface.
        keyboard (Keyboard): Keyboard operations interface.
        screenshot (Screenshot): Screenshot operations interface.
        display (Display): Display operations interface.
        recording (RecordingService): Screen recording operations interface.
    """

    def __init__(
        self,
        api_client: ComputerUseApi,
        ensure_toolbox_url: Callable[[], None],
    ):
        self._api_client: ComputerUseApi = api_client

        self.mouse: Mouse = Mouse(api_client)
        self.keyboard: Keyboard = Keyboard(api_client)
        self.screenshot: Screenshot = Screenshot(api_client)
        self.display: Display = Display(api_client)
        self.recording: RecordingService = RecordingService(api_client, ensure_toolbox_url)

    @intercept_errors(message_prefix="Failed to start computer use: ")
    @with_instrumentation()
    def start(self) -> ComputerUseStartResponse:
        """Starts all computer use processes (Xvfb, xfce4, x11vnc, novnc).

        Returns:
            ComputerUseStartResponse: Computer use start response.

        Example:
            ```python
            result = sandbox.computer_use.start()
            print("Computer use processes started:", result.message)
            ```
        """
        response = self._api_client.start_computer_use()
        return response

    @intercept_errors(message_prefix="Failed to stop computer use: ")
    @with_instrumentation()
    def stop(self) -> ComputerUseStopResponse:
        """Stops all computer use processes.

        Returns:
            ComputerUseStopResponse: Computer use stop response.

        Example:
            ```python
            result = sandbox.computer_use.stop()
            print("Computer use processes stopped:", result.message)
            ```
        """
        response = self._api_client.stop_computer_use()
        return response

    @intercept_errors(message_prefix="Failed to get computer use status: ")
    @with_instrumentation()
    def get_status(self) -> ComputerUseStatusResponse:
        """Gets the status of all computer use processes.

        Returns:
            ComputerUseStatusResponse: Status information about all VNC desktop processes.

        Example:
            ```python
            response = sandbox.computer_use.get_status()
            print("Computer use status:", response.status)
            ```
        """
        return self._api_client.get_computer_use_status()

    @intercept_errors(message_prefix="Failed to get process status: ")
    @with_instrumentation()
    def get_process_status(self, process_name: str) -> ProcessStatusResponse:
        """Gets the status of a specific VNC process.

        Args:
            process_name (str): Name of the process to check.

        Returns:
            ProcessStatusResponse: Status information about the specific process.

        Example:
            ```python
            xvfb_status = sandbox.computer_use.get_process_status("xvfb")
            no_vnc_status = sandbox.computer_use.get_process_status("novnc")
            ```
        """
        response = self._api_client.get_process_status(process_name=process_name)
        return response

    @intercept_errors(message_prefix="Failed to restart process: ")
    @with_instrumentation()
    def restart_process(self, process_name: str) -> ProcessRestartResponse:
        """Restarts a specific VNC process.

        Args:
            process_name (str): Name of the process to restart.

        Returns:
            ProcessRestartResponse: Process restart response.

        Example:
            ```python
            result = sandbox.computer_use.restart_process("xfce4")
            print("XFCE4 process restarted:", result.message)
            ```
        """
        response = self._api_client.restart_process(process_name=process_name)
        return response

    @intercept_errors(message_prefix="Failed to get process logs: ")
    @with_instrumentation()
    def get_process_logs(self, process_name: str) -> ProcessLogsResponse:
        """Gets logs for a specific VNC process.

        Args:
            process_name (str): Name of the process to get logs for.

        Returns:
            ProcessLogsResponse: Process logs.

        Example:
            ```python
            logs = sandbox.computer_use.get_process_logs("novnc")
            print("NoVNC logs:", logs)
            ```
        """
        response = self._api_client.get_process_logs(process_name=process_name)
        return response

    @intercept_errors(message_prefix="Failed to get process errors: ")
    @with_instrumentation()
    def get_process_errors(self, process_name: str) -> ProcessErrorsResponse:
        """Gets error logs for a specific VNC process.

        Args:
            process_name (str): Name of the process to get error logs for.

        Returns:
            ProcessErrorsResponse: Process error logs.

        Example:
            ```python
            errors = sandbox.computer_use.get_process_errors("x11vnc")
            print("X11VNC errors:", errors)
            ```
        """
        response = self._api_client.get_process_errors(process_name=process_name)
        return response
