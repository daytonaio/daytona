/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  ComputerUseApi,
  MousePositionResponse,
  MouseMoveRequest,
  MouseClickRequest,
  MouseClickResponse,
  MouseDragRequest,
  MouseDragResponse,
  MouseScrollRequest,
  KeyboardTypeRequest,
  KeyboardPressRequest,
  KeyboardHotkeyRequest,
  ScreenshotResponse,
  DisplayInfoResponse,
  WindowsResponse,
  ComputerUseStartResponse,
  ComputerUseStopResponse,
  ComputerUseStatusResponse,
  ProcessStatusResponse,
  ProcessRestartResponse,
  ProcessLogsResponse,
  ProcessErrorsResponse,
} from '@daytonaio/toolbox-api-client'
import { WithInstrumentation } from './utils/otel.decorator'

/**
 * Interface for region coordinates used in screenshot operations
 */
export interface ScreenshotRegion {
  x: number
  y: number
  width: number
  height: number
}

/**
 * Interface for screenshot compression options
 */
export interface ScreenshotOptions {
  showCursor?: boolean
  format?: string
  quality?: number
  scale?: number
}

/**
 * Mouse operations for computer use functionality
 */
export class Mouse {
  constructor(private readonly apiClient: ComputerUseApi) {}

  /**
   * Gets the current mouse cursor position
   *
   * @returns {Promise<MousePositionResponse>} Current mouse position with x and y coordinates
   *
   * @example
   * ```typescript
   * const position = await sandbox.computerUse.mouse.getPosition();
   * console.log(`Mouse is at: ${position.x}, ${position.y}`);
   * ```
   */
  @WithInstrumentation()
  public async getPosition(): Promise<MousePositionResponse> {
    const response = await this.apiClient.getMousePosition()
    return response.data
  }

  /**
   * Moves the mouse cursor to the specified coordinates
   *
   * @param {number} x - The x coordinate to move to
   * @param {number} y - The y coordinate to move to
   * @returns {Promise<MousePositionResponse>} Position after move
   *
   * @example
   * ```typescript
   * const result = await sandbox.computerUse.mouse.move(100, 200);
   * console.log(`Mouse moved to: ${result.x}, ${result.y}`);
   * ```
   */
  @WithInstrumentation()
  public async move(x: number, y: number): Promise<MousePositionResponse> {
    const request: MouseMoveRequest = { x, y }
    const response = await this.apiClient.moveMouse(request)
    return response.data
  }

  /**
   * Clicks the mouse at the specified coordinates
   *
   * @param {number} x - The x coordinate to click at
   * @param {number} y - The y coordinate to click at
   * @param {string} [button='left'] - The mouse button to click ('left', 'right', 'middle')
   * @param {boolean} [double=false] - Whether to perform a double-click
   * @returns {Promise<MouseClickResponse>} Click operation result
   *
   * @example
   * ```typescript
   * // Single left click
   * const result = await sandbox.computerUse.mouse.click(100, 200);
   *
   * // Double click
   * const doubleClick = await sandbox.computerUse.mouse.click(100, 200, 'left', true);
   *
   * // Right click
   * const rightClick = await sandbox.computerUse.mouse.click(100, 200, 'right');
   * ```
   */
  @WithInstrumentation()
  public async click(x: number, y: number, button = 'left', double = false): Promise<MouseClickResponse> {
    const request: MouseClickRequest = { x, y, button, double }
    const response = await this.apiClient.click(request)
    return response.data
  }

  /**
   * Drags the mouse from start coordinates to end coordinates
   *
   * @param {number} startX - The starting x coordinate
   * @param {number} startY - The starting y coordinate
   * @param {number} endX - The ending x coordinate
   * @param {number} endY - The ending y coordinate
   * @param {string} [button='left'] - The mouse button to use for dragging
   * @returns {Promise<MouseDragResponse>} Drag operation result
   *
   * @example
   * ```typescript
   * const result = await sandbox.computerUse.mouse.drag(50, 50, 150, 150);
   * console.log(`Dragged from ${result.from.x},${result.from.y} to ${result.to.x},${result.to.y}`);
   * ```
   */
  @WithInstrumentation()
  public async drag(
    startX: number,
    startY: number,
    endX: number,
    endY: number,
    button = 'left',
  ): Promise<MouseDragResponse> {
    const request: MouseDragRequest = { startX, startY, endX, endY, button }
    const response = await this.apiClient.drag(request)
    return response.data
  }

  /**
   * Scrolls the mouse wheel at the specified coordinates
   *
   * @param {number} x - The x coordinate to scroll at
   * @param {number} y - The y coordinate to scroll at
   * @param {'up' | 'down'} direction - The direction to scroll
   * @param {number} [amount=1] - The amount to scroll
   * @returns {Promise<boolean>} Whether the scroll operation was successful
   *
   * @example
   * ```typescript
   * // Scroll up
   * const scrollUp = await sandbox.computerUse.mouse.scroll(100, 200, 'up', 3);
   *
   * // Scroll down
   * const scrollDown = await sandbox.computerUse.mouse.scroll(100, 200, 'down', 5);
   * ```
   */
  @WithInstrumentation()
  public async scroll(x: number, y: number, direction: 'up' | 'down', amount = 1): Promise<boolean> {
    const request: MouseScrollRequest = { x, y, direction, amount }
    const response = await this.apiClient.scroll(request)
    return response.data.success
  }
}

/**
 * Keyboard operations for computer use functionality
 */
export class Keyboard {
  constructor(private readonly apiClient: ComputerUseApi) {}

  /**
   * Types the specified text
   *
   * @param {string} text - The text to type
   * @param {number} [delay=0] - Delay between characters in milliseconds
   * @throws {DaytonaError} If the type operation fails
   *
   * @example
   * ```typescript
   * try {
   *   await sandbox.computerUse.keyboard.type('Hello, World!');
   *   console.log('Operation success');
   * } catch (e) {
   *   console.log('Operation failed:', e);
   * }
   *
   * // With delay between characters
   * try {
   *   await sandbox.computerUse.keyboard.type('Slow typing', 100);
   *   console.log('Operation success');
   * } catch (e) {
   *   console.log('Operation failed:', e);
   * }
   * ```
   */
  @WithInstrumentation()
  public async type(text: string, delay?: number): Promise<void> {
    const request: KeyboardTypeRequest = { text, delay }
    await this.apiClient.typeText(request)
  }

  /**
   * Presses a key with optional modifiers
   *
   * @param {string} key - The key to press (e.g., 'Enter', 'Escape', 'Tab', 'a', 'A')
   * @param {string[]} [modifiers=[]] - Modifier keys ('ctrl', 'alt', 'meta', 'shift')
   * @throws {DaytonaError} If the press operation fails
   *
   * @example
   * ```typescript
   * // Press Enter
   * try {
   *   await sandbox.computerUse.keyboard.press('Return');
   *   console.log('Operation success');
   * } catch (e) {
   *   console.log('Operation failed:', e);
   * }
   *
   * // Press Ctrl+C
   * try {
   *   await sandbox.computerUse.keyboard.press('c', ['ctrl']);
   *   console.log('Operation success');
   * } catch (e) {
   *   console.log('Operation failed:', e);
   * }
   *
   * // Press Ctrl+Shift+T
   * try {
   *   await sandbox.computerUse.keyboard.press('t', ['ctrl', 'shift']);
   *   console.log('Operation success');
   * } catch (e) {
   *   console.log('Operation failed:', e);
   * }
   * ```
   */
  @WithInstrumentation()
  public async press(key: string, modifiers: string[] = []): Promise<void> {
    const request: KeyboardPressRequest = { key, modifiers }
    await this.apiClient.pressKey(request)
  }

  /**
   * Presses a hotkey combination
   *
   * @param {string} keys - The hotkey combination (e.g., 'ctrl+c', 'alt+tab', 'cmd+shift+t')
   * @throws {DaytonaError} If the hotkey operation fails
   *
   * @example
   * ```typescript
   * // Copy
   * try {
   *   await sandbox.computerUse.keyboard.hotkey('ctrl+c');
   *   console.log('Operation success');
   * } catch (e) {
   *   console.log('Operation failed:', e);
   * }
   *
   * // Paste
   * try {
   *   await sandbox.computerUse.keyboard.hotkey('ctrl+v');
   *   console.log('Operation success');
   * } catch (e) {
   *   console.log('Operation failed:', e);
   * }
   *
   * // Alt+Tab
   * try {
   *   await sandbox.computerUse.keyboard.hotkey('alt+tab');
   *   console.log('Operation success');
   * } catch (e) {
   *   console.log('Operation failed:', e);
   * }
   * ```
   */
  @WithInstrumentation()
  public async hotkey(keys: string): Promise<void> {
    const request: KeyboardHotkeyRequest = { keys }
    await this.apiClient.pressHotkey(request)
  }
}

/**
 * Screenshot operations for computer use functionality
 */
export class Screenshot {
  constructor(private readonly apiClient: ComputerUseApi) {}

  /**
   * Takes a screenshot of the entire screen
   *
   * @param {boolean} [showCursor=false] - Whether to show the cursor in the screenshot
   * @returns {Promise<ScreenshotResponse>} Screenshot data with base64 encoded image
   *
   * @example
   * ```typescript
   * const screenshot = await sandbox.computerUse.screenshot.takeFullScreen();
   * console.log(`Screenshot size: ${screenshot.width}x${screenshot.height}`);
   *
   * // With cursor visible
   * const withCursor = await sandbox.computerUse.screenshot.takeFullScreen(true);
   * ```
   */
  @WithInstrumentation()
  public async takeFullScreen(showCursor = false): Promise<ScreenshotResponse> {
    const response = await this.apiClient.takeScreenshot(showCursor)
    return response.data
  }

  /**
   * Takes a screenshot of a specific region
   *
   * @param {ScreenshotRegion} region - The region to capture
   * @param {boolean} [showCursor=false] - Whether to show the cursor in the screenshot
   * @returns {Promise<RegionScreenshotResponse>} Screenshot data with base64 encoded image
   *
   * @example
   * ```typescript
   * const region = { x: 100, y: 100, width: 300, height: 200 };
   * const screenshot = await sandbox.computerUse.screenshot.takeRegion(region);
   * console.log(`Captured region: ${screenshot.region.width}x${screenshot.region.height}`);
   * ```
   */
  @WithInstrumentation()
  public async takeRegion(region: ScreenshotRegion, showCursor = false): Promise<ScreenshotResponse> {
    const response = await this.apiClient.takeRegionScreenshot(
      region.height,
      region.width,
      region.y,
      region.x,
      showCursor,
    )
    return response.data
  }

  /**
   * Takes a compressed screenshot of the entire screen
   *
   * @param {ScreenshotOptions} [options={}] - Compression and display options
   * @returns {Promise<CompressedScreenshotResponse>} Compressed screenshot data
   *
   * @example
   * ```typescript
   * // Default compression
   * const screenshot = await sandbox.computerUse.screenshot.takeCompressed();
   *
   * // High quality JPEG
   * const jpeg = await sandbox.computerUse.screenshot.takeCompressed({
   *   format: 'jpeg',
   *   quality: 95,
   *   showCursor: true
   * });
   *
   * // Scaled down PNG
   * const scaled = await sandbox.computerUse.screenshot.takeCompressed({
   *   format: 'png',
   *   scale: 0.5
   * });
   * ```
   */
  @WithInstrumentation()
  public async takeCompressed(options: ScreenshotOptions = {}): Promise<ScreenshotResponse> {
    const response = await this.apiClient.takeCompressedScreenshot(
      options.showCursor,
      options.format,
      options.quality,
      options.scale,
    )
    return response.data
  }

  /**
   * Takes a compressed screenshot of a specific region
   *
   * @param {ScreenshotRegion} region - The region to capture
   * @param {ScreenshotOptions} [options={}] - Compression and display options
   * @returns {Promise<CompressedScreenshotResponse>} Compressed screenshot data
   *
   * @example
   * ```typescript
   * const region = { x: 0, y: 0, width: 800, height: 600 };
   * const screenshot = await sandbox.computerUse.screenshot.takeCompressedRegion(region, {
   *   format: 'webp',
   *   quality: 80,
   *   showCursor: true
   * });
   * console.log(`Compressed size: ${screenshot.size_bytes} bytes`);
   * ```
   */
  @WithInstrumentation()
  public async takeCompressedRegion(
    region: ScreenshotRegion,
    options: ScreenshotOptions = {},
  ): Promise<ScreenshotResponse> {
    const response = await this.apiClient.takeCompressedRegionScreenshot(
      region.x,
      region.y,
      region.width,
      region.height,
      options.showCursor,
      options.format,
      options.quality,
      options.scale,
    )
    return response.data
  }
}

/**
 * Display operations for computer use functionality
 */
export class Display {
  constructor(private readonly apiClient: ComputerUseApi) {}

  /**
   * Gets information about the displays
   *
   * @returns {Promise<DisplayInfoResponse>} Display information including primary display and all available displays
   *
   * @example
   * ```typescript
   * const info = await sandbox.computerUse.display.getInfo();
   * console.log(`Primary display: ${info.primary_display.width}x${info.primary_display.height}`);
   * console.log(`Total displays: ${info.total_displays}`);
   * info.displays.forEach((display, index) => {
   *   console.log(`Display ${index}: ${display.width}x${display.height} at ${display.x},${display.y}`);
   * });
   * ```
   */
  @WithInstrumentation()
  public async getInfo(): Promise<DisplayInfoResponse> {
    const response = await this.apiClient.getDisplayInfo()
    return response.data
  }

  /**
   * Gets the list of open windows
   *
   * @returns {Promise<WindowsResponse>} List of open windows with their IDs and titles
   *
   * @example
   * ```typescript
   * const windows = await sandbox.computerUse.display.getWindows();
   * console.log(`Found ${windows.count} open windows:`);
   * windows.windows.forEach(window => {
   *   console.log(`- ${window.title} (ID: ${window.id})`);
   * });
   * ```
   */
  @WithInstrumentation()
  public async getWindows(): Promise<WindowsResponse> {
    const response = await this.apiClient.getWindows()
    return response.data
  }
}

/**
 * Computer Use functionality for interacting with the desktop environment.
 *
 * Provides access to mouse, keyboard, screenshot, and display operations
 * for automating desktop interactions within a sandbox.
 *
 * @property {Mouse} mouse - Mouse operations interface
 * @property {Keyboard} keyboard - Keyboard operations interface
 * @property {Screenshot} screenshot - Screenshot operations interface
 * @property {Display} display - Display operations interface
 *
 * @class
 */
export class ComputerUse {
  public readonly mouse: Mouse
  public readonly keyboard: Keyboard
  public readonly screenshot: Screenshot
  public readonly display: Display

  constructor(private readonly apiClient: ComputerUseApi) {
    this.mouse = new Mouse(apiClient)
    this.keyboard = new Keyboard(apiClient)
    this.screenshot = new Screenshot(apiClient)
    this.display = new Display(apiClient)
  }

  /**
   * Starts all computer use processes (Xvfb, xfce4, x11vnc, novnc)
   *
   * @returns {Promise<ComputerUseStartResponse>} Computer use start response
   *
   * @example
   * ```typescript
   * const result = await sandbox.computerUse.start();
   * console.log('Computer use processes started:', result.message);
   * ```
   */
  @WithInstrumentation()
  public async start(): Promise<ComputerUseStartResponse> {
    const response = await this.apiClient.startComputerUse()
    return response.data
  }

  /**
   * Stops all computer use processes
   *
   * @returns {Promise<ComputerUseStopResponse>} Computer use stop response
   *
   * @example
   * ```typescript
   * const result = await sandbox.computerUse.stop();
   * console.log('Computer use processes stopped:', result.message);
   * ```
   */
  @WithInstrumentation()
  public async stop(): Promise<ComputerUseStopResponse> {
    const response = await this.apiClient.stopComputerUse()
    return response.data
  }

  /**
   * Gets the status of all computer use processes
   *
   * @returns {Promise<ComputerUseStatusResponse>} Status information about all VNC desktop processes
   *
   * @example
   * ```typescript
   * const status = await sandbox.computerUse.getStatus();
   * console.log('Computer use status:', status.status);
   * ```
   */
  @WithInstrumentation()
  public async getStatus(): Promise<ComputerUseStatusResponse> {
    const response = await this.apiClient.getComputerUseStatus()
    return response.data
  }

  /**
   * Gets the status of a specific VNC process
   *
   * @param {string} processName - Name of the process to check
   * @returns {Promise<ProcessStatusResponse>} Status information about the specific process
   *
   * @example
   * ```typescript
   * const xvfbStatus = await sandbox.computerUse.getProcessStatus('xvfb');
   * const noVncStatus = await sandbox.computerUse.getProcessStatus('novnc');
   * ```
   */
  @WithInstrumentation()
  public async getProcessStatus(processName: string): Promise<ProcessStatusResponse> {
    const response = await this.apiClient.getProcessStatus(processName)
    return response.data
  }

  /**
   * Restarts a specific VNC process
   *
   * @param {string} processName - Name of the process to restart
   * @returns {Promise<ProcessRestartResponse>} Process restart response
   *
   * @example
   * ```typescript
   * const result = await sandbox.computerUse.restartProcess('xfce4');
   * console.log('XFCE4 process restarted:', result.message);
   * ```
   */
  @WithInstrumentation()
  public async restartProcess(processName: string): Promise<ProcessRestartResponse> {
    const response = await this.apiClient.restartProcess(processName)
    return response.data
  }

  /**
   * Gets logs for a specific VNC process
   *
   * @param {string} processName - Name of the process to get logs for
   * @returns {Promise<ProcessLogsResponse>} Process logs
   *
   * @example
   * ```typescript
   * const logsResp = await sandbox.computerUse.getProcessLogs('novnc');
   * console.log('NoVNC logs:', logsResp.logs);
   * ```
   */
  @WithInstrumentation()
  public async getProcessLogs(processName: string): Promise<ProcessLogsResponse> {
    const response = await this.apiClient.getProcessLogs(processName)
    return response.data
  }

  /**
   * Gets error logs for a specific VNC process
   *
   * @param {string} processName - Name of the process to get error logs for
   * @returns {Promise<ProcessErrorsResponse>} Process error logs
   *
   * @example
   * ```typescript
   * const errorsResp = await sandbox.computerUse.getProcessErrors('x11vnc');
   * console.log('X11VNC errors:', errorsResp.errors);
   * ```
   */
  @WithInstrumentation()
  public async getProcessErrors(processName: string): Promise<ProcessErrorsResponse> {
    const response = await this.apiClient.getProcessErrors(processName)
    return response.data
  }
}
