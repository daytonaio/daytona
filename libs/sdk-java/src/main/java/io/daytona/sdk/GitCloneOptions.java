// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import java.util.List;

/**
 * Optional parameters for cloning a Git repository.
 */
public class GitCloneOptions {
    private String branch;
    private String commitId;
    private String username;
    private String password;
    private Integer depth;
    private Boolean singleBranch;
    private String shallowSince;
    private Boolean noTags;
    private String filter;
    private Boolean sparse;
    private List<String> sparsePaths;
    private String referencePath;
    private Boolean dissociate;
    private Boolean recurseSubmodules;
    private Boolean shallowSubmodules;
    private Boolean filterSubmodules;

    public String getBranch() {
        return branch;
    }

    public GitCloneOptions branch(String branch) {
        this.branch = branch;
        return this;
    }

    public String getCommitId() {
        return commitId;
    }

    public GitCloneOptions commitId(String commitId) {
        this.commitId = commitId;
        return this;
    }

    public String getUsername() {
        return username;
    }

    public GitCloneOptions username(String username) {
        this.username = username;
        return this;
    }

    public String getPassword() {
        return password;
    }

    public GitCloneOptions password(String password) {
        this.password = password;
        return this;
    }

    public Integer getDepth() {
        return depth;
    }

    public GitCloneOptions depth(Integer depth) {
        this.depth = depth;
        return this;
    }

    public Boolean getSingleBranch() {
        return singleBranch;
    }

    public GitCloneOptions singleBranch(Boolean singleBranch) {
        this.singleBranch = singleBranch;
        return this;
    }

    public String getShallowSince() {
        return shallowSince;
    }

    public GitCloneOptions shallowSince(String shallowSince) {
        this.shallowSince = shallowSince;
        return this;
    }

    public Boolean getNoTags() {
        return noTags;
    }

    public GitCloneOptions noTags(Boolean noTags) {
        this.noTags = noTags;
        return this;
    }

    public String getFilter() {
        return filter;
    }

    public GitCloneOptions filter(String filter) {
        this.filter = filter;
        return this;
    }

    public Boolean getSparse() {
        return sparse;
    }

    public GitCloneOptions sparse(Boolean sparse) {
        this.sparse = sparse;
        return this;
    }

    public List<String> getSparsePaths() {
        return sparsePaths;
    }

    public GitCloneOptions sparsePaths(List<String> sparsePaths) {
        this.sparsePaths = sparsePaths;
        return this;
    }

    public String getReferencePath() {
        return referencePath;
    }

    public GitCloneOptions referencePath(String referencePath) {
        this.referencePath = referencePath;
        return this;
    }

    public Boolean getDissociate() {
        return dissociate;
    }

    public GitCloneOptions dissociate(Boolean dissociate) {
        this.dissociate = dissociate;
        return this;
    }

    public Boolean getRecurseSubmodules() {
        return recurseSubmodules;
    }

    public GitCloneOptions recurseSubmodules(Boolean recurseSubmodules) {
        this.recurseSubmodules = recurseSubmodules;
        return this;
    }

    public Boolean getShallowSubmodules() {
        return shallowSubmodules;
    }

    public GitCloneOptions shallowSubmodules(Boolean shallowSubmodules) {
        this.shallowSubmodules = shallowSubmodules;
        return this;
    }

    public Boolean getFilterSubmodules() {
        return filterSubmodules;
    }

    public GitCloneOptions filterSubmodules(Boolean filterSubmodules) {
        this.filterSubmodules = filterSubmodules;
        return this;
    }
}
