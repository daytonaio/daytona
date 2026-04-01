from __future__ import annotations

from unittest.mock import MagicMock, patch

import pytest

from daytona.common.daytona import (
    CodeLanguage,
    CreateSandboxFromImageParams,
    CreateSandboxFromSnapshotParams,
    DaytonaConfig,
)
from daytona.common.errors import DaytonaError
from daytona.common.sandbox import Resources

SYNC_MODULE = "daytona._sync.daytona"


def _make_daytona(config=None):
    from daytona_api_client import Configuration

    from daytona._sync.daytona import Daytona

    with (
        patch(f"{SYNC_MODULE}.ApiClient") as mock_api_cls,
        patch(f"{SYNC_MODULE}.ToolboxApiClient") as mock_toolbox_cls,
    ):
        mock_api_instance = MagicMock()
        mock_api_instance.configuration = Configuration(host="https://test.daytona.io/api")
        mock_api_instance.default_headers = {}
        mock_api_instance.user_agent = ""
        mock_api_cls.return_value = mock_api_instance

        mock_toolbox_instance = MagicMock()
        mock_toolbox_instance.default_headers = {}
        mock_toolbox_cls.return_value = mock_toolbox_instance

        d = Daytona(config)
    return d


class TestDaytonaInit:
    def test_init_with_config(self):
        config = DaytonaConfig(
            api_key="test-key",
            api_url="https://api.test.io",
            target="us",
        )
        d = _make_daytona(config)
        assert d._api_key == "test-key"
        assert d._api_url == "https://api.test.io"
        assert d._target == "us"

    def test_init_with_env_vars(self, env_with_api_key):
        d = _make_daytona()
        assert d._api_key == "test-api-key-123"
        assert d._api_url == "https://test.daytona.io/api"
        assert d._target == "us"

    def test_init_with_jwt_token(self, env_with_jwt):
        d = _make_daytona()
        assert d._jwt_token == "test-jwt-token-123"
        assert d._organization_id == "test-org-id"

    @patch("daytona._utils.env.dotenv_values", return_value={})
    def test_init_jwt_without_org_id_raises(self, mock_dotenv, monkeypatch):
        monkeypatch.setenv("DAYTONA_JWT_TOKEN", "jwt-token")
        monkeypatch.setenv("DAYTONA_API_URL", "https://api.test.io")
        monkeypatch.setenv("DAYTONA_TARGET", "us")
        monkeypatch.delenv("DAYTONA_API_KEY", raising=False)
        monkeypatch.delenv("DAYTONA_ORGANIZATION_ID", raising=False)

        with pytest.raises(DaytonaError, match="Organization ID is required"):
            _make_daytona()

    @patch("daytona._utils.env.dotenv_values", return_value={})
    def test_init_without_credentials_raises(self, mock_dotenv, monkeypatch):
        monkeypatch.delenv("DAYTONA_API_KEY", raising=False)
        monkeypatch.delenv("DAYTONA_JWT_TOKEN", raising=False)
        monkeypatch.delenv("DAYTONA_API_URL", raising=False)
        monkeypatch.delenv("DAYTONA_TARGET", raising=False)
        monkeypatch.delenv("DAYTONA_SERVER_URL", raising=False)
        monkeypatch.delenv("DAYTONA_ORGANIZATION_ID", raising=False)

        from daytona._sync.daytona import Daytona

        with pytest.raises(DaytonaError, match="API key or JWT token is required"):
            Daytona()

    def test_default_api_url(self, monkeypatch):
        monkeypatch.setenv("DAYTONA_API_KEY", "key")
        monkeypatch.delenv("DAYTONA_API_URL", raising=False)
        monkeypatch.delenv("DAYTONA_SERVER_URL", raising=False)
        d = _make_daytona()
        assert d._api_url == "https://app.daytona.io/api"


class TestDaytonaGetCodeToolbox:
    def test_python_toolbox(self, env_with_api_key):
        from daytona.code_toolbox.sandbox_python_code_toolbox import SandboxPythonCodeToolbox

        d = _make_daytona()
        tb = d._get_code_toolbox(CodeLanguage.PYTHON)
        assert isinstance(tb, SandboxPythonCodeToolbox)

    def test_javascript_toolbox(self, env_with_api_key):
        from daytona.code_toolbox.sandbox_js_code_toolbox import SandboxJsCodeToolbox

        d = _make_daytona()
        tb = d._get_code_toolbox(CodeLanguage.JAVASCRIPT)
        assert isinstance(tb, SandboxJsCodeToolbox)

    def test_typescript_toolbox(self, env_with_api_key):
        from daytona.code_toolbox.sandbox_ts_code_toolbox import SandboxTsCodeToolbox

        d = _make_daytona()
        tb = d._get_code_toolbox(CodeLanguage.TYPESCRIPT)
        assert isinstance(tb, SandboxTsCodeToolbox)

    def test_none_defaults_to_python(self, env_with_api_key):
        from daytona.code_toolbox.sandbox_python_code_toolbox import SandboxPythonCodeToolbox

        d = _make_daytona()
        tb = d._get_code_toolbox(None)
        assert isinstance(tb, SandboxPythonCodeToolbox)

    def test_unsupported_language_raises(self, env_with_api_key):
        d = _make_daytona()
        with pytest.raises(DaytonaError, match="Unsupported language"):
            d._get_code_toolbox("ruby")


class TestDaytonaCreateValidation:
    def test_negative_timeout_raises(self, env_with_api_key):
        d = _make_daytona()
        with pytest.raises(DaytonaError, match="Timeout must be a non-negative number"):
            d._create(CreateSandboxFromSnapshotParams(language="python"), timeout=-1)

    def test_negative_auto_stop_raises(self, env_with_api_key):
        d = _make_daytona()
        params = CreateSandboxFromSnapshotParams(language="python", auto_stop_interval=-1)
        with pytest.raises(DaytonaError, match="auto_stop_interval must be a non-negative"):
            d._create(params, timeout=60)

    def test_negative_auto_archive_raises(self, env_with_api_key):
        d = _make_daytona()
        params = CreateSandboxFromSnapshotParams(language="python", auto_archive_interval=-1)
        with pytest.raises(DaytonaError, match="auto_archive_interval must be a non-negative"):
            d._create(params, timeout=60)


class TestDaytonaGet:
    def test_get_empty_id_raises(self, env_with_api_key):
        d = _make_daytona()
        with pytest.raises(DaytonaError, match="sandbox_id_or_name is required"):
            d.get("")

    def test_get_returns_sandbox(self, env_with_api_key, sandbox_dto):
        from daytona._sync.sandbox import Sandbox

        d = _make_daytona()
        d._sandbox_api = MagicMock()
        d._sandbox_api.get_sandbox.return_value = sandbox_dto
        sandbox = d.get("test-sandbox-id")
        assert isinstance(sandbox, Sandbox)
        assert sandbox.id == "test-sandbox-id"


class TestDaytonaList:
    def test_list_invalid_page_raises(self, env_with_api_key):
        d = _make_daytona()
        with pytest.raises(DaytonaError, match="page must be a positive integer"):
            d.list(page=0)

    def test_list_invalid_limit_raises(self, env_with_api_key):
        d = _make_daytona()
        with pytest.raises(DaytonaError, match="limit must be a positive integer"):
            d.list(limit=0)


class TestDaytonaValidateLanguageLabel:
    def test_none_returns_python(self, env_with_api_key):
        d = _make_daytona()
        assert d._validate_language_label(None) == CodeLanguage.PYTHON

    def test_valid_language(self, env_with_api_key):
        d = _make_daytona()
        assert d._validate_language_label("typescript") == CodeLanguage.TYPESCRIPT

    def test_invalid_language_raises(self, env_with_api_key):
        d = _make_daytona()
        with pytest.raises(DaytonaError, match="Invalid code-toolbox-language"):
            d._validate_language_label("ruby")
