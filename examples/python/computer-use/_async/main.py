"""
Computer Use Example (Async)

This example demonstrates the computer use functionality of the Daytona SDK
using async/await, including mouse control, keyboard input, screenshots,
and display information.

Computer use enables desktop automation within a sandbox environment.
"""

import asyncio
import base64
import os

from daytona import AsyncDaytona, CreateSandboxFromImageParams, ScreenshotOptions, ScreenshotRegion


def save_screenshot(screenshot_data: str, filename: str) -> None:
    """Save a base64 encoded screenshot to a file."""
    image_bytes = base64.b64decode(screenshot_data)
    with open(filename, "wb") as f:
        f.write(image_bytes)
    print(f"  Saved screenshot to {filename} ({len(image_bytes)} bytes)")


async def main():
    async with AsyncDaytona() as daytona:
        # Create a sandbox with desktop environment support
        # Computer use requires a sandbox with a graphical environment
        sandbox_params = CreateSandboxFromImageParams(
            image="daytonaio/ai-sandbox:latest",
        )
        sandbox = await daytona.create(sandbox_params, timeout=120)
        print(f"Created sandbox with ID: {sandbox.id}")

        # --- Display Information ---
        print("\n=== Display Information ===")
        display_info = await sandbox.computer_use.display.get_info()
        print(f"Displays: {display_info}")

        # Get list of open windows
        windows = await sandbox.computer_use.display.get_windows()
        print(f"Open windows: {len(windows.windows)} window(s)")
        for window in windows.windows[:5]:  # Show first 5 windows
            print(f"  - {window.title} ({window.width}x{window.height})")

        # --- Screenshots ---
        print("\n=== Taking Screenshots ===")

        # Take a full screen screenshot
        print("Taking full screen screenshot...")
        screenshot = await sandbox.computer_use.screenshot.take_full_screen()
        print(f"  Screenshot size: {screenshot.size_bytes} bytes")
        save_screenshot(screenshot.screenshot, "screenshot_full.png")

        # Take a screenshot with cursor visible
        print("Taking screenshot with cursor...")
        screenshot_cursor = await sandbox.computer_use.screenshot.take_full_screen(show_cursor=True)
        save_screenshot(screenshot_cursor.screenshot, "screenshot_with_cursor.png")

        # Take a region screenshot
        print("Taking region screenshot...")
        try:
            region = ScreenshotRegion(x=100, y=100, width=400, height=300)
            screenshot_region = await sandbox.computer_use.screenshot.take_region(region)
            save_screenshot(screenshot_region.screenshot, "screenshot_region.png")
        except Exception as e:
            print(f"  Region screenshot not supported on this platform: {e}")

        # Take a compressed screenshot (JPEG with quality)
        print("Taking compressed screenshot...")
        try:
            options = ScreenshotOptions(show_cursor=False, fmt="jpeg", quality=75, scale=0.5)
            screenshot_compressed = await sandbox.computer_use.screenshot.take_compressed(options)
            print(f"  Compressed size: {screenshot_compressed.size_bytes} bytes")
            save_screenshot(screenshot_compressed.screenshot, "screenshot_compressed.jpg")
        except Exception as e:
            print(f"  Compressed screenshot not supported on this platform: {e}")

        # --- Mouse Operations ---
        print("\n=== Mouse Operations ===")

        # Get current mouse position
        position = await sandbox.computer_use.mouse.get_position()
        print(f"Current mouse position: ({position.x}, {position.y})")

        # Move mouse to a specific position
        new_position = await sandbox.computer_use.mouse.move(400, 300)
        print(f"Moved mouse to: ({new_position.x}, {new_position.y})")

        # Perform a click
        click_result = await sandbox.computer_use.mouse.click(500, 400, button="left")
        print(f"Clicked at: ({click_result.x}, {click_result.y})")

        # Double click
        double_click = await sandbox.computer_use.mouse.click(500, 400, button="left", double=True)
        print(f"Double-clicked at: ({double_click.x}, {double_click.y})")

        # Right click
        right_click = await sandbox.computer_use.mouse.click(600, 400, button="right")
        print(f"Right-clicked at: ({right_click.x}, {right_click.y})")

        # Scroll the mouse wheel
        scroll_result = await sandbox.computer_use.mouse.scroll(500, 400, direction="down", amount=3)
        print(f"Scrolled: {scroll_result}")

        # Drag operation
        drag_result = await sandbox.computer_use.mouse.drag(100, 100, 300, 200, button="left")
        print(f"Dragged: {drag_result}")

        # --- Keyboard Operations ---
        print("\n=== Keyboard Operations ===")

        # Type some text
        print("Typing text...")
        await sandbox.computer_use.keyboard.type("Hello from Daytona SDK!")
        print("  Typed: 'Hello from Daytona SDK!'")

        # Press Enter key
        print("Pressing Enter key...")
        await sandbox.computer_use.keyboard.press("Return")
        print("  Pressed: Enter")

        # Use keyboard hotkeys (e.g., Ctrl+A to select all)
        print("Pressing hotkey Ctrl+A...")
        await sandbox.computer_use.keyboard.hotkey("ctrl+a")
        print("  Pressed: Ctrl+A (Select All)")

        # More hotkey examples
        print("Pressing hotkey Ctrl+C...")
        await sandbox.computer_use.keyboard.hotkey("ctrl+c")
        print("  Pressed: Ctrl+C (Copy)")

        # Take final screenshot to see results
        print("\n=== Final Screenshot ===")
        final_screenshot = await sandbox.computer_use.screenshot.take_full_screen(show_cursor=True)
        save_screenshot(final_screenshot.screenshot, "screenshot_final.png")

        # --- Stop Computer Use ---
        print("\n=== Stopping Computer Use ===")
        stop_result = await sandbox.computer_use.stop()
        print(f"Stop result: {stop_result}")

        # Clean up local screenshot files
        print("\n=== Cleanup ===")
        for filename in [
            "screenshot_full.png",
            "screenshot_with_cursor.png",
            "screenshot_region.png",
            "screenshot_compressed.jpg",
            "screenshot_final.png",
        ]:
            if os.path.exists(filename):
                os.remove(filename)
                print(f"  Removed {filename}")

        # Delete the sandbox
        await daytona.delete(sandbox)
        print(f"\nDeleted sandbox {sandbox.id}")


if __name__ == "__main__":
    asyncio.run(main())
