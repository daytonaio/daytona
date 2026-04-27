// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.api.client.api.VolumesApi;
import io.daytona.api.client.model.VolumeDto;
import io.daytona.api.client.model.VolumeState;
import io.daytona.sdk.exception.DaytonaNotFoundException;
import io.daytona.sdk.exception.DaytonaServerException;
import io.daytona.sdk.model.Volume;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Arrays;
import java.util.Collections;
import java.util.stream.Stream;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.isNull;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class VolumeServiceTest {

    @Mock
    private VolumesApi volumesApi;

    private VolumeService service;

    @BeforeEach
    void setUp() {
        service = new VolumeService(volumesApi);
    }

    @Test
    void createMapsResponse() {
        when(volumesApi.createVolume(any(), isNull())).thenReturn(volumeDto("vol-1", "cache", VolumeState.READY));

        Volume volume = service.create("cache");

        assertThat(volume.getId()).isEqualTo("vol-1");
        assertThat(volume.getName()).isEqualTo("cache");
        assertThat(volume.getState()).isEqualTo("ready");
    }

    @Test
    void createReturnsEmptyModelWhenApiReturnsNull() {
        when(volumesApi.createVolume(any(), isNull())).thenReturn(null);

        Volume volume = service.create("cache");

        assertThat(volume.getId()).isNull();
        assertThat(volume.getName()).isNull();
        assertThat(volume.getState()).isNull();
    }

    @Test
    void listMapsAllItems() {
        when(volumesApi.listVolumes(isNull(), isNull())).thenReturn(Arrays.asList(
                volumeDto("vol-1", "cache", VolumeState.READY),
                volumeDto("vol-2", "artifacts", VolumeState.READY)
        ));

        assertThat(service.list())
                .extracting(Volume::getName)
                .containsExactly("cache", "artifacts");
    }

    @Test
    void listReturnsEmptyListWhenApiReturnsNull() {
        when(volumesApi.listVolumes(isNull(), isNull())).thenReturn(null);

        assertThat(service.list()).isEqualTo(Collections.<Volume>emptyList());
    }

    @Test
    void getByNameMapsResponse() {
        when(volumesApi.getVolumeByName("cache", null)).thenReturn(volumeDto("vol-1", "cache", VolumeState.READY));

        Volume volume = service.getByName("cache");

        assertThat(volume.getId()).isEqualTo("vol-1");
        assertThat(volume.getState()).isEqualTo("ready");
    }

    @Test
    void getByNameReturnsEmptyModelWhenApiReturnsNull() {
        when(volumesApi.getVolumeByName("cache", null)).thenReturn(null);

        Volume volume = service.getByName("cache");

        assertThat(volume.getId()).isNull();
        assertThat(volume.getName()).isNull();
        assertThat(volume.getState()).isNull();
    }

    @Test
    void deleteDelegatesToApi() {
        service.delete("vol-1");

        verify(volumesApi).deleteVolume("vol-1", null);
    }

    @ParameterizedTest
    @MethodSource("mappedMainApiExceptions")
    void getByNameMapsApiErrors(int status, Class<? extends RuntimeException> type) {
        when(volumesApi.getVolumeByName("cache", null))
                .thenThrow(new io.daytona.api.client.ApiException(status, "boom", null, "{\"message\":\"mapped\"}"));

        assertThatThrownBy(() -> service.getByName("cache"))
                .isInstanceOf(type)
                .hasMessage("mapped");
    }

    private static Stream<Arguments> mappedMainApiExceptions() {
        return Stream.of(
                Arguments.of(404, DaytonaNotFoundException.class),
                Arguments.of(500, DaytonaServerException.class)
        );
    }

    private static VolumeDto volumeDto(String id, String name, VolumeState state) {
        VolumeDto dto = new VolumeDto();
        dto.setId(id);
        dto.setName(name);
        dto.setState(state);
        return dto;
    }
}
