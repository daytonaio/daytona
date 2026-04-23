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

        with pytest.raises(
            DaytonaAuthenticationError, match="Authentication credentials not found. Set DAYTONA_API_KEY"
        ):
            AsyncDaytona()

    def test_jwt_without_organization_id_raises(self):
        with pytest.raises(DaytonaAuthenticationError, match="DAYTONA_ORGANIZATION_ID is required"):
            _make_async_daytona(DaytonaConfig(jwt_token="jwt", api_url="https://api.test.io", target="us"))


class TestAsyncDaytonaContextManager:
    @pytest.mark.asyncio
    async def test_context_manager_closes_clients(self, env_with_api_key):
        from daytona._async.daytona import AsyncDaytona
        from daytona_api_client_async import Configuration

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


class TestAsyncDaytonaSharedSession:
    """Guards the SharedAiohttpSession contract: a single session shared across both
    rest_clients regardless of which one fires first."""

    @pytest.mark.asyncio
    async def test_first_request_propagates_session_to_other_rest_client(self, monkeypatch):
        """The coordinator pre-creates one shared session and propagates it to every
        attached rest_client before their lazy-create branch ever runs.  This is what
        lets the SDK's runtime ``happy_eyeballs_delay`` decision win over the value
        baked into the generated ``rest.py``.
        """
        import aiohttp

        from daytona.internal import shared_session as ss
        from daytona.internal.shared_session import SharedAiohttpSession

        fake_session = MagicMock(spec=aiohttp.ClientSession)
        fake_session.closed = False

        def fake_ensure(self: SharedAiohttpSession, rest_client: object) -> None:
            if self._session is None:
                self._session = fake_session
                for client in self._rest_clients:
                    client.pool_manager = fake_session

        monkeypatch.setattr(ss.SharedAiohttpSession, "_ensure_session", fake_ensure)

        rc_main = MagicMock(spec=["request", "pool_manager"])
        rc_main.pool_manager = None
        rc_tb = MagicMock(spec=["request", "pool_manager"])
        rc_tb.pool_manager = None

        async def stub_request(*_args, **_kwargs):
            return None

        rc_main.request = stub_request

        coordinator = SharedAiohttpSession()
        coordinator.attach(rc_main)
        coordinator.attach(rc_tb)

        await rc_main.request("GET", "/anything")

        assert coordinator.session is fake_session
        assert rc_main.pool_manager is fake_session
        assert rc_tb.pool_manager is fake_session

    @pytest.mark.asyncio
    async def test_wrapper_does_not_override_callers_request_timeout(self):
        """Wrapper passes ``_request_timeout`` through verbatim and never injects a
        default — calls without one fall through to aiohttp's session timeout."""
        from daytona.internal.shared_session import SharedAiohttpSession

        seen_kwargs: dict = {}
        rc = MagicMock(spec=["request", "pool_manager"])
        rc.pool_manager = None

        async def stub_request(*_args, **kwargs):
            seen_kwargs.update(kwargs)
            return None

        rc.request = stub_request

        coordinator = SharedAiohttpSession()
        coordinator.attach(rc)

        # No timeout supplied — wrapper must NOT inject one.
        await rc.request("GET", "/anything")
        assert "_request_timeout" not in seen_kwargs or seen_kwargs.get("_request_timeout") is None

        # Caller-supplied timeout — wrapper must pass it through verbatim.
        sentinel = object()
        await rc.request("GET", "/anything", _request_timeout=sentinel)
        assert seen_kwargs.get("_request_timeout") is sentinel

    @pytest.mark.asyncio
    async def test_toolbox_first_propagates_to_main(self):
        """If the toolbox rest_client fires first, adoption still propagates to main."""
        import aiohttp

        from daytona.internal.shared_session import SharedAiohttpSession

        rc_main = MagicMock(spec=["request", "pool_manager"])
        rc_main.pool_manager = None
        rc_tb = MagicMock(spec=["request", "pool_manager"])
        rc_tb.pool_manager = None

        coordinator = SharedAiohttpSession()
        coordinator.attach(rc_main)
        coordinator.attach(rc_tb)

        fake_session = MagicMock(spec=aiohttp.ClientSession)
        fake_session.closed = False
        rc_tb.pool_manager = fake_session
        coordinator._adopt(fake_session)

        assert rc_main.pool_manager is fake_session
        assert rc_tb.pool_manager is fake_session

    @pytest.mark.asyncio
    async def test_first_writer_wins_under_race(self):
        """First adopted session sticks while it's open; later candidates are ignored."""
        import aiohttp

        from daytona.internal.shared_session import SharedAiohttpSession

        coordinator = SharedAiohttpSession()
        first = MagicMock(spec=aiohttp.ClientSession)
        first.closed = False
        second = MagicMock(spec=aiohttp.ClientSession)
        second.closed = False

        coordinator._adopt(first)
        coordinator._adopt(second)

        assert coordinator.session is first

    @pytest.mark.asyncio
    async def test_attach_rejects_invalid_rest_client(self):
        """Attach fails loudly at init if the rest_client surface changes."""
        from daytona.internal.shared_session import SharedAiohttpSession

        coordinator = SharedAiohttpSession()
        with pytest.raises(RuntimeError, match="rest_client API surface changed"):
            coordinator.attach(object())

    @pytest.mark.asyncio
    @pytest.mark.parametrize("method", ["GET", "HEAD", "OPTIONS", "TRACE", "PUT", "DELETE"])
    async def test_idempotent_methods_retry_on_server_disconnect(self, method, monkeypatch):
        """ServerDisconnectedError on an idempotent method is retried."""
        import aiohttp

        from daytona.internal import shared_session as ss
        from daytona.internal.shared_session import SharedAiohttpSession

        monkeypatch.setattr(ss.SharedAiohttpSession, "_ensure_session", lambda self, rc: None)

        rc = MagicMock(spec=["request", "pool_manager"])
        rc.pool_manager = MagicMock()
        rc.pool_manager.closed = False

        calls = 0
        sentinel = object()

        async def stub_request(*_args, **_kwargs):
            nonlocal calls
            calls += 1
            if calls == 1:
                raise aiohttp.ServerDisconnectedError("simulated stale keep-alive")
            return sentinel

        rc.request = stub_request

        coordinator = SharedAiohttpSession()
        coordinator.attach(rc)

        result = await rc.request(method, "/anything")

        assert result is sentinel
        assert calls == 2

    @pytest.mark.asyncio
    @pytest.mark.parametrize("method", ["POST", "PATCH"])
    async def test_non_idempotent_methods_do_not_retry_on_server_disconnect(self, method, monkeypatch):
        """ServerDisconnectedError on a non-idempotent method propagates immediately."""
        import aiohttp

        from daytona.internal import shared_session as ss
        from daytona.internal.shared_session import SharedAiohttpSession

        monkeypatch.setattr(ss.SharedAiohttpSession, "_ensure_session", lambda self, rc: None)

        rc = MagicMock(spec=["request", "pool_manager"])
        rc.pool_manager = MagicMock()
        rc.pool_manager.closed = False

        calls = 0

        async def stub_request(*_args, **_kwargs):
            nonlocal calls
            calls += 1
            raise aiohttp.ServerDisconnectedError("simulated stale keep-alive")

        rc.request = stub_request

        coordinator = SharedAiohttpSession()
        coordinator.attach(rc)

        with pytest.raises(aiohttp.ServerDisconnectedError):
            await rc.request(method, "/anything")

        assert calls == 1  # no retries

    @pytest.mark.asyncio
    @pytest.mark.parametrize("method", ["POST", "PATCH", "PUT", "DELETE", "GET"])
    async def test_connector_error_retries_on_any_method(self, method, monkeypatch):
        """ClientConnectorError (TCP connect() failed) is safe to retry on any
        method since the request was never sent."""
        import aiohttp
        from aiohttp.client_reqrep import ConnectionKey

        from daytona.internal import shared_session as ss
        from daytona.internal.shared_session import SharedAiohttpSession

        monkeypatch.setattr(ss.SharedAiohttpSession, "_ensure_session", lambda self, rc: None)

        rc = MagicMock(spec=["request", "pool_manager"])
        rc.pool_manager = MagicMock()
        rc.pool_manager.closed = False

        conn_key = ConnectionKey(
            host="example.invalid",
            port=443,
            is_ssl=True,
            ssl=True,
            proxy=None,
            proxy_auth=None,
            proxy_headers_hash=None,
        )

        calls = 0
        sentinel = object()

        async def stub_request(*_args, **_kwargs):
            nonlocal calls
            calls += 1
            if calls == 1:
                raise aiohttp.ClientConnectorError(conn_key, OSError("simulated connect failure"))
            return sentinel

        rc.request = stub_request

        coordinator = SharedAiohttpSession()
        coordinator.attach(rc)

        result = await rc.request(method, "/anything")

        assert result is sentinel
        assert calls == 2

    @pytest.mark.asyncio
    async def test_retry_does_not_observe_first_attempt_header_mutation(self, monkeypatch):
        """Mutable kwargs are shallow-copied before each attempt so a retry
        cannot see headers the first attempt added or deleted.

        The generated rest_client mutates ``headers`` in place (adds
        ``Content-Type`` for JSON, deletes it for multipart).  Without a
        per-attempt copy, attempt 2 would re-use the mutated dict.
        """
        import aiohttp

        from daytona.internal import shared_session as ss
        from daytona.internal.shared_session import SharedAiohttpSession

        monkeypatch.setattr(ss.SharedAiohttpSession, "_ensure_session", lambda self, rc: None)

        rc = MagicMock(spec=["request", "pool_manager"])
        rc.pool_manager = MagicMock()
        rc.pool_manager.closed = False

        calls = 0
        seen_headers: list[dict[str, str]] = []
        sentinel = object()

        async def stub_request(*_args, **kwargs):
            nonlocal calls
            calls += 1
            headers = kwargs.get("headers")
            # Snapshot the headers dict we received this attempt.
            seen_headers.append(dict(headers) if isinstance(headers, dict) else {})
            # Mutate the dict in place, mirroring the generated client's behavior.
            if isinstance(headers, dict):
                headers["Content-Type"] = "application/json"
            if calls == 1:
                raise aiohttp.ServerDisconnectedError("simulated")
            return sentinel

        rc.request = stub_request

        coordinator = SharedAiohttpSession()
        coordinator.attach(rc)

        caller_headers = {"Authorization": "Bearer xxx"}
        result = await rc.request("GET", "/anything", headers=caller_headers)

        assert result is sentinel
        assert calls == 2
        # Second attempt must NOT see the Content-Type that the first attempt set.
        assert "Content-Type" not in seen_headers[1]
        # The caller's original dict must NOT have been mutated either.
        assert "Content-Type" not in caller_headers

    def test_retry_backoff_has_jitter_and_grows_with_attempt(self):
        """Backoff is ``base * attempt + uniform(0, jitter)`` — must not be
        a constant (we'd lock-step concurrent retries) and must grow with
        attempt within the jitter window."""
        from daytona.internal.shared_session import (
            _RETRY_BACKOFF_BASE_S,
            _RETRY_BACKOFF_JITTER_S,
            _retry_backoff_seconds,
        )

        # Bounds for any given attempt.
        for attempt in (1, 2):
            samples = [_retry_backoff_seconds(attempt) for _ in range(200)]
            lo = _RETRY_BACKOFF_BASE_S * attempt
            hi = lo + _RETRY_BACKOFF_JITTER_S
            assert all(
                lo <= s <= hi for s in samples
            ), f"attempt {attempt}: samples out of [{lo}, {hi}] -> {min(samples)}..{max(samples)}"
            # Jitter must actually vary — refuse identical samples.
            assert len(set(samples)) > 1, f"attempt {attempt}: backoff is constant"


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

    def test_list_returns_async_iterator(self, env_with_api_key):
        import inspect

        from daytona._async.daytona import AsyncDaytona

        # ``list`` is an async generator function — calling it returns an
        # async iterator without performing any network I/O.
        assert inspect.isasyncgenfunction(AsyncDaytona.list)

    @pytest.mark.asyncio
    async def test_list_serializes_labels(self, env_with_api_key, sandbox_dto):
        from daytona import ListSandboxesQuery

        response = MagicMock(items=[sandbox_dto], next_cursor=None)
        daytona = _make_async_daytona()
        daytona._sandbox_api.list_sandboxes = AsyncMock(return_value=response)

        sandboxes = []
        async for sb in daytona.list(ListSandboxesQuery(labels={"project": "test"}, limit=10)):
            sandboxes.append(sb)

        assert len(sandboxes) == 1
        kwargs = daytona._sandbox_api.list_sandboxes.call_args.kwargs
        assert json.loads(kwargs["labels"]) == {"project": "test"}
        assert kwargs["limit"] == 10
        assert kwargs["cursor"] is None

    @pytest.mark.asyncio
    async def test_list_paginates_via_cursor(self, env_with_api_key, sandbox_dto):
        page1 = MagicMock(items=[sandbox_dto, sandbox_dto], next_cursor="cursor-2")
        page2 = MagicMock(items=[sandbox_dto], next_cursor=None)

        daytona = _make_async_daytona()
        daytona._sandbox_api.list_sandboxes = AsyncMock(side_effect=[page1, page2])

        collected = []
        async for sb in daytona.list():
            collected.append(sb)

        assert len(collected) == 3
        assert daytona._sandbox_api.list_sandboxes.call_count == 2
        second_call_kwargs = daytona._sandbox_api.list_sandboxes.call_args_list[1].kwargs
        assert second_call_kwargs["cursor"] == "cursor-2"

    @pytest.mark.asyncio
    async def test_list_early_termination_stops_fetching(self, env_with_api_key, sandbox_dto):
        page1 = MagicMock(items=[sandbox_dto, sandbox_dto], next_cursor="cursor-2")
        page2 = MagicMock(items=[sandbox_dto], next_cursor=None)

        daytona = _make_async_daytona()
        daytona._sandbox_api.list_sandboxes = AsyncMock(side_effect=[page1, page2])

        async for _ in daytona.list():
            break

        assert daytona._sandbox_api.list_sandboxes.call_count == 1


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


class TestResolveHappyEyeballsDelay:
    @pytest.mark.parametrize("raw", [None, "", "   "])
    def test_unset_returns_missing_sentinel(self, raw):
        from daytona._async.daytona import _MISSING_HAPPY_EYEBALLS_DELAY, _resolve_happy_eyeballs_delay

        assert _resolve_happy_eyeballs_delay(raw) is _MISSING_HAPPY_EYEBALLS_DELAY

    @pytest.mark.parametrize("raw", ["none", "None", "NONE", "  none  "])
    def test_none_disables(self, raw):
        from daytona._async.daytona import _resolve_happy_eyeballs_delay

        assert _resolve_happy_eyeballs_delay(raw) is None

    @pytest.mark.parametrize(
        "raw,expected",
        [
            ("0", 0.0),
            ("0.5", 0.5),
            ("1", 1.0),
            ("10.5", 10.5),
        ],
    )
    def test_valid_floats(self, raw, expected):
        from daytona._async.daytona import _resolve_happy_eyeballs_delay

        assert _resolve_happy_eyeballs_delay(raw) == expected

    @pytest.mark.parametrize(
        "raw",
        [
            "abc",
            "-1",
            "-0.5",
            "1.0.0",
            "true",
            "false",
            # float() accepts these without raising; the resolver must catch them.
            "nan",
            "NaN",
            "inf",
            "Infinity",
            "-inf",
        ],
    )
    def test_invalid_values_raise(self, raw):
        from daytona._async.daytona import _resolve_happy_eyeballs_delay

        with pytest.raises(
            DaytonaValidationError,
            match="DAYTONA_HAPPY_EYEBALLS_DELAY must be a finite non-negative float or 'none'",
        ):
            _resolve_happy_eyeballs_delay(raw)
