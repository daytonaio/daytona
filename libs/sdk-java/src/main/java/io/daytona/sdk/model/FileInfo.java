// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
public class FileInfo {
    @JsonProperty("name")
    private String name;
    @JsonProperty("size")
    private Long size;
    @JsonProperty("mode")
    private String mode;
    @JsonProperty("modTime")
    private String modTime;
    @JsonProperty("isDir")
    private Boolean isDir;

    public String getName() { return name; }
    public void setName(String name) { this.name = name; }
    public Long getSize() { return size == null ? 0L : size; }
    public void setSize(Long size) { this.size = size; }
    public String getMode() { return mode; }
    public void setMode(String mode) { this.mode = mode; }
    public String getModTime() { return modTime; }
    public void setModTime(String modTime) { this.modTime = modTime; }
    public boolean isDir() { return isDir != null && isDir; }
    public void setDir(Boolean dir) { isDir = dir; }
}