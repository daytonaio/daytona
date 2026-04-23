// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.api.client.api.SnapshotsApi;
import io.daytona.api.client.model.CreateSnapshot;
import io.daytona.api.client.model.SnapshotDto;
import io.daytona.api.client.model.SnapshotState;
import io.daytona.api.client.model.Url;
import io.daytona.sdk.exception.DaytonaException;
import io.daytona.sdk.exception.DaytonaNotFoundException;
import io.daytona.sdk.exception.DaytonaServerException;
import io.daytona.sdk.model.PaginatedSnapshots;
import io.daytona.sdk.model.Resources;
import io.daytona.sdk.model.Snapshot;
import okhttp3.OkHttpClient;
import okhttp3.mockwebserver.MockResponse;
import okhttp3.mockwebserver.MockWebServer;
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
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.stream.Stream;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.isNull;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class SnapshotServiceTest {

    @Mock
    private SnapshotsApi snapshotsApi;

    private SnapshotService snapshotService;

    @BeforeEach
    void setUp() {
        snapshotService = new SnapshotService(snapshotsApi, new OkHttpClient(), "test-key");
    }

    @Test
    void createFromImageNameMapsResponse() {
        when(snapshotsApi.createSnapshot(any(), isNull())).thenReturn(snapshotDto("snap-1", "snapshot", SnapshotState.ACTIVE));

        Snapshot snapshot = snapshotService.create("snapshot", "python:3.12");

        assertThat(snapshot.getId()).isEqualTo("snap-1");
        assertThat(snapshot.getName()).isEqualTo("snapshot");
        assertThat(snapshot.getState()).isEqualTo("active");
    }

    @Test
    void createFromImageNameBuildsRequest() {
        when(snapshotsApi.createSnapshot(any(), isNull())).thenReturn(snapshotDto("snap-1", "snapshot", SnapshotState.ACTIVE));

        snapshotService.create("snapshot", "python:3.12");

        ArgumentCaptor<CreateSnapshot> captor = ArgumentCaptor.forClass(CreateSnapshot.class);
        verify(snapshotsApi).createSnapshot(captor.capture(), isNull());
        assertThat(captor.getValue().getName()).isEqualTo("snapshot");
        assertThat(captor.getValue().getImageName()).isEqualTo("python:3.12");
        assertThat(captor.getValue().getBuildInfo()).isNull();
    }

    @Test
    void createFromDeclarativeImageBuildsDockerfileAndResources() {
        SnapshotDto snapshotDto = snapshotDto("snap-1", "snapshot", SnapshotState.ACTIVE);
        when(snapshotsApi.createSnapshot(any(), isNull())).thenReturn(snapshotDto);

        Resources resources = new Resources();
        resources.setCpu(2);
        resources.setMemory(4);
        resources.setDisk(8);

        List<String> logs = new java.util.ArrayList<String>();
        Snapshot snapshot = snapshotService.create("snapshot", Image.base("python:3.12").runCommands("echo hi"), resources, logs::add);

        assertThat(snapshot.getId()).isEqualTo("snap-1");
        assertThat(logs).hasSize(2);

        ArgumentCaptor<io.daytona.api.client.model.CreateSnapshot> captor = ArgumentCaptor.forClass(io.daytona.api.client.model.CreateSnapshot.class);
        verify(snapshotsApi).createSnapshot(captor.capture(), isNull());
        assertThat(captor.getValue().getBuildInfo().getDockerfileContent()).contains("FROM python:3.12\nRUN echo hi\n");
        assertThat(captor.getValue().getCpu()).isEqualTo(2);
        assertThat(captor.getValue().getMemory()).isEqualTo(4);
        assertThat(captor.getValue().getDisk()).isEqualTo(8);
    }

    @Test
    void createFromDeclarativeImageStreamsLogsUntilSnapshotBecomesActive() throws Exception {
        try (MockWebServer server = new MockWebServer()) {
            server.enqueue(new MockResponse().setBody("log-1\nlog-2\n"));
            SnapshotDto pending = snapshotDto("snap-2", "pending-snapshot", SnapshotState.BUILDING);
            SnapshotDto active = snapshotDto("snap-2", "pending-snapshot", SnapshotState.ACTIVE);
            when(snapshotsApi.createSnapshot(any(), isNull())).thenReturn(pending);
            when(snapshotsApi.getSnapshot("pending-snapshot", null)).thenReturn(pending, active);
            when(snapshotsApi.getSnapshotBuildLogsUrl("snap-2", null)).thenReturn(new Url().url(server.url("/snapshot-logs").toString()));

            List<String> logs = new ArrayList<String>();
            Snapshot snapshot = snapshotService.create("pending-snapshot", Image.base("python:3.12"), null, logs::add);

            assertThat(snapshot.getState()).isEqualTo("active");
            assertThat(logs).contains("Creating snapshot pending-snapshot (building)", "log-1", "Created snapshot pending-snapshot (active)");
            assertThat(server.takeRequest().getPath()).isEqualTo("/snapshot-logs?follow=true");
        }
    }

    @Test
    void createFromDeclarativeImageThrowsWhenApiReturnsNullSnapshot() {
        when(snapshotsApi.createSnapshot(any(), isNull())).thenReturn(null);

        assertThatThrownBy(() -> snapshotService.create("snapshot", Image.base("python:3.12"), null, null))
                .isInstanceOf(DaytonaException.class)
                .hasMessageContaining("no response from API");
    }

    @Test
    void createFromDeclarativeImageThrowsForErrorState() {
        when(snapshotsApi.createSnapshot(any(), isNull())).thenReturn(snapshotDto("snap-1", "snapshot", SnapshotState.ERROR));

        assertThatThrownBy(() -> snapshotService.create("snapshot", Image.base("python:3.12"), logs -> { }))
                .isInstanceOf(DaytonaException.class)
                .hasMessageContaining("Snapshot build failed");
    }

    @Test
    void createFromDeclarativeImageAcceptsPartialResources() {
        when(snapshotsApi.createSnapshot(any(), isNull())).thenReturn(snapshotDto("snap-1", "snapshot", SnapshotState.ACTIVE));
        Resources resources = new Resources();
        resources.setCpu(2);

        snapshotService.create("snapshot", Image.base("python:3.12"), resources, null);

        ArgumentCaptor<CreateSnapshot> captor = ArgumentCaptor.forClass(CreateSnapshot.class);
        verify(snapshotsApi).createSnapshot(captor.capture(), isNull());
        assertThat(captor.getValue().getCpu()).isEqualTo(2);
        assertThat(captor.getValue().getMemory()).isNull();
        assertThat(captor.getValue().getDisk()).isNull();
    }

    @Test
    void listMapsPagination() {
        io.daytona.api.client.model.PaginatedSnapshots response = new io.daytona.api.client.model.PaginatedSnapshots();
        response.setItems(Arrays.asList(
                snapshotDto("snap-1", "one", SnapshotState.ACTIVE),
                snapshotDto("snap-2", "two", SnapshotState.ACTIVE)
        ));
        response.setTotal(BigDecimal.valueOf(2));
        response.setPage(BigDecimal.ONE);
        response.setTotalPages(BigDecimal.ONE);
        when(snapshotsApi.getAllSnapshots(isNull(), any(), any(), isNull(), isNull(), isNull())).thenReturn(response);

        PaginatedSnapshots snapshots = snapshotService.list(null, null);

        assertThat(snapshots.getItems()).extracting(Snapshot::getName).containsExactly("one", "two");
        assertThat(snapshots.getTotal()).isEqualTo(2);
        assertThat(snapshots.getPage()).isEqualTo(1);
    }

    @Test
    void listReturnsDefaultsWhenApiReturnsNull() {
        when(snapshotsApi.getAllSnapshots(isNull(), any(), any(), isNull(), isNull(), isNull())).thenReturn(null);

        PaginatedSnapshots snapshots = snapshotService.list(2, 3);

        assertThat(snapshots.getItems()).isEmpty();
        assertThat(snapshots.getTotal()).isZero();
        assertThat(snapshots.getPage()).isZero();
        assertThat(snapshots.getTotalPages()).isZero();
    }

    @Test
    void getAndDeleteDelegate() {
        when(snapshotsApi.getSnapshot("snapshot", null)).thenReturn(snapshotDto("snap-1", "snapshot", SnapshotState.ACTIVE));

        Snapshot snapshot = snapshotService.get("snapshot");
        snapshotService.delete("snap-1");

        assertThat(snapshot.getName()).isEqualTo("snapshot");
        verify(snapshotsApi).removeSnapshot("snap-1", null);
    }

    @ParameterizedTest
    @MethodSource("mappedMainApiExceptions")
    void getMapsApiErrors(int status, Class<? extends RuntimeException> type) {
        when(snapshotsApi.getSnapshot("snapshot", null))
                .thenThrow(new io.daytona.api.client.ApiException(status, "boom", null, "{\"message\":\"mapped\"}"));

        assertThatThrownBy(() -> snapshotService.get("snapshot"))
                .isInstanceOf(type)
                .hasMessage("mapped");
    }

    private static Stream<Arguments> mappedMainApiExceptions() {
        return Stream.of(
                Arguments.of(404, DaytonaNotFoundException.class),
                Arguments.of(500, DaytonaServerException.class)
        );
    }

    private static SnapshotDto snapshotDto(String id, String name, SnapshotState state) {
        SnapshotDto dto = new SnapshotDto();
        dto.setId(id);
        dto.setName(name);
        dto.setImageName("python:3.12");
        dto.setState(state);
        return dto;
    }
}
