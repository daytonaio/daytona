// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk;

import io.daytona.sdk.model.GitCommitResponse;
import io.daytona.sdk.model.GitStatus;
import io.daytona.toolbox.client.api.GitApi;
import io.daytona.toolbox.client.model.GitAddRemoteRequest;
import io.daytona.toolbox.client.model.GitAddRequest;
import io.daytona.toolbox.client.model.GitAuthenticateRequest;
import io.daytona.toolbox.client.model.GitBranchRequest;
import io.daytona.toolbox.client.model.GitCheckoutRequest;
import io.daytona.toolbox.client.model.GitCloneRequest;
import io.daytona.toolbox.client.model.GitCommitRequest;
import io.daytona.toolbox.client.model.GitConfigResponse;
import io.daytona.toolbox.client.model.GitConfigureUserRequest;
import io.daytona.toolbox.client.model.GitDeleteBranchRequest;
import io.daytona.toolbox.client.model.GitInitRequest;
import io.daytona.toolbox.client.model.GitPullRequest;
import io.daytona.toolbox.client.model.GitPushRequest;
import io.daytona.toolbox.client.model.GitRemote;
import io.daytona.toolbox.client.model.GitResetRequest;
import io.daytona.toolbox.client.model.GitRestoreRequest;
import io.daytona.toolbox.client.model.GitSetConfigRequest;
import io.daytona.toolbox.client.model.ListBranchResponse;
import io.daytona.toolbox.client.model.ListRemotesResponse;

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
        clone(url, path, null, null, null, null, null);
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
        clone(url, path, branch, commitId, username, password, null);
    }

    /**
     * Clones a repository with optional branch, commit, credentials, and TLS verification override.
     *
     * @param url repository URL
     * @param path destination path in the Sandbox
     * @param branch branch to clone; {@code null} uses default branch
     * @param commitId commit SHA to checkout after clone; {@code null} skips detached checkout
     * @param username username for authenticated remotes
     * @param password password or token for authenticated remotes
     * @param insecureSkipTls when {@code true}, skip TLS certificate verification (insecure).
     *   Use only for trusted internal Git servers with self-signed or private-CA certs;
     *   credentials, if supplied, are transmitted over an unverified TLS connection.
     *   {@code null} or {@code false} keeps strict TLS verification.
     * @throws io.daytona.sdk.exception.DaytonaException if cloning fails
     */
    public void clone(String url, String path, String branch, String commitId, String username, String password, Boolean insecureSkipTls) {
        clone(url, path, branch, commitId, username, password, insecureSkipTls, null);
    }

    /**
     * Clones a repository with optional branch, commit, credentials, TLS verification override, and shallow depth.
     *
     * @param url repository URL
     * @param path destination path in the Sandbox
     * @param branch branch to clone; {@code null} uses default branch
     * @param commitId commit SHA to checkout after clone; {@code null} skips detached checkout
     * @param username username for authenticated remotes
     * @param password password or token for authenticated remotes
     * @param insecureSkipTls when {@code true}, skip TLS certificate verification (insecure)
     * @param depth create a shallow clone truncated to the given number of commits; {@code null} clones full history
     * @throws io.daytona.sdk.exception.DaytonaException if cloning fails
     */
    public void clone(String url, String path, String branch, String commitId, String username, String password, Boolean insecureSkipTls, Integer depth) {
        GitCloneRequest request = new GitCloneRequest().url(url).path(path);
        if (branch != null) request.branch(branch);
        if (commitId != null) request.commitId(commitId);
        if (username != null) request.username(username);
        if (password != null) request.password(password);
        if (insecureSkipTls != null) request.insecureSkipTls(insecureSkipTls);
        if (depth != null) request.depth(depth);
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
            status.setDetached(response.getDetached());
            status.setUpstream(response.getUpstream());

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
        push(path, null, null, null, null, null);
    }

    /**
     * Pushes local commits to remote with optional credentials, branch, remote, and upstream tracking.
     *
     * @param path repository path in the Sandbox
     * @param username username for authenticated remotes
     * @param password password or token for authenticated remotes
     * @param branch branch to push; {@code null} uses the current branch
     * @param remote remote to push to; {@code null} uses "origin"
     * @param setUpstream when {@code true}, record the pushed branch as the upstream tracking branch
     * @throws io.daytona.sdk.exception.DaytonaException if push fails
     */
    public void push(String path, String username, String password, String branch, String remote, Boolean setUpstream) {
        GitPushRequest request = new GitPushRequest().path(path);
        if (username != null) request.username(username);
        if (password != null) request.password(password);
        if (branch != null) request.branch(branch);
        if (remote != null) request.remote(remote);
        if (setUpstream != null) request.setUpstream(setUpstream);
        ExceptionMapper.runToolbox(() -> gitApi.pushChanges(request));
    }

    /**
     * Pulls updates from remote.
     *
     * @param path repository path in the Sandbox
     * @throws io.daytona.sdk.exception.DaytonaException if pull fails
     */
    public void pull(String path) {
        pull(path, null, null, null, null);
    }

    /**
     * Pulls updates from remote with optional credentials, branch, and remote.
     *
     * @param path repository path in the Sandbox
     * @param username username for authenticated remotes
     * @param password password or token for authenticated remotes
     * @param branch branch to pull; {@code null} uses the current branch's upstream
     * @param remote remote to pull from; {@code null} uses "origin"
     * @throws io.daytona.sdk.exception.DaytonaException if pull fails
     */
    public void pull(String path, String username, String password, String branch, String remote) {
        GitPullRequest request = new GitPullRequest().path(path);
        if (username != null) request.username(username);
        if (password != null) request.password(password);
        if (branch != null) request.branch(branch);
        if (remote != null) request.remote(remote);
        ExceptionMapper.runToolbox(() -> gitApi.pullChanges(request));
    }

    /**
     * Creates a new branch at the current HEAD.
     *
     * @param path repository path in the Sandbox
     * @param name name of the new branch
     * @throws io.daytona.sdk.exception.DaytonaException if branch creation fails
     */
    public void createBranch(String path, String name) {
        ExceptionMapper.runToolbox(() -> gitApi.createBranch(new GitBranchRequest().path(path).name(name)));
    }

    /**
     * Checks out a branch or commit.
     *
     * @param path repository path in the Sandbox
     * @param branch branch name or commit SHA to checkout
     * @throws io.daytona.sdk.exception.DaytonaException if checkout fails
     */
    public void checkoutBranch(String path, String branch) {
        ExceptionMapper.runToolbox(() -> gitApi.checkoutBranch(new GitCheckoutRequest().path(path).branch(branch)));
    }

    /**
     * Deletes a branch.
     *
     * @param path repository path in the Sandbox
     * @param name name of the branch to delete
     * @throws io.daytona.sdk.exception.DaytonaException if deletion fails
     */
    public void deleteBranch(String path, String name) {
        GitDeleteBranchRequest request = new GitDeleteBranchRequest().path(path).name(name);
        ExceptionMapper.runToolbox(() -> gitApi.deleteBranch(request));
    }

    /**
     * Initializes a new Git repository at the specified path.
     *
     * @param path destination path in the Sandbox
     * @throws io.daytona.sdk.exception.DaytonaException if initialization fails
     */
    public void init(String path) {
        init(path, null, null);
    }

    /**
     * Initializes a new Git repository with optional bare mode and initial branch.
     *
     * @param path destination path in the Sandbox
     * @param bare when {@code true}, create a bare repository without a working tree
     * @param initialBranch name of the initial branch; {@code null} uses the Git default
     * @throws io.daytona.sdk.exception.DaytonaException if initialization fails
     */
    public void init(String path, Boolean bare, String initialBranch) {
        GitInitRequest request = new GitInitRequest().path(path);
        if (bare != null) request.bare(bare);
        if (initialBranch != null) request.initialBranch(initialBranch);
        ExceptionMapper.runToolbox(() -> gitApi.initRepository(request));
    }

    /**
     * Resets the current HEAD to HEAD (mixed reset, unstaging changes).
     *
     * @param path repository path in the Sandbox
     * @throws io.daytona.sdk.exception.DaytonaException if reset fails
     */
    public void reset(String path) {
        reset(path, null, null, null);
    }

    /**
     * Resets the current HEAD to the specified state.
     *
     * @param path repository path in the Sandbox
     * @param mode reset mode: "soft", "mixed" (default), "hard", "merge" or "keep"; {@code null} uses "mixed"
     * @param target revision to reset to; {@code null} uses HEAD
     * @param files constrain the reset to the given paths; {@code null} resets all
     * @throws io.daytona.sdk.exception.DaytonaException if reset fails
     */
    public void reset(String path, String mode, String target, List<String> files) {
        GitResetRequest request = new GitResetRequest().path(path);
        if (mode != null) request.mode(mode);
        if (target != null) request.target(target);
        if (files != null) request.files(files);
        ExceptionMapper.runToolbox(() -> gitApi.resetChanges(request));
    }

    /**
     * Restores working tree files for the given paths (discarding local changes).
     *
     * @param path repository path in the Sandbox
     * @param files file paths to restore
     * @throws io.daytona.sdk.exception.DaytonaException if restore fails
     */
    public void restore(String path, List<String> files) {
        restore(path, files, null, null, null);
    }

    /**
     * Restores working tree files or unstages changes.
     *
     * @param path repository path in the Sandbox
     * @param files file paths to restore
     * @param staged restore the staging index for the given files
     * @param worktree restore the working tree for the given files; defaults to {@code true} when both are {@code null}
     * @param source restore file contents from the given revision instead of the index
     * @throws io.daytona.sdk.exception.DaytonaException if restore fails
     */
    public void restore(String path, List<String> files, Boolean staged, Boolean worktree, String source) {
        GitRestoreRequest request = new GitRestoreRequest().path(path).files(files);
        if (staged != null) request.staged(staged);
        if (worktree != null) request.worktree(worktree);
        if (source != null) request.source(source);
        ExceptionMapper.runToolbox(() -> gitApi.restoreFiles(request));
    }

    /**
     * Adds a remote to the repository.
     *
     * @param path repository path in the Sandbox
     * @param name name of the remote
     * @param url URL of the remote
     * @throws io.daytona.sdk.exception.DaytonaException if adding the remote fails
     */
    public void remoteAdd(String path, String name, String url) {
        remoteAdd(path, name, url, null, null);
    }

    /**
     * Adds (or overwrites) a remote in the repository.
     *
     * @param path repository path in the Sandbox
     * @param name name of the remote
     * @param url URL of the remote
     * @param fetch when {@code true}, fetch from the remote immediately after adding it
     * @param overwrite when {@code true}, replace an existing remote with the same name
     * @throws io.daytona.sdk.exception.DaytonaException if adding the remote fails
     */
    public void remoteAdd(String path, String name, String url, Boolean fetch, Boolean overwrite) {
        GitAddRemoteRequest request = new GitAddRemoteRequest().path(path).name(name).url(url);
        if (fetch != null) request.fetch(fetch);
        if (overwrite != null) request.overwrite(overwrite);
        ExceptionMapper.runToolbox(() -> gitApi.addRemote(request));
    }

    /**
     * Lists the remotes configured in the repository.
     *
     * @param path repository path in the Sandbox
     * @return the configured remotes (name + URL)
     * @throws io.daytona.sdk.exception.DaytonaException if listing remotes fails
     */
    public List<GitRemote> remotes(String path) {
        ListRemotesResponse response = ExceptionMapper.callToolbox(() -> gitApi.listRemotes(path));
        return response == null ? new ArrayList<GitRemote>() : response.getRemotes();
    }

    /**
     * Gets the URL of a remote, or {@code null} when it does not exist.
     *
     * @param path repository path in the Sandbox
     * @param name name of the remote
     * @return the remote URL, or {@code null} when the remote does not exist
     * @throws io.daytona.sdk.exception.DaytonaException if the operation fails
     */
    public String remoteGet(String path, String name) {
        ListRemotesResponse response = ExceptionMapper.callToolbox(() -> gitApi.listRemotes(path));
        if (response == null || response.getRemotes() == null) {
            return null;
        }
        for (GitRemote remote : response.getRemotes()) {
            if (name.equals(remote.getName())) {
                return remote.getUrl();
            }
        }
        return null;
    }

    /**
     * Sets a Git config value at the global scope.
     *
     * @param key config key in dotted form (e.g. "user.name")
     * @param value config value
     * @throws io.daytona.sdk.exception.DaytonaException if setting config fails
     */
    public void setConfig(String key, String value) {
        setConfig(key, value, "global", null);
    }

    /**
     * Sets a Git config value at the given scope.
     *
     * @param key config key in dotted form (e.g. "user.name")
     * @param value config value
     * @param scope config scope: "global" (default), "local" or "system"
     * @param path repository path, required when scope is "local"
     * @throws io.daytona.sdk.exception.DaytonaException if setting config fails
     */
    public void setConfig(String key, String value, String scope, String path) {
        GitSetConfigRequest request = new GitSetConfigRequest().key(key).value(value);
        if (scope != null) request.scope(scope);
        if (path != null) request.path(path);
        ExceptionMapper.runToolbox(() -> gitApi.setGitConfig(request));
    }

    /**
     * Gets a Git config value at the global scope, or {@code null} when unset.
     *
     * @param key config key in dotted form (e.g. "user.name")
     * @return the config value, or {@code null} when the key is not set
     * @throws io.daytona.sdk.exception.DaytonaException if getting config fails
     */
    public String getConfig(String key) {
        return getConfig(key, "global", null);
    }

    /**
     * Gets a Git config value at the given scope, or {@code null} when unset.
     *
     * @param key config key in dotted form (e.g. "user.name")
     * @param scope config scope: "global" (default), "local" or "system"
     * @param path repository path, required when scope is "local"
     * @return the config value, or {@code null} when the key is not set
     * @throws io.daytona.sdk.exception.DaytonaException if getting config fails
     */
    public String getConfig(String key, String scope, String path) {
        GitConfigResponse response = ExceptionMapper.callToolbox(() -> gitApi.getGitConfig(key, path, scope));
        return response == null ? null : response.getValue();
    }

    /**
     * Configures the Git user name and email at the global scope.
     *
     * @param name user name (user.name)
     * @param email user email (user.email)
     * @throws io.daytona.sdk.exception.DaytonaException if configuring user fails
     */
    public void configureUser(String name, String email) {
        configureUser(name, email, "global", null);
    }

    /**
     * Configures the Git user name and email at the given scope.
     *
     * @param name user name (user.name)
     * @param email user email (user.email)
     * @param scope config scope: "global" (default), "local" or "system"
     * @param path repository path, required when scope is "local"
     * @throws io.daytona.sdk.exception.DaytonaException if configuring user fails
     */
    public void configureUser(String name, String email, String scope, String path) {
        GitConfigureUserRequest request = new GitConfigureUserRequest().name(name).email(email);
        if (scope != null) request.scope(scope);
        if (path != null) request.path(path);
        ExceptionMapper.runToolbox(() -> gitApi.configureUser(request));
    }

    /**
     * Persists Git credentials globally so that subsequent operations against github.com authenticate automatically.
     *
     * <p>This stores the password in plaintext on disk via the Git credential store.
     *
     * @param username Git username
     * @param password Git password or token
     * @throws io.daytona.sdk.exception.DaytonaException if authentication fails
     */
    public void dangerouslyAuthenticate(String username, String password) {
        dangerouslyAuthenticate(username, password, null, null);
    }

    /**
     * Persists Git credentials globally so that subsequent operations against the given host authenticate automatically.
     *
     * <p>This stores the password in plaintext on disk via the Git credential store.
     *
     * @param username Git username
     * @param password Git password or token
     * @param host host to authenticate against; {@code null} uses "github.com"
     * @param protocol protocol to authenticate against; {@code null} uses "https"
     * @throws io.daytona.sdk.exception.DaytonaException if authentication fails
     */
    public void dangerouslyAuthenticate(String username, String password, String host, String protocol) {
        GitAuthenticateRequest request = new GitAuthenticateRequest().username(username).password(password);
        if (host != null) request.host(host);
        if (protocol != null) request.protocol(protocol);
        ExceptionMapper.runToolbox(() -> gitApi.authenticate(request));
    }
}
