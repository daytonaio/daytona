# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import warnings
from unittest.mock import MagicMock

import pytest

from daytona.common.errors import DaytonaError, DaytonaValidationError
from daytona_api_client import SandboxState

from .conftest import make_sandbox_dto


def make_sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api):
    from daytona._sync.sandbox import Sandbox

    return Sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api, "python", http_client=MagicMock())


class TestSandboxInit:
    def test_sandbox_properties(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api):
        sandbox = make_sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api)
        assert sandbox.id == "test-sandbox-id"
        assert sandbox.name == "test-sandbox"
        assert sandbox.state == SandboxState.STARTED
        assert sandbox.cpu == 4
        assert sandbox.memory == 8
        assert sandbox.disk == 30
        assert sandbox.user == "daytona"
        assert sandbox.public is False

    def test_sandbox_has_subsystems(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api):
        sandbox = make_sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api)
        assert sandbox.fs is not None
        assert sandbox.git is not None
        assert sandbox.process is not None
        assert sandbox.computer_use is not None
        assert sandbox.code_interpreter is not None


class TestSandboxLifecycleSettings:
    def test_negative_autostop_interval_raises(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api):
        sandbox = make_sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api)
        with pytest.raises(DaytonaValidationError, match="Auto-stop interval must be a non-negative"):
            sandbox.set_autostop_interval(-1)

    def test_valid_autostop_interval(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api):
        sandbox = make_sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api)
        sandbox.set_autostop_interval(30)
        assert sandbox.auto_stop_interval == 30
        mock_sandbox_api.set_autostop_interval.assert_called_once_with(sandbox.id, 30)

    def test_negative_auto_archive_interval_raises(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api):
        sandbox = make_sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api)
        with pytest.raises(DaytonaValidationError, match="Auto-archive interval must be a non-negative"):
            sandbox.set_auto_archive_interval(-1)

    def test_valid_auto_archive_interval(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api):
        sandbox = make_sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api)
        sandbox.set_auto_archive_interval(60)
        assert sandbox.auto_archive_interval == 60

    def test_auto_delete_interval(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api):
        sandbox = make_sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api)
        sandbox.set_auto_delete_interval(120)
        assert sandbox.auto_delete_interval == 120
        sandbox.set_auto_delete_interval(-1)
        assert sandbox.auto_delete_interval == -1


class TestSandboxOperations:
    def test_set_labels(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api):
        sandbox = make_sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api)
        new_labels = {"project": "test", "env": "dev"}
        mock_response = MagicMock(labels=new_labels)
        mock_sandbox_api.replace_labels.return_value = mock_response
        result = sandbox.set_labels(new_labels)
        assert result == new_labels
        assert sandbox.labels == new_labels

    def test_create_lsp_server(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api):
        from daytona._sync.lsp_server import LspServer

        sandbox = make_sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api)
        lsp = sandbox.create_lsp_server("python", "/workspace/project")
        assert isinstance(lsp, LspServer)

    def test_refresh_data(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api):
        sandbox = make_sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api)
        mock_sandbox_api.get_sandbox.return_value = make_sandbox_dto(state=SandboxState.STOPPED, cpu=8)
        sandbox.refresh_data()
        assert sandbox.state == SandboxState.STOPPED
        assert sandbox.cpu == 8

    def test_get_user_home_dir(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api):
        sandbox = make_sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api)
        sandbox._info_api = MagicMock(get_user_home_dir=MagicMock(return_value=MagicMock(dir="/home/daytona")))
        assert sandbox.get_user_home_dir() == "/home/daytona"

    def test_get_work_dir(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api):
        sandbox = make_sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api)
        sandbox._info_api = MagicMock(get_work_dir=MagicMock(return_value=MagicMock(dir="/workspace")))
        assert sandbox.get_work_dir() == "/workspace"

    def test_get_user_root_dir_delegates_to_home_dir(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api):
        sandbox = make_sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api)
        sandbox._info_api = MagicMock(get_user_home_dir=MagicMock(return_value=MagicMock(dir="/home/daytona")))

        with warnings.catch_warnings(record=True):
            warnings.simplefilter("always")
            assert sandbox.get_user_root_dir() == "/home/daytona"

    def test_preview_and_ssh_operations_delegate(self, sandbox_dto, mock_toolbox_api_client, mock_sandbox_api):
        sandbox = make_sandbox(sandbox_dto, mock_toolbox_api_client, mock_sandbox_api)
        mock_sandbox_api.get_port_preview_url.return_value = MagicMock(url="https://preview")
        mock_sandbox_api.create_ssh_access.return_value = MagicMock(token="ssh-token")
        mock_sandbox_api.validate_ssh_access.return_value = MagicMock(valid=True)

        assert sandbox.get_preview_link(3000).url == "https://preview"
        assert sandbox.create_ssh_access(10).token == "ssh-token"
        assert sandbox.validate_ssh_access("token").valid is True
        sandbox.revoke_ssh_access("token")
        sandbox.refresh_activity()

        mock_sandbox_api.get_port_preview_url.assert_called_once_with(sandbox.id, 3000)
        mock_sandbox_api.revoke_ssh_access.assert_called_once_with(sandbox.id, "token")
        mock_sandbox_api.update_last_activity.assert_called_once_with(sandbox.id)


class TestSandboxWaitForStart:
    def test_error_state_raises(self, mock_toolbox_api_client, mock_sandbox_api):
        error_dto = make_sandbox_dto(state=SandboxState.ERROR, error_reason="build failed")
        sandbox = make_sandbox(error_dto, mock_toolbox_api_client, mock_sandbox_api)
        mock_sandbox_api.get_sandbox.return_value = error_dto
        with pytest.raises(DaytonaError, match="failed to start"):
            sandbox.wait_for_sandbox_start(timeout=0)
