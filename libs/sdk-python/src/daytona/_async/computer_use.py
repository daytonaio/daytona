# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from typing import List, Optional

from daytona_api_client_async import (
    CompressedScreenshotResponse,
    ComputerUseStartResponse,
    ComputerUseStatusResponse,
    ComputerUseStopResponse,
    DisplayInfoResponse,
    KeyboardHotkeyRequest,
    KeyboardPressRequest,
    KeyboardTypeRequest,
    MouseClickRequest,
    MouseClickResponse,
    MouseDragRequest,
    MouseDragResponse,
    MouseMoveRequest,
    MouseMoveResponse,
    MousePosition,
    MouseScrollRequest,
    ProcessErrorsResponse,
    ProcessLogsResponse,
    ProcessRestartResponse,
    ProcessStatusResponse,
    RegionScreenshotResponse,
    ScreenshotResponse,
    ToolboxApi,
    WindowsResponse,
)

from .._utils.errors import intercept_errors
from ..common.computer_use import ScreenshotOptions, ScreenshotRegion


class AsyncMouse:
    """Mouse operations for computer use functionality."""

    def __init__(self, sandbox_id: str, toolbox_api: ToolboxApi):
        self._sandbox_id = sandbox_id
        self._toolbox_api = toolbox_api

    @intercept_errors(message_prefix="Failed to get mouse position: ")
    async def get_position(self) -> MousePosition:
        """Gets the current mouse cursor position.

        Returns:
            MousePosition: Current mouse position with x and y coordinates.

        Example:
            ```python
            position = await sandbox.computer_use.mouse.get_position()
            print(f"Mouse is at: {position.x}, {position.y}")
            ```
        """
        response = await self._toolbox_api.get_mouse_position(self._sandbox_id)
        return response

    @intercept_errors(message_prefix="Failed to move mouse: ")
    async def move(self, x: int, y: int) -> MouseMoveResponse:
        """Moves the mouse cursor to the specified coordinates.

        Args:
            x (int): The x coordinate to move to.
            y (int): The y coordinate to move to.

        Returns:
            MouseMoveResponse: Move operation result.

        Example:
            ```python
            result = await sandbox.computer_use.mouse.move(100, 200)
            print(f"Mouse moved to: {result.x}, {result.y}")
            ```
        """
        request = MouseMoveRequest(x=x, y=y)
        response = await self._toolbox_api.move_mouse(self._sandbox_id, request)
        return response

    @intercept_errors(message_prefix="Failed to click mouse: ")
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
        response = await self._toolbox_api.click_mouse(self._sandbox_id, request)
        return response

    @intercept_errors(message_prefix="Failed to drag mouse: ")
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
            print(f"Dragged from {result.from_x},{result.from_y} to {result.to_x},{result.to_y}")
            ```
        """
        request = MouseDragRequest(start_x=start_x, start_y=start_y, end_x=end_x, end_y=end_y, button=button)
        response = await self._toolbox_api.drag_mouse(self._sandbox_id, request)
        return response

    @intercept_errors(message_prefix="Failed to scroll mouse: ")
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
        response = await self._toolbox_api.scroll_mouse(self._sandbox_id, request)
        return response


class AsyncKeyboard:
    """Keyboard operations for computer use functionality."""

    def __init__(self, sandbox_id: str, toolbox_api: ToolboxApi):
        self._sandbox_id = sandbox_id
        self._toolbox_api = toolbox_api

    @intercept_errors(message_prefix="Failed to type text: ")
    async def type(self, text: str, delay: Optional[int] = None) -> None:
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
        await self._toolbox_api.type_text(self._sandbox_id, request)

    @intercept_errors(message_prefix="Failed to press key: ")
    async def press(self, key: str, modifiers: Optional[List[str]] = None) -> None:
        """Presses a key with optional modifiers.

        Args:
            key (str): The key to press (e.g., 'Enter', 'Escape', 'Tab', 'a', 'A').
            modifiers (List[str]): Modifier keys ('ctrl', 'alt', 'meta', 'shift').

        Raises:
            DaytonaError: If the press operation fails.

        Example:
            ```python
            # Press Enter
            try:
                await sandbox.computer_use.keyboard.press("Return")
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
        await self._toolbox_api.press_key(self._sandbox_id, request)

    @intercept_errors(message_prefix="Failed to press hotkey: ")
    async def hotkey(self, keys: str) -> None:
        """Presses a hotkey combination.

        Args:
            keys (str): The hotkey combination (e.g., 'ctrl+c', 'alt+tab', 'cmd+shift+t').

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
        await self._toolbox_api.press_hotkey(self._sandbox_id, request)


class AsyncScreenshot:
    """Screenshot operations for computer use functionality."""

    def __init__(self, sandbox_id: str, toolbox_api: ToolboxApi):
        self._sandbox_id = sandbox_id
        self._toolbox_api = toolbox_api

    @intercept_errors(message_prefix="Failed to take screenshot: ")
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
        response = await self._toolbox_api.take_screenshot(self._sandbox_id, show_cursor=show_cursor)
        return response

    @intercept_errors(message_prefix="Failed to take region screenshot: ")
    async def take_region(self, region: ScreenshotRegion, show_cursor: bool = False) -> RegionScreenshotResponse:
        """Takes a screenshot of a specific region.

        Args:
            region (ScreenshotRegion): The region to capture.
            show_cursor (bool): Whether to show the cursor in the screenshot.

        Returns:
            RegionScreenshotResponse: Screenshot data with base64 encoded image.

        Example:
            ```python
            region = ScreenshotRegion(x=100, y=100, width=300, height=200)
            screenshot = await sandbox.computer_use.screenshot.take_region(region)
            print(f"Captured region: {screenshot.region.width}x{screenshot.region.height}")
            ```
        """
        response = await self._toolbox_api.take_region_screenshot(
            self._sandbox_id, region.height, region.width, region.y, region.x, show_cursor=show_cursor
        )
        return response

    @intercept_errors(message_prefix="Failed to take compressed screenshot: ")
    async def take_compressed(self, options: Optional[ScreenshotOptions] = None) -> CompressedScreenshotResponse:
        """Takes a compressed screenshot of the entire screen.

        Args:
            options (ScreenshotOptions): Compression and display options.

        Returns:
            CompressedScreenshotResponse: Compressed screenshot data.

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

        response = await self._toolbox_api.take_compressed_screenshot(
            self._sandbox_id,
            scale=options.scale,
            quality=options.quality,
            format=options.fmt,
            show_cursor=options.show_cursor,
        )
        return response

    @intercept_errors(message_prefix="Failed to take compressed region screenshot: ")
    async def take_compressed_region(
        self, region: ScreenshotRegion, options: Optional[ScreenshotOptions] = None
    ) -> CompressedScreenshotResponse:
        """Takes a compressed screenshot of a specific region.

        Args:
            region (ScreenshotRegion): The region to capture.
            options (ScreenshotOptions): Compression and display options.

        Returns:
            CompressedScreenshotResponse: Compressed screenshot data.

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

        response = await self._toolbox_api.take_compressed_region_screenshot(
            self._sandbox_id,
            region.height,
            region.width,
            region.y,
            region.x,
            scale=options.scale,
            quality=options.quality,
            format=options.fmt,
            show_cursor=options.show_cursor,
        )
        return response


class AsyncDisplay:
    """Display operations for computer use functionality."""

    def __init__(self, sandbox_id: str, toolbox_api: ToolboxApi):
        self._sandbox_id = sandbox_id
        self._toolbox_api = toolbox_api

    @intercept_errors(message_prefix="Failed to get display info: ")
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
        response = await self._toolbox_api.get_display_info(self._sandbox_id)
        return response

    @intercept_errors(message_prefix="Failed to get windows: ")
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
        response = await self._toolbox_api.get_windows(self._sandbox_id)
        return response


class AsyncComputerUse:
    """Computer Use functionality for interacting with the desktop environment.

    Provides access to mouse, keyboard, screenshot, and display operations
    for automating desktop interactions within a sandbox.

    Attributes:
        mouse (AsyncMouse): Mouse operations interface.
        keyboard (AsyncKeyboard): Keyboard operations interface.
        screenshot (AsyncScreenshot): Screenshot operations interface.
        display (AsyncDisplay): Display operations interface.
    """

    def __init__(self, sandbox_id: str, toolbox_api: ToolboxApi):
        self._sandbox_id = sandbox_id
        self._toolbox_api = toolbox_api

        self.mouse = AsyncMouse(sandbox_id, toolbox_api)
        self.keyboard = AsyncKeyboard(sandbox_id, toolbox_api)
        self.screenshot = AsyncScreenshot(sandbox_id, toolbox_api)
        self.display = AsyncDisplay(sandbox_id, toolbox_api)

    @intercept_errors(message_prefix="Failed to start computer use: ")
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
        response = await self._toolbox_api.start_computer_use(self._sandbox_id)
        return response

    @intercept_errors(message_prefix="Failed to stop computer use: ")
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
        response = await self._toolbox_api.stop_computer_use(self._sandbox_id)
        return response

    @intercept_errors(message_prefix="Failed to get computer use status: ")
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
        return await self._toolbox_api.get_computer_use_status(self._sandbox_id)

    @intercept_errors(message_prefix="Failed to get process status: ")
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
        response = await self._toolbox_api.get_process_status(process_name, self._sandbox_id)
        return response

    @intercept_errors(message_prefix="Failed to restart process: ")
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
        response = await self._toolbox_api.restart_process(process_name, self._sandbox_id)
        return response

    @intercept_errors(message_prefix="Failed to get process logs: ")
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
        response = await self._toolbox_api.get_process_logs(process_name, self._sandbox_id)
        return response

    @intercept_errors(message_prefix="Failed to get process errors: ")
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
        response = await self._toolbox_api.get_process_errors(process_name, self._sandbox_id)
        return response
