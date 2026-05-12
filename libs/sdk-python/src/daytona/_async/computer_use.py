# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import os

import aiofiles

from daytona_toolbox_api_client_async import (
    AccessibilityInvokeRequest,
    AccessibilityNodeRequest,
    AccessibilityNodesResponse,
    AccessibilitySetValueRequest,
    AccessibilityTreeResponse,
    ComputerUseApi,
    ComputerUseStartResponse,
    ComputerUseStatusResponse,
    ComputerUseStopResponse,
    DisplayInfoResponse,
    FindAccessibilityNodesRequest,
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
from ..internal.http_client import aiohttp_request_timeout as _request_timeout
from ..internal.shared_session import http_session_of


class AsyncMouse:
    """Mouse operations for computer use functionality."""

    def __init__(self, api_client: ComputerUseApi):
        self._api_client: ComputerUseApi = api_client

    @intercept_errors(message_prefix="Failed to get mouse position: ")
    @with_instrumentation()
    async def get_position(self) -> MousePositionResponse:
        """Gets the current mouse cursor position.

        Returns:
            MousePositionResponse: Current mouse position with x and y coordinates.

        Example:
            ```python
            position = await sandbox.computer_use.mouse.get_position()
            print(f"Mouse is at: {position.x}, {position.y}")
            ```
        """
        response = await self._api_client.get_mouse_position()
        return response

    @intercept_errors(message_prefix="Failed to move mouse: ")
    @with_instrumentation()
    async def move(self, x: int, y: int) -> MousePositionResponse:
        """Moves the mouse cursor to the specified coordinates.

        Args:
            x (int): The x coordinate to move to.
            y (int): The y coordinate to move to.

        Returns:
            MousePositionResponse: Position after move.

        Example:
            ```python
            result = await sandbox.computer_use.mouse.move(100, 200)
            print(f"Mouse moved to: {result.x}, {result.y}")
            ```
        """
        request = MouseMoveRequest(x=x, y=y)
        response = await self._api_client.move_mouse(request)
        return response

    @intercept_errors(message_prefix="Failed to click mouse: ")
    @with_instrumentation()
    async def click(self, x: int, y: int, button: str = "left", double: bool = False) -> MouseClickResponse:
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
            result = await sandbox.computer_use.mouse.click(100, 200)

            # Double click
            double_click = await sandbox.computer_use.mouse.click(100, 200, "left", True)

            # Right click
            right_click = await sandbox.computer_use.mouse.click(100, 200, "right")
            ```
        """
        request = MouseClickRequest(x=x, y=y, button=button, double=double)
        response = await self._api_client.click(request)
        return response

    @intercept_errors(message_prefix="Failed to drag mouse: ")
    @with_instrumentation()
    async def drag(self, start_x: int, start_y: int, end_x: int, end_y: int, button: str = "left") -> MouseDragResponse:
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
            result = await sandbox.computer_use.mouse.drag(50, 50, 150, 150)
            print(f"Drag ended at {result.x}, {result.y}")
            ```
        """
        request = MouseDragRequest(start_x=start_x, start_y=start_y, end_x=end_x, end_y=end_y, button=button)
        response = await self._api_client.drag(request=request)
        return response

    @intercept_errors(message_prefix="Failed to scroll mouse: ")
    @with_instrumentation()
    async def scroll(self, x: int, y: int, direction: str, amount: int = 1) -> bool:
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
            scroll_up = await sandbox.computer_use.mouse.scroll(100, 200, "up", 3)

            # Scroll down
            scroll_down = await sandbox.computer_use.mouse.scroll(100, 200, "down", 5)
            ```
        """
        request = MouseScrollRequest(x=x, y=y, direction=direction, amount=amount)
        response = await self._api_client.scroll(request=request)
        return response.success is True


class AsyncKeyboard:
    """Keyboard operations for computer use functionality."""

    def __init__(self, api_client: ComputerUseApi):
        self._api_client: ComputerUseApi = api_client

    @intercept_errors(message_prefix="Failed to type text: ")
    @with_instrumentation()
    async def type(self, text: str, delay: int | None = None) -> None:
        """Types the specified text.

        Args:
            text (str): The text to type.
            delay (int): Delay between characters in milliseconds.

        Raises:
            DaytonaError: If the type operation fails.

        Example:
            ```python
            try:
                await sandbox.computer_use.keyboard.type("Hello, World!")
                print(f"Operation success")
            except Exception as e:
                print(f"Operation failed: {e}")

            # With delay between characters
            try:
                await sandbox.computer_use.keyboard.type("Slow typing", 100)
                print(f"Operation success")
            except Exception as e:
                print(f"Operation failed: {e}")
            ```
        """
        request = KeyboardTypeRequest(text=text, delay=delay)
        _ = await self._api_client.type_text(request=request)

    @intercept_errors(message_prefix="Failed to press key: ")
    @with_instrumentation()
    async def press(self, key: str, modifiers: list[str] | None = None) -> None:
        """Presses a key with optional modifiers.

        Args:
            key (str): The key to press. Canonical names include 'enter', 'escape',
                'tab', letters, digits, unshifted punctuation, function keys, and
                grammar-safe numpad names such as 'num_plus'. Named keys are
                case-insensitive, and common aliases such as 'Return' and 'Escape'
                are normalized.
            modifiers (list[str]): Canonical modifier names are 'ctrl', 'alt',
                'shift', and 'cmd'. Common aliases such as 'control', 'option',
                'meta', and 'win' are normalized.

        Raises:
            DaytonaError: If the press operation fails.

        Example:
            ```python
            # Press Enter
            try:
                await sandbox.computer_use.keyboard.press("enter")
                print(f"Operation success")
            except Exception as e:
                print(f"Operation failed: {e}")

            # Press Ctrl+C
            try:
                await sandbox.computer_use.keyboard.press("c", ["ctrl"])
                print(f"Operation success")

            # Press Ctrl+Shift+T
            try:
                await sandbox.computer_use.keyboard.press("t", ["ctrl", "shift"])
                print(f"Operation success")
            except Exception as e:
                print(f"Operation failed: {e}")
            ```
        """
        request = KeyboardPressRequest(key=key, modifiers=modifiers or [])
        _ = await self._api_client.press_key(request=request)

    @intercept_errors(message_prefix="Failed to press hotkey: ")
    @with_instrumentation()
    async def hotkey(self, keys: str) -> None:
        """Presses a hotkey combination.

        Args:
            keys (str): A single atomic hotkey chord (e.g., 'ctrl+c', 'alt+tab',
                'cmd+shift+t', 'ctrl + c', 'shift'). Uses the same normalized key
                contract as ``press()``.

        Raises:
            DaytonaError: If the hotkey operation fails.

        Example:
            ```python
            # Copy
            try:
                await sandbox.computer_use.keyboard.hotkey("ctrl+c")
                print(f"Operation success")
            except Exception as e:
                print(f"Operation failed: {e}")

            # Paste
            try:
                await sandbox.computer_use.keyboard.hotkey("ctrl+v")
                print(f"Operation success")
            except Exception as e:
                print(f"Operation failed: {e}")

            # Alt+Tab
            try:
                await sandbox.computer_use.keyboard.hotkey("alt+tab")
                print(f"Operation success")
            except Exception as e:
                print(f"Operation failed: {e}")
            ```
        """
        request = KeyboardHotkeyRequest(keys=keys)
        _ = await self._api_client.press_hotkey(request=request)


class AsyncScreenshot:
    """Screenshot operations for computer use functionality."""

    def __init__(self, api_client: ComputerUseApi):
        self._api_client: ComputerUseApi = api_client

    @intercept_errors(message_prefix="Failed to take screenshot: ")
    @with_instrumentation()
    async def take_full_screen(self, show_cursor: bool = False) -> ScreenshotResponse:
        """Takes a screenshot of the entire screen.

        Args:
            show_cursor (bool): Whether to show the cursor in the screenshot.

        Returns:
            ScreenshotResponse: Screenshot data with base64 encoded image.

        Example:
            ```python
            screenshot = await sandbox.computer_use.screenshot.take_full_screen()
            print(f"Screenshot size: {screenshot.width}x{screenshot.height}")

            # With cursor visible
            with_cursor = await sandbox.computer_use.screenshot.take_full_screen(True)
            ```
        """
        response = await self._api_client.take_screenshot(show_cursor=show_cursor)
        return response

    @intercept_errors(message_prefix="Failed to take region screenshot: ")
    @with_instrumentation()
    async def take_region(self, region: ScreenshotRegion, show_cursor: bool = False) -> ScreenshotResponse:
        """Takes a screenshot of a specific region.

        Args:
            region (ScreenshotRegion): The region to capture.
            show_cursor (bool): Whether to show the cursor in the screenshot.

        Returns:
            ScreenshotResponse: Screenshot data with base64 encoded image.

        Example:
            ```python
            region = ScreenshotRegion(x=100, y=100, width=300, height=200)
            screenshot = await sandbox.computer_use.screenshot.take_region(region)
            print(f"Captured region: {screenshot.region.width}x{screenshot.region.height}")
            ```
        """
        response = await self._api_client.take_region_screenshot(
            height=region.height, width=region.width, y=region.y, x=region.x, show_cursor=show_cursor
        )
        return response

    @intercept_errors(message_prefix="Failed to take compressed screenshot: ")
    @with_instrumentation()
    async def take_compressed(self, options: ScreenshotOptions | None = None) -> ScreenshotResponse:
        """Takes a compressed screenshot of the entire screen.

        Args:
            options (ScreenshotOptions | None): Compression and display options.

        Returns:
            ScreenshotResponse: Compressed screenshot data.

        Example:
            ```python
            # Default compression
            screenshot = await sandbox.computer_use.screenshot.take_compressed()

            # High quality JPEG
            jpeg = await sandbox.computer_use.screenshot.take_compressed(
                ScreenshotOptions(format="jpeg", quality=95, show_cursor=True)
            )

            # Scaled down PNG
            scaled = await sandbox.computer_use.screenshot.take_compressed(
                ScreenshotOptions(format="png", scale=0.5)
            )
            ```
        """
        if options is None:
            options = ScreenshotOptions()

        response = await self._api_client.take_compressed_screenshot(
            scale=options.scale,
            quality=options.quality,
            format=options.fmt,
            show_cursor=options.show_cursor,
        )
        return response

    @intercept_errors(message_prefix="Failed to take compressed region screenshot: ")
    @with_instrumentation()
    async def take_compressed_region(
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
            screenshot = await sandbox.computer_use.screenshot.take_compressed_region(
                region,
                ScreenshotOptions(format="webp", quality=80, show_cursor=True)
            )
            print(f"Compressed size: {screenshot.size_bytes} bytes")
            ```
        """
        if options is None:
            options = ScreenshotOptions()

        response = await self._api_client.take_compressed_region_screenshot(
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


class AsyncDisplay:
    """Display operations for computer use functionality."""

    def __init__(self, api_client: ComputerUseApi):
        self._api_client: ComputerUseApi = api_client

    @intercept_errors(message_prefix="Failed to get display info: ")
    @with_instrumentation()
    async def get_info(self) -> DisplayInfoResponse:
        """Gets information about the displays.

        Returns:
            DisplayInfoResponse: Display information including primary display and all available displays.

        Example:
            ```python
            info = await sandbox.computer_use.display.get_info()
            print(f"Primary display: {info.primary_display.width}x{info.primary_display.height}")
            print(f"Total displays: {info.total_displays}")
            for i, display in enumerate(info.displays):
                print(f"Display {i}: {display.width}x{display.height} at {display.x},{display.y}")
            ```
        """
        response = await self._api_client.get_display_info()
        return response

    @intercept_errors(message_prefix="Failed to get windows: ")
    @with_instrumentation()
    async def get_windows(self) -> WindowsResponse:
        """Gets the list of open windows.

        Returns:
            WindowsResponse: List of open windows with their IDs and titles.

        Example:
            ```python
            windows = await sandbox.computer_use.display.get_windows()
            print(f"Found {windows.count} open windows:")
            for window in windows.windows:
                print(f"- {window.title} (ID: {window.id})")
            ```
        """
        response = await self._api_client.get_windows()
        return response


class AsyncRecordingService:
    """Recording operations for computer use functionality."""

    def __init__(
        self,
        api_client: ComputerUseApi,
    ):
        self._api_client: ComputerUseApi = api_client

    @intercept_errors(message_prefix="Failed to start recording: ")
    @with_instrumentation()
    async def start(self, label: str | None = None) -> Recording:
        """Starts a new screen recording session.

        Args:
            label (str | None): Optional custom label for the recording.

        Returns:
            Recording: Recording start response.

        Example:
            ```python
            # Start a recording with a label
            recording = await sandbox.computer_use.recording.start("my-test-recording")
            print(f"Recording started: {recording.id}")
            print(f"File: {recording.file_path}")
            ```
        """
        request = StartRecordingRequest(label=label)
        return await self._api_client.start_recording(request=request)

    @intercept_errors(message_prefix="Failed to stop recording: ")
    @with_instrumentation()
    async def stop(self, recording_id: str) -> Recording:
        """Stops an active screen recording session.

        Args:
            recording_id (str): The ID of the recording to stop.

        Returns:
            Recording: Recording stop response.

        Example:
            ```python
            result = await sandbox.computer_use.recording.stop(recording.id)
            print(f"Recording stopped: {result.duration_seconds} seconds")
            print(f"Saved to: {result.file_path}")
            ```
        """
        request = StopRecordingRequest(id=recording_id)
        return await self._api_client.stop_recording(request=request)

    @intercept_errors(message_prefix="Failed to list recordings: ")
    @with_instrumentation()
    async def list(self) -> ListRecordingsResponse:
        """Lists all recordings (active and completed).

        Returns:
            ListRecordingsResponse: List of all recordings.

        Example:
            ```python
            recordings = await sandbox.computer_use.recording.list()
            print(f"Found {len(recordings.recordings)} recordings")
            for rec in recordings.recordings:
                print(f"- {rec.file_name}: {rec.status}")
            ```
        """
        return await self._api_client.list_recordings()

    @intercept_errors(message_prefix="Failed to get recording: ")
    @with_instrumentation()
    async def get(self, recording_id: str) -> Recording:
        """Gets details of a specific recording by ID.

        Args:
            recording_id (str): The ID of the recording to retrieve.

        Returns:
            Recording: Recording details.

        Example:
            ```python
            recording = await sandbox.computer_use.recording.get(recording_id)
            print(f"Recording: {recording.file_name}")
            print(f"Status: {recording.status}")
            print(f"Duration: {recording.duration_seconds} seconds")
            ```
        """
        return await self._api_client.get_recording(id=recording_id)

    @intercept_errors(message_prefix="Failed to delete recording: ")
    @with_instrumentation()
    async def delete(self, recording_id: str) -> None:
        """Deletes a recording by ID.

        Args:
            recording_id (str): The ID of the recording to delete.

        Example:
            ```python
            await sandbox.computer_use.recording.delete(recording_id)
            print("Recording deleted")
            ```
        """
        await self._api_client.delete_recording(id=recording_id)

    @intercept_errors(message_prefix="Failed to download recording: ")
    @with_instrumentation()
    async def download(self, recording_id: str, local_path: str) -> None:
        """Downloads a recording file from the Sandbox and saves it to a local file.

        The file is streamed directly to disk without loading the entire content into memory.

        Args:
            recording_id (str): The ID of the recording to download.
            local_path (str): Path to save the recording file locally.

        Example:
            ```python
            # Download recording to file
            await sandbox.computer_use.recording.download(recording_id, "local_recording.mp4")
            print("Recording downloaded")
            ```
        """
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

        download_timeout = 30 * 60
        async with http_session_of(self._api_client.api_client).request(
            method,
            url,
            headers=headers,
            timeout=_request_timeout(download_timeout),
        ) as response:
            response.raise_for_status()

            async with aiofiles.open(local_path, "wb") as f:
                async for chunk in response.content.iter_chunked(64 * 1024):
                    _ = await f.write(chunk)


class AsyncAccessibility:
    """Accessibility operations for computer use functionality.

    This service exposes thin wrappers over the toolbox AT-SPI accessibility
    API. Start computer use before calling these methods.
    """

    def __init__(self, api_client: ComputerUseApi):
        self._api_client: ComputerUseApi = api_client

    @intercept_errors(message_prefix="Failed to get accessibility tree: ")
    @with_instrumentation()
    async def get_tree(
        self,
        scope: str | None = None,
        pid: int | None = None,
        max_depth: int | None = None,
    ) -> AccessibilityTreeResponse:
        """Fetches the AT-SPI accessibility tree.

        Args:
            scope (str | None): Tree scope to inspect: ``focused``, ``pid``, or ``all``.
            pid (int | None): Process ID when ``scope`` is ``pid``.
            max_depth (int | None): Maximum depth to descend. Use ``0`` for the root only.

        Returns:
            AccessibilityTreeResponse: Accessibility tree rooted at the requested scope.

        Example:
            ```python
            tree = await sandbox.computer_use.accessibility.get_tree(scope="all", max_depth=3)
            print(tree.root.name)
            ```
        """
        return await self._api_client.get_accessibility_tree(scope=scope, pid=pid, max_depth=max_depth)

    @intercept_errors(message_prefix="Failed to find accessibility nodes: ")
    @with_instrumentation()
    async def find_nodes(
        self,
        scope: str | None = None,
        pid: int | None = None,
        role: str | None = None,
        name: str | None = None,
        name_match: str | None = None,
        states: list[str] | None = None,
        limit: int | None = None,
    ) -> AccessibilityNodesResponse:
        """Finds AT-SPI accessibility nodes matching the provided filters.

        Args:
            scope (str | None): Search scope: ``focused``, ``pid``, or ``all``.
            pid (int | None): Process ID when ``scope`` is ``pid``.
            role (str | None): Accessibility role to match, such as ``button``.
            name (str | None): Accessible name to match.
            name_match (str | None): Name match mode, such as ``exact`` or ``substring``.
            states (list[str] | None): Required accessibility states.
            limit (int | None): Maximum number of matches. Use ``0`` to let the API apply its default.

        Returns:
            AccessibilityNodesResponse: Matching accessibility nodes.

        Example:
            ```python
            buttons = await sandbox.computer_use.accessibility.find_nodes(
                scope="all",
                role="button",
                name="Submit",
                name_match="substring",
            )
            print(len(buttons.matches))
            ```
        """
        request = FindAccessibilityNodesRequest(
            scope=scope,
            pid=pid,
            role=role,
            name=name,
            name_match=name_match,
            states=states,
            limit=limit,
        )
        return await self._api_client.find_accessibility_nodes(request=request)

    @intercept_errors(message_prefix="Failed to focus accessibility node: ")
    @with_instrumentation()
    async def focus_node(self, node_id: str) -> None:
        """Focuses an AT-SPI accessibility node.

        Args:
            node_id (str): Accessibility node ID returned by ``get_tree`` or ``find_nodes``.

        Raises:
            DaytonaError: If the focus operation fails. API failures may use a more specific subclass.

        Example:
            ```python
            await sandbox.computer_use.accessibility.focus_node(node.id)
            ```
        """
        request = AccessibilityNodeRequest(id=node_id)
        _ = await self._api_client.focus_accessibility_node(request=request)

    @intercept_errors(message_prefix="Failed to invoke accessibility node: ")
    @with_instrumentation()
    async def invoke_node(self, node_id: str, action: str | None = None) -> None:
        """Invokes an AT-SPI accessibility node action.

        Args:
            node_id (str): Accessibility node ID returned by ``get_tree`` or ``find_nodes``.
            action (str | None): Action name to invoke. If omitted, the API invokes the primary action.

        Raises:
            DaytonaError: If the invoke operation fails. API failures may use a more specific subclass.

        Example:
            ```python
            await sandbox.computer_use.accessibility.invoke_node(node.id, action="click")
            ```
        """
        request = AccessibilityInvokeRequest(id=node_id, action=action)
        _ = await self._api_client.invoke_accessibility_node(request=request)

    @intercept_errors(message_prefix="Failed to set accessibility node value: ")
    @with_instrumentation()
    async def set_node_value(self, node_id: str, value: str) -> None:
        """Sets an AT-SPI accessibility node value.

        Args:
            node_id (str): Accessibility node ID returned by ``get_tree`` or ``find_nodes``.
            value (str): Value to write to the node.

        Raises:
            DaytonaError: If the value update fails. API failures may use a more specific subclass.

        Example:
            ```python
            await sandbox.computer_use.accessibility.set_node_value(node.id, "hello")
            ```
        """
        request = AccessibilitySetValueRequest(id=node_id, value=value)
        _ = await self._api_client.set_accessibility_node_value(request=request)


class AsyncComputerUse:
    """Computer Use functionality for interacting with the desktop environment.

    Provides access to mouse, keyboard, screenshot, display, recording, and accessibility operations
    for automating desktop interactions within a sandbox.

    Attributes:
        mouse (AsyncMouse): Mouse operations interface.
        keyboard (AsyncKeyboard): Keyboard operations interface.
        screenshot (AsyncScreenshot): Screenshot operations interface.
        display (AsyncDisplay): Display operations interface.
        recording (AsyncRecordingService): Screen recording operations interface.
        accessibility (AsyncAccessibility): Accessibility operations interface.
    """

    def __init__(
        self,
        api_client: ComputerUseApi,
    ):
        self._api_client: ComputerUseApi = api_client

        self.mouse: AsyncMouse = AsyncMouse(api_client)
        self.keyboard: AsyncKeyboard = AsyncKeyboard(api_client)
        self.screenshot: AsyncScreenshot = AsyncScreenshot(api_client)
        self.display: AsyncDisplay = AsyncDisplay(api_client)
        self.recording: AsyncRecordingService = AsyncRecordingService(api_client)
        self.accessibility: AsyncAccessibility = AsyncAccessibility(api_client)

    @intercept_errors(message_prefix="Failed to start computer use: ")
    @with_instrumentation()
    async def start(self) -> ComputerUseStartResponse:
        """Starts all computer use processes (Xvfb, xfce4, x11vnc, novnc).

        Returns:
            ComputerUseStartResponse: Computer use start response.

        Example:
            ```python
            result = await sandbox.computer_use.start()
            print("Computer use processes started:", result.message)
            ```
        """
        response = await self._api_client.start_computer_use()
        return response

    @intercept_errors(message_prefix="Failed to stop computer use: ")
    @with_instrumentation()
    async def stop(self) -> ComputerUseStopResponse:
        """Stops all computer use processes.

        Returns:
            ComputerUseStopResponse: Computer use stop response.

        Example:
            ```python
            result = await sandbox.computer_use.stop()
            print("Computer use processes stopped:", result.message)
            ```
        """
        response = await self._api_client.stop_computer_use()
        return response

    @intercept_errors(message_prefix="Failed to get computer use status: ")
    @with_instrumentation()
    async def get_status(self) -> ComputerUseStatusResponse:
        """Gets the status of all computer use processes.

        Returns:
            ComputerUseStatusResponse: Status information about all VNC desktop processes.

        Example:
            ```python
            response = await sandbox.computer_use.get_status()
            print("Computer use status:", response.status)
            ```
        """
        return await self._api_client.get_computer_use_status()

    @intercept_errors(message_prefix="Failed to get process status: ")
    @with_instrumentation()
    async def get_process_status(self, process_name: str) -> ProcessStatusResponse:
        """Gets the status of a specific VNC process.

        Args:
            process_name (str): Name of the process to check.

        Returns:
            ProcessStatusResponse: Status information about the specific process.

        Example:
            ```python
            xvfb_status = await sandbox.computer_use.get_process_status("xvfb")
            no_vnc_status = await sandbox.computer_use.get_process_status("novnc")
            ```
        """
        response = await self._api_client.get_process_status(process_name=process_name)
        return response

    @intercept_errors(message_prefix="Failed to restart process: ")
    @with_instrumentation()
    async def restart_process(self, process_name: str) -> ProcessRestartResponse:
        """Restarts a specific VNC process.

        Args:
            process_name (str): Name of the process to restart.

        Returns:
            ProcessRestartResponse: Process restart response.

        Example:
            ```python
            result = await sandbox.computer_use.restart_process("xfce4")
            print("XFCE4 process restarted:", result.message)
            ```
        """
        response = await self._api_client.restart_process(process_name=process_name)
        return response

    @intercept_errors(message_prefix="Failed to get process logs: ")
    @with_instrumentation()
    async def get_process_logs(self, process_name: str) -> ProcessLogsResponse:
        """Gets logs for a specific VNC process.

        Args:
            process_name (str): Name of the process to get logs for.

        Returns:
            ProcessLogsResponse: Process logs.

        Example:
            ```python
            logs = await sandbox.computer_use.get_process_logs("novnc")
            print("NoVNC logs:", logs)
            ```
        """
        response = await self._api_client.get_process_logs(process_name=process_name)
        return response

    @intercept_errors(message_prefix="Failed to get process errors: ")
    @with_instrumentation()
    async def get_process_errors(self, process_name: str) -> ProcessErrorsResponse:
        """Gets error logs for a specific VNC process.

        Args:
            process_name (str): Name of the process to get error logs for.

        Returns:
            ProcessErrorsResponse: Process error logs.

        Example:
            ```python
            errors = await sandbox.computer_use.get_process_errors("x11vnc")
            print("X11VNC errors:", errors)
            ```
        """
        response = await self._api_client.get_process_errors(process_name=process_name)
        return response
