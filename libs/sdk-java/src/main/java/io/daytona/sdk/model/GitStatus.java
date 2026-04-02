// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.ArrayList;
import java.util.List;

@JsonIgnoreProperties(ignoreUnknown = true)
/**
 * Git repository status information.
 */
public class GitStatus {
    @JsonProperty("currentBranch")
    private String currentBranch;
    @JsonProperty("ahead")
    private Integer ahead;
    @JsonProperty("behind")
    private Integer behind;
    @JsonProperty("branchPublished")
    private Boolean branchPublished;
    @JsonProperty("fileStatus")
    private List<FileStatus> fileStatus;

    @JsonIgnoreProperties(ignoreUnknown = true)
    /**
     * Per-file Git status entry.
     */
    public static class FileStatus {
        @JsonProperty("path")
        private String path;
        @JsonProperty("status")
        private String status;

        /**
         * Returns file path relative to repository root.
         *
         * @return file path
         */
        public String getPath() { return path; }

        /**
         * Sets file path relative to repository root.
         *
         * @param path file path
         */
        public void setPath(String path) { this.path = path; }

        /**
         * Returns combined Git status string.
         *
         * @return status descriptor
         */
        public String getStatus() { return status; }

        /**
         * Sets combined Git status string.
         *
         * @param status status descriptor
         */
        public void setStatus(String status) { this.status = status; }
    }

    /**
     * Returns current branch name.
     *
     * @return current branch
     */
    public String getCurrentBranch() { return currentBranch; }

    /**
     * Sets current branch name.
     *
     * @param currentBranch current branch
     */
    public void setCurrentBranch(String currentBranch) { this.currentBranch = currentBranch; }

    /**
     * Returns number of commits ahead of upstream.
     *
     * @return ahead count
     */
    public int getAhead() { return ahead == null ? 0 : ahead; }

    /**
     * Sets number of commits ahead of upstream.
     *
     * @param ahead ahead count
     */
    public void setAhead(Integer ahead) { this.ahead = ahead; }

    /**
     * Returns number of commits behind upstream.
     *
     * @return behind count
     */
    public int getBehind() { return behind == null ? 0 : behind; }

    /**
     * Sets number of commits behind upstream.
     *
     * @param behind behind count
     */
    public void setBehind(Integer behind) { this.behind = behind; }

    /**
     * Returns whether current branch is published to remote.
     *
     * @return {@code true} if branch is published
     */
    public boolean isBranchPublished() { return branchPublished != null && branchPublished; }

    /**
     * Sets whether current branch is published to remote.
     *
     * @param branchPublished publication flag
     */
    public void setBranchPublished(Boolean branchPublished) { this.branchPublished = branchPublished; }

    /**
     * Returns per-file status entries.
     *
     * @return file status list
     */
    public List<FileStatus> getFileStatus() { return fileStatus == null ? new ArrayList<FileStatus>() : fileStatus; }

    /**
     * Sets per-file status entries.
     *
     * @param fileStatus file status list
     */
    public void setFileStatus(List<FileStatus> fileStatus) { this.fileStatus = fileStatus; }
}
