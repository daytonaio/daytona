# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from typing import List, Optional

from daytona_toolbox_api_client_async import (
    GitAddRequest,
    GitApi,
    GitBranchRequest,
    GitCheckoutRequest,
    GitCloneRequest,
    GitCommitRequest,
    GitGitDeleteBranchRequest,
    GitRepoRequest,
    GitStatus,
    ListBranchResponse,
)

from .._utils.errors import intercept_errors
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
        self._api_client = api_client

    @intercept_errors(message_prefix="Failed to add files: ")
    async def add(self, path: str, files: List[str]) -> None:
        """Stages the specified files for the next commit, similar to
        running 'git add' on the command line.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            files (List[str]): List of file paths or directories to stage, relative to the repository root.

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
    async def clone(
        self,
        url: str,
        path: str,
        branch: Optional[str] = None,
        commit_id: Optional[str] = None,
        username: Optional[str] = None,
        password: Optional[str] = None,
    ) -> None:
        """Clones a Git repository into the specified path. It supports
        cloning specific branches or commits, and can authenticate with the remote
        repository if credentials are provided.

        Args:
            url (str): Repository URL to clone from.
            path (str): Path where the repository should be cloned. Relative paths are resolved
            based on the sandbox working directory.
            branch (Optional[str]): Specific branch to clone. If not specified,
                clones the default branch.
            commit_id (Optional[str]): Specific commit to clone. If specified,
                the repository will be left in a detached HEAD state at this commit.
            username (Optional[str]): Git username for authentication.
            password (Optional[str]): Git password or token for authentication.

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
                commitId=commit_id,
            ),
        )

    @intercept_errors(message_prefix="Failed to commit changes: ")
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
            allow_empty (bool, optional): Allow creating an empty commit when no changes are staged. Defaults to False.

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
    async def push(
        self,
        path: str,
        username: Optional[str] = None,
        password: Optional[str] = None,
    ) -> None:
        """Pushes all local commits on the current branch to the remote
        repository. If the remote repository requires authentication, provide
        username and password/token.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            username (Optional[str]): Git username for authentication.
            password (Optional[str]): Git password or token for authentication.

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
            ```
        """
        await self._api_client.push_changes(
            request=GitRepoRequest(
                path=path,
                username=username,
                password=password,
            ),
        )

    @intercept_errors(message_prefix="Failed to pull changes: ")
    async def pull(
        self,
        path: str,
        username: Optional[str] = None,
        password: Optional[str] = None,
    ) -> None:
        """Pulls changes from the remote repository. If the remote repository requires authentication,
        provide username and password/token.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on
            the sandbox working directory.
            username (Optional[str]): Git username for authentication.
            password (Optional[str]): Git password or token for authentication.

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
            ```
        """
        await self._api_client.pull_changes(
            request=GitRepoRequest(
                path=path,
                username=username,
                password=password,
            ),
        )

    @intercept_errors(message_prefix="Failed to get status: ")
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
            request=GitGitDeleteBranchRequest(
                path=path,
                name=name,
            ),
        )
