// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.model;

import java.time.OffsetDateTime;
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
    private List<SandboxState> states;
    /** Filter by snapshot names */
    private List<String> snapshots;
    /** Filter by targets */
    private List<String> targets;
    /** Filter by minimum CPU */
    private Integer minCpu;
    /** Filter by maximum CPU */
    private Integer maxCpu;
    /** Filter by minimum memory in GiB */
    private Integer minMemoryGib;
    /** Filter by maximum memory in GiB */
    private Integer maxMemoryGib;
    /** Filter by minimum disk space in GiB */
    private Integer minDiskGib;
    /** Filter by maximum disk space in GiB */
    private Integer maxDiskGib;
    /** Filter by public status */
    private Boolean isPublic;
    /** Filter by recoverable status */
    private Boolean isRecoverable;
    /** Include sandboxes created after this timestamp */
    private OffsetDateTime createdAtAfter;
    /** Include sandboxes created before this timestamp */
    private OffsetDateTime createdAtBefore;
    /** Include sandboxes with last activity after this timestamp */
    private OffsetDateTime lastActivityAfter;
    /** Include sandboxes with last activity before this timestamp */
    private OffsetDateTime lastActivityBefore;
    /** Sort by field */
    private SandboxListSortField sort;
    /** Sort direction */
    private SandboxListSortDirection order;

    public Integer getLimit() { return limit; }
    public void setLimit(Integer limit) { this.limit = limit; }
    public String getId() { return id; }
    public void setId(String id) { this.id = id; }
    public String getName() { return name; }
    public void setName(String name) { this.name = name; }
    public Map<String, String> getLabels() { return labels; }
    public void setLabels(Map<String, String> labels) { this.labels = labels; }
    public List<SandboxState> getStates() { return states; }
    public void setStates(List<SandboxState> states) { this.states = states; }
    public List<String> getSnapshots() { return snapshots; }
    public void setSnapshots(List<String> snapshots) { this.snapshots = snapshots; }
    public List<String> getTargets() { return targets; }
    public void setTargets(List<String> targets) { this.targets = targets; }
    public Integer getMinCpu() { return minCpu; }
    public void setMinCpu(Integer minCpu) { this.minCpu = minCpu; }
    public Integer getMaxCpu() { return maxCpu; }
    public void setMaxCpu(Integer maxCpu) { this.maxCpu = maxCpu; }
    public Integer getMinMemoryGib() { return minMemoryGib; }
    public void setMinMemoryGib(Integer minMemoryGib) { this.minMemoryGib = minMemoryGib; }
    public Integer getMaxMemoryGib() { return maxMemoryGib; }
    public void setMaxMemoryGib(Integer maxMemoryGib) { this.maxMemoryGib = maxMemoryGib; }
    public Integer getMinDiskGib() { return minDiskGib; }
    public void setMinDiskGib(Integer minDiskGib) { this.minDiskGib = minDiskGib; }
    public Integer getMaxDiskGib() { return maxDiskGib; }
    public void setMaxDiskGib(Integer maxDiskGib) { this.maxDiskGib = maxDiskGib; }
    public Boolean getIsPublic() { return isPublic; }
    public void setIsPublic(Boolean isPublic) { this.isPublic = isPublic; }
    public Boolean getIsRecoverable() { return isRecoverable; }
    public void setIsRecoverable(Boolean isRecoverable) { this.isRecoverable = isRecoverable; }
    public OffsetDateTime getCreatedAtAfter() { return createdAtAfter; }
    public void setCreatedAtAfter(OffsetDateTime createdAtAfter) { this.createdAtAfter = createdAtAfter; }
    public OffsetDateTime getCreatedAtBefore() { return createdAtBefore; }
    public void setCreatedAtBefore(OffsetDateTime createdAtBefore) { this.createdAtBefore = createdAtBefore; }
    public OffsetDateTime getLastActivityAfter() { return lastActivityAfter; }
    public void setLastActivityAfter(OffsetDateTime lastActivityAfter) { this.lastActivityAfter = lastActivityAfter; }
    public OffsetDateTime getLastActivityBefore() { return lastActivityBefore; }
    public void setLastActivityBefore(OffsetDateTime lastActivityBefore) { this.lastActivityBefore = lastActivityBefore; }
    public SandboxListSortField getSort() { return sort; }
    public void setSort(SandboxListSortField sort) { this.sort = sort; }
    public SandboxListSortDirection getOrder() { return order; }
    public void setOrder(SandboxListSortDirection order) { this.order = order; }
}
