// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.ArrayList;
import java.util.List;

@JsonIgnoreProperties(ignoreUnknown = true)
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
    public static class FileStatus {
        @JsonProperty("path")
        private String path;
        @JsonProperty("status")
        private String status;

        public String getPath() { return path; }
        public void setPath(String path) { this.path = path; }
        public String getStatus() { return status; }
        public void setStatus(String status) { this.status = status; }
    }

    public String getCurrentBranch() { return currentBranch; }
    public void setCurrentBranch(String currentBranch) { this.currentBranch = currentBranch; }
    public int getAhead() { return ahead == null ? 0 : ahead; }
    public void setAhead(Integer ahead) { this.ahead = ahead; }
    public int getBehind() { return behind == null ? 0 : behind; }
    public void setBehind(Integer behind) { this.behind = behind; }
    public boolean isBranchPublished() { return branchPublished != null && branchPublished; }
    public void setBranchPublished(Boolean branchPublished) { this.branchPublished = branchPublished; }
    public List<FileStatus> getFileStatus() { return fileStatus == null ? new ArrayList<FileStatus>() : fileStatus; }
    public void setFileStatus(List<FileStatus> fileStatus) { this.fileStatus = fileStatus; }
}