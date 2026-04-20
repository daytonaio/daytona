// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import java.util.List;
import java.util.Map;

/**
 * Query parameters for filtering and sorting when listing Sandboxes.
 */
public class ListSandboxesQuery {
    /** Per-page fetch size. Does NOT limit the total number of Sandboxes returned. */
    private Integer limit;
    /** Filter by ID prefix (case-insensitive) */
    private String id;
    /** Filter by name prefix (case-insensitive) */
    private String name;
    /** Filter by labels */
    private Map<String, String> labels;
    /** Filter by states */
    private List<String> states;
    /** Filter by snapshot names */
    private List<String> snapshots;
    /** Filter by targets */
    private List<String> targets;
    /** Filter by minimum CPU */
    private Integer minCpu;
    /** Filter by maximum CPU */
    private Integer maxCpu;
    /** Filter by minimum memory in GiB */
    private Integer minMemoryGiB;
    /** Filter by maximum memory in GiB */
    private Integer maxMemoryGiB;
    /** Filter by minimum disk space in GiB */
    private Integer minDiskGiB;
    /** Filter by maximum disk space in GiB */
    private Integer maxDiskGiB;
    /** Filter by public status */
    private Boolean isPublic;
    /** Filter by recoverable status */
    private Boolean isRecoverable;
    /** Include sandboxes created after this timestamp */
    private String createdAtAfter;
    /** Include sandboxes created before this timestamp */
    private String createdAtBefore;
    /** Include sandboxes with last activity after this timestamp */
    private String lastActivityAfter;
    /** Include sandboxes with last activity before this timestamp */
    private String lastActivityBefore;
    /** Sort by field (name, cpu, memoryGiB, diskGiB, lastActivityAt, createdAt) */
    private String sort;
    /** Sort direction (asc, desc) */
    private String order;

    public Integer getLimit() { return limit; }
    public void setLimit(Integer limit) { this.limit = limit; }
    public String getId() { return id; }
    public void setId(String id) { this.id = id; }
    public String getName() { return name; }
    public void setName(String name) { this.name = name; }
    public Map<String, String> getLabels() { return labels; }
    public void setLabels(Map<String, String> labels) { this.labels = labels; }
    public List<String> getStates() { return states; }
    public void setStates(List<String> states) { this.states = states; }
    public List<String> getSnapshots() { return snapshots; }
    public void setSnapshots(List<String> snapshots) { this.snapshots = snapshots; }
    public List<String> getTargets() { return targets; }
    public void setTargets(List<String> targets) { this.targets = targets; }
    public Integer getMinCpu() { return minCpu; }
    public void setMinCpu(Integer minCpu) { this.minCpu = minCpu; }
    public Integer getMaxCpu() { return maxCpu; }
    public void setMaxCpu(Integer maxCpu) { this.maxCpu = maxCpu; }
    public Integer getMinMemoryGiB() { return minMemoryGiB; }
    public void setMinMemoryGiB(Integer minMemoryGiB) { this.minMemoryGiB = minMemoryGiB; }
    public Integer getMaxMemoryGiB() { return maxMemoryGiB; }
    public void setMaxMemoryGiB(Integer maxMemoryGiB) { this.maxMemoryGiB = maxMemoryGiB; }
    public Integer getMinDiskGiB() { return minDiskGiB; }
    public void setMinDiskGiB(Integer minDiskGiB) { this.minDiskGiB = minDiskGiB; }
    public Integer getMaxDiskGiB() { return maxDiskGiB; }
    public void setMaxDiskGiB(Integer maxDiskGiB) { this.maxDiskGiB = maxDiskGiB; }
    public Boolean getIsPublic() { return isPublic; }
    public void setIsPublic(Boolean isPublic) { this.isPublic = isPublic; }
    public Boolean getIsRecoverable() { return isRecoverable; }
    public void setIsRecoverable(Boolean isRecoverable) { this.isRecoverable = isRecoverable; }
    public String getCreatedAtAfter() { return createdAtAfter; }
    public void setCreatedAtAfter(String createdAtAfter) { this.createdAtAfter = createdAtAfter; }
    public String getCreatedAtBefore() { return createdAtBefore; }
    public void setCreatedAtBefore(String createdAtBefore) { this.createdAtBefore = createdAtBefore; }
    public String getLastActivityAfter() { return lastActivityAfter; }
    public void setLastActivityAfter(String lastActivityAfter) { this.lastActivityAfter = lastActivityAfter; }
    public String getLastActivityBefore() { return lastActivityBefore; }
    public void setLastActivityBefore(String lastActivityBefore) { this.lastActivityBefore = lastActivityBefore; }
    public String getSort() { return sort; }
    public void setSort(String sort) { this.sort = sort; }
    public String getOrder() { return order; }
    public void setOrder(String order) { this.order = order; }
}
