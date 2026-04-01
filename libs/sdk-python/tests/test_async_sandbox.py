from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest

from daytona_api_client import SandboxState

from daytona.common.errors import DaytonaError

from .conftest import make_sandbox_dto


def make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox):
    from daytona._async.sandbox import AsyncSandbox

    return AsyncSandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox)


class TestAsyncSandboxInit:
    def test_sandbox_properties(
        self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox
    ):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox)
        assert sandbox.id == "test-sandbox-id"
        assert sandbox.name == "test-sandbox"
        assert sandbox.state == SandboxState.STARTED
        assert sandbox.cpu == 4

    def test_sandbox_has_subsystems(
        self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox
    ):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox)
        assert sandbox.fs is not None
        assert sandbox.git is not None
        assert sandbox.process is not None
        assert sandbox.computer_use is not None
        assert sandbox.code_interpreter is not None


class TestAsyncSandboxSetAutostopInterval:
    @pytest.mark.asyncio
    async def test_negative_interval_raises(
        self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox
    ):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox)
        with pytest.raises(DaytonaError, match="Auto-stop interval must be a non-negative"):
            await sandbox.set_autostop_interval(-1)

    @pytest.mark.asyncio
    async def test_valid_interval(
        self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox
    ):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox)
        mock_async_sandbox_api.set_autostop_interval = AsyncMock(return_value=None)
        await sandbox.set_autostop_interval(30)
        assert sandbox.auto_stop_interval == 30


class TestAsyncSandboxSetAutoArchiveInterval:
    @pytest.mark.asyncio
    async def test_negative_interval_raises(
        self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox
    ):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox)
        with pytest.raises(DaytonaError, match="Auto-archive interval must be a non-negative"):
            await sandbox.set_auto_archive_interval(-1)

    @pytest.mark.asyncio
    async def test_valid_interval(
        self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox
    ):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox)
        mock_async_sandbox_api.set_auto_archive_interval = AsyncMock(return_value=None)
        await sandbox.set_auto_archive_interval(60)
        assert sandbox.auto_archive_interval == 60


class TestAsyncSandboxSetAutoDeleteInterval:
    @pytest.mark.asyncio
    async def test_valid_interval(
        self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox
    ):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox)
        mock_async_sandbox_api.set_auto_delete_interval = AsyncMock(return_value=None)
        await sandbox.set_auto_delete_interval(120)
        assert sandbox.auto_delete_interval == 120


class TestAsyncSandboxSetLabels:
    @pytest.mark.asyncio
    async def test_set_labels(
        self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox
    ):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox)
        new_labels = {"project": "test", "env": "dev"}
        mock_response = MagicMock()
        mock_response.labels = new_labels
        mock_async_sandbox_api.replace_labels = AsyncMock(return_value=mock_response)
        result = await sandbox.set_labels(new_labels)
        assert result == new_labels


class TestAsyncSandboxRefreshData:
    @pytest.mark.asyncio
    async def test_refresh_data(
        self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox
    ):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox)
        updated_dto = make_sandbox_dto(state=SandboxState.STOPPED, cpu=8)
        mock_async_sandbox_api.get_sandbox = AsyncMock(return_value=updated_dto)
        await sandbox.refresh_data()
        assert sandbox.state == SandboxState.STOPPED
        assert sandbox.cpu == 8


class TestAsyncSandboxGetUserHomeDir:
    @pytest.mark.asyncio
    async def test_get_user_home_dir(
        self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox
    ):
        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox)
        mock_info_api = AsyncMock()
        mock_info_api.get_user_home_dir.return_value = MagicMock(dir="/home/daytona")
        sandbox._info_api = mock_info_api
        result = await sandbox.get_user_home_dir()
        assert result == "/home/daytona"


class TestAsyncSandboxCreateLspServer:
    def test_create_lsp_server(
        self, sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox
    ):
        from daytona._async.lsp_server import AsyncLspServer

        sandbox = make_async_sandbox(sandbox_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox)
        lsp = sandbox.create_lsp_server("python", "/workspace/project")
        assert isinstance(lsp, AsyncLspServer)


class TestAsyncSandboxWaitForStart:
    @pytest.mark.asyncio
    async def test_error_state_raises(
        self, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox
    ):
        error_dto = make_sandbox_dto(state=SandboxState.ERROR, error_reason="build failed")
        sandbox = make_async_sandbox(error_dto, mock_async_toolbox_api_client, mock_async_sandbox_api, mock_code_toolbox)
        mock_async_sandbox_api.get_sandbox = AsyncMock(return_value=error_dto)
        with pytest.raises(DaytonaError, match="failed to start"):
            await sandbox.wait_for_sandbox_start(timeout=0)
