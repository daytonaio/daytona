// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
public class VolumeMount {
    @JsonProperty("volumeId")
    private String volumeId;
    @JsonProperty("mountPath")
    private String mountPath;

    public String getVolumeId() { return volumeId; }
    public void setVolumeId(String volumeId) { this.volumeId = volumeId; }
    public String getMountPath() { return mountPath; }
    public void setMountPath(String mountPath) { this.mountPath = mountPath; }
}