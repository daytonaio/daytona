// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
public class Resources {
    @JsonProperty("cpu")
    private Integer cpu;
    @JsonProperty("gpu")
    private Integer gpu;
    @JsonProperty("memory")
    private Integer memory;
    @JsonProperty("disk")
    private Integer disk;

    public Integer getCpu() { return cpu; }
    public void setCpu(Integer cpu) { this.cpu = cpu; }
    public Integer getGpu() { return gpu; }
    public void setGpu(Integer gpu) { this.gpu = gpu; }
    public Integer getMemory() { return memory; }
    public void setMemory(Integer memory) { this.memory = memory; }
    public Integer getDisk() { return disk; }
    public void setDisk(Integer disk) { this.disk = disk; }
}