// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
/**
 * File metadata for Sandbox file-system operations.
 */
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

    /**
     * Returns file or directory name.
     *
     * @return item name
     */
    public String getName() { return name; }

    /**
     * Sets file or directory name.
     *
     * @param name item name
     */
    public void setName(String name) { this.name = name; }

    /**
     * Returns item size in bytes.
     *
     * @return file size in bytes
     */
    public Long getSize() { return size == null ? 0L : size; }

    /**
     * Sets item size in bytes.
     *
     * @param size file size in bytes
     */
    public void setSize(Long size) { this.size = size; }

    /**
     * Returns mode string for this item.
     *
     * @return mode value
     */
    public String getMode() { return mode; }

    /**
     * Sets mode string for this item.
     *
     * @param mode mode value
     */
    public void setMode(String mode) { this.mode = mode; }

    /**
     * Returns modification timestamp.
     *
     * @return modification time
     */
    public String getModTime() { return modTime; }

    /**
     * Sets modification timestamp.
     *
     * @param modTime modification time
     */
    public void setModTime(String modTime) { this.modTime = modTime; }

    /**
     * Returns whether this item is a directory.
     *
     * @return {@code true} if directory
     */
    public boolean isDir() { return isDir != null && isDir; }

    /**
     * Sets directory flag.
     *
     * @param dir directory flag
     */
    public void setDir(Boolean dir) { isDir = dir; }
}
