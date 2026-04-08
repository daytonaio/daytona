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

/**
 * Git operations facade for a specific Sandbox.
 *
 * <p>Provides repository clone, branch, commit, status, and sync operations mapped to Daytona
 * toolbox Git endpoints.
 */
public class Git {
    private final GitApi gitApi;

    Git(GitApi gitApi) {
        this.gitApi = gitApi;
    }

    /**
     * Clones a Git repository into the specified path.
     *
     * @param url repository URL
     * @param path destination path in the Sandbox
     * @throws io.daytona.sdk.exception.DaytonaException if cloning fails
     */
    public void clone(String url, String path) {
        clone(url, path, null, null, null, null);
    }

    /**
     * Clones a repository with optional branch, commit, and credentials.
     *
     * @param url repository URL
     * @param path destination path in the Sandbox
     * @param branch branch to clone; {@code null} uses default branch
     * @param commitId commit SHA to checkout after clone; {@code null} skips detached checkout
     * @param username username for authenticated remotes
     * @param password password or token for authenticated remotes
     * @throws io.daytona.sdk.exception.DaytonaException if cloning fails
     */
    public void clone(String url, String path, String branch, String commitId, String username, String password) {
        GitCloneRequest request = new GitCloneRequest().url(url).path(path);
        if (branch != null) request.branch(branch);
        if (commitId != null) request.commitId(commitId);
        if (username != null) request.username(username);
        if (password != null) request.password(password);
        ExceptionMapper.runToolbox(() -> gitApi.cloneRepository(request));
    }

    /**
     * Lists branches in a repository.
     *
     * @param path repository path in the Sandbox
     * @return map containing {@code branches} list
     * @throws io.daytona.sdk.exception.DaytonaException if the operation fails
     */
    public Map<String, Object> branches(String path) {
        ListBranchResponse response = ExceptionMapper.callToolbox(() -> gitApi.listBranches(path));
        Map<String, Object> result = new HashMap<String, Object>();
        result.put("branches", response == null ? new ArrayList<String>() : response.getBranches());
        return result;
    }

    /**
     * Stages files for commit.
     *
     * @param path repository path in the Sandbox
     * @param files file paths to stage relative to repository root
     * @throws io.daytona.sdk.exception.DaytonaException if staging fails
     */
    public void add(String path, List<String> files) {
        ExceptionMapper.runToolbox(() -> gitApi.addFiles(new GitAddRequest().path(path).files(files)));
    }

    /**
     * Creates a commit from staged changes.
     *
     * @param path repository path in the Sandbox
     * @param message commit message
     * @param author author display name
     * @param email author email address
     * @return commit metadata containing resulting hash
     * @throws io.daytona.sdk.exception.DaytonaException if commit fails
     */
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

    /**
     * Retrieves Git status for a repository.
     *
     * @param path repository path in the Sandbox
     * @return repository status including branch divergence and file status entries
     * @throws io.daytona.sdk.exception.DaytonaException if the operation fails
     */
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

    /**
     * Pushes local commits to remote.
     *
     * @param path repository path in the Sandbox
     * @throws io.daytona.sdk.exception.DaytonaException if push fails
     */
    public void push(String path) {
        ExceptionMapper.runToolbox(() -> gitApi.pushChanges(new GitRepoRequest().path(path)));
    }

    /**
     * Pulls updates from remote.
     *
     * @param path repository path in the Sandbox
     * @throws io.daytona.sdk.exception.DaytonaException if pull fails
     */
    public void pull(String path) {
        ExceptionMapper.runToolbox(() -> gitApi.pullChanges(new GitRepoRequest().path(path)));
    }
}
