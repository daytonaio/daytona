// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.api.client.ApiClient;
import io.daytona.api.client.api.SandboxApi;
import io.daytona.api.client.model.CreateSandboxSnapshot;
import io.daytona.api.client.model.ForkSandbox;
import io.daytona.api.client.model.SandboxState;
import io.daytona.api.client.model.ToolboxProxyUrl;
import io.daytona.sdk.exception.DaytonaBadRequestException;
import io.daytona.sdk.exception.DaytonaConflictException;
import io.daytona.sdk.exception.DaytonaForbiddenException;
import io.daytona.sdk.exception.DaytonaNotFoundException;
import io.daytona.sdk.exception.DaytonaRateLimitException;
import io.daytona.sdk.exception.DaytonaServerException;
import io.daytona.sdk.exception.DaytonaValidationException;
import io.daytona.toolbox.client.api.InfoApi;
import io.daytona.toolbox.client.model.UserHomeDirResponse;
import io.daytona.toolbox.client.model.WorkDirResponse;
import okhttp3.Call;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.math.BigDecimal;
import java.util.Collections;
import java.util.HashMap;
import java.util.Map;
import java.util.stream.Stream;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.ArgumentMatchers.isNull;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class SandboxTest {

    @Mock
    private SandboxApi sandboxApi;

    @Mock
    private InfoApi infoApi;

    private Sandbox sandbox;

    @BeforeEach
    void setUp() {
        sandbox = new Sandbox(sandboxApi, TestSupport.config(), TestSupport.mainSandbox("sb-1", SandboxState.STARTED));
    }

    @Test
    void constructorPopulatesFieldsAndChildServices() {
        assertThat(sandbox.getId()).isEqualTo("sb-1");
        assertThat(sandbox.getName()).isEqualTo("sandbox-sb-1");
        assertThat(sandbox.getState()).isEqualTo("started");
        assertThat(sandbox.getProcess()).isNotNull();
        assertThat(sandbox.getFs()).isNotNull();
        assertThat(sandbox.getGit()).isNotNull();
        assertThat(sandbox.computerUse).isNotNull();
        assertThat(sandbox.codeInterpreter).isNotNull();
        assertThat(sandbox.createLspServer("python", "/project")).isNotNull();
        assertThat(sandbox.getLanguage()).isEqualTo("python");
    }

    @Test
    void constructorLoadsToolboxProxyUrlWhenMissing() {
        io.daytona.api.client.model.Sandbox model = TestSupport.mainSandbox("sb-2", SandboxState.STARTED);
        model.setToolboxProxyUrl("");
        when(sandboxApi.getToolboxProxyUrl("sb-2", null)).thenReturn(new ToolboxProxyUrl().url("https://proxy.example"));

        Sandbox loaded = new Sandbox(sandboxApi, TestSupport.config(), model);

        assertThat(loaded.getToolboxProxyUrl()).isEmpty();
        assertThat(loaded.getToolboxApiClient().getBasePath()).isEqualTo("https://proxy.example/sb-2");
    }

    @Test
    void constructorUsesLanguageFromLabelsAndFallsBackToEmptyStrings() {
        io.daytona.api.client.model.Sandbox model = new io.daytona.api.client.model.Sandbox();
        model.setId("sb-3");
        model.setState(SandboxState.STARTED);
        model.setLabels(Collections.singletonMap(Daytona.CODE_TOOLBOX_LANGUAGE_LABEL, "javascript"));
        model.setToolboxProxyUrl("https://proxy.example/");

        Sandbox loaded = new Sandbox(sandboxApi, TestSupport.config(), model);

        assertThat(loaded.getLanguage()).isEqualTo("javascript");
        assertThat(loaded.getName()).isEmpty();
        assertThat(loaded.getEnv()).isEmpty();
        assertThat(loaded.getToolboxApiClient().getBasePath()).isEqualTo("https://proxy.example/sb-3");
    }

    @Test
    void startUpdatesStateAndWaitsUntilStarted() {
        when(sandboxApi.startSandbox("sb-1", null)).thenReturn(TestSupport.mainSandbox("sb-1", SandboxState.STARTING));
        when(sandboxApi.getSandbox("sb-1", null, null)).thenReturn(TestSupport.mainSandbox("sb-1", SandboxState.STARTED));

        sandbox.start(1);

        assertThat(sandbox.getState()).isEqualTo("started");
    }

    @Test
    void waitUntilStartedRejectsInvalidTimeoutAndFailureState() {
        assertThatThrownBy(() -> sandbox.waitUntilStarted(-1))
                .hasMessageContaining("Timeout must be non-negative");

        when(sandboxApi.getSandbox("sb-1", null, null)).thenReturn(TestSupport.mainSandbox("sb-1", SandboxState.ERROR));
        TestSupport.setField(sandbox, "state", "starting");

        assertThatThrownBy(() -> sandbox.waitUntilStarted(1))
                .hasMessageContaining("Sandbox entered failure state");
    }

    @Test
    void waitUntilStartedTimesOutWhenStateNeverChanges() {
        when(sandboxApi.getSandbox("sb-1", null, null)).thenReturn(TestSupport.mainSandbox("sb-1", SandboxState.STARTING));
        TestSupport.setField(sandbox, "state", "starting");

        assertThatThrownBy(() -> sandbox.waitUntilStarted(1))
                .hasMessageContaining("Sandbox failed to become started before timeout");
    }

    @Test
    void stopRefreshesAndWaitsUntilStopped() {
        when(sandboxApi.getSandbox("sb-1", null, null)).thenReturn(TestSupport.mainSandbox("sb-1", SandboxState.STOPPED));

        sandbox.stop(1);

        assertThat(sandbox.getState()).isEqualTo("stopped");
        verify(sandboxApi).stopSandbox("sb-1", null, null);
    }

    @Test
    void waitUntilStoppedRejectsFailureAndInvalidTimeout() {
        assertThatThrownBy(() -> sandbox.waitUntilStopped(-1))
                .hasMessageContaining("Timeout must be non-negative");

        when(sandboxApi.getSandbox("sb-1", null, null)).thenReturn(TestSupport.mainSandbox("sb-1", SandboxState.ERROR));
        TestSupport.setField(sandbox, "state", "stopping");

        assertThatThrownBy(() -> sandbox.waitUntilStopped(1))
                .hasMessageContaining("Sandbox entered error state while stopping");
    }

    @Test
    void waitUntilStoppedAllowsDestroyedState() {
        when(sandboxApi.getSandbox("sb-1", null, null)).thenReturn(TestSupport.mainSandbox("sb-1", SandboxState.DESTROYED));
        TestSupport.setField(sandbox, "state", "stopping");

        sandbox.waitUntilStopped(1);

        assertThat(sandbox.getState()).isEqualTo("destroyed");
    }

    @Test
    void deleteAndRefreshDelegate() {
        when(sandboxApi.getSandbox("sb-1", null, null)).thenReturn(TestSupport.mainSandbox("sb-1", SandboxState.STARTED));

        sandbox.refreshData();
        sandbox.delete(5);

        verify(sandboxApi).getSandbox("sb-1", null, null);
        verify(sandboxApi).deleteSandbox("sb-1", null);
    }

    @Test
    void refreshDataIgnoresNullApiPayload() {
        TestSupport.setField(sandbox, "name", "before");
        when(sandboxApi.getSandbox("sb-1", null, null)).thenReturn(null);

        sandbox.refreshData();

        assertThat(sandbox.getName()).isEqualTo("before");
    }

    @Test
    void setLabelsAndIntervalsUpdateLocalState() throws Exception {
        Call call = org.mockito.Mockito.mock(Call.class);
        ApiClient apiClient = org.mockito.Mockito.mock(ApiClient.class);
        when(sandboxApi.replaceLabelsCall(eq("sb-1"), any(), isNull(), isNull())).thenReturn(call);
        when(sandboxApi.getApiClient()).thenReturn(apiClient);
        when(apiClient.execute(eq(call), isNull())).thenReturn(null);

        io.daytona.api.client.model.Sandbox refreshed = TestSupport.mainSandbox("sb-1", SandboxState.STARTED);
        refreshed.setLabels(Collections.singletonMap("team", "sdk"));
        refreshed.setAutoStopInterval(BigDecimal.ONE);
        refreshed.setAutoArchiveInterval(BigDecimal.valueOf(2));
        refreshed.setAutoDeleteInterval(BigDecimal.valueOf(3));
        when(sandboxApi.getSandbox("sb-1", null, null)).thenReturn(refreshed);
        when(sandboxApi.setAutostopInterval("sb-1", BigDecimal.valueOf(1), null)).thenReturn(refreshed);
        when(sandboxApi.setAutoArchiveInterval("sb-1", BigDecimal.valueOf(2), null)).thenReturn(refreshed);
        when(sandboxApi.setAutoDeleteInterval("sb-1", BigDecimal.valueOf(3), null)).thenReturn(refreshed);

        Map<String, String> labels = sandbox.setLabels(Collections.singletonMap("team", "sdk"));
        sandbox.setAutostopInterval(1);
        sandbox.setAutoArchiveInterval(2);
        sandbox.setAutoDeleteInterval(3);

        assertThat(labels).containsEntry("team", "sdk");
        assertThat(sandbox.getAutoStopInterval()).isEqualTo(1);
        assertThat(sandbox.getAutoArchiveInterval()).isEqualTo(2);
        assertThat(sandbox.getAutoDeleteInterval()).isEqualTo(3);
    }

    @Test
    void userDirectoryMethodsUseInfoApi() {
        TestSupport.setField(sandbox, "infoApi", infoApi);
        when(infoApi.getUserHomeDir()).thenReturn(new UserHomeDirResponse().dir("/home/daytona"));
        when(infoApi.getWorkDir()).thenReturn(new WorkDirResponse().dir("/workspace"));

        assertThat(sandbox.getUserHomeDir()).isEqualTo("/home/daytona");
        assertThat(sandbox.getWorkDir()).isEqualTo("/workspace");
    }

    @Test
    void userDirectoryMethodsReturnEmptyStringForNullResponses() {
        TestSupport.setField(sandbox, "infoApi", infoApi);
        when(infoApi.getUserHomeDir()).thenReturn(null);
        when(infoApi.getWorkDir()).thenReturn(null);

        assertThat(sandbox.getUserHomeDir()).isEmpty();
        assertThat(sandbox.getWorkDir()).isEmpty();
    }

    @Test
    void experimentalForkAndSnapshotDelegate() {
        when(sandboxApi.forkSandbox(eq("sb-1"), any(ForkSandbox.class), isNull())).thenReturn(TestSupport.mainSandbox("sb-2", SandboxState.STARTED));
        io.daytona.api.client.model.Sandbox snapshotting = TestSupport.mainSandbox("sb-1", SandboxState.SNAPSHOTTING);
        io.daytona.api.client.model.Sandbox started = TestSupport.mainSandbox("sb-1", SandboxState.STARTED);
        when(sandboxApi.getSandbox("sb-1", null, null)).thenReturn(snapshotting, started);

        Sandbox forked = sandbox.experimentalFork("forked", 1);
        sandbox.experimentalCreateSnapshot("snap-1", 1);

        assertThat(forked.getId()).isEqualTo("sb-2");
        verify(sandboxApi).createSandboxSnapshot(eq("sb-1"), any(CreateSandboxSnapshot.class), isNull());
    }

    @Test
    void experimentalForkDefaultArgumentsOmitName() {
        when(sandboxApi.forkSandbox(eq("sb-1"), any(ForkSandbox.class), isNull())).thenReturn(TestSupport.mainSandbox("sb-4", SandboxState.STARTED));

        sandbox.experimentalFork();

        ArgumentCaptor<ForkSandbox> captor = ArgumentCaptor.forClass(ForkSandbox.class);
        verify(sandboxApi).forkSandbox(eq("sb-1"), captor.capture(), isNull());
        assertThat(captor.getValue().getName()).isNull();
    }

    @Test
    void experimentalCreateSnapshotFailsForErrorState() {
        io.daytona.api.client.model.Sandbox snapshotting = TestSupport.mainSandbox("sb-1", SandboxState.SNAPSHOTTING);
        io.daytona.api.client.model.Sandbox error = TestSupport.mainSandbox("sb-1", SandboxState.ERROR);
        when(sandboxApi.getSandbox("sb-1", null, null)).thenReturn(snapshotting, error);

        assertThatThrownBy(() -> sandbox.experimentalCreateSnapshot("snap-err", 1))
                .hasMessageContaining("Sandbox snapshot failed with state: error");
    }

    @Test
    void experimentalCreateSnapshotTimesOutWhenSnapshottingPersists() {
        io.daytona.api.client.model.Sandbox snapshotting = TestSupport.mainSandbox("sb-1", SandboxState.SNAPSHOTTING);
        when(sandboxApi.getSandbox("sb-1", null, null)).thenReturn(snapshotting, snapshotting, snapshotting, snapshotting, snapshotting);

        assertThatThrownBy(() -> sandbox.experimentalCreateSnapshot("snap-timeout", 1))
                .hasMessageContaining("Sandbox snapshot did not complete before timeout");
    }

    @ParameterizedTest
    @MethodSource("mappedMainApiExceptions")
    void lifecycleMethodsMapApiErrors(int status, Class<? extends RuntimeException> type) {
        when(sandboxApi.startSandbox("sb-1", null))
                .thenThrow(new io.daytona.api.client.ApiException(status, "boom", null, "{\"message\":\"mapped\"}"));

        assertThatThrownBy(() -> sandbox.start(1))
                .isInstanceOf(type)
                .hasMessage("mapped");
    }

    private static Stream<Arguments> mappedMainApiExceptions() {
        return Stream.of(
                Arguments.of(400, DaytonaBadRequestException.class),
                Arguments.of(403, DaytonaForbiddenException.class),
                Arguments.of(404, DaytonaNotFoundException.class),
                Arguments.of(409, DaytonaConflictException.class),
                Arguments.of(422, DaytonaValidationException.class),
                Arguments.of(429, DaytonaRateLimitException.class),
                Arguments.of(500, DaytonaServerException.class)
        );
    }
}
