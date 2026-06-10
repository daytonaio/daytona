# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from daytona_toolbox_api_client_async import (
    GitAddRemoteRequest,
    GitAddRequest,
    GitApi,
    GitAuthenticateRequest,
    GitBranchRequest,
    GitCheckoutRequest,
    GitCloneRequest,
    GitCommitRequest,
    GitConfigureUserRequest,
    GitDeleteBranchRequest,
    GitInitRequest,
    GitPullRequest,
    GitPushRequest,
    GitResetRequest,
    GitRestoreRequest,
    GitSetConfigRequest,
    GitStatus,
    ListBranchResponse,
    ListRemotesResponse,
)

from .._utils.errors import intercept_errors
from .._utils.otel_decorator import with_instrumentation
from ..common.git import GitCommitResponse


class AsyncGit:
    """Provides Git operations within a Sandbox.

    Example:
        ```python
        # Clone a repository
        await sandbox.git.clone(
            url="https://github.com/user/repo.git",
            path="workspace/repo"
        )

        # Check repository status
        status = await sandbox.git.status("workspace/repo")
        print(f"Modified files: {status.modified}")

        # Stage and commit changes
        await sandbox.git.add("workspace/repo", ["file.txt"])
        await sandbox.git.commit(
            path="workspace/repo",
            message="Update file",
            author="John Doe",
            email="john@example.com"
        )
        ```
    """

    def __init__(
        self,
        api_client: GitApi,
    ):
        """Initializes a new Git handler instance.

        Args:
            api_client (GitApi): API client for Sandbox Git operations.
        """
        self._api_client: GitApi = api_client

    @intercept_errors(message_prefix="Failed to add files: ")
    @with_instrumentation()
    async def add(self, path: str, files: list[str]) -> None:
        """Stages the specified files for the next commit, similar to
        running 'git add' on the command line.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            files (list[str]): List of file paths or directories to stage, relative to the repository root.

        Example:
            ```python
            # Stage a single file
            await sandbox.git.add("workspace/repo", ["file.txt"])

            # Stage multiple files
            await sandbox.git.add("workspace/repo", [
                "src/main.py",
                "tests/test_main.py",
                "README.md"
            ])
            ```
        """
        await self._api_client.add_files(
            request=GitAddRequest(path=path, files=files),
        )

    @intercept_errors(message_prefix="Failed to list branches: ")
    @with_instrumentation()
    async def branches(self, path: str) -> ListBranchResponse:
        """Lists branches in the repository.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.

        Returns:
            ListBranchResponse: List of branches in the repository.

        Example:
            ```python
            response = await sandbox.git.branches("workspace/repo")
            print(f"Branches: {response.branches}")
            ```
        """
        return await self._api_client.list_branches(
            path=path,
        )

    @intercept_errors(message_prefix="Failed to clone repository: ")
    @with_instrumentation()
    async def clone(
        self,
        url: str,
        path: str,
        branch: str | None = None,
        commit_id: str | None = None,
        username: str | None = None,
        password: str | None = None,
        insecure_skip_tls: bool | None = None,
        depth: int | None = None,
    ) -> None:
        """Clones a Git repository into the specified path. It supports
        cloning specific branches or commits, and can authenticate with the remote
        repository if credentials are provided.

        Args:
            url (str): Repository URL to clone from.
            path (str): Path where the repository should be cloned. Relative paths are resolved
            based on the sandbox working directory.
            branch (str | None): Specific branch to clone. If not specified,
                clones the default branch.
            commit_id (str | None): Specific commit to clone. If specified,
                the repository will be left in a detached HEAD state at this commit.
            username (str | None): Git username for authentication.
            password (str | None): Git password or token for authentication.
            insecure_skip_tls (bool | None): Skip TLS certificate verification (insecure).
                Use only for trusted internal Git servers with self-signed or private-CA certs;
                credentials, if supplied, are transmitted over an unverified TLS connection.
            depth (int | None): Create a shallow clone truncated to the given number of commits.

        Example:
            ```python
            # Clone the default branch
            await sandbox.git.clone(
                url="https://github.com/user/repo.git",
                path="workspace/repo"
            )

            # Clone a specific branch with authentication
            await sandbox.git.clone(
                url="https://github.com/user/private-repo.git",
                path="workspace/private",
                branch="develop",
                username="user",
                password="token"
            )

            # Clone a specific commit
            await sandbox.git.clone(
                url="https://github.com/user/repo.git",
                path="workspace/repo-old",
                commit_id="abc123"
            )
            ```
        """
        await self._api_client.clone_repository(
            request=GitCloneRequest(
                url=url,
                branch=branch,
                path=path,
                username=username,
                password=password,
                commit_id=commit_id,
                insecure_skip_tls=insecure_skip_tls,
                depth=depth,
            ),
        )

    @intercept_errors(message_prefix="Failed to commit changes: ")
    @with_instrumentation()
    async def commit(
        self, path: str, message: str, author: str, email: str, allow_empty: bool = False
    ) -> GitCommitResponse:
        """Creates a new commit with the staged changes. Make sure to stage
        changes using the add() method before committing.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            message (str): Commit message describing the changes.
            author (str): Name of the commit author.
            email (str): Email address of the commit author.
            allow_empty (bool): Allow creating an empty commit when no changes are staged. Defaults to False.

        Example:
            ```python
            # Stage and commit changes
            await sandbox.git.add("workspace/repo", ["README.md"])
            await sandbox.git.commit(
                path="workspace/repo",
                message="Update documentation",
                author="John Doe",
                email="john@example.com",
                allow_empty=True
            )
            ```
        """
        response = await self._api_client.commit_changes(
            request=GitCommitRequest(
                path=path,
                message=message,
                author=author,
                email=email,
                allow_empty=allow_empty,
            ),
        )
        return GitCommitResponse(sha=response.hash)

    @intercept_errors(message_prefix="Failed to push changes: ")
    @with_instrumentation()
    async def push(
        self,
        path: str,
        username: str | None = None,
        password: str | None = None,
        branch: str | None = None,
        remote: str | None = None,
        set_upstream: bool = False,
    ) -> None:
        """Pushes all local commits on the current branch to the remote
        repository. If the remote repository requires authentication, provide
        username and password/token.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            username (str | None): Git username for authentication.
            password (str | None): Git password or token for authentication.
            branch (str | None): Branch to push. Defaults to the current branch.
            remote (str | None): Remote to push to. Defaults to "origin".
            set_upstream (bool, optional): Record the pushed branch as the upstream tracking
                branch. Defaults to False.

        Example:
            ```python
            # Push without authentication (for public repos or SSH)
            await sandbox.git.push("workspace/repo")

            # Push with authentication
            await sandbox.git.push(
                path="workspace/repo",
                username="user",
                password="github_token"
            )

            # Push a new branch and set its upstream
            await sandbox.git.push("workspace/repo", branch="feature", set_upstream=True)
            ```
        """
        await self._api_client.push_changes(
            request=GitPushRequest(
                path=path,
                username=username,
                password=password,
                branch=branch,
                remote=remote,
                set_upstream=set_upstream,
            ),
        )

    @intercept_errors(message_prefix="Failed to pull changes: ")
    @with_instrumentation()
    async def pull(
        self,
        path: str,
        username: str | None = None,
        password: str | None = None,
        branch: str | None = None,
        remote: str | None = None,
    ) -> None:
        """Pulls changes from the remote repository. If the remote repository requires authentication,
        provide username and password/token.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            username (str | None): Git username for authentication.
            password (str | None): Git password or token for authentication.
            branch (str | None): Branch to pull. Defaults to the current branch's upstream.
            remote (str | None): Remote to pull from. Defaults to "origin".

        Example:
            ```python
            # Pull without authentication
            await sandbox.git.pull("workspace/repo")

            # Pull with authentication
            await sandbox.git.pull(
                path="workspace/repo",
                username="user",
                password="github_token"
            )

            # Pull a specific branch from a specific remote
            await sandbox.git.pull("workspace/repo", remote="upstream", branch="main")
            ```
        """
        await self._api_client.pull_changes(
            request=GitPullRequest(
                path=path,
                username=username,
                password=password,
                branch=branch,
                remote=remote,
            ),
        )

    @intercept_errors(message_prefix="Failed to get status: ")
    @with_instrumentation()
    async def status(self, path: str) -> GitStatus:
        """Gets the current Git repository status.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.

        Returns:
            GitStatus: Repository status information including:
                - current_branch: Current branch name
                - file_status: List of file statuses
                - ahead: Number of local commits not pushed to remote
                - behind: Number of remote commits not pulled locally
                - branch_published: Whether the branch has been published to the remote repository

        Example:
            ```python
            status = await sandbox.git.status("workspace/repo")
            print(f"On branch: {status.current_branch}")
            print(f"Commits ahead: {status.ahead}")
            print(f"Commits behind: {status.behind}")
            ```
        """
        return await self._api_client.get_status(
            path=path,
        )

    @intercept_errors(message_prefix="Failed to checkout branch: ")
    @with_instrumentation()
    async def checkout_branch(self, path: str, branch: str) -> None:
        """Checkout branch in the repository.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            branch (str): Name of the branch to checkout

        Example:
            ```python
            # Checkout a branch
            await sandbox.git.checkout_branch("workspace/repo", "feature-branch")
            ```
        """
        await self._api_client.checkout_branch(
            request=GitCheckoutRequest(
                path=path,
                branch=branch,
            ),
        )

    @intercept_errors(message_prefix="Failed to create branch: ")
    @with_instrumentation()
    async def create_branch(self, path: str, name: str) -> None:
        """Create branch in the repository.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            name (str): Name of the new branch to create

        Example:
            ```python
            # Create a new branch
            await sandbox.git.create_branch("workspace/repo", "new-feature")
            ```
        """
        await self._api_client.create_branch(
            request=GitBranchRequest(
                path=path,
                name=name,
            ),
        )

    @intercept_errors(message_prefix="Failed to delete branch: ")
    @with_instrumentation()
    async def delete_branch(self, path: str, name: str) -> None:
        """Delete branch in the repository.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            name (str): Name of the branch to delete

        Example:
            ```python
            # Delete a branch
            await sandbox.git.delete_branch("workspace/repo", "old-feature")
            ```
        """
        await self._api_client.delete_branch(
            request=GitDeleteBranchRequest(
                path=path,
                name=name,
            ),
        )

    @intercept_errors(message_prefix="Failed to initialize repository: ")
    @with_instrumentation()
    async def init(self, path: str, bare: bool = False, initial_branch: str | None = None) -> None:
        """Initializes a new Git repository at the specified path.

        Args:
            path (str): Path where the repository should be initialized. Relative paths are
            resolved based on the sandbox working directory.
            bare (bool, optional): Create a bare repository without a working tree. Defaults to False.
            initial_branch (str | None): Name of the initial branch. If not specified, uses the
                Git default.

        Example:
            ```python
            await sandbox.git.init("workspace/repo", initial_branch="main")
            ```
        """
        await self._api_client.init_repository(
            request=GitInitRequest(
                path=path,
                bare=bare,
                initial_branch=initial_branch,
            ),
        )

    @intercept_errors(message_prefix="Failed to reset: ")
    @with_instrumentation()
    async def reset(
        self,
        path: str,
        mode: str | None = None,
        target: str | None = None,
        files: list[str] | None = None,
    ) -> None:
        """Resets the current HEAD to the specified state.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            mode (str | None): Reset mode, one of "soft", "mixed" (default), "hard", "merge" or "keep".
            target (str | None): Revision to reset to. Defaults to HEAD.
            files (list[str] | None): Constrain the reset to the given paths.

        Example:
            ```python
            # Unstage all changes (mixed reset to HEAD)
            await sandbox.git.reset("workspace/repo")

            # Hard reset to a previous commit
            await sandbox.git.reset("workspace/repo", mode="hard", target="HEAD~1")
            ```
        """
        await self._api_client.reset_changes(
            request=GitResetRequest(
                path=path,
                mode=mode,
                target=target,
                files=files,
            ),
        )

    @intercept_errors(message_prefix="Failed to restore files: ")
    @with_instrumentation()
    async def restore(
        self,
        path: str,
        files: list[str],
        staged: bool | None = None,
        worktree: bool | None = None,
        source: str | None = None,
    ) -> None:
        """Restores working tree files or unstages changes.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            files (list[str]): File paths to restore.
            staged (bool | None): Restore the staging index for the given files.
            worktree (bool | None): Restore the working tree for the given files. Defaults to
                True when neither staged nor worktree is provided.
            source (str | None): Restore file contents from the given revision instead of the index.

        Example:
            ```python
            # Discard working tree changes
            await sandbox.git.restore("workspace/repo", ["file.txt"])

            # Unstage changes
            await sandbox.git.restore("workspace/repo", ["file.txt"], staged=True)
            ```
        """
        await self._api_client.restore_files(
            request=GitRestoreRequest(
                path=path,
                files=files,
                staged=staged,
                worktree=worktree,
                source=source,
            ),
        )

    @intercept_errors(message_prefix="Failed to add remote: ")
    @with_instrumentation()
    async def remote_add(
        self,
        path: str,
        name: str,
        url: str,
        fetch: bool = False,
        overwrite: bool = False,
    ) -> None:
        """Adds (or overwrites) a remote in the repository.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            name (str): Name of the remote.
            url (str): URL of the remote.
            fetch (bool, optional): Fetch from the remote immediately after adding it. Defaults to False.
            overwrite (bool, optional): Replace an existing remote with the same name. Defaults to False.

        Example:
            ```python
            await sandbox.git.remote_add("workspace/repo", "origin", "https://github.com/user/repo.git")
            ```
        """
        await self._api_client.add_remote(
            request=GitAddRemoteRequest(
                path=path,
                name=name,
                url=url,
                fetch=fetch,
                overwrite=overwrite,
            ),
        )

    @intercept_errors(message_prefix="Failed to list remotes: ")
    @with_instrumentation()
    async def remotes(self, path: str) -> ListRemotesResponse:
        """Lists the remotes configured in the repository.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.

        Returns:
            ListRemotesResponse: The configured remotes (name + URL).

        Example:
            ```python
            response = await sandbox.git.remotes("workspace/repo")
            for remote in response.remotes:
                print(f"{remote.name}: {remote.url}")
            ```
        """
        return await self._api_client.list_remotes(
            path=path,
        )

    @intercept_errors(message_prefix="Failed to get remote: ")
    @with_instrumentation()
    async def remote_get(self, path: str, name: str) -> str | None:
        """Gets the URL of a remote, or None when it does not exist.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            name (str): Name of the remote.

        Returns:
            str | None: The remote URL, or None when the remote does not exist.

        Example:
            ```python
            url = await sandbox.git.remote_get("workspace/repo", "origin")
            ```
        """
        response = await self._api_client.list_remotes(path=path)
        for remote in response.remotes:
            if remote.name == name:
                return remote.url
        return None

    @intercept_errors(message_prefix="Failed to set config: ")
    @with_instrumentation()
    async def set_config(self, key: str, value: str, scope: str = "global", path: str | None = None) -> None:
        """Sets a Git config value at the given scope.

        Args:
            key (str): Config key in dotted form (e.g. "user.name").
            value (str): Config value.
            scope (str, optional): Config scope, one of "global" (default), "local" or "system".
            path (str | None): Repository path, required when scope is "local".

        Example:
            ```python
            await sandbox.git.set_config("user.name", "John Doe")
            ```
        """
        await self._api_client.set_git_config(
            request=GitSetConfigRequest(
                path=path,
                key=key,
                value=value,
                scope=scope,
            ),
        )

    @intercept_errors(message_prefix="Failed to get config: ")
    @with_instrumentation()
    async def get_config(self, key: str, scope: str = "global", path: str | None = None) -> str | None:
        """Gets a Git config value at the given scope, or None when unset.

        Args:
            key (str): Config key in dotted form (e.g. "user.name").
            scope (str, optional): Config scope, one of "global" (default), "local" or "system".
            path (str | None): Repository path, required when scope is "local".

        Returns:
            str | None: The config value, or None when the key is not set.

        Example:
            ```python
            name = await sandbox.git.get_config("user.name")
            ```
        """
        response = await self._api_client.get_git_config(key=key, scope=scope, path=path)
        return response.value

    @intercept_errors(message_prefix="Failed to configure user: ")
    @with_instrumentation()
    async def configure_user(self, name: str, email: str, scope: str = "global", path: str | None = None) -> None:
        """Configures the Git user name and email at the given scope.

        Args:
            name (str): User name (user.name).
            email (str): User email (user.email).
            scope (str, optional): Config scope, one of "global" (default), "local" or "system".
            path (str | None): Repository path, required when scope is "local".

        Example:
            ```python
            await sandbox.git.configure_user("John Doe", "john@example.com")
            ```
        """
        await self._api_client.configure_user(
            request=GitConfigureUserRequest(
                path=path,
                name=name,
                email=email,
                scope=scope,
            ),
        )

    @intercept_errors(message_prefix="Failed to authenticate: ")
    @with_instrumentation()
    async def dangerously_authenticate(
        self,
        username: str,
        password: str,
        host: str | None = None,
        protocol: str | None = None,
    ) -> None:
        """Persists Git credentials globally so that subsequent operations against the
        given host authenticate automatically.

        Warning:
            This stores the password in plaintext on disk via the Git credential store.

        Args:
            username (str): Git username.
            password (str): Git password or token.
            host (str | None): Host to authenticate against. Defaults to "github.com".
            protocol (str | None): Protocol to authenticate against. Defaults to "https".

        Example:
            ```python
            await sandbox.git.dangerously_authenticate("user", "github_token")
            ```
        """
        await self._api_client.authenticate(
            request=GitAuthenticateRequest(
                username=username,
                password=password,
                host=host,
                protocol=protocol,
            ),
        )
