---
title: "AsyncComputerUse"
hideTitleOnPage: true
---

## AsyncComputerUse

```python
class AsyncComputerUse()
```

Computer Use functionality for interacting with the desktop environment.

Provides access to mouse, keyboard, screenshot, display, and recording operations
for automating desktop interactions within a sandbox.

**Attributes**:

- `mouse` _AsyncMouse_ - Mouse operations interface.
- `keyboard` _AsyncKeyboard_ - Keyboard operations interface.
- `screenshot` _AsyncScreenshot_ - Screenshot operations interface.
- `display` _AsyncDisplay_ - Display operations interface.
- `recording` _AsyncRecordingService_ - Screen recording operations interface.

#### AsyncComputerUse.start

```python
@intercept_errors(message_prefix="Failed to start computer use: ")
@with_instrumentation()
async def start() -> ComputerUseStartResponse
```

Starts all computer use processes (Xvfb, xfce4, x11vnc, novnc).

**Returns**:

- `ComputerUseStartResponse` - Computer use start response.
  

**Example**:

```python
result = await sandbox.computer_use.start()
print("Computer use processes started:", result.message)
```

#### AsyncComputerUse.stop

```python
@intercept_errors(message_prefix="Failed to stop computer use: ")
@with_instrumentation()
async def stop() -> ComputerUseStopResponse
```

Stops all computer use processes.

**Returns**:

- `ComputerUseStopResponse` - Computer use stop response.
  

**Example**:

```python
result = await sandbox.computer_use.stop()
print("Computer use processes stopped:", result.message)
```

#### AsyncComputerUse.get\_status

```python
@intercept_errors(message_prefix="Failed to get computer use status: ")
@with_instrumentation()
async def get_status() -> ComputerUseStatusResponse
```

Gets the status of all computer use processes.

**Returns**:

- `ComputerUseStatusResponse` - Status information about all VNC desktop processes.
  

**Example**:

```python
response = await sandbox.computer_use.get_status()
print("Computer use status:", response.status)
```

#### AsyncComputerUse.get\_process\_status

```python
@intercept_errors(message_prefix="Failed to get process status: ")
@with_instrumentation()
async def get_process_status(process_name: str) -> ProcessStatusResponse
```

Gets the status of a specific VNC process.

**Arguments**:

- `process_name` _str_ - Name of the process to check.
  

**Returns**:

- `ProcessStatusResponse` - Status information about the specific process.
  

**Example**:

```python
xvfb_status = await sandbox.computer_use.get_process_status("xvfb")
no_vnc_status = await sandbox.computer_use.get_process_status("novnc")
```

#### AsyncComputerUse.restart\_process

```python
@intercept_errors(message_prefix="Failed to restart process: ")
@with_instrumentation()
async def restart_process(process_name: str) -> ProcessRestartResponse
```

Restarts a specific VNC process.

**Arguments**:

- `process_name` _str_ - Name of the process to restart.
  

**Returns**:

- `ProcessRestartResponse` - Process restart response.
  

**Example**:

```python
result = await sandbox.computer_use.restart_process("xfce4")
print("XFCE4 process restarted:", result.message)
```

#### AsyncComputerUse.get\_process\_logs

```python
@intercept_errors(message_prefix="Failed to get process logs: ")
@with_instrumentation()
async def get_process_logs(process_name: str) -> ProcessLogsResponse
```

Gets logs for a specific VNC process.

**Arguments**:

- `process_name` _str_ - Name of the process to get logs for.
  

**Returns**:

- `ProcessLogsResponse` - Process logs.
  

**Example**:

```python
logs = await sandbox.computer_use.get_process_logs("novnc")
print("NoVNC logs:", logs)
```

#### AsyncComputerUse.get\_process\_errors

```python
@intercept_errors(message_prefix="Failed to get process errors: ")
@with_instrumentation()
async def get_process_errors(process_name: str) -> ProcessErrorsResponse
```

Gets error logs for a specific VNC process.

**Arguments**:

- `process_name` _str_ - Name of the process to get error logs for.
  

**Returns**:

- `ProcessErrorsResponse` - Process error logs.
  

**Example**:

```python
errors = await sandbox.computer_use.get_process_errors("x11vnc")
print("X11VNC errors:", errors)
```


## AsyncMouse

```python
class AsyncMouse()
```

Mouse operations for computer use functionality.

#### AsyncMouse.get\_position

```python
@intercept_errors(message_prefix="Failed to get mouse position: ")
@with_instrumentation()
async def get_position() -> MousePositionResponse
```

Gets the current mouse cursor position.

**Returns**:

- `MousePositionResponse` - Current mouse position with x and y coordinates.
  

**Example**:

```python
position = await sandbox.computer_use.mouse.get_position()
print(f"Mouse is at: {position.x}, {position.y}")
```

#### AsyncMouse.move

```python
@intercept_errors(message_prefix="Failed to move mouse: ")
@with_instrumentation()
async def move(x: int, y: int) -> MousePositionResponse
```

Moves the mouse cursor to the specified coordinates.

**Arguments**:

- `x` _int_ - The x coordinate to move to.
- `y` _int_ - The y coordinate to move to.
  

**Returns**:

- `MousePositionResponse` - Position after move.
  

**Example**:

```python
result = await sandbox.computer_use.mouse.move(100, 200)
print(f"Mouse moved to: {result.x}, {result.y}")
```

#### AsyncMouse.click

```python
@intercept_errors(message_prefix="Failed to click mouse: ")
@with_instrumentation()
async def click(x: int,
                y: int,
                button: str = "left",
                double: bool = False) -> MouseClickResponse
```

Clicks the mouse at the specified coordinates.

**Arguments**:

- `x` _int_ - The x coordinate to click at.
- `y` _int_ - The y coordinate to click at.
- `button` _str_ - The mouse button to click ('left', 'right', 'middle').
- `double` _bool_ - Whether to perform a double-click.
  

**Returns**:

- `MouseClickResponse` - Click operation result.
  

**Example**:

```python
# Single left click
result = await sandbox.computer_use.mouse.click(100, 200)

# Double click
double_click = await sandbox.computer_use.mouse.click(100, 200, "left", True)

# Right click
right_click = await sandbox.computer_use.mouse.click(100, 200, "right")
```

#### AsyncMouse.drag

```python
@intercept_errors(message_prefix="Failed to drag mouse: ")
@with_instrumentation()
async def drag(start_x: int,
               start_y: int,
               end_x: int,
               end_y: int,
               button: str = "left") -> MouseDragResponse
```

Drags the mouse from start coordinates to end coordinates.

**Arguments**:

- `start_x` _int_ - The starting x coordinate.
- `start_y` _int_ - The starting y coordinate.
- `end_x` _int_ - The ending x coordinate.
- `end_y` _int_ - The ending y coordinate.
- `button` _str_ - The mouse button to use for dragging.
  

**Returns**:

- `MouseDragResponse` - Drag operation result.
  

**Example**:

```python
result = await sandbox.computer_use.mouse.drag(50, 50, 150, 150)
print(f"Dragged from {result.from_x},{result.from_y} to {result.to_x},{result.to_y}")
```

#### AsyncMouse.scroll

```python
@intercept_errors(message_prefix="Failed to scroll mouse: ")
@with_instrumentation()
async def scroll(x: int, y: int, direction: str, amount: int = 1) -> bool
```

Scrolls the mouse wheel at the specified coordinates.

**Arguments**:

- `x` _int_ - The x coordinate to scroll at.
- `y` _int_ - The y coordinate to scroll at.
- `direction` _str_ - The direction to scroll ('up' or 'down').
- `amount` _int_ - The amount to scroll.
  

**Returns**:

- `bool` - Whether the scroll operation was successful.
  

**Example**:

```python
# Scroll up
scroll_up = await sandbox.computer_use.mouse.scroll(100, 200, "up", 3)

# Scroll down
scroll_down = await sandbox.computer_use.mouse.scroll(100, 200, "down", 5)
```

## AsyncKeyboard

```python
class AsyncKeyboard()
```

Keyboard operations for computer use functionality.

#### AsyncKeyboard.type

```python
@intercept_errors(message_prefix="Failed to type text: ")
@with_instrumentation()
async def type(text: str, delay: int | None = None) -> None
```

Types the specified text.

**Arguments**:

- `text` _str_ - The text to type.
- `delay` _int_ - Delay between characters in milliseconds.
  

**Raises**:

- `DaytonaError` - If the type operation fails.
  

**Example**:

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

#### AsyncKeyboard.press

```python
@intercept_errors(message_prefix="Failed to press key: ")
@with_instrumentation()
async def press(key: str, modifiers: list[str] | None = None) -> None
```

Presses a key with optional modifiers.

**Arguments**:

- `key` _str_ - The key to press (e.g., 'Enter', 'Escape', 'Tab', 'a', 'A').
- `modifiers` _list[str]_ - Modifier keys ('ctrl', 'alt', 'meta', 'shift').
  

**Raises**:

- `DaytonaError` - If the press operation fails.
  

**Example**:

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

#### AsyncKeyboard.hotkey

```python
@intercept_errors(message_prefix="Failed to press hotkey: ")
@with_instrumentation()
async def hotkey(keys: str) -> None
```

Presses a hotkey combination.

**Arguments**:

- `keys` _str_ - The hotkey combination (e.g., 'ctrl+c', 'alt+tab', 'cmd+shift+t').
  

**Raises**:

- `DaytonaError` - If the hotkey operation fails.
  

**Example**:

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

## AsyncScreenshot

```python
class AsyncScreenshot()
```

Screenshot operations for computer use functionality.

#### AsyncScreenshot.take\_full\_screen

```python
@intercept_errors(message_prefix="Failed to take screenshot: ")
@with_instrumentation()
async def take_full_screen(show_cursor: bool = False) -> ScreenshotResponse
```

Takes a screenshot of the entire screen.

**Arguments**:

- `show_cursor` _bool_ - Whether to show the cursor in the screenshot.
  

**Returns**:

- `ScreenshotResponse` - Screenshot data with base64 encoded image.
  

**Example**:

```python
screenshot = await sandbox.computer_use.screenshot.take_full_screen()
print(f"Screenshot size: {screenshot.width}x{screenshot.height}")

# With cursor visible
with_cursor = await sandbox.computer_use.screenshot.take_full_screen(True)
```

#### AsyncScreenshot.take\_region

```python
@intercept_errors(message_prefix="Failed to take region screenshot: ")
@with_instrumentation()
async def take_region(region: ScreenshotRegion,
                      show_cursor: bool = False) -> ScreenshotResponse
```

Takes a screenshot of a specific region.

**Arguments**:

- `region` _ScreenshotRegion_ - The region to capture.
- `show_cursor` _bool_ - Whether to show the cursor in the screenshot.
  

**Returns**:

- `ScreenshotResponse` - Screenshot data with base64 encoded image.
  

**Example**:

```python
region = ScreenshotRegion(x=100, y=100, width=300, height=200)
screenshot = await sandbox.computer_use.screenshot.take_region(region)
print(f"Captured region: {screenshot.region.width}x{screenshot.region.height}")
```

#### AsyncScreenshot.take\_compressed

```python
@intercept_errors(message_prefix="Failed to take compressed screenshot: ")
@with_instrumentation()
async def take_compressed(
        options: ScreenshotOptions | None = None) -> ScreenshotResponse
```

Takes a compressed screenshot of the entire screen.

**Arguments**:

- `options` _ScreenshotOptions | None_ - Compression and display options.
  

**Returns**:

- `ScreenshotResponse` - Compressed screenshot data.
  

**Example**:

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

#### AsyncScreenshot.take\_compressed\_region

```python
@intercept_errors(
    message_prefix="Failed to take compressed region screenshot: ")
@with_instrumentation()
async def take_compressed_region(
        region: ScreenshotRegion,
        options: ScreenshotOptions | None = None) -> ScreenshotResponse
```

Takes a compressed screenshot of a specific region.

**Arguments**:

- `region` _ScreenshotRegion_ - The region to capture.
- `options` _ScreenshotOptions | None_ - Compression and display options.
  

**Returns**:

- `ScreenshotResponse` - Compressed screenshot data.
  

**Example**:

```python
region = ScreenshotRegion(x=0, y=0, width=800, height=600)
screenshot = await sandbox.computer_use.screenshot.take_compressed_region(
    region,
    ScreenshotOptions(format="webp", quality=80, show_cursor=True)
)
print(f"Compressed size: {screenshot.size_bytes} bytes")
```

## AsyncDisplay

```python
class AsyncDisplay()
```

Display operations for computer use functionality.

#### AsyncDisplay.get\_info

```python
@intercept_errors(message_prefix="Failed to get display info: ")
@with_instrumentation()
async def get_info() -> DisplayInfoResponse
```

Gets information about the displays.

**Returns**:

- `DisplayInfoResponse` - Display information including primary display and all available displays.
  

**Example**:

```python
info = await sandbox.computer_use.display.get_info()
print(f"Primary display: {info.primary_display.width}x{info.primary_display.height}")
print(f"Total displays: {info.total_displays}")
for i, display in enumerate(info.displays):
    print(f"Display {i}: {display.width}x{display.height} at {display.x},{display.y}")
```

#### AsyncDisplay.get\_windows

```python
@intercept_errors(message_prefix="Failed to get windows: ")
@with_instrumentation()
async def get_windows() -> WindowsResponse
```

Gets the list of open windows.

**Returns**:

- `WindowsResponse` - List of open windows with their IDs and titles.
  

**Example**:

```python
windows = await sandbox.computer_use.display.get_windows()
print(f"Found {windows.count} open windows:")
for window in windows.windows:
    print(f"- {window.title} (ID: {window.id})")
```

## AsyncRecordingService

```python
class AsyncRecordingService()
```

Recording operations for computer use functionality.

#### AsyncRecordingService.start

```python
@intercept_errors(message_prefix="Failed to start recording: ")
@with_instrumentation()
async def start(label: str | None = None) -> Recording
```

Starts a new screen recording session.

**Arguments**:

- `label` _str | None_ - Optional custom label for the recording.
  

**Returns**:

- `Recording` - Recording start response.
  

**Example**:

```python
# Start a recording with a label
recording = await sandbox.computer_use.recording.start("my-test-recording")
print(f"Recording started: {recording.id}")
print(f"File: {recording.file_path}")
```

#### AsyncRecordingService.stop

```python
@intercept_errors(message_prefix="Failed to stop recording: ")
@with_instrumentation()
async def stop(recording_id: str) -> Recording
```

Stops an active screen recording session.

**Arguments**:

- `recording_id` _str_ - The ID of the recording to stop.
  

**Returns**:

- `Recording` - Recording stop response.
  

**Example**:

```python
result = await sandbox.computer_use.recording.stop(recording.id)
print(f"Recording stopped: {result.duration_seconds} seconds")
print(f"Saved to: {result.file_path}")
```

#### AsyncRecordingService.list

```python
@intercept_errors(message_prefix="Failed to list recordings: ")
@with_instrumentation()
async def list() -> ListRecordingsResponse
```

Lists all recordings (active and completed).

**Returns**:

- `ListRecordingsResponse` - List of all recordings.
  

**Example**:

```python
recordings = await sandbox.computer_use.recording.list()
print(f"Found {len(recordings.recordings)} recordings")
for rec in recordings.recordings:
    print(f"- {rec.file_name}: {rec.status}")
```

#### AsyncRecordingService.get

```python
@intercept_errors(message_prefix="Failed to get recording: ")
@with_instrumentation()
async def get(recording_id: str) -> Recording
```

Gets details of a specific recording by ID.

**Arguments**:

- `recording_id` _str_ - The ID of the recording to retrieve.
  

**Returns**:

- `Recording` - Recording details.
  

**Example**:

```python
recording = await sandbox.computer_use.recording.get(recording_id)
print(f"Recording: {recording.file_name}")
print(f"Status: {recording.status}")
print(f"Duration: {recording.duration_seconds} seconds")
```

#### AsyncRecordingService.delete

```python
@intercept_errors(message_prefix="Failed to delete recording: ")
@with_instrumentation()
async def delete(recording_id: str) -> None
```

Deletes a recording by ID.

**Arguments**:

- `recording_id` _str_ - The ID of the recording to delete.
  

**Example**:

```python
await sandbox.computer_use.recording.delete(recording_id)
print("Recording deleted")
```

#### AsyncRecordingService.download

```python
@intercept_errors(message_prefix="Failed to download recording: ")
@with_instrumentation()
async def download(recording_id: str, local_path: str) -> None
```

Downloads a recording file from the Sandbox and saves it to a local file.

The file is streamed directly to disk without loading the entire content into memory.

**Arguments**:

- `recording_id` _str_ - The ID of the recording to download.
- `local_path` _str_ - Path to save the recording file locally.
  

**Example**:

```python
# Download recording to file
await sandbox.computer_use.recording.download(recording_id, "local_recording.mp4")
print("Recording downloaded")
```

## ScreenshotRegion

```python
class ScreenshotRegion(BaseModel)
```

Region coordinates for screenshot operations.

**Attributes**:

- `x` _int_ - X coordinate of the region.
- `y` _int_ - Y coordinate of the region.
- `width` _int_ - Width of the region.
- `height` _int_ - Height of the region.

## ScreenshotOptions

```python
class ScreenshotOptions(BaseModel)
```

Options for screenshot compression and display.

**Attributes**:

- `show_cursor` _bool | None_ - Whether to show the cursor in the screenshot.
- `fmt` _str | None_ - Image format (e.g., 'png', 'jpeg', 'webp').
- `quality` _int | None_ - Compression quality (0-100).
- `scale` _float | None_ - Scale factor for the screenshot.

