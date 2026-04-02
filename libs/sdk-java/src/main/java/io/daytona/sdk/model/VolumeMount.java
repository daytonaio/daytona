// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
/**
 * Volume mount definition for Sandbox creation.
 */
public class VolumeMount {
    @JsonProperty("volumeId")
    private String volumeId;
    @JsonProperty("mountPath")
    private String mountPath;

    /**
     * Returns mounted volume identifier.
     *
     * @return volume ID
     */
    public String getVolumeId() { return volumeId; }

    /**
     * Sets mounted volume identifier.
     *
     * @param volumeId volume ID
     */
    public void setVolumeId(String volumeId) { this.volumeId = volumeId; }

    /**
     * Returns mount path inside the Sandbox.
     *
     * @return Sandbox mount path
     */
    public String getMountPath() { return mountPath; }

    /**
     * Sets mount path inside the Sandbox.
     *
     * @param mountPath Sandbox mount path
     */
    public void setMountPath(String mountPath) { this.mountPath = mountPath; }
}
