# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import json
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from daytona.common.daytona import CreateSandboxFromImageParams, CreateSandboxFromSnapshotParams, DaytonaConfig
from daytona.common.errors import DaytonaAuthenticationError, DaytonaValidationError
from daytona.common.sandbox import Resources

ASYNC_MODULE = "daytona._async.daytona"


def _make_async_daytona(config=None):
    from daytona._async.daytona import AsyncDaytona
    from daytona_api_client_async import Configuration

    with patch(f"{ASYNC_MODULE}.ApiClient") as mock_api_cls, patch(
        f"{ASYNC_MODULE}.ToolboxApiClient"
    ) as mock_toolbox_cls:
        mock_api_instance = MagicMock()
        mock_api_instance.configuration = Configuration(host="https://test.daytona.io/api")
        mock_api_instance.default_headers = {}
        mock_api_instance.user_agent = ""
        mock_api_instance.close = AsyncMock()
        mock_api_cls.return_value = mock_api_instance

        mock_toolbox_instance = MagicMock()
        mock_toolbox_instance.default_headers = {}
        mock_toolbox_instance.close = AsyncMock()
        mock_toolbox_cls.return_value = mock_toolbox_instance

        return AsyncDaytona(config)


class TestAsyncDaytonaInit:
    def test_init_with_config(self):
        daytona = _make_async_daytona(DaytonaConfig(api_key="test-key", api_url="https://api.test.io", target="us"))
        assert daytona._api_key == "test-key"
        assert daytona._api_url == "https://api.test.io"
        assert daytona._target == "us"

    def test_init_with_env_vars(self, env_with_api_key):
        daytona = _make_async_daytona()
        assert daytona._api_key == "test-api-key-123"
        assert daytona._api_url == "https://test.daytona.io/api"

    def test_init_with_jwt(self, env_with_jwt):
        daytona = _make_async_daytona()
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

        from daytona._async.daytona import AsyncDaytona

        with pytest.raises(DaytonaAuthenticationError, match="API key or JWT token is required"):
            AsyncDaytona()

    def test_jwt_without_organization_id_raises(self):
        with pytest.raises(DaytonaAuthenticationError, match="Organization ID is required"):
            _make_async_daytona(DaytonaConfig(jwt_token="jwt", api_url="https://api.test.io", target="us"))


class TestAsyncDaytonaContextManager:
    @pytest.mark.asyncio
    async def test_context_manager_closes_clients(self, env_with_api_key):
        from daytona._async.daytona import AsyncDaytona
        from daytona_api_client_async import Configuration

        with patch(f"{ASYNC_MODULE}.ApiClient") as mock_api_cls, patch(
            f"{ASYNC_MODULE}.ToolboxApiClient"
        ) as mock_toolbox_cls:
            mock_api_instance = MagicMock()
            mock_api_instance.configuration = Configuration(host="https://test.daytona.io/api")
            mock_api_instance.default_headers = {}
            mock_api_instance.user_agent = ""
            mock_api_instance.close = AsyncMock()
            mock_api_cls.return_value = mock_api_instance

            mock_toolbox_instance = MagicMock()
            mock_toolbox_instance.default_headers = {}
            mock_toolbox_instance.close = AsyncMock()
            mock_toolbox_cls.return_value = mock_toolbox_instance

            async with AsyncDaytona() as daytona:
                assert daytona is not None

        mock_api_instance.close.assert_awaited_once()
        mock_toolbox_instance.close.assert_awaited_once()

    @pytest.mark.asyncio
    async def test_close_shuts_down_tracer_provider(self, env_with_api_key):
        daytona = _make_async_daytona()
        tracer_provider = MagicMock()
        daytona._tracer_provider = tracer_provider

        await daytona.close()

        tracer_provider.shutdown.assert_called_once()
        daytona._api_client.close.assert_awaited_once()
        daytona._toolbox_api_client.close.assert_awaited_once()


class TestAsyncDaytonaCreateValidation:
    @pytest.mark.asyncio
    async def test_negative_timeout_raises(self, env_with_api_key):
        daytona = _make_async_daytona()
        with pytest.raises(DaytonaValidationError, match="Timeout must be a non-negative number"):
            await daytona._create(CreateSandboxFromSnapshotParams(language="python"), timeout=-1)

    @pytest.mark.asyncio
    async def test_negative_auto_stop_raises(self, env_with_api_key):
        daytona = _make_async_daytona()
        with pytest.raises(DaytonaValidationError, match="auto_stop_interval must be a non-negative"):
            await daytona._create(CreateSandboxFromSnapshotParams(language="python", auto_stop_interval=-1), timeout=60)

    @pytest.mark.asyncio
    async def test_negative_auto_archive_raises(self, env_with_api_key):
        daytona = _make_async_daytona()
        with pytest.raises(DaytonaValidationError, match="auto_archive_interval must be a non-negative"):
            await daytona._create(
                CreateSandboxFromSnapshotParams(language="python", auto_archive_interval=-1), timeout=60
            )

    @pytest.mark.asyncio
    async def test_create_defaults_language_and_sets_label(self, env_with_api_key, sandbox_dto):
        from daytona.common.daytona import CODE_TOOLBOX_LANGUAGE_LABEL

        daytona = _make_async_daytona()
        daytona._sandbox_api.create_sandbox = AsyncMock(return_value=sandbox_dto)
        sandbox = await daytona.create()
        create_request = daytona._sandbox_api.create_sandbox.call_args.args[0]
        assert create_request.labels[CODE_TOOLBOX_LANGUAGE_LABEL] == "python"
        assert sandbox.id == sandbox_dto.id

    @pytest.mark.asyncio
    async def test_create_from_image_sets_resources(self, env_with_api_key, sandbox_dto):
        daytona = _make_async_daytona()
        daytona._sandbox_api.create_sandbox = AsyncMock(return_value=sandbox_dto)
        params = CreateSandboxFromImageParams(image="python:3.12", resources=Resources(cpu=2, memory=4, disk=8, gpu=1))
        await daytona.create(params)
        create_request = daytona._sandbox_api.create_sandbox.call_args.args[0]
        assert create_request.cpu == 2
        assert create_request.memory == 4
        assert create_request.disk == 8
        assert create_request.gpu == 1

    @pytest.mark.asyncio
    async def test_create_from_snapshot_sets_snapshot_and_volume_mounts(self, env_with_api_key, sandbox_dto):
        from daytona.common.volume import VolumeMount

        daytona = _make_async_daytona()
        daytona._sandbox_api.create_sandbox = AsyncMock(return_value=sandbox_dto)
        params = CreateSandboxFromSnapshotParams(
            snapshot="snap-1",
            volumes=[VolumeMount(volume_id="vol-1", mount_path="/data", subpath="logs")],
        )

        await daytona.create(params)

        create_request = daytona._sandbox_api.create_sandbox.call_args.args[0]
        assert create_request.snapshot == "snap-1"
        assert create_request.volumes[0].volume_id == "vol-1"
        assert create_request.volumes[0].subpath == "logs"


class TestAsyncDaytonaGetAndList:
    @pytest.mark.asyncio
    async def test_get_empty_id_raises(self, env_with_api_key):
        daytona = _make_async_daytona()
        with pytest.raises(DaytonaValidationError, match="sandbox_id_or_name is required"):
            await daytona.get("")

    @pytest.mark.asyncio
    async def test_get_returns_sandbox(self, env_with_api_key, sandbox_dto):
        from daytona._async.sandbox import AsyncSandbox

        daytona = _make_async_daytona()
        daytona._sandbox_api = AsyncMock()
        daytona._sandbox_api.get_sandbox.return_value = sandbox_dto
        sandbox = await daytona.get("test-sandbox-id")
        assert isinstance(sandbox, AsyncSandbox)
        assert sandbox.id == "test-sandbox-id"

    @pytest.mark.asyncio
    async def test_list_invalid_page_raises(self, env_with_api_key):
        daytona = _make_async_daytona()
        with pytest.raises(DaytonaValidationError, match="page must be a positive integer"):
            await daytona.list(page=0)

    @pytest.mark.asyncio
    async def test_list_invalid_limit_raises(self, env_with_api_key):
        daytona = _make_async_daytona()
        with pytest.raises(DaytonaValidationError, match="limit must be a positive integer"):
            await daytona.list(limit=0)

    @pytest.mark.asyncio
    async def test_list_serializes_labels(self, env_with_api_key, sandbox_dto):
        response = MagicMock(items=[sandbox_dto], total=1, page=1, total_pages=1)
        daytona = _make_async_daytona()
        daytona._sandbox_api.list_sandboxes_paginated = AsyncMock(return_value=response)
        result = await daytona.list(labels={"project": "test"}, page=1, limit=10)
        assert result.total == 1
        kwargs = daytona._sandbox_api.list_sandboxes_paginated.call_args.kwargs
        assert json.loads(kwargs["labels"]) == {"project": "test"}
        assert kwargs["page"] == 1
        assert kwargs["limit"] == 10


class TestAsyncDaytonaValidateLanguageLabel:
    @pytest.mark.parametrize("value", [None, "python", "typescript", "javascript"])
    def test_valid_language_values(self, env_with_api_key, value):
        daytona = _make_async_daytona()
        result = daytona._validate_language_label(value)
        assert str(result) == (value or "python")

    def test_invalid_language_raises(self, env_with_api_key):
        from daytona.common.daytona import CODE_TOOLBOX_LANGUAGE_LABEL

        daytona = _make_async_daytona()
        with pytest.raises(DaytonaValidationError, match=f"Invalid {CODE_TOOLBOX_LANGUAGE_LABEL}"):
            daytona._validate_language_label("ruby")
