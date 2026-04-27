# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""Shared fixtures and mock setup for Daytona Python SDK tests."""

from __future__ import annotations

import sys
from pathlib import Path
from typing import Any, cast
from unittest.mock import AsyncMock, MagicMock

import pytest

from daytona_api_client import Sandbox as SyncSandboxDto
from daytona_api_client import SandboxState

SDK_SRC = Path(__file__).resolve().parents[1] / "src"
if str(SDK_SRC) not in sys.path:
    sys.path.insert(0, str(SDK_SRC))


def make_sandbox_dto(
    sandbox_id: str = "test-sandbox-id",
    name: str = "test-sandbox",
    state: SandboxState = SandboxState.STARTED,
    organization_id: str = "test-org-id",
    user: str = "daytona",
    target: str = "us",
    cpu: int = 4,
    gpu: int = 0,
    memory: int = 8,
    disk: int = 30,
    **kwargs: Any,
) -> SyncSandboxDto:
    """Create a mock SandboxDto with sensible defaults."""
    return SyncSandboxDto(
        id=cast(str, kwargs.get("id", sandbox_id)),
        name=cast(str, kwargs.get("name", name)),
        organization_id=cast(str, kwargs.get("organization_id", organization_id)),
        snapshot=cast(str | None, kwargs.get("snapshot", "default-snapshot")),
        user=cast(str, kwargs.get("user", user)),
        env=cast(dict[str, str], kwargs.get("env", {"PATH": "/usr/bin"})),
        labels=cast(dict[str, str], kwargs.get("labels", {"code-toolbox-language": "python"})),
        public=cast(bool, kwargs.get("public", False)),
        target=cast(str, kwargs.get("target", target)),
        cpu=cast(int, kwargs.get("cpu", cpu)),
        gpu=cast(int, kwargs.get("gpu", gpu)),
        memory=cast(int, kwargs.get("memory", memory)),
        disk=cast(int, kwargs.get("disk", disk)),
        state=cast(SandboxState, kwargs.get("state", state)),
        error_reason=cast(str | None, kwargs.get("error_reason", None)),
        recoverable=cast(bool | None, kwargs.get("recoverable", None)),
        backup_state=cast(str | None, kwargs.get("backup_state", None)),
        backup_created_at=cast(str | None, kwargs.get("backup_created_at", None)),
        auto_stop_interval=cast(int | None, kwargs.get("auto_stop_interval", 15)),
        auto_archive_interval=cast(int | None, kwargs.get("auto_archive_interval", 10080)),
        auto_delete_interval=cast(int | None, kwargs.get("auto_delete_interval", -1)),
        volumes=cast(list[Any] | None, kwargs.get("volumes", None)),
        build_info=cast(Any, kwargs.get("build_info", None)),
        created_at=cast(str | None, kwargs.get("created_at", "2025-01-01T00:00:00Z")),
        updated_at=cast(str | None, kwargs.get("updated_at", "2025-01-01T00:00:00Z")),
        network_block_all=cast(bool, kwargs.get("network_block_all", False)),
        network_allow_list=cast(str | None, kwargs.get("network_allow_list", None)),
        toolbox_proxy_url=cast(str, kwargs.get("toolbox_proxy_url", "http://localhost:2280")),
    )


@pytest.fixture
def sandbox_dto():
    """Return a started SandboxDto."""
    return make_sandbox_dto()


@pytest.fixture
def stopped_sandbox_dto():
    """Return a stopped SandboxDto."""
    return make_sandbox_dto(state=SandboxState.STOPPED)


@pytest.fixture
def mock_toolbox_api_client():
    """Return a MagicMock for the sync ToolboxApiClient."""
    client = MagicMock()
    client.default_headers = {"Authorization": "Bearer test-key"}
    return client


@pytest.fixture
def mock_async_toolbox_api_client():
    """Return a MagicMock for the async ToolboxApiClient."""
    client = MagicMock()
    client.default_headers = {"Authorization": "Bearer test-key"}
    return client


@pytest.fixture
def mock_sandbox_api():
    """Return a MagicMock for the sync SandboxApi."""
    api = MagicMock()
    return api


@pytest.fixture
def mock_async_sandbox_api():
    """Return an AsyncMock for the async SandboxApi."""
    api = AsyncMock()
    return api


@pytest.fixture
def mock_code_toolbox():
    """Return a MagicMock for the SandboxCodeToolbox."""
    toolbox = MagicMock()
    toolbox.get_run_command.return_value = 'python3 -c "print(1)"'
    return toolbox


@pytest.fixture
def env_with_api_key(monkeypatch):
    """Set standard env vars for Daytona client initialization."""
    monkeypatch.setenv("DAYTONA_API_KEY", "test-api-key-123")
    monkeypatch.setenv("DAYTONA_API_URL", "https://test.daytona.io/api")
    monkeypatch.setenv("DAYTONA_TARGET", "us")


@pytest.fixture
def env_with_jwt(monkeypatch):
    """Set JWT-based env vars for Daytona client initialization."""
    monkeypatch.setenv("DAYTONA_JWT_TOKEN", "test-jwt-token-123")
    monkeypatch.setenv("DAYTONA_ORGANIZATION_ID", "test-org-id")
    monkeypatch.setenv("DAYTONA_API_URL", "https://test.daytona.io/api")
    monkeypatch.setenv("DAYTONA_TARGET", "us")
