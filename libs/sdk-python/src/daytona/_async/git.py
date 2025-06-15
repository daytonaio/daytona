# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from typing import Awaitable, Callable, List, Optional

from daytona_api_client_async import (
    GitAddRequest,
    GitCloneRequest,
    GitCommitRequest,
    GitRepoRequest,
    GitStatus,
    ListBranchResponse,
    ToolboxApi,
)

from .._utils.errors import intercept_errors
from .._utils.path import prefix_relative_path
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
        sandbox_id: str,
        toolbox_api: ToolboxApi,
        get_root_dir: Callable[[], Awaitable[str]],
    ):
        """Initializes a new Git handler instance.

        Args:
            sandbox_id (str): The Sandbox ID.
            toolbox_api (ToolboxApi): API client for Sandbox operations.
            get_root_dir (Callable[[], str]): A function to get the default root directory of the Sandbox.
        """
        self._sandbox_id = sandbox_id
        self._toolbox_api = toolbox_api
        self._get_root_dir = get_root_dir

    @intercept_errors(message_prefix="Failed to add files: ")
    async def add(self, path: str, files: List[str]) -> None:
        """Stages the specified files for the next commit, similar to
        running 'git add' on the command line.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on the user's
            root directory.
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
        await self._toolbox_api.git_add_files(
            self._sandbox_id,
            git_add_request=GitAddRequest(path=prefix_relative_path(await self._get_root_dir(), path), files=files),
        )

    @intercept_errors(message_prefix="Failed to list branches: ")
    async def branches(self, path: str) -> ListBranchResponse:
        """Lists branches in the repository.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on the user's
            root directory.

        Returns:
            ListBranchResponse: List of branches in the repository.

        Example:
            ```python
            response = await sandbox.git.branches("workspace/repo")
            print(f"Branches: {response.branches}")
            ```
        """
        return await self._toolbox_api.git_list_branches(
            self._sandbox_id,
            path=prefix_relative_path(await self._get_root_dir(), path),
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
            path (str): Path where the repository should be cloned. Relative paths are resolved based on the user's
            root directory.
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
        await self._toolbox_api.git_clone_repository(
            self._sandbox_id,
            git_clone_request=GitCloneRequest(
                url=url,
                branch=branch,
                path=prefix_relative_path(await self._get_root_dir(), path),
                username=username,
                password=password,
                commitId=commit_id,
            ),
        )

    @intercept_errors(message_prefix="Failed to commit changes: ")
    async def commit(self, path: str, message: str, author: str, email: str) -> GitCommitResponse:
        """Creates a new commit with the staged changes. Make sure to stage
        changes using the add() method before committing.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on the user's
            root directory.
            message (str): Commit message describing the changes.
            author (str): Name of the commit author.
            email (str): Email address of the commit author.

        Example:
            ```python
            # Stage and commit changes
            await sandbox.git.add("workspace/repo", ["README.md"])
            await sandbox.git.commit(
                path="workspace/repo",
                message="Update documentation",
                author="John Doe",
                email="john@example.com"
            )
            ```
        """
        response = await self._toolbox_api.git_commit_changes(
            self._sandbox_id,
            git_commit_request=GitCommitRequest(
                path=prefix_relative_path(await self._get_root_dir(), path),
                message=message,
                author=author,
                email=email,
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
            path (str): Path to the Git repository root. Relative paths are resolved based on the user's
            root directory.
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
        await self._toolbox_api.git_push_changes(
            self._sandbox_id,
            git_repo_request=GitRepoRequest(
                path=prefix_relative_path(await self._get_root_dir(), path),
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
            path (str): Path to the Git repository root. Relative paths are resolved based on the user's
            root directory.
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
        await self._toolbox_api.git_pull_changes(
            self._sandbox_id,
            git_repo_request=GitRepoRequest(
                path=prefix_relative_path(await self._get_root_dir(), path),
                username=username,
                password=password,
            ),
        )

    @intercept_errors(message_prefix="Failed to get status: ")
    async def status(self, path: str) -> GitStatus:
        """Gets the current Git repository status.

        Args:
            path (str): Path to the Git repository root. Relative paths are resolved based on the user's
            root directory.

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
        return await self._toolbox_api.git_get_status(
            self._sandbox_id,
            path=prefix_relative_path(await self._get_root_dir(), path),
        )
