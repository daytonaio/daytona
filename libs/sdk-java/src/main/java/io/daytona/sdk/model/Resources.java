// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
/**
 * Resource allocation values for a Sandbox.
 */
public class Resources {
    @JsonProperty("cpu")
    private Integer cpu;
    @JsonProperty("gpu")
    private Integer gpu;
    @JsonProperty("memory")
    private Integer memory;
    @JsonProperty("disk")
    private Integer disk;

    /**
     * Returns CPU allocation in cores.
     *
     * @return CPU cores
     */
    public Integer getCpu() { return cpu; }

    /**
     * Sets CPU allocation in cores.
     *
     * @param cpu CPU cores
     */
    public void setCpu(Integer cpu) { this.cpu = cpu; }

    /**
     * Returns GPU allocation.
     *
     * @return GPU units
     */
    public Integer getGpu() { return gpu; }

    /**
     * Sets GPU allocation.
     *
     * @param gpu GPU units
     */
    public void setGpu(Integer gpu) { this.gpu = gpu; }

    /**
     * Returns memory allocation in GiB.
     *
     * @return memory allocation
     */
    public Integer getMemory() { return memory; }

    /**
     * Sets memory allocation in GiB.
     *
     * @param memory memory allocation
     */
    public void setMemory(Integer memory) { this.memory = memory; }

    /**
     * Returns disk allocation in GiB.
     *
     * @return disk allocation
     */
    public Integer getDisk() { return disk; }

    /**
     * Sets disk allocation in GiB.
     *
     * @param disk disk allocation
     */
    public void setDisk(Integer disk) { this.disk = disk; }
}
