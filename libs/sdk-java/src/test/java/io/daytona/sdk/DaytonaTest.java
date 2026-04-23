// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.api.client.api.SandboxApi;
import io.daytona.api.client.model.CreateBuildInfo;
import io.daytona.api.client.model.CreateSandbox;
import io.daytona.api.client.model.PaginatedSandboxes;
import io.daytona.api.client.model.SandboxState;
import io.daytona.api.client.model.Url;
import io.daytona.sdk.exception.DaytonaAuthenticationException;
import io.daytona.sdk.exception.DaytonaBadRequestException;
import io.daytona.sdk.exception.DaytonaConflictException;
import io.daytona.sdk.exception.DaytonaException;
import io.daytona.sdk.exception.DaytonaForbiddenException;
import io.daytona.sdk.exception.DaytonaNotFoundException;
import io.daytona.sdk.exception.DaytonaRateLimitException;
import io.daytona.sdk.exception.DaytonaServerException;
import io.daytona.sdk.exception.DaytonaValidationException;
import io.daytona.sdk.model.CreateSandboxFromImageParams;
import io.daytona.sdk.model.CreateSandboxFromSnapshotParams;
import io.daytona.sdk.model.Resources;
import io.daytona.sdk.model.VolumeMount;
import okhttp3.mockwebserver.MockResponse;
import okhttp3.mockwebserver.MockWebServer;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

import java.math.BigDecimal;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Stream;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.ArgumentMatchers.isNull;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class DaytonaTest {

    @Mock
    private SandboxApi sandboxApi;

    private Daytona daytona;

    @BeforeEach
    void setUp() {
        daytona = new Daytona(TestSupport.config());
        TestSupport.setField(daytona, "sandboxApi", sandboxApi);
    }

    @Test
    void constructorValidatesConfiguration() {
        assertThatThrownBy(() -> new Daytona((DaytonaConfig) null))
                .isInstanceOf(DaytonaException.class)
                .hasMessage("DAYTONA_API_KEY is required");

        assertThatThrownBy(() -> new Daytona(new DaytonaConfig.Builder().apiKey("").build()))
                .isInstanceOf(DaytonaException.class)
                .hasMessage("DAYTONA_API_KEY is required");
    }

    @Test
    void constructorConfiguresUnderlyingApiClient() {
        io.daytona.api.client.ApiClient apiClient = TestSupport.getField(daytona, "apiClient", io.daytona.api.client.ApiClient.class);

        assertThat(apiClient.getBasePath()).isEqualTo("https://example.com/api");
        assertThat(apiClient.getAuthentications()).containsKey("oauth2");
    }

    @Test
    void createUsesDefaultSnapshotParamsAndWaitsUntilStarted() {
        when(sandboxApi.createSandbox(any(), isNull())).thenReturn(TestSupport.mainSandbox("sb-1", SandboxState.STARTING));
        when(sandboxApi.getSandbox("sb-1", null, null)).thenReturn(TestSupport.mainSandbox("sb-1", SandboxState.STARTED));

        Sandbox sandbox = daytona.create();

        assertThat(sandbox.getId()).isEqualTo("sb-1");
        ArgumentCaptor<CreateSandbox> captor = ArgumentCaptor.forClass(CreateSandbox.class);
        org.mockito.Mockito.verify(sandboxApi).createSandbox(captor.capture(), isNull());
        assertThat(captor.getValue().getLabels()).containsEntry(Daytona.CODE_TOOLBOX_LANGUAGE_LABEL, "python");
        assertThat(captor.getValue().getTarget()).isEqualTo("eu");
    }

    @Test
    void createFromImageStringBuildsDockerfile() {
        when(sandboxApi.createSandbox(any(), isNull())).thenReturn(TestSupport.mainSandbox("sb-1", SandboxState.STARTED));
        when(sandboxApi.getSandbox("sb-1", null, null)).thenReturn(TestSupport.mainSandbox("sb-1", SandboxState.STARTED));

        CreateSandboxFromImageParams params = new CreateSandboxFromImageParams();
        params.setImage("python:3.12-slim");
        daytona.create(params, 1);

        ArgumentCaptor<CreateSandbox> captor = ArgumentCaptor.forClass(CreateSandbox.class);
        org.mockito.Mockito.verify(sandboxApi).createSandbox(captor.capture(), isNull());
        assertThat(captor.getValue().getBuildInfo().getDockerfileContent()).isEqualTo("FROM python:3.12-slim\n");
    }

    @Test
    void createFromImageObjectAddsResources() {
        when(sandboxApi.createSandbox(any(), isNull())).thenReturn(TestSupport.mainSandbox("sb-1", SandboxState.STARTED));
        when(sandboxApi.getSandbox("sb-1", null, null)).thenReturn(TestSupport.mainSandbox("sb-1", SandboxState.STARTED));

        CreateSandboxFromImageParams params = new CreateSandboxFromImageParams();
        params.setImage(Image.base("python:3.12").runCommands("echo hi"));
        Resources resources = new Resources();
        resources.setCpu(2);
        resources.setGpu(1);
        resources.setMemory(4);
        resources.setDisk(8);
        params.setResources(resources);
        daytona.create(params, 1);

        ArgumentCaptor<CreateSandbox> captor = ArgumentCaptor.forClass(CreateSandbox.class);
        org.mockito.Mockito.verify(sandboxApi, org.mockito.Mockito.times(1)).createSandbox(captor.capture(), isNull());
        assertThat(captor.getValue().getBuildInfo().getDockerfileContent()).contains("RUN echo hi\n");
        assertThat(captor.getValue().getCpu()).isEqualTo(2);
        assertThat(captor.getValue().getGpu()).isEqualTo(1);
        assertThat(captor.getValue().getMemory()).isEqualTo(4);
        assertThat(captor.getValue().getDisk()).isEqualTo(8);
    }

    @Test
    void createFromSnapshotCopiesAllCommonFieldsAndNormalizesLanguage() {
        when(sandboxApi.createSandbox(any(), isNull())).thenReturn(TestSupport.mainSandbox("sb-9", SandboxState.STARTED));

        CreateSandboxFromSnapshotParams params = new CreateSandboxFromSnapshotParams();
        params.setName("sandbox-name");
        params.setUser("custom-user");
        params.setLanguage("typescript");
        params.setEnvVars(Collections.singletonMap("A", "1"));
        params.setLabels(Collections.singletonMap("team", "sdk"));
        params.setPublic(true);
        params.setAutoStopInterval(7);
        params.setAutoArchiveInterval(8);
        params.setAutoDeleteInterval(9);
        params.setNetworkBlockAll(true);
        params.setSnapshot("snap-1");
        VolumeMount mount = new VolumeMount();
        mount.setVolumeId("vol-1");
        mount.setMountPath("/workspace");
        params.setVolumes(Collections.singletonList(mount));

        daytona.create(params, 1);

        ArgumentCaptor<CreateSandbox> captor = ArgumentCaptor.forClass(CreateSandbox.class);
        org.mockito.Mockito.verify(sandboxApi).createSandbox(captor.capture(), isNull());
        CreateSandbox body = captor.getValue();
        assertThat(body.getName()).isEqualTo("sandbox-name");
        assertThat(body.getUser()).isEqualTo("custom-user");
        assertThat(body.getEnv()).containsEntry("A", "1");
        assertThat(body.getLabels())
                .containsEntry("team", "sdk")
                .containsEntry(Daytona.CODE_TOOLBOX_LANGUAGE_LABEL, "typescript");
        assertThat(body.getPublic()).isTrue();
        assertThat(body.getAutoStopInterval()).isEqualTo(7);
        assertThat(body.getAutoArchiveInterval()).isEqualTo(8);
        assertThat(body.getAutoDeleteInterval()).isEqualTo(9);
        assertThat(body.getNetworkBlockAll()).isTrue();
        assertThat(body.getSnapshot()).isEqualTo("snap-1");
        assertThat(body.getVolumes()).singleElement().satisfies(volume -> {
            assertThat(volume.getVolumeId()).isEqualTo("vol-1");
            assertThat(volume.getMountPath()).isEqualTo("/workspace");
        });
    }

    @Test
    void createFromSnapshotRejectsUnsupportedLanguage() {
        CreateSandboxFromSnapshotParams params = new CreateSandboxFromSnapshotParams();
        params.setLanguage("ruby");

        assertThatThrownBy(() -> daytona.create(params, 1))
                .isInstanceOf(DaytonaException.class)
                .hasMessageContaining("Invalid code-toolbox-language: ruby");
    }

    @Test
    void createFromImageWithoutImageLeavesBuildInfoUnset() {
        when(sandboxApi.createSandbox(any(), isNull())).thenReturn(TestSupport.mainSandbox("sb-10", SandboxState.STARTED));
        when(sandboxApi.getSandbox("sb-10", null, null)).thenReturn(TestSupport.mainSandbox("sb-10", SandboxState.STARTED));

        CreateSandboxFromImageParams params = new CreateSandboxFromImageParams();
        params.setImage("");

        daytona.create(params, 1);

        ArgumentCaptor<CreateSandbox> captor = ArgumentCaptor.forClass(CreateSandbox.class);
        org.mockito.Mockito.verify(sandboxApi).createSandbox(captor.capture(), isNull());
        assertThat(captor.getValue().getBuildInfo()).isNull();
        assertThat(captor.getValue().getLabels()).containsEntry(Daytona.CODE_TOOLBOX_LANGUAGE_LABEL, "python");
    }

    @Test
    void createFromImageStreamsBuildLogsForPendingBuildSandboxes() throws Exception {
        try (MockWebServer server = new MockWebServer()) {
            server.enqueue(new MockResponse().setBody("build-line-1\nbuild-line-2\n"));
            io.daytona.api.client.model.Sandbox created = TestSupport.mainSandbox("sb-build", SandboxState.PENDING_BUILD);
            io.daytona.api.client.model.Sandbox starting = TestSupport.mainSandbox("sb-build", SandboxState.STARTING);
            io.daytona.api.client.model.Sandbox started = TestSupport.mainSandbox("sb-build", SandboxState.STARTED);
            when(sandboxApi.createSandbox(any(), isNull())).thenReturn(created);
            when(sandboxApi.getSandbox("sb-build", null, null)).thenReturn(starting, started, started);
            when(sandboxApi.getBuildLogsUrl("sb-build", null)).thenReturn(new Url().url(server.url("/logs").toString()));

            List<String> lines = new ArrayList<String>();
            Sandbox sandbox = daytona.create(new CreateSandboxFromImageParams(), 2, lines::add);

            assertThat(sandbox.getId()).isEqualTo("sb-build");
            assertThat(lines).contains("build-line-1");
            assertThat(server.takeRequest().getPath()).isEqualTo("/logs?follow=true");
        }
    }

    @Test
    void getWrapsSandboxModel() {
        when(sandboxApi.getSandbox("sandbox-1", null, null)).thenReturn(TestSupport.mainSandbox("sandbox-1", SandboxState.STARTED));

        Sandbox sandbox = daytona.get("sandbox-1");

        assertThat(sandbox.getId()).isEqualTo("sandbox-1");
        assertThat(sandbox.getState()).isEqualTo("started");
    }

    @Test
    void listMapsPagination() {
        PaginatedSandboxes response = new PaginatedSandboxes();
        response.setItems(Collections.singletonList(TestSupport.mainSandbox("sb-1", SandboxState.STARTED)));
        response.setTotal(BigDecimal.ONE);
        response.setPage(BigDecimal.valueOf(2));
        response.setTotalPages(BigDecimal.valueOf(3));
        doReturn(response).when(sandboxApi).listSandboxesPaginated(
                any(), any(), any(), any(), any(), any(), any(), any(), any(), any(),
                any(), any(), any(), any(), any(), any(), any(), any(), any(), any());

        io.daytona.sdk.model.PaginatedSandboxes sandboxes = daytona.list(Collections.singletonMap("team", "sdk"), 2, 5);

        assertThat(sandboxes.getItems()).singleElement().satisfies(item -> assertThat(item).containsEntry("id", "sb-1"));
        assertThat(sandboxes.getTotal()).isEqualTo(1);
        assertThat(sandboxes.getPage()).isEqualTo(2);
        assertThat(sandboxes.getTotalPages()).isEqualTo(3);
    }

    @Test
    void listUsesDefaultPagingWhenArgumentsAreNull() {
        doReturn(new PaginatedSandboxes()).when(sandboxApi).listSandboxesPaginated(
                any(), any(), any(), any(), any(), any(), any(), any(), any(), any(),
                any(), any(), any(), any(), any(), any(), any(), any(), any(), any());

        daytona.list();

        org.mockito.Mockito.verify(sandboxApi).listSandboxesPaginated(
                isNull(),
                eq(BigDecimal.ONE),
                eq(BigDecimal.TEN),
                isNull(), isNull(), isNull(), isNull(), isNull(), isNull(), isNull(),
                isNull(), isNull(), isNull(), isNull(), isNull(), isNull(), isNull(), isNull(), isNull(), isNull());
    }

    @Test
    void listReturnsEmptyPaginationWhenApiReturnsNull() {
        doReturn(null).when(sandboxApi).listSandboxesPaginated(
                any(), any(), any(), any(), any(), any(), any(), any(), any(), any(),
                any(), any(), any(), any(), any(), any(), any(), any(), any(), any());

        io.daytona.sdk.model.PaginatedSandboxes sandboxes = daytona.list(Collections.<String, String>emptyMap(), null, null);

        assertThat(sandboxes.getItems()).isEmpty();
        assertThat(sandboxes.getTotal()).isZero();
        assertThat(sandboxes.getPage()).isZero();
        assertThat(sandboxes.getTotalPages()).isZero();
    }

    @ParameterizedTest
    @MethodSource("mappedApiExceptions")
    void getMapsApiErrors(int status, Class<? extends RuntimeException> type) {
        when(sandboxApi.getSandbox(anyString(), isNull(), isNull()))
                .thenThrow(new io.daytona.api.client.ApiException(status, "boom", null, "{\"message\":\"mapped\"}"));

        assertThatThrownBy(() -> daytona.get("sandbox-1"))
                .isInstanceOf(type)
                .hasMessage("mapped");
    }

    @Test
    void closeHandlesNullHttpClientCacheAndUtilityHelpers() {
        assertThat(Daytona.urlEncodePathSegment("a b/c")).isEqualTo("a+b%2Fc".replace("+", "%20"));
        assertThat(Daytona.urlEncodeQuery("a b")).isEqualTo("a+b");
        assertThat(Daytona.castStringMap(Collections.singletonMap(1, null))).containsEntry("1", "");
        assertThat(Daytona.sandboxToMap(TestSupport.mainSandbox("sb-map", SandboxState.STARTED)))
                .containsEntry("id", "sb-map")
                .containsEntry("state", "started");
        Daytona.shutdownHttpClient(null);

        io.daytona.api.client.ApiClient apiClient = TestSupport.getField(daytona, "apiClient", io.daytona.api.client.ApiClient.class);
        daytona.close();
        assertThat(apiClient.getHttpClient().dispatcher().executorService().isShutdown()).isTrue();
    }

    @Test
    void closeShutsDownHttpClient() {
        io.daytona.api.client.ApiClient apiClient = TestSupport.getField(daytona, "apiClient", io.daytona.api.client.ApiClient.class);

        daytona.close();

        assertThat(apiClient.getHttpClient().dispatcher().executorService().isShutdown()).isTrue();
    }

    private static Stream<Arguments> mappedApiExceptions() {
        return Stream.of(
                Arguments.of(400, DaytonaBadRequestException.class),
                Arguments.of(401, DaytonaAuthenticationException.class),
                Arguments.of(403, DaytonaForbiddenException.class),
                Arguments.of(404, DaytonaNotFoundException.class),
                Arguments.of(409, DaytonaConflictException.class),
                Arguments.of(422, DaytonaValidationException.class),
                Arguments.of(429, DaytonaRateLimitException.class),
                Arguments.of(500, DaytonaServerException.class)
        );
    }
}
