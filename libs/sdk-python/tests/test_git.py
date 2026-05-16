# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest

from daytona.common.git import GitCommitResponse


class TestSyncGit:
    def _make_git(self):
        from daytona._sync.git import Git

        mock_api = MagicMock()
        return Git(mock_api), mock_api

    def test_add(self):
        git, api = self._make_git()
        api.add_files.return_value = None
        git.add("workspace/repo", ["file.txt", "src/main.py"])
        api.add_files.assert_called_once()
        call_args = api.add_files.call_args
        assert call_args.kwargs["request"].path == "workspace/repo"
        assert call_args.kwargs["request"].files == ["file.txt", "src/main.py"]

    def test_branches(self):
        git, api = self._make_git()
        mock_response = MagicMock()
        mock_response.branches = ["main", "develop"]
        api.list_branches.return_value = mock_response
        result = git.branches("workspace/repo")
        assert result.branches == ["main", "develop"]

    def test_clone(self):
        git, api = self._make_git()
        api.clone_repository.return_value = None
        git.clone("https://github.com/user/repo.git", "workspace/repo")
        api.clone_repository.assert_called_once()
        call_args = api.clone_repository.call_args
        assert call_args.kwargs["request"].url == "https://github.com/user/repo.git"
        assert call_args.kwargs["request"].path == "workspace/repo"

    def test_clone_with_branch(self):
        git, api = self._make_git()
        api.clone_repository.return_value = None
        git.clone("https://github.com/user/repo.git", "workspace/repo", branch="develop")
        call_args = api.clone_repository.call_args
        assert call_args.kwargs["request"].branch == "develop"

    def test_clone_with_credentials(self):
        git, api = self._make_git()
        api.clone_repository.return_value = None
        git.clone(
            "https://github.com/user/repo.git",
            "workspace/repo",
            username="user",
            password="token",
        )
        call_args = api.clone_repository.call_args
        assert call_args.kwargs["request"].username == "user"
        assert call_args.kwargs["request"].password == "token"

    def test_clone_with_advanced_options(self):
        git, api = self._make_git()
        api.clone_repository.return_value = None
        git.clone(
            "https://github.com/user/repo.git",
            "workspace/repo",
            branch="main",
            commit_id="abc123",
            username="user",
            password="token",
            depth=1,
            single_branch=False,
            shallow_since="2025-01-01",
            no_tags=True,
            filter="blob:none",
            sparse=True,
            sparse_paths=["src", "README.md"],
            reference_path="/cache/repo.git",
            dissociate=True,
            recurse_submodules=True,
            shallow_submodules=True,
            filter_submodules=True,
        )
        request = api.clone_repository.call_args.kwargs["request"]
        assert request.branch == "main"
        assert request.commit_id == "abc123"
        assert request.depth == 1
        assert request.single_branch is False
        assert request.shallow_since == "2025-01-01"
        assert request.no_tags is True
        assert request.filter == "blob:none"
        assert request.sparse is True
        assert request.sparse_paths == ["src", "README.md"]
        assert request.reference_path == "/cache/repo.git"
        assert request.dissociate is True
        assert request.recurse_submodules is True
        assert request.shallow_submodules is True
        assert request.filter_submodules is True

    def test_commit(self):
        git, api = self._make_git()
        mock_response = MagicMock()
        mock_response.hash = "abc123"
        api.commit_changes.return_value = mock_response
        result = git.commit(
            path="workspace/repo",
            message="initial commit",
            author="Test",
            email="test@example.com",
        )
        assert isinstance(result, GitCommitResponse)
        assert result.sha == "abc123"

    def test_commit_allow_empty_and_push_with_credentials(self):
        git, api = self._make_git()
        api.commit_changes.return_value = MagicMock(hash="abc123")

        git.commit("workspace/repo", "msg", "Author", "a@b.com", allow_empty=True)
        git.push("workspace/repo", username="user", password="token")

        commit_request = api.commit_changes.call_args.kwargs["request"]
        assert commit_request.allow_empty is True
        push_request = api.push_changes.call_args.kwargs["request"]
        assert push_request.username == "user"
        assert push_request.password == "token"

    def test_push(self):
        git, api = self._make_git()
        api.push_changes.return_value = None
        git.push("workspace/repo")
        api.push_changes.assert_called_once()

    def test_pull(self):
        git, api = self._make_git()
        api.pull_changes.return_value = None
        git.pull("workspace/repo")
        api.pull_changes.assert_called_once()

    def test_status(self):
        git, api = self._make_git()
        mock_status = MagicMock()
        mock_status.current_branch = "main"
        mock_status.ahead = 0
        mock_status.behind = 0
        api.get_status.return_value = mock_status
        result = git.status("workspace/repo")
        assert result.current_branch == "main"

    def test_checkout_branch(self):
        git, api = self._make_git()
        api.checkout_branch.return_value = None
        git.checkout_branch("workspace/repo", "feature-branch")
        api.checkout_branch.assert_called_once()

    def test_create_branch(self):
        git, api = self._make_git()
        api.create_branch.return_value = None
        git.create_branch("workspace/repo", "new-feature")
        api.create_branch.assert_called_once()

    def test_delete_branch(self):
        git, api = self._make_git()
        api.delete_branch.return_value = None
        git.delete_branch("workspace/repo", "old-feature")
        api.delete_branch.assert_called_once()


class TestAsyncGit:
    def _make_git(self):
        from daytona._async.git import AsyncGit

        mock_api = AsyncMock()
        return AsyncGit(mock_api), mock_api

    @pytest.mark.asyncio
    async def test_add(self):
        git, api = self._make_git()
        await git.add("workspace/repo", ["file.txt"])
        api.add_files.assert_called_once()

    @pytest.mark.asyncio
    async def test_clone(self):
        git, api = self._make_git()
        await git.clone("https://github.com/user/repo.git", "workspace/repo")
        api.clone_repository.assert_called_once()

    @pytest.mark.asyncio
    async def test_clone_with_advanced_options(self):
        git, api = self._make_git()
        await git.clone(
            "https://github.com/user/repo.git",
            "workspace/repo",
            branch="main",
            commit_id="abc123",
            username="user",
            password="token",
            depth=1,
            single_branch=False,
            shallow_since="2025-01-01",
            no_tags=True,
            filter="blob:none",
            sparse=True,
            sparse_paths=["src", "README.md"],
            reference_path="/cache/repo.git",
            dissociate=True,
            recurse_submodules=True,
            shallow_submodules=True,
            filter_submodules=True,
        )
        request = api.clone_repository.call_args.kwargs["request"]
        assert request.branch == "main"
        assert request.commit_id == "abc123"
        assert request.depth == 1
        assert request.single_branch is False
        assert request.shallow_since == "2025-01-01"
        assert request.no_tags is True
        assert request.filter == "blob:none"
        assert request.sparse is True
        assert request.sparse_paths == ["src", "README.md"]
        assert request.reference_path == "/cache/repo.git"
        assert request.dissociate is True
        assert request.recurse_submodules is True
        assert request.shallow_submodules is True
        assert request.filter_submodules is True

    @pytest.mark.asyncio
    async def test_commit(self):
        git, api = self._make_git()
        mock_response = MagicMock()
        mock_response.hash = "def456"
        api.commit_changes.return_value = mock_response
        result = await git.commit("workspace/repo", "msg", "Author", "a@b.com")
        assert result.sha == "def456"

    @pytest.mark.asyncio
    async def test_commit_allow_empty_and_push_with_credentials(self):
        git, api = self._make_git()
        api.commit_changes.return_value = MagicMock(hash="def456")

        await git.commit("workspace/repo", "msg", "Author", "a@b.com", allow_empty=True)
        await git.push("workspace/repo", username="user", password="token")

        commit_request = api.commit_changes.call_args.kwargs["request"]
        assert commit_request.allow_empty is True
        push_request = api.push_changes.call_args.kwargs["request"]
        assert push_request.username == "user"
        assert push_request.password == "token"

    @pytest.mark.asyncio
    async def test_branches(self):
        git, api = self._make_git()
        mock_response = MagicMock()
        mock_response.branches = ["main"]
        api.list_branches.return_value = mock_response
        result = await git.branches("workspace/repo")
        assert result.branches == ["main"]

    @pytest.mark.asyncio
    async def test_status(self):
        git, api = self._make_git()
        mock_status = MagicMock()
        mock_status.current_branch = "develop"
        api.get_status.return_value = mock_status
        result = await git.status("workspace/repo")
        assert result.current_branch == "develop"

    @pytest.mark.asyncio
    async def test_push(self):
        git, api = self._make_git()
        await git.push("workspace/repo")
        api.push_changes.assert_called_once()

    @pytest.mark.asyncio
    async def test_pull(self):
        git, api = self._make_git()
        await git.pull("workspace/repo")
        api.pull_changes.assert_called_once()

    @pytest.mark.asyncio
    async def test_checkout_branch(self):
        git, api = self._make_git()
        await git.checkout_branch("workspace/repo", "feature")
        api.checkout_branch.assert_called_once()

    @pytest.mark.asyncio
    async def test_create_branch(self):
        git, api = self._make_git()
        await git.create_branch("workspace/repo", "new")
        api.create_branch.assert_called_once()

    @pytest.mark.asyncio
    async def test_delete_branch(self):
        git, api = self._make_git()
        await git.delete_branch("workspace/repo", "old")
        api.delete_branch.assert_called_once()
