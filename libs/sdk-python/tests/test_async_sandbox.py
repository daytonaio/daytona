# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import warnings
from unittest.mock import AsyncMock, MagicMock

import pytest

from daytona.common.errors import DaytonaError, DaytonaValidationError
from daytona_api_client import SandboxState

from .conftest import make_sandbox_dto


def make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api):
    from daytona._async.sandbox import AsyncSandbox

    return AsyncSandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, "python")


class TestAsyncSandboxInit:
    def test_sandbox_properties(self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api)
        assert sandbox.id == "test-sandbox-id"
        assert sandbox.name == "test-sandbox"
        assert sandbox.state == SandboxState.STARTED
        assert sandbox.cpu == 4

    def test_sandbox_has_subsystems(self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api)
        assert sandbox.fs is not None
        assert sandbox.git is not None
        assert sandbox.process is not None
        assert sandbox.computer_use is not None
        assert sandbox.code_interpreter is not None


class TestAsyncSandboxLifecycleSettings:
    @pytest.mark.asyncio
    async def test_negative_autostop_interval_raises(
        self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api
    ):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api)
        with pytest.raises(DaytonaValidationError, match="Auto-stop interval must be a non-negative"):
            await sandbox.set_autostop_interval(-1)

    @pytest.mark.asyncio
    async def test_valid_autostop_interval(self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api)
        mock_async_sandbox_api.set_autostop_interval = AsyncMock(return_value=None)
        await sandbox.set_autostop_interval(30)
        assert sandbox.auto_stop_interval == 30

    @pytest.mark.asyncio
    async def test_negative_auto_archive_interval_raises(
        self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api
    ):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api)
        with pytest.raises(DaytonaValidationError, match="Auto-archive interval must be a non-negative"):
            await sandbox.set_auto_archive_interval(-1)

    @pytest.mark.asyncio
    async def test_valid_auto_archive_interval(
        self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api
    ):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api)
        mock_async_sandbox_api.set_auto_archive_interval = AsyncMock(return_value=None)
        await sandbox.set_auto_archive_interval(60)
        assert sandbox.auto_archive_interval == 60

    @pytest.mark.asyncio
    async def test_auto_delete_interval(self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api)
        mock_async_sandbox_api.set_auto_delete_interval = AsyncMock(return_value=None)
        await sandbox.set_auto_delete_interval(120)
        assert sandbox.auto_delete_interval == 120


class TestAsyncSandboxOperations:
    @pytest.mark.asyncio
    async def test_set_labels(self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api)
        new_labels = {"project": "test", "env": "dev"}
        mock_async_sandbox_api.replace_labels = AsyncMock(return_value=MagicMock(labels=new_labels))
        result = await sandbox.set_labels(new_labels)
        assert result == new_labels

    def test_create_lsp_server(self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api):
        from daytona._async.lsp_server import AsyncLspServer

        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api)
        lsp = sandbox.create_lsp_server("python", "/workspace/project")
        assert isinstance(lsp, AsyncLspServer)

    @pytest.mark.asyncio
    async def test_refresh_data(self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api)
        mock_async_sandbox_api.get_sandbox = AsyncMock(return_value=make_sandbox_dto(state=SandboxState.STOPPED, cpu=8))
        await sandbox.refresh_data()
        assert sandbox.state == SandboxState.STOPPED
        assert sandbox.cpu == 8

    @pytest.mark.asyncio
    async def test_get_user_home_dir(self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api)
        sandbox._info_api = AsyncMock(get_user_home_dir=AsyncMock(return_value=MagicMock(dir="/home/daytona")))
        assert await sandbox.get_user_home_dir() == "/home/daytona"

    @pytest.mark.asyncio
    async def test_get_work_dir(self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api)
        sandbox._info_api = AsyncMock(get_work_dir=AsyncMock(return_value=MagicMock(dir="/workspace")))
        assert await sandbox.get_work_dir() == "/workspace"

    @pytest.mark.asyncio
    async def test_get_user_root_dir_delegates_to_home_dir(
        self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api
    ):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api)
        sandbox._info_api = AsyncMock(get_user_home_dir=AsyncMock(return_value=MagicMock(dir="/home/daytona")))

        with warnings.catch_warnings(record=True):
            warnings.simplefilter("always")
            assert await sandbox.get_user_root_dir() == "/home/daytona"

    @pytest.mark.asyncio
    async def test_preview_and_ssh_operations_delegate(
        self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api
    ):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api)
        mock_async_sandbox_api.get_port_preview_url = AsyncMock(return_value=MagicMock(url="https://preview"))
        mock_async_sandbox_api.create_ssh_access = AsyncMock(return_value=MagicMock(token="ssh-token"))
        mock_async_sandbox_api.validate_ssh_access = AsyncMock(return_value=MagicMock(valid=True))
        mock_async_sandbox_api.revoke_ssh_access = AsyncMock()
        mock_async_sandbox_api.update_last_activity = AsyncMock()

        assert (await sandbox.get_preview_link(3000)).url == "https://preview"
        assert (await sandbox.create_ssh_access(10)).token == "ssh-token"
        assert (await sandbox.validate_ssh_access("token")).valid is True
        await sandbox.revoke_ssh_access("token")
        await sandbox.refresh_activity()

        mock_async_sandbox_api.get_port_preview_url.assert_awaited_once_with(sandbox.id, 3000)
        mock_async_sandbox_api.revoke_ssh_access.assert_awaited_once_with(sandbox.id, "token")
        mock_async_sandbox_api.update_last_activity.assert_awaited_once_with(sandbox.id)


class TestAsyncSandboxWaitForStart:
    @pytest.mark.asyncio
    async def test_error_state_raises(self, mock_async_toolbox_api_client, mock_async_sandbox_api):
        error_dto = make_sandbox_dto(state=SandboxState.ERROR, error_reason="build failed")
        sandbox = make_async_sandbox(error_dto, mock_async_toolbox_api_client, mock_async_sandbox_api)
        mock_async_sandbox_api.get_sandbox = AsyncMock(return_value=error_dto)
        with pytest.raises(DaytonaError, match="failed to start"):
            await sandbox.wait_for_sandbox_start(timeout=0)
