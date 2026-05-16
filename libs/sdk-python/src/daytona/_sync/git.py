# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from daytona_toolbox_api_client import (
    GitAddRequest,
    GitApi,
    GitBranchRequest,
    GitCheckoutRequest,
    GitCloneRequest,
    GitCommitRequest,
    GitDeleteBranchRequest,
    GitRepoRequest,
    GitStatus,
    ListBranchResponse,
)

from .._utils.errors import intercept_errors
from .._utils.otel_decorator import with_instrumentation
from ..common.git import GitCommitResponse


class Git:
    """Provides Git operations within a Sandbox.

    Example:
        ```python
        # Clone a repository
        sandbox.git.clone(
            url="https://github.com/user/repo.git",
            path="workspace/repo"
        )

        # Check repository status
        status = sandbox.git.status("workspace/repo")
        print(f"Modified files: {status.modified}")

        # Stage and commit changes
        sandbox.git.add("workspace/repo", ["file.txt"])
        sandbox.git.commit(
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
    def add(self, path: str, files: list[str]) -> None:
        """Stages the specified files for the next commit, similar to
        running 'git add' on the command line.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            files (list[str]): List of file paths or directories to stage, relative to the repository root.

        Example:
            ```python
            # Stage a single file
            sandbox.git.add("workspace/repo", ["file.txt"])

            # Stage multiple files
            sandbox.git.add("workspace/repo", [
                "src/main.py",
                "tests/test_main.py",
                "README.md"
            ])
            ```
        """
        self._api_client.add_files(
            request=GitAddRequest(path=path, files=files),
        )

    @intercept_errors(message_prefix="Failed to list branches: ")
    @with_instrumentation()
    def branches(self, path: str) -> ListBranchResponse:
        """Lists branches in the repository.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.

        Returns:
            ListBranchResponse: List of branches in the repository.

        Example:
            ```python
            response = sandbox.git.branches("workspace/repo")
            print(f"Branches: {response.branches}")
            ```
        """
        return self._api_client.list_branches(
            path=path,
        )

    @intercept_errors(message_prefix="Failed to clone repository: ")
    @with_instrumentation()
    def clone(
        self,
        url: str,
        path: str,
        branch: str | None = None,
        commit_id: str | None = None,
        username: str | None = None,
        password: str | None = None,
        depth: int | None = None,
        single_branch: bool | None = None,
        shallow_since: str | None = None,
        no_tags: bool | None = None,
        filter: str | None = None,
        sparse: bool | None = None,
        sparse_paths: list[str] | None = None,
        reference_path: str | None = None,
        dissociate: bool | None = None,
        recurse_submodules: bool | None = None,
        shallow_submodules: bool | None = None,
        filter_submodules: bool | None = None,
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
            depth (int | None): Number of commits to fetch for a shallow clone.
            single_branch (bool | None): Whether to restrict history to one branch.
            shallow_since (str | None): Fetch only history newer than this date.
            no_tags (bool | None): Skip fetching tags.
            filter (str | None): Partial clone filter, such as "blob:none".
            sparse (bool | None): Initialize sparse checkout.
            sparse_paths (list[str] | None): Sparse checkout paths to include.
            reference_path (str | None): Local Git object store to borrow from if available.
            dissociate (bool | None): Copy borrowed reference objects into the clone.
            recurse_submodules (bool | None): Clone submodules recursively.
            shallow_submodules (bool | None): Use shallow clones for submodules.
            filter_submodules (bool | None): Apply the partial clone filter to submodules.

        Example:
            ```python
            # Clone the default branch
            sandbox.git.clone(
                url="https://github.com/user/repo.git",
                path="workspace/repo"
            )

            # Clone a specific branch with authentication
            sandbox.git.clone(
                url="https://github.com/user/private-repo.git",
                path="workspace/private",
                branch="develop",
                username="user",
                password="token"
            )

            # Clone a specific commit
            sandbox.git.clone(
                url="https://github.com/user/repo.git",
                path="workspace/repo-old",
                commit_id="abc123"
            )
            ```
        """
        self._api_client.clone_repository(
            request=GitCloneRequest(
                url=url,
                branch=branch,
                path=path,
                username=username,
                password=password,
                commit_id=commit_id,
                depth=depth,
                single_branch=single_branch,
                shallow_since=shallow_since,
                no_tags=no_tags,
                filter=filter,
                sparse=sparse,
                sparse_paths=sparse_paths,
                reference_path=reference_path,
                dissociate=dissociate,
                recurse_submodules=recurse_submodules,
                shallow_submodules=shallow_submodules,
                filter_submodules=filter_submodules,
            ),
        )

    @intercept_errors(message_prefix="Failed to commit changes: ")
    @with_instrumentation()
    def commit(self, path: str, message: str, author: str, email: str, allow_empty: bool = False) -> GitCommitResponse:
        """Creates a new commit with the staged changes. Make sure to stage
        changes using the add() method before committing.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            message (str): Commit message describing the changes.
            author (str): Name of the commit author.
            email (str): Email address of the commit author.
            allow_empty (bool, optional): Allow creating an empty commit when no changes are staged. Defaults to False.

        Example:
            ```python
            # Stage and commit changes
            sandbox.git.add("workspace/repo", ["README.md"])
            sandbox.git.commit(
                path="workspace/repo",
                message="Update documentation",
                author="John Doe",
                email="john@example.com",
                allow_empty=True
            )
            ```
        """
        response = self._api_client.commit_changes(
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
    def push(
        self,
        path: str,
        username: str | None = None,
        password: str | None = None,
    ) -> None:
        """Pushes all local commits on the current branch to the remote
        repository. If the remote repository requires authentication, provide
        username and password/token.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            username (str | None): Git username for authentication.
            password (str | None): Git password or token for authentication.

        Example:
            ```python
            # Push without authentication (for public repos or SSH)
            sandbox.git.push("workspace/repo")

            # Push with authentication
            sandbox.git.push(
                path="workspace/repo",
                username="user",
                password="github_token"
            )
            ```
        """
        self._api_client.push_changes(
            request=GitRepoRequest(
                path=path,
                username=username,
                password=password,
            ),
        )

    @intercept_errors(message_prefix="Failed to pull changes: ")
    @with_instrumentation()
    def pull(
        self,
        path: str,
        username: str | None = None,
        password: str | None = None,
    ) -> None:
        """Pulls changes from the remote repository. If the remote repository requires authentication,
        provide username and password/token.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            username (str | None): Git username for authentication.
            password (str | None): Git password or token for authentication.

        Example:
            ```python
            # Pull without authentication
            sandbox.git.pull("workspace/repo")

            # Pull with authentication
            sandbox.git.pull(
                path="workspace/repo",
                username="user",
                password="github_token"
            )
            ```
        """
        self._api_client.pull_changes(
            request=GitRepoRequest(
                path=path,
                username=username,
                password=password,
            ),
        )

    @intercept_errors(message_prefix="Failed to get status: ")
    @with_instrumentation()
    def status(self, path: str) -> GitStatus:
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
            status = sandbox.git.status("workspace/repo")
            print(f"On branch: {status.current_branch}")
            print(f"Commits ahead: {status.ahead}")
            print(f"Commits behind: {status.behind}")
            ```
        """
        return self._api_client.get_status(
            path=path,
        )

    @intercept_errors(message_prefix="Failed to checkout branch: ")
    @with_instrumentation()
    def checkout_branch(self, path: str, branch: str) -> None:
        """Checkout branch in the repository.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            branch (str): Name of the branch to checkout

        Example:
            ```python
            # Checkout a branch
            sandbox.git.checkout_branch("workspace/repo", "feature-branch")
            ```
        """
        self._api_client.checkout_branch(
            request=GitCheckoutRequest(
                path=path,
                branch=branch,
            ),
        )

    @intercept_errors(message_prefix="Failed to create branch: ")
    @with_instrumentation()
    def create_branch(self, path: str, name: str) -> None:
        """Create branch in the repository.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            name (str): Name of the new branch to create

        Example:
            ```python
            # Create a new branch
            sandbox.git.create_branch("workspace/repo", "new-feature")
            ```
        """
        self._api_client.create_branch(
            request=GitBranchRequest(
                path=path,
                name=name,
            ),
        )

    @intercept_errors(message_prefix="Failed to delete branch: ")
    @with_instrumentation()
    def delete_branch(self, path: str, name: str) -> None:
        """Delete branch in the repository.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            name (str): Name of the branch to delete

        Example:
            ```python
            # Delete a branch
            sandbox.git.delete_branch("workspace/repo", "old-feature")
            ```
        """
        self._api_client.delete_branch(
            request=GitDeleteBranchRequest(
                path=path,
                name=name,
            ),
        )
