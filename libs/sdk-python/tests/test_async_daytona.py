from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from daytona.common.daytona import (
    CodeLanguage,
    CreateSandboxFromSnapshotParams,
    DaytonaConfig,
)
from daytona.common.errors import DaytonaError

ASYNC_MODULE = "daytona._async.daytona"


def _make_async_daytona(config=None):
    from daytona_api_client_async import Configuration

    from daytona._async.daytona import AsyncDaytona

    with (
        patch(f"{ASYNC_MODULE}.ApiClient") as mock_api_cls,
        patch(f"{ASYNC_MODULE}.ToolboxApiClient") as mock_toolbox_cls,
    ):
        mock_api_instance = MagicMock()
        mock_api_instance.configuration = Configuration(host="https://test.daytona.io/api")
        mock_api_instance.default_headers = {}
        mock_api_instance.user_agent = ""
        mock_api_cls.return_value = mock_api_instance

        mock_toolbox_instance = MagicMock()
        mock_toolbox_instance.default_headers = {}
        mock_toolbox_cls.return_value = mock_toolbox_instance

        d = AsyncDaytona(config)
    return d


class TestAsyncDaytonaInit:
    def test_init_with_config(self):
        config = DaytonaConfig(
            api_key="test-key",
            api_url="https://api.test.io",
            target="us",
        )
        d = _make_async_daytona(config)
        assert d._api_key == "test-key"
        assert d._api_url == "https://api.test.io"
        assert d._target == "us"

    def test_init_with_env_vars(self, env_with_api_key):
        d = _make_async_daytona()
        assert d._api_key == "test-api-key-123"
        assert d._api_url == "https://test.daytona.io/api"

    def test_init_with_jwt(self, env_with_jwt):
        d = _make_async_daytona()
        assert d._jwt_token == "test-jwt-token-123"
        assert d._organization_id == "test-org-id"

    @patch("daytona._utils.env.dotenv_values", return_value={})
    def test_init_without_credentials_raises(self, mock_dotenv, monkeypatch):
        monkeypatch.delenv("DAYTONA_API_KEY", raising=False)
        monkeypatch.delenv("DAYTONA_JWT_TOKEN", raising=False)
        monkeypatch.delenv("DAYTONA_API_URL", raising=False)
        monkeypatch.delenv("DAYTONA_TARGET", raising=False)
        monkeypatch.delenv("DAYTONA_SERVER_URL", raising=False)
        monkeypatch.delenv("DAYTONA_ORGANIZATION_ID", raising=False)

        from daytona._async.daytona import AsyncDaytona

        with pytest.raises(DaytonaError, match="API key or JWT token is required"):
            AsyncDaytona()


class TestAsyncDaytonaContextManager:
    @pytest.mark.asyncio
    async def test_context_manager(self, env_with_api_key):
        from daytona_api_client_async import Configuration

        from daytona._async.daytona import AsyncDaytona

        with (
            patch(f"{ASYNC_MODULE}.ApiClient") as mock_api_cls,
            patch(f"{ASYNC_MODULE}.ToolboxApiClient") as mock_toolbox_cls,
        ):
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

            async with AsyncDaytona() as d:
                assert d is not None


class TestAsyncDaytonaCreateValidation:
    @pytest.mark.asyncio
    async def test_negative_timeout_raises(self, env_with_api_key):
        d = _make_async_daytona()
        with pytest.raises(DaytonaError, match="Timeout must be a non-negative number"):
            await d._create(CreateSandboxFromSnapshotParams(language="python"), timeout=-1)

    @pytest.mark.asyncio
    async def test_negative_auto_stop_raises(self, env_with_api_key):
        d = _make_async_daytona()
        params = CreateSandboxFromSnapshotParams(language="python", auto_stop_interval=-1)
        with pytest.raises(DaytonaError, match="auto_stop_interval must be a non-negative"):
            await d._create(params, timeout=60)

    @pytest.mark.asyncio
    async def test_negative_auto_archive_raises(self, env_with_api_key):
        d = _make_async_daytona()
        params = CreateSandboxFromSnapshotParams(language="python", auto_archive_interval=-1)
        with pytest.raises(DaytonaError, match="auto_archive_interval must be a non-negative"):
            await d._create(params, timeout=60)


class TestAsyncDaytonaGet:
    @pytest.mark.asyncio
    async def test_get_empty_id_raises(self, env_with_api_key):
        d = _make_async_daytona()
        with pytest.raises(DaytonaError, match="sandbox_id_or_name is required"):
            await d.get("")

    @pytest.mark.asyncio
    async def test_get_returns_sandbox(self, env_with_api_key, sandbox_dto):
        from daytona._async.sandbox import AsyncSandbox

        d = _make_async_daytona()
        d._sandbox_api = AsyncMock()
        d._sandbox_api.get_sandbox = AsyncMock(return_value=sandbox_dto)
        sandbox = await d.get("test-sandbox-id")
        assert isinstance(sandbox, AsyncSandbox)
        assert sandbox.id == "test-sandbox-id"


class TestAsyncDaytonaList:
    @pytest.mark.asyncio
    async def test_list_invalid_page_raises(self, env_with_api_key):
        d = _make_async_daytona()
        with pytest.raises(DaytonaError, match="page must be a positive integer"):
            await d.list(page=0)

    @pytest.mark.asyncio
    async def test_list_invalid_limit_raises(self, env_with_api_key):
        d = _make_async_daytona()
        with pytest.raises(DaytonaError, match="limit must be a positive integer"):
            await d.list(limit=0)


class TestAsyncDaytonaGetCodeToolbox:
    def test_unsupported_language_raises(self, env_with_api_key):
        d = _make_async_daytona()
        with pytest.raises(DaytonaError, match="Unsupported language"):
            d._get_code_toolbox("ruby")

    def test_none_defaults_to_python(self, env_with_api_key):
        from daytona.code_toolbox.sandbox_python_code_toolbox import SandboxPythonCodeToolbox

        d = _make_async_daytona()
        tb = d._get_code_toolbox(None)
        assert isinstance(tb, SandboxPythonCodeToolbox)
