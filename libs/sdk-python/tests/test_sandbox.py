from __future__ import annotations

from unittest.mock import MagicMock

import pytest

from daytona_api_client import SandboxState

from daytona.common.errors import DaytonaError
from daytona.common.sandbox import Resources

from .conftest import make_sandbox_dto


class TestSandboxInit:
    def test_sandbox_properties(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox):
        from daytona._sync.sandbox import Sandbox

        sandbox = Sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox)
        assert sandbox.id == "test-sandbox-id"
        assert sandbox.name == "test-sandbox"
        assert sandbox.state == SandboxState.STARTED
        assert sandbox.cpu == 4
        assert sandbox.memory == 8
        assert sandbox.disk == 30
        assert sandbox.user == "daytona"
        assert sandbox.public is False

    def test_sandbox_has_fs_git_process(
        self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox
    ):
        from daytona._sync.sandbox import Sandbox

        sandbox = Sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox)
        assert sandbox.fs is not None
        assert sandbox.git is not None
        assert sandbox.process is not None
        assert sandbox.computer_use is not None
        assert sandbox.code_interpreter is not None


class TestSandboxSetAutostopInterval:
    def test_negative_interval_raises(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox):
        from daytona._sync.sandbox import Sandbox

        sandbox = Sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox)
        with pytest.raises(DaytonaError, match="Auto-stop interval must be a non-negative"):
            sandbox.set_autostop_interval(-1)

    def test_valid_interval(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox):
        from daytona._sync.sandbox import Sandbox

        sandbox = Sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox)
        mock_sandbox_api.set_autostop_interval.return_value = None
        sandbox.set_autostop_interval(30)
        assert sandbox.auto_stop_interval == 30
        mock_sandbox_api.set_autostop_interval.assert_called_once_with(sandbox.id, 30)

    def test_zero_disables_autostop(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox):
        from daytona._sync.sandbox import Sandbox

        sandbox = Sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox)
        mock_sandbox_api.set_autostop_interval.return_value = None
        sandbox.set_autostop_interval(0)
        assert sandbox.auto_stop_interval == 0


class TestSandboxSetAutoArchiveInterval:
    def test_negative_interval_raises(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox):
        from daytona._sync.sandbox import Sandbox

        sandbox = Sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox)
        with pytest.raises(DaytonaError, match="Auto-archive interval must be a non-negative"):
            sandbox.set_auto_archive_interval(-1)

    def test_valid_interval(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox):
        from daytona._sync.sandbox import Sandbox

        sandbox = Sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox)
        mock_sandbox_api.set_auto_archive_interval.return_value = None
        sandbox.set_auto_archive_interval(60)
        assert sandbox.auto_archive_interval == 60


class TestSandboxSetAutoDeleteInterval:
    def test_valid_interval(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox):
        from daytona._sync.sandbox import Sandbox

        sandbox = Sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox)
        mock_sandbox_api.set_auto_delete_interval.return_value = None
        sandbox.set_auto_delete_interval(120)
        assert sandbox.auto_delete_interval == 120

    def test_negative_disables(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox):
        from daytona._sync.sandbox import Sandbox

        sandbox = Sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox)
        mock_sandbox_api.set_auto_delete_interval.return_value = None
        sandbox.set_auto_delete_interval(-1)
        assert sandbox.auto_delete_interval == -1


class TestSandboxSetLabels:
    def test_set_labels(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox):
        from daytona._sync.sandbox import Sandbox

        sandbox = Sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox)
        new_labels = {"project": "test", "env": "dev"}
        mock_response = MagicMock()
        mock_response.labels = new_labels
        mock_sandbox_api.replace_labels.return_value = mock_response
        result = sandbox.set_labels(new_labels)
        assert result == new_labels
        assert sandbox.labels == new_labels


class TestSandboxCreateLspServer:
    def test_create_lsp_server(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox):
        from daytona._sync.sandbox import Sandbox
        from daytona._sync.lsp_server import LspServer

        sandbox = Sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox)
        lsp = sandbox.create_lsp_server("python", "/workspace/project")
        assert isinstance(lsp, LspServer)


class TestSandboxRefreshData:
    def test_refresh_data(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox):
        from daytona._sync.sandbox import Sandbox

        sandbox = Sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox)
        updated_dto = make_sandbox_dto(state=SandboxState.STOPPED, cpu=8)
        mock_sandbox_api.get_sandbox.return_value = updated_dto
        sandbox.refresh_data()
        assert sandbox.state == SandboxState.STOPPED
        assert sandbox.cpu == 8


class TestSandboxGetUserHomeDir:
    def test_get_user_home_dir(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox):
        from daytona._sync.sandbox import Sandbox

        sandbox = Sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox)
        mock_info_api = MagicMock()
        mock_info_api.get_user_home_dir.return_value = MagicMock(dir="/home/daytona")
        sandbox._info_api = mock_info_api
        result = sandbox.get_user_home_dir()
        assert result == "/home/daytona"


class TestSandboxGetWorkDir:
    def test_get_work_dir(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox):
        from daytona._sync.sandbox import Sandbox

        sandbox = Sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, mock_code_toolbox)
        mock_info_api = MagicMock()
        mock_info_api.get_work_dir.return_value = MagicMock(dir="/workspace")
        sandbox._info_api = mock_info_api
        result = sandbox.get_work_dir()
        assert result == "/workspace"
