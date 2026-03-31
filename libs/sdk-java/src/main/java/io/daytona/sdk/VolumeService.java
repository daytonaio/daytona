// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.api.client.api.VolumesApi;
import io.daytona.api.client.model.CreateVolume;
import io.daytona.sdk.model.Volume;

import java.util.List;
import java.util.ArrayList;

public class VolumeService {
    private final VolumesApi volumesApi;

    VolumeService(VolumesApi volumesApi) {
        this.volumesApi = volumesApi;
    }

    public Volume create(String name) {
        io.daytona.api.client.model.VolumeDto volumeDto = ExceptionMapper.callMain(
                () -> volumesApi.createVolume(new CreateVolume().name(name), null)
        );
        return toVolume(volumeDto);
    }

    public List<Volume> list() {
        List<io.daytona.api.client.model.VolumeDto> volumes = ExceptionMapper.callMain(() -> volumesApi.listVolumes(null, null));
        List<Volume> result = new ArrayList<Volume>();
        if (volumes != null) {
            for (io.daytona.api.client.model.VolumeDto volume : volumes) {
                result.add(toVolume(volume));
            }
        }
        return result;
    }

    public Volume getByName(String name) {
        io.daytona.api.client.model.VolumeDto volumeDto = ExceptionMapper.callMain(() -> volumesApi.getVolumeByName(name, null));
        return toVolume(volumeDto);
    }

    public void delete(String id) {
        ExceptionMapper.runMain(() -> volumesApi.deleteVolume(id, null));
    }

    private Volume toVolume(io.daytona.api.client.model.VolumeDto source) {
        Volume volume = new Volume();
        if (source != null) {
            volume.setId(source.getId());
            volume.setName(source.getName());
            volume.setState(source.getState() == null ? null : source.getState().getValue());
        }
        return volume;
    }
}