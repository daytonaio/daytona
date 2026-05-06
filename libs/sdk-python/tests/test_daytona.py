# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import json
from unittest.mock import MagicMock, patch

import pytest

from daytona.common.daytona import CreateSandboxFromImageParams, CreateSandboxFromSnapshotParams, DaytonaConfig
from daytona.common.errors import DaytonaAuthenticationError, DaytonaValidationError
from daytona.common.sandbox import Resources

SYNC_MODULE = "daytona._sync.daytona"


def _make_daytona(config=None):
    from daytona._sync.daytona import Daytona
    from daytona_api_client import Configuration

    with patch(f"{SYNC_MODULE}.ApiClient") as mock_api_cls, patch(
        f"{SYNC_MODULE}.ToolboxApiClient"
    ) as mock_toolbox_cls:
        mock_api_instance = MagicMock()
        mock_api_instance.configuration = Configuration(host="https://test.daytona.io/api")
        mock_api_instance.default_headers = {}
        mock_api_instance.user_agent = ""
        mock_api_cls.return_value = mock_api_instance

        mock_toolbox_instance = MagicMock()
        mock_toolbox_instance.default_headers = {}
        mock_toolbox_cls.return_value = mock_toolbox_instance

        return Daytona(config)


class TestDaytonaInit:
    def test_init_with_config(self):
        daytona = _make_daytona(DaytonaConfig(api_key="test-key", api_url="https://api.test.io", target="us"))
        assert daytona._api_key == "test-key"
        assert daytona._api_url == "https://api.test.io"
        assert daytona._target == "us"

    def test_init_with_env_vars(self, env_with_api_key):
        daytona = _make_daytona()
        assert daytona._api_key == "test-api-key-123"
        assert daytona._api_url == "https://test.daytona.io/api"
        assert daytona._target == "us"

    def test_init_with_jwt(self, env_with_jwt):
        daytona = _make_daytona()
        assert daytona._jwt_token == "test-jwt-token-123"
        assert daytona._organization_id == "test-org-id"

    @patch("daytona._utils.env.dotenv_values", return_value={})
    def test_init_without_credentials_raises(self, _mock_dotenv, monkeypatch):
        for key in [
            "DAYTONA_API_KEY",
            "DAYTONA_JWT_TOKEN",
            "DAYTONA_API_URL",
            "DAYTONA_TARGET",
            "DAYTONA_SERVER_URL",
            "DAYTONA_ORGANIZATION_ID",
        ]:
            monkeypatch.delenv(key, raising=False)

        from daytona._sync.daytona import Daytona

        with pytest.raises(
            DaytonaAuthenticationError, match="Authentication credentials not found. Set DAYTONA_API_KEY"
        ):
            Daytona()

    @patch("daytona._utils.env.dotenv_values", return_value={})
    def test_default_api_url(self, _mock_dotenv, monkeypatch):
        monkeypatch.setenv("DAYTONA_API_KEY", "key")
        monkeypatch.setenv("DAYTONA_TARGET", "us")
        monkeypatch.delenv("DAYTONA_API_URL", raising=False)
        monkeypatch.delenv("DAYTONA_SERVER_URL", raising=False)
        daytona = _make_daytona()
        assert daytona._api_url == "https://app.daytona.io/api"

    @patch("daytona._utils.env.dotenv_values", return_value={})
    def test_env_server_url_warns_when_api_url_missing(self, _mock_dotenv, monkeypatch):
        monkeypatch.setenv("DAYTONA_API_KEY", "key")
        monkeypatch.setenv("DAYTONA_TARGET", "us")
        monkeypatch.setenv("DAYTONA_SERVER_URL", "https://server.daytona.io/api")
        monkeypatch.delenv("DAYTONA_API_URL", raising=False)

        with pytest.warns(DeprecationWarning, match="DAYTONA_SERVER_URL"):
            daytona = _make_daytona()

        assert daytona._api_url == "https://server.daytona.io/api"

    def test_jwt_without_organization_id_raises(self):
        with pytest.raises(DaytonaAuthenticationError, match="DAYTONA_ORGANIZATION_ID is required"):
            _make_daytona(DaytonaConfig(jwt_token="jwt", api_url="https://api.test.io", target="us"))


class TestDaytonaCreateValidation:
    def test_negative_timeout_raises(self, env_with_api_key):
        daytona = _make_daytona()
        with pytest.raises(DaytonaValidationError, match="Timeout must be a non-negative number"):
            daytona._create(CreateSandboxFromSnapshotParams(language="python"), timeout=-1)

    def test_negative_auto_stop_raises(self, env_with_api_key):
        daytona = _make_daytona()
        with pytest.raises(DaytonaValidationError, match="auto_stop_interval must be a non-negative"):
            daytona._create(CreateSandboxFromSnapshotParams(language="python", auto_stop_interval=-1), timeout=60)

    def test_negative_auto_archive_raises(self, env_with_api_key):
        daytona = _make_daytona()
        with pytest.raises(DaytonaValidationError, match="auto_archive_interval must be a non-negative"):
            daytona._create(CreateSandboxFromSnapshotParams(language="python", auto_archive_interval=-1), timeout=60)

    def test_create_defaults_language_and_sets_label(self, env_with_api_key, sandbox_dto):
        from daytona.common.daytona import CODE_TOOLBOX_LANGUAGE_LABEL

        daytona = _make_daytona()
        daytona._sandbox_api = MagicMock()
        daytona._sandbox_api.create_sandbox.return_value = sandbox_dto
        sandbox = daytona.create()

        request = daytona._sandbox_api.create_sandbox.call_args.kwargs["_request_timeout"]
        assert request == 60
        create_request = daytona._sandbox_api.create_sandbox.call_args.args[0]
        assert create_request.labels[CODE_TOOLBOX_LANGUAGE_LABEL] == "python"
        assert sandbox.id == sandbox_dto.id

    def test_create_from_image_sets_resources(self, env_with_api_key, sandbox_dto):
        daytona = _make_daytona()
        daytona._sandbox_api = MagicMock()
        daytona._sandbox_api.create_sandbox.return_value = sandbox_dto
        params = CreateSandboxFromImageParams(image="python:3.12", resources=Resources(cpu=2, memory=4, disk=8, gpu=1))
        daytona.create(params)
        create_request = daytona._sandbox_api.create_sandbox.call_args.args[0]
        assert create_request.cpu == 2
        assert create_request.memory == 4
        assert create_request.disk == 8
        assert create_request.gpu == 1

    def test_create_from_snapshot_sets_snapshot_and_volume_mounts(self, env_with_api_key, sandbox_dto):
        from daytona.common.volume import VolumeMount

        daytona = _make_daytona()
        daytona._sandbox_api = MagicMock()
        daytona._sandbox_api.create_sandbox.return_value = sandbox_dto
        params = CreateSandboxFromSnapshotParams(
            snapshot="snap-1",
            volumes=[VolumeMount(volume_id="vol-1", mount_path="/data", subpath="logs")],
        )

        daytona.create(params)

        create_request = daytona._sandbox_api.create_sandbox.call_args.args[0]
        assert create_request.snapshot == "snap-1"
        assert create_request.volumes[0].volume_id == "vol-1"
        assert create_request.volumes[0].subpath == "logs"


class TestDaytonaGetAndList:
    def test_get_empty_id_raises(self, env_with_api_key):
        daytona = _make_daytona()
        with pytest.raises(DaytonaValidationError, match="sandbox_id_or_name is required"):
            daytona.get("")

    def test_get_returns_sandbox(self, env_with_api_key, sandbox_dto):
        from daytona._sync.sandbox import Sandbox

        daytona = _make_daytona()
        daytona._sandbox_api = MagicMock()
        daytona._sandbox_api.get_sandbox.return_value = sandbox_dto
        sandbox = daytona.get("test-sandbox-id")
        assert isinstance(sandbox, Sandbox)
        assert sandbox.id == "test-sandbox-id"

    def test_list_invalid_page_raises(self, env_with_api_key):
        daytona = _make_daytona()
        with pytest.raises(DaytonaValidationError, match="page must be a positive integer"):
            daytona.list(page=0)

    def test_list_invalid_limit_raises(self, env_with_api_key):
        daytona = _make_daytona()
        with pytest.raises(DaytonaValidationError, match="limit must be a positive integer"):
            daytona.list(limit=0)

    def test_list_serializes_labels(self, env_with_api_key, sandbox_dto):
        response = MagicMock(items=[sandbox_dto], total=1, page=1, total_pages=1)
        daytona = _make_daytona()
        daytona._sandbox_api = MagicMock()
        daytona._sandbox_api.list_sandboxes_paginated.return_value = response
        result = daytona.list(labels={"project": "test"}, page=1, limit=10)
        assert result.total == 1
        kwargs = daytona._sandbox_api.list_sandboxes_paginated.call_args.kwargs
        assert json.loads(kwargs["labels"]) == {"project": "test"}
        assert kwargs["page"] == 1
        assert kwargs["limit"] == 10


class TestDaytonaValidateLanguageLabel:
    def test_none_returns_python(self, env_with_api_key):
        from daytona.common.daytona import CodeLanguage

        daytona = _make_daytona()
        assert daytona._validate_language_label(None) == CodeLanguage.PYTHON

    @pytest.mark.parametrize("value", ["python", "typescript", "javascript"])
    def test_valid_language(self, env_with_api_key, value):
        daytona = _make_daytona()
        assert str(daytona._validate_language_label(value)) == value

    def test_invalid_language_raises(self, env_with_api_key):
        from daytona.common.daytona import CODE_TOOLBOX_LANGUAGE_LABEL

        daytona = _make_daytona()
        with pytest.raises(DaytonaValidationError, match=f"Invalid {CODE_TOOLBOX_LANGUAGE_LABEL}"):
            daytona._validate_language_label("ruby")
