# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

"""SDK contract tests for `daytona.session.*`.

Each test starts as ``pytest.skip(...)`` until the corresponding implementation lands.
Once unskipped, runs against a real Daytona API configured via ``DAYTONA_API_URL`` +
``DAYTONA_API_KEY`` (same env contract as the existing ``test_e2e.py``).

Mirrored across ``_sync`` and ``_async``. Smaller scope than the Go suite — only
validates that the SDK correctly surfaces the API contract: ergonomics, typed-error
mapping, callback semantics.
"""

from __future__ import annotations

import os

import pytest

pytestmark = [pytest.mark.e2e]


_ENV_OK = bool(os.getenv("DAYTONA_API_URL")) and bool(os.getenv("DAYTONA_API_KEY"))


def _skip_until(todo_id: str) -> None:
    pytest.skip(f"not yet implemented: {todo_id}")


# ---------------------------------------------------------------------------
# Sync
# ---------------------------------------------------------------------------


@pytest.mark.skipif(not _ENV_OK, reason="DAYTONA_API_URL / DAYTONA_API_KEY not set")
class TestSessionSync:
    def test_run_returns_stdout_and_duration(self) -> None:
        _skip_until("ts-sdk-session / python-sdk-session")
        # from daytona import Daytona
        # daytona = Daytona()
        # result = daytona.session.run("print(1)")
        # assert result.stdout == "1\n"
        # assert result.duration_ms > 0
        # assert result.error is None
        # assert result.displays == []

    def test_create_session_then_run_with_context(self) -> None:
        _skip_until("python-sdk-session")
        # from daytona import Daytona
        # daytona = Daytona()
        # ctx = daytona.session.create_session(template="python-default", language="python")
        # try:
        #     daytona.session.run("x = 42", context=ctx)
        #     assert daytona.session.run("print(x)", context=ctx).stdout == "42\n"
        # finally:
        #     daytona.session.delete_session(ctx)

    def test_context_invalidated_error_typed(self) -> None:
        _skip_until("python-sdk-session")
        # from daytona import Daytona
        # from daytona.common.errors import SessionInvalidatedError
        # daytona = Daytona()
        # ctx = daytona.session.create_session(template="python-default", language="python")
        # # invalidate via test infra ...
        # with pytest.raises(SessionInvalidatedError) as exc:
        #     daytona.session.run("print(1)", context=ctx)
        # assert exc.value.session_id == ctx.id
        # assert exc.value.invalidated_at is not None

    def test_context_expired_error_typed(self) -> None:
        _skip_until("python-sdk-session")
        # from daytona.common.errors import SessionExpiredError
        # ...
        # with pytest.raises(SessionExpiredError) as exc:
        #     daytona.session.run("print(1)", context=ctx)
        # assert exc.value.reason in ("idle", "absolute")

    def test_list_sessions_parses_expires_at(self) -> None:
        _skip_until("python-sdk-session")
        # from datetime import datetime, timezone
        # daytona = Daytona()
        # ctx = daytona.session.create_session(template="python-default", language="python")
        # try:
        #     listed = daytona.session.list_sessions()
        #     match = next(c for c in listed if c.id == ctx.id)
        #     assert isinstance(match.expires_at, datetime)
        #     assert match.expires_at.tzinfo is not None
        # finally:
        #     daytona.session.delete_session(ctx)


# ---------------------------------------------------------------------------
# Async
# ---------------------------------------------------------------------------


@pytest.mark.asyncio
@pytest.mark.skipif(not _ENV_OK, reason="DAYTONA_API_URL / DAYTONA_API_KEY not set")
class TestSessionAsync:
    async def test_run_async(self) -> None:
        _skip_until("python-sdk-session")
        # from daytona import AsyncDaytona
        # async with AsyncDaytona() as daytona:
        #     result = await daytona.session.run("print(1)")
        #     assert result.stdout == "1\n"

    async def test_run_stream_callbacks_async(self) -> None:
        _skip_until("python-sdk-session")
        # collected = []
        # async with AsyncDaytona() as daytona:
        #     final = await daytona.session.run_stream(
        #         "print('a'); print('b')",
        #         language="python",
        #         on_stdout=lambda c: collected.append(("stdout", c)),
        #     )
        # assert any(c == "a\n" for kind, c in collected if kind == "stdout")
        # assert final.stdout == "a\nb\n"
