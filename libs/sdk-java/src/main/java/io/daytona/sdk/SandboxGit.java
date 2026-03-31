// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.model.GitCommitResponse;
import io.daytona.sdk.model.GitStatus;
import io.daytona.toolbox.client.api.GitApi;
import io.daytona.toolbox.client.model.GitAddRequest;
import io.daytona.toolbox.client.model.GitCloneRequest;
import io.daytona.toolbox.client.model.GitCommitRequest;
import io.daytona.toolbox.client.model.GitRepoRequest;
import io.daytona.toolbox.client.model.ListBranchResponse;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class SandboxGit {
    private final GitApi gitApi;

    SandboxGit(GitApi gitApi) {
        this.gitApi = gitApi;
    }

    public void clone(String url, String path) {
        clone(url, path, null, null, null, null);
    }

    public void clone(String url, String path, String branch, String commitId, String username, String password) {
        GitCloneRequest request = new GitCloneRequest().url(url).path(path);
        if (branch != null) request.branch(branch);
        if (commitId != null) request.commitId(commitId);
        if (username != null) request.username(username);
        if (password != null) request.password(password);
        ExceptionMapper.runToolbox(() -> gitApi.cloneRepository(request));
    }

    public Map<String, Object> branches(String path) {
        ListBranchResponse response = ExceptionMapper.callToolbox(() -> gitApi.listBranches(path));
        Map<String, Object> result = new HashMap<String, Object>();
        result.put("branches", response == null ? new ArrayList<String>() : response.getBranches());
        return result;
    }

    public void add(String path, List<String> files) {
        ExceptionMapper.runToolbox(() -> gitApi.addFiles(new GitAddRequest().path(path).files(files)));
    }

    public GitCommitResponse commit(String path, String message, String author, String email) {
        io.daytona.toolbox.client.model.GitCommitResponse response = ExceptionMapper.callToolbox(
                () -> gitApi.commitChanges(new GitCommitRequest().path(path).message(message).author(author).email(email))
        );
        GitCommitResponse output = new GitCommitResponse();
        if (response != null) {
            output.setHash(response.getHash());
        }
        return output;
    }

    public GitStatus status(String path) {
        io.daytona.toolbox.client.model.GitStatus response = ExceptionMapper.callToolbox(() -> gitApi.getStatus(path));
        GitStatus status = new GitStatus();
        if (response != null) {
            status.setCurrentBranch(response.getCurrentBranch());
            status.setAhead(response.getAhead());
            status.setBehind(response.getBehind());
            status.setBranchPublished(response.getBranchPublished());

            List<GitStatus.FileStatus> fileStatuses = new ArrayList<GitStatus.FileStatus>();
            if (response.getFileStatus() != null) {
                for (io.daytona.toolbox.client.model.FileStatus item : response.getFileStatus()) {
                    GitStatus.FileStatus fs = new GitStatus.FileStatus();
                    fs.setPath(item.getName());
                    String staging = item.getStaging() == null ? "" : item.getStaging().getValue();
                    String worktree = item.getWorktree() == null ? "" : item.getWorktree().getValue();
                    fs.setStatus(staging + (worktree.isEmpty() ? "" : "/" + worktree));
                    fileStatuses.add(fs);
                }
            }
            status.setFileStatus(fileStatuses);
        }
        return status;
    }

    public void push(String path) {
        ExceptionMapper.runToolbox(() -> gitApi.pushChanges(new GitRepoRequest().path(path)));
    }

    public void pull(String path) {
        ExceptionMapper.runToolbox(() -> gitApi.pullChanges(new GitRepoRequest().path(path)));
    }
}