// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.toolbox.client.api.ComputerUseApi;
import io.daytona.toolbox.client.model.ListRecordingsResponse;
import io.daytona.toolbox.client.model.MouseDragResponse;
import io.daytona.toolbox.client.model.MousePositionResponse;
import io.daytona.toolbox.client.model.Recording;
import io.daytona.toolbox.client.model.ScreenshotResponse;
import io.daytona.toolbox.client.model.ComputerUseStatusResponse;
import io.daytona.toolbox.client.model.DisplayInfoResponse;
import io.daytona.toolbox.client.model.KeyboardHotkeyRequest;
import io.daytona.toolbox.client.model.MouseClickRequest;
import io.daytona.toolbox.client.model.MouseClickResponse;
import io.daytona.toolbox.client.model.ScrollResponse;
import io.daytona.toolbox.client.model.WindowsResponse;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.io.File;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class ComputerUseTest {

    @Mock
    private ComputerUseApi computerUseApi;

    private ComputerUse computerUse;

    @BeforeEach
    void setUp() {
        computerUse = new ComputerUse(computerUseApi);
    }

    @Test
    void startStopAndStatusDelegate() {
        ComputerUseStatusResponse response = new ComputerUseStatusResponse();
        when(computerUseApi.getComputerUseStatus()).thenReturn(response);

        assertThat(computerUse.getStatus()).isSameAs(response);
        computerUse.start();
        computerUse.stop();

        verify(computerUseApi).startComputerUse();
        verify(computerUseApi).stopComputerUse();
    }

    @Test
    void screenshotOperationsDelegate() {
        ScreenshotResponse screenshot = new ScreenshotResponse();
        when(computerUseApi.takeScreenshot(false)).thenReturn(screenshot);
        when(computerUseApi.takeScreenshot(true)).thenReturn(screenshot);
        when(computerUseApi.takeRegionScreenshot(1, 2, 3, 4, false)).thenReturn(screenshot);
        when(computerUseApi.takeCompressedScreenshot(false, "png", 75, java.math.BigDecimal.valueOf(0.5))).thenReturn(screenshot);

        assertThat(computerUse.takeScreenshot()).isSameAs(screenshot);
        assertThat(computerUse.takeScreenshot(true)).isSameAs(screenshot);
        assertThat(computerUse.takeRegionScreenshot(1, 2, 3, 4)).isSameAs(screenshot);
        assertThat(computerUse.takeCompressedScreenshot("png", 75, 0.5)).isSameAs(screenshot);
    }

    @Test
    void clickAndDoubleClickBuildRequests() {
        MouseClickResponse response = new MouseClickResponse();
        when(computerUseApi.click(any())).thenReturn(response);

        assertThat(computerUse.click(10, 20, "right")).isSameAs(response);
        assertThat(computerUse.doubleClick(1, 2)).isSameAs(response);

        ArgumentCaptor<MouseClickRequest> captor = ArgumentCaptor.forClass(MouseClickRequest.class);
        verify(computerUseApi, org.mockito.Mockito.times(2)).click(captor.capture());
        assertThat(captor.getAllValues().get(0).getButton()).isEqualTo("right");
        assertThat(captor.getAllValues().get(0).getDouble()).isFalse();
        assertThat(captor.getAllValues().get(1).getButton()).isEqualTo("left");
        assertThat(captor.getAllValues().get(1).getDouble()).isTrue();
    }

    @Test
    void scrollAndHotkeyBuildRequests() {
        ScrollResponse response = new ScrollResponse();
        when(computerUseApi.scroll(any())).thenReturn(response);

        assertThat(computerUse.scroll(10, 20, 0, -4)).isSameAs(response);
        computerUse.pressHotkey("ctrl", "shift", "t");

        verify(computerUseApi).scroll(argThat(request -> "up".equals(request.getDirection()) && request.getAmount() == 4));
        ArgumentCaptor<KeyboardHotkeyRequest> captor = ArgumentCaptor.forClass(KeyboardHotkeyRequest.class);
        verify(computerUseApi).pressHotkey(captor.capture());
        assertThat(captor.getValue().getKeys()).isEqualTo("ctrl+shift+t");
    }

    @ParameterizedTest
    @CsvSource({"5,0,down,5", "0,4,down,4", "0,-3,up,3"})
    void scrollUsesDeltaYOrDeltaXFallback(int deltaX, int deltaY, String direction, int amount) {
        when(computerUseApi.scroll(any())).thenReturn(new ScrollResponse());

        computerUse.scroll(10, 20, deltaX, deltaY);

        verify(computerUseApi).scroll(argThat(request -> direction.equals(request.getDirection()) && amount == request.getAmount()));
    }

    @Test
    void delegatesReadOnlyDesktopQueries() {
        DisplayInfoResponse displayInfo = new DisplayInfoResponse();
        WindowsResponse windows = new WindowsResponse();
        MousePositionResponse position = new MousePositionResponse();
        MouseDragResponse drag = new MouseDragResponse();
        when(computerUseApi.getDisplayInfo()).thenReturn(displayInfo);
        when(computerUseApi.getWindows()).thenReturn(windows);
        when(computerUseApi.getMousePosition()).thenReturn(position);
        when(computerUseApi.drag(any())).thenReturn(drag);

        assertThat(computerUse.getDisplayInfo()).isSameAs(displayInfo);
        assertThat(computerUse.getWindows()).isSameAs(windows);
        assertThat(computerUse.getMousePosition()).isSameAs(position);
        assertThat(computerUse.drag(1, 2, 3, 4)).isSameAs(drag);
    }

    @Test
    void keyboardAndMouseMovementDelegate() {
        MousePositionResponse position = new MousePositionResponse();
        when(computerUseApi.moveMouse(any())).thenReturn(position);

        assertThat(computerUse.moveMouse(10, 11)).isSameAs(position);
        computerUse.typeText("hello");
        computerUse.pressKey("Enter");

        verify(computerUseApi).moveMouse(argThat(request -> request.getX() == 10 && request.getY() == 11));
        verify(computerUseApi).typeText(argThat(request -> "hello".equals(request.getText())));
        verify(computerUseApi).pressKey(argThat(request -> "Enter".equals(request.getKey())));
    }

    @Test
    void recordingOperationsDelegate() throws Exception {
        Recording recording = new Recording();
        recording.setId("rec-1");
        File file = File.createTempFile("recording", ".mp4");
        ListRecordingsResponse list = new ListRecordingsResponse();
        when(computerUseApi.startRecording(any())).thenReturn(recording);
        when(computerUseApi.stopRecording(any())).thenReturn(recording);
        when(computerUseApi.listRecordings()).thenReturn(list);
        when(computerUseApi.getRecording("rec-1")).thenReturn(recording);
        when(computerUseApi.downloadRecording("rec-1")).thenReturn(file);

        assertThat(computerUse.startRecording()).isSameAs(recording);
        assertThat(computerUse.startRecording("demo")).isSameAs(recording);
        assertThat(computerUse.stopRecording("rec-1")).isSameAs(recording);
        assertThat(computerUse.listRecordings()).isSameAs(list);
        assertThat(computerUse.getRecording("rec-1")).isSameAs(recording);
        assertThat(computerUse.downloadRecording("rec-1")).isEqualTo(file);
        computerUse.deleteRecording("rec-1");

        verify(computerUseApi).deleteRecording("rec-1");
    }

    private static <T> T argThat(org.mockito.ArgumentMatcher<T> matcher) {
        return org.mockito.ArgumentMatchers.argThat(matcher);
    }
}
