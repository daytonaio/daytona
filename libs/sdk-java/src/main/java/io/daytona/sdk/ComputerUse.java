// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.toolbox.client.api.ComputerUseApi;
import io.daytona.toolbox.client.model.ComputerUseStartResponse;
import io.daytona.toolbox.client.model.ComputerUseStatusResponse;
import io.daytona.toolbox.client.model.ComputerUseStopResponse;
import io.daytona.toolbox.client.model.DisplayInfoResponse;
import io.daytona.toolbox.client.model.KeyboardHotkeyRequest;
import io.daytona.toolbox.client.model.KeyboardPressRequest;
import io.daytona.toolbox.client.model.KeyboardTypeRequest;
import io.daytona.toolbox.client.model.ListRecordingsResponse;
import io.daytona.toolbox.client.model.MouseClickRequest;
import io.daytona.toolbox.client.model.MouseClickResponse;
import io.daytona.toolbox.client.model.MouseDragRequest;
import io.daytona.toolbox.client.model.MouseDragResponse;
import io.daytona.toolbox.client.model.MouseMoveRequest;
import io.daytona.toolbox.client.model.MousePositionResponse;
import io.daytona.toolbox.client.model.MouseScrollRequest;
import io.daytona.toolbox.client.model.Recording;
import io.daytona.toolbox.client.model.ScreenshotResponse;
import io.daytona.toolbox.client.model.ScrollResponse;
import io.daytona.toolbox.client.model.StartRecordingRequest;
import io.daytona.toolbox.client.model.StopRecordingRequest;
import io.daytona.toolbox.client.model.WindowsResponse;

import java.io.File;
import java.math.BigDecimal;
import java.util.Arrays;

/**
 * Desktop automation operations for a Sandbox.
 *
 * <p>Provides a Java facade for computer-use features including desktop session management,
 * screenshots, mouse and keyboard automation, display/window inspection, and screen recording.
 */
public class ComputerUse {
    private final ComputerUseApi computerUseApi;

    ComputerUse(ComputerUseApi computerUseApi) {
        this.computerUseApi = computerUseApi;
    }

    /**
     * Starts the computer-use desktop stack (VNC/noVNC and related processes).
     *
     * @return start response containing process status details
     */
    public ComputerUseStartResponse start() {
        return ExceptionMapper.callToolbox(computerUseApi::startComputerUse);
    }

    /**
     * Stops all computer-use desktop processes.
     *
     * @return stop response containing process status details
     */
    public ComputerUseStopResponse stop() {
        return ExceptionMapper.callToolbox(computerUseApi::stopComputerUse);
    }

    /**
     * Returns current computer-use status.
     *
     * @return overall computer-use status
     */
    public ComputerUseStatusResponse getStatus() {
        return ExceptionMapper.callToolbox(computerUseApi::getComputerUseStatus);
    }

    /**
     * Captures a full-screen screenshot without cursor.
     *
     * @return screenshot payload (base64 image and metadata)
     */
    public ScreenshotResponse takeScreenshot() {
        return takeScreenshot(false);
    }

    /**
     * Captures a full-screen screenshot.
     *
     * @param showCursor whether to render cursor in the screenshot
     * @return screenshot payload (base64 image and metadata)
     */
    public ScreenshotResponse takeScreenshot(boolean showCursor) {
        return ExceptionMapper.callToolbox(() -> computerUseApi.takeScreenshot(showCursor));
    }

    /**
     * Captures a screenshot of a rectangular region without cursor.
     *
     * @param x region top-left X coordinate
     * @param y region top-left Y coordinate
     * @param width region width in pixels
     * @param height region height in pixels
     * @return region screenshot payload
     */
    public ScreenshotResponse takeRegionScreenshot(int x, int y, int width, int height) {
        return ExceptionMapper.callToolbox(() -> computerUseApi.takeRegionScreenshot(x, y, width, height, false));
    }

    /**
     * Captures a compressed full-screen screenshot.
     *
     * @param format output image format (for example: {@code png}, {@code jpeg}, {@code webp})
     * @param quality compression quality (typically 1-100, format dependent)
     * @param scale screenshot scale factor (for example: {@code 0.5} for 50%)
     * @return compressed screenshot payload
     */
    public ScreenshotResponse takeCompressedScreenshot(String format, int quality, double scale) {
        return ExceptionMapper.callToolbox(
                () -> computerUseApi.takeCompressedScreenshot(false, format, quality, BigDecimal.valueOf(scale))
        );
    }

    /**
     * Performs a left mouse click at the given coordinates.
     *
     * @param x target X coordinate
     * @param y target Y coordinate
     * @return click response with resulting cursor position
     */
    public MouseClickResponse click(int x, int y) {
        return click(x, y, "left");
    }

    /**
     * Performs a mouse click at the given coordinates with a specific button.
     *
     * @param x target X coordinate
     * @param y target Y coordinate
     * @param button button type ({@code left}, {@code right}, {@code middle})
     * @return click response with resulting cursor position
     */
    public MouseClickResponse click(int x, int y, String button) {
        MouseClickRequest request = new MouseClickRequest()
                .x(x)
                .y(y)
                .button(button)
                ._double(false);
        return ExceptionMapper.callToolbox(() -> computerUseApi.click(request));
    }

    /**
     * Performs a double left-click at the given coordinates.
     *
     * @param x target X coordinate
     * @param y target Y coordinate
     * @return click response with resulting cursor position
     */
    public MouseClickResponse doubleClick(int x, int y) {
        MouseClickRequest request = new MouseClickRequest()
                .x(x)
                .y(y)
                .button("left")
                ._double(true);
        return ExceptionMapper.callToolbox(() -> computerUseApi.click(request));
    }

    /**
     * Moves the mouse cursor to the given coordinates.
     *
     * @param x target X coordinate
     * @param y target Y coordinate
     * @return new mouse position
     */
    public MousePositionResponse moveMouse(int x, int y) {
        MouseMoveRequest request = new MouseMoveRequest().x(x).y(y);
        return ExceptionMapper.callToolbox(() -> computerUseApi.moveMouse(request));
    }

    /**
     * Returns current mouse position.
     *
     * @return current mouse cursor coordinates
     */
    public MousePositionResponse getMousePosition() {
        return ExceptionMapper.callToolbox(computerUseApi::getMousePosition);
    }

    /**
     * Drags the mouse from one point to another using the left button.
     *
     * @param startX drag start X coordinate
     * @param startY drag start Y coordinate
     * @param endX drag end X coordinate
     * @param endY drag end Y coordinate
     * @return drag response with resulting cursor position
     */
    public MouseDragResponse drag(int startX, int startY, int endX, int endY) {
        MouseDragRequest request = new MouseDragRequest()
                .startX(startX)
                .startY(startY)
                .endX(endX)
                .endY(endY)
                .button("left");
        return ExceptionMapper.callToolbox(() -> computerUseApi.drag(request));
    }

    /**
     * Scrolls at the given coordinates.
     *
     * <p>The current toolbox API supports directional scrolling ({@code up}/{@code down}) with an
     * amount. This method maps {@code deltaY} to vertical scroll direction and magnitude.
     * If {@code deltaY} is {@code 0}, {@code deltaX} is used as a fallback.
     *
     * @param x anchor X coordinate
     * @param y anchor Y coordinate
     * @param deltaX horizontal delta (used only when {@code deltaY == 0})
     * @param deltaY vertical delta
     * @return scroll response indicating operation success
     */
    public ScrollResponse scroll(int x, int y, int deltaX, int deltaY) {
        int effectiveDelta = deltaY != 0 ? deltaY : deltaX;
        String direction = effectiveDelta < 0 ? "up" : "down";
        int amount = Math.abs(effectiveDelta);

        MouseScrollRequest request = new MouseScrollRequest()
                .x(x)
                .y(y)
                .direction(direction)
                .amount(amount);
        return ExceptionMapper.callToolbox(() -> computerUseApi.scroll(request));
    }

    /**
     * Types text using keyboard automation.
     *
     * @param text text to type
     */
    public void typeText(String text) {
        KeyboardTypeRequest request = new KeyboardTypeRequest().text(text);
        ExceptionMapper.callToolbox(() -> computerUseApi.typeText(request));
    }

    /**
     * Presses a single key.
     *
     * @param key key to press (for example: {@code Enter}, {@code Escape}, {@code a})
     */
    public void pressKey(String key) {
        KeyboardPressRequest request = new KeyboardPressRequest().key(key);
        ExceptionMapper.callToolbox(() -> computerUseApi.pressKey(request));
    }

    /**
     * Presses a key combination as a hotkey sequence.
     *
     * <p>Keys are joined with {@code +} before being sent (for example,
     * {@code pressHotkey("ctrl", "shift", "t") -> "ctrl+shift+t"}).
     *
     * @param keys hotkey parts to combine
     */
    public void pressHotkey(String... keys) {
        String joined = String.join("+", Arrays.asList(keys));
        KeyboardHotkeyRequest request = new KeyboardHotkeyRequest().keys(joined);
        ExceptionMapper.callToolbox(() -> computerUseApi.pressHotkey(request));
    }

    /**
     * Returns display configuration information.
     *
     * @return display information including available displays and their geometry
     */
    public DisplayInfoResponse getDisplayInfo() {
        return ExceptionMapper.callToolbox(computerUseApi::getDisplayInfo);
    }

    /**
     * Returns currently open windows.
     *
     * @return window list and metadata
     */
    public WindowsResponse getWindows() {
        return ExceptionMapper.callToolbox(computerUseApi::getWindows);
    }

    /**
     * Starts a recording with default options.
     *
     * @return newly started recording metadata
     */
    public Recording startRecording() {
        return startRecording(null);
    }

    /**
     * Starts a recording with an optional label.
     *
     * @param label optional recording label
     * @return newly started recording metadata
     */
    public Recording startRecording(String label) {
        StartRecordingRequest request = new StartRecordingRequest().label(label);
        return ExceptionMapper.callToolbox(() -> computerUseApi.startRecording(request));
    }

    /**
     * Stops an active recording.
     *
     * @param id recording identifier
     * @return finalized recording metadata
     */
    public Recording stopRecording(String id) {
        StopRecordingRequest request = new StopRecordingRequest().id(id);
        return ExceptionMapper.callToolbox(() -> computerUseApi.stopRecording(request));
    }

    /**
     * Lists all recordings for the current sandbox session.
     *
     * @return recordings list response
     */
    public ListRecordingsResponse listRecordings() {
        return ExceptionMapper.callToolbox(computerUseApi::listRecordings);
    }

    /**
     * Returns metadata for a specific recording.
     *
     * @param id recording identifier
     * @return recording details
     */
    public Recording getRecording(String id) {
        return ExceptionMapper.callToolbox(() -> computerUseApi.getRecording(id));
    }

    /**
     * Downloads a recording file.
     *
     * @param id recording identifier
     * @return downloaded temporary/local file handle returned by the API client
     */
    public File downloadRecording(String id) {
        return ExceptionMapper.callToolbox(() -> computerUseApi.downloadRecording(id));
    }

    /**
     * Deletes a recording.
     *
     * @param id recording identifier
     */
    public void deleteRecording(String id) {
        ExceptionMapper.runToolbox(() -> computerUseApi.deleteRecording(id));
    }
}
