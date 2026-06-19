# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0
"""
End-to-end tests for the async SDK (``AsyncDaytona``).

Mirrors ``test_e2e.py`` but exercises the asyncio code path.  Run with::

    DAYTONA_API_KEY=dtn_... DAYTONA_API_URL=https://app.daytona.io/api \\
        pytest tests/test_async_e2e.py -m e2e

The suite intentionally starts small (lifecycle / list / get / delete plus the
connection-resilience smoke test that previously lived in
``test_conn_resilience.py``).  Add more cases as new async surface area is
introduced or when a regression needs a guard.
"""
from __future__ import annotations

import asyncio
import os
import time
import uuid
from collections import Counter
from collections.abc import AsyncIterator

import pytest
import pytest_asyncio

from daytona import AsyncDaytona, CreateSandboxFromSnapshotParams, ListSandboxesQuery
from daytona.common.errors import DaytonaConnectionError, DaytonaError, DaytonaNotFoundError

if not os.getenv("DAYTONA_API_KEY"):
    raise RuntimeError("DAYTONA_API_KEY environment variable is required for E2E tests")

# Module-scoped loop is opt-in here (vs. the suite-wide default of function-scoped)
# so the module-scoped ``async_daytona_client`` / ``async_sandbox`` fixtures
# below can outlive a single test function without ``RuntimeError: Session is
# closed``.  The fixture's ``loop_scope`` must match the test marker's.
pytestmark = [pytest.mark.e2e, pytest.mark.asyncio(loop_scope="module")]


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------


@pytest_asyncio.fixture(loop_scope="module", scope="module")
async def async_daytona_client() -> AsyncIterator[AsyncDaytona]:
    async with AsyncDaytona() as daytona:
        yield daytona


@pytest_asyncio.fixture(loop_scope="module", scope="module")
async def async_sandbox(async_daytona_client: AsyncDaytona):
    params = CreateSandboxFromSnapshotParams(language="python")
    sb = await async_daytona_client.create(params, timeout=120)
    try:
        yield sb
    finally:
        try:
            await async_daytona_client.delete(sb)
        except Exception:
            pass


# ===========================================================================
# Sandbox Lifecycle
# ===========================================================================


async def test_async_sandbox_has_valid_id(async_sandbox):
    assert async_sandbox.id, "Sandbox should have a non-empty ID"


async def test_async_sandbox_has_valid_name(async_sandbox):
    assert async_sandbox.name, "Sandbox should have a non-empty name"


async def test_async_sandbox_state_is_started(async_sandbox):
    state = str(getattr(async_sandbox.state, "value", async_sandbox.state)).lower()
    assert state == "started", f"Expected 'started', got {state!r}"


async def test_async_sandbox_has_resource_properties(async_sandbox):
    assert async_sandbox.cpu > 0, f"Expected cpu > 0, got {async_sandbox.cpu}"
    assert async_sandbox.memory > 0, f"Expected memory > 0, got {async_sandbox.memory}"
    assert async_sandbox.disk > 0, f"Expected disk > 0, got {async_sandbox.disk}"


async def test_async_sandbox_refresh_data_preserves_id(async_sandbox):
    old_id = async_sandbox.id
    await async_sandbox.refresh_data()
    assert async_sandbox.id == old_id
    assert async_sandbox.state is not None


# ===========================================================================
# AsyncDaytona Client Operations
# ===========================================================================


async def test_async_get_sandbox_by_id(async_daytona_client, async_sandbox):
    fetched = await async_daytona_client.get(async_sandbox.id)
    assert fetched.id == async_sandbox.id
    assert fetched.name == async_sandbox.name


async def test_async_list_sandboxes_contains_created(async_daytona_client, async_sandbox):
    sandboxes = [s async for s in async_daytona_client.list()]
    assert len(sandboxes) > 0
    assert any(
        s.id == async_sandbox.id for s in sandboxes
    ), f"Expected created sandbox {async_sandbox.id} to appear in list"


async def test_async_list_with_pagination(async_daytona_client, async_sandbox):
    yielded = 0
    async for _ in async_daytona_client.list(ListSandboxesQuery(limit=1)):
        yielded += 1
        if yielded >= 1:
            break
    assert yielded >= 1


async def test_async_get_unknown_sandbox_raises_not_found(async_daytona_client):
    name = f"async-e2e-missing-{uuid.uuid4().hex[:12]}"
    with pytest.raises(DaytonaNotFoundError):
        await async_daytona_client.get(name)


# ===========================================================================
# Connection Resilience
# ===========================================================================
#
# Hammers ``daytona.get()`` with names that do not exist and verifies that no
# transient connection errors leak through as ``DaytonaConnectionError`` or
# the generic ``DaytonaError``.  Guards against regressions of the retry
# wrapper installed by ``SharedAiohttpSession``.
#
# Tune at runtime with ``CONN_TEST_CONCURRENCY`` (default 50) and
# ``CONN_TEST_ROUNDS`` (default 200).

_CONCURRENCY = int(os.environ.get("CONN_TEST_CONCURRENCY", "50"))
_ROUNDS = int(os.environ.get("CONN_TEST_ROUNDS", "200"))


async def _get_nonexistent(daytona: AsyncDaytona, sem: asyncio.Semaphore) -> str:
    async with sem:
        name = f"conn-test-{uuid.uuid4().hex[:12]}"
        try:
            await daytona.get(name)
            return "unexpected_found"
        except DaytonaNotFoundError:
            return "not_found"
        except DaytonaConnectionError as e:
            return f"conn_error:{type(e).__name__}:{e}"
        except DaytonaError as e:
            return f"daytona_error:{type(e).__name__}:{e}"
        except Exception as e:  # pragma: no cover - last-resort catch-all
            return f"other:{type(e).__name__}:{e}"


async def test_async_concurrent_get_no_connection_errors(async_daytona_client):
    sem = asyncio.Semaphore(_CONCURRENCY)

    t0 = time.monotonic()
    outcomes = await asyncio.gather(*[_get_nonexistent(async_daytona_client, sem) for _ in range(_ROUNDS)])
    elapsed = time.monotonic() - t0

    results: Counter[str] = Counter(o.split(":", 1)[0] for o in outcomes)
    print(f"\n--- Async connection resilience ({_ROUNDS} requests, {_CONCURRENCY} concurrency, {elapsed:.1f}s) ---")
    for k, v in results.most_common():
        print(f"  {k}: {v}")

    leaked = results.get("conn_error", 0) + results.get("daytona_error", 0) + results.get("other", 0)
    assert leaked == 0, f"{leaked} connection/transport errors leaked through. Full results: {dict(results)}"


# ===========================================================================
# Signed URL Operations
# ===========================================================================

ASYNC_SIGNED_URL_DIR = "e2e-async-signed-url-test"


async def test_async_signed_url_setup(async_sandbox):
    try:
        await async_sandbox.fs.delete_file(ASYNC_SIGNED_URL_DIR, recursive=True)
    except Exception:
        pass
    await async_sandbox.fs.create_folder(ASYNC_SIGNED_URL_DIR, "755")
    await async_sandbox.fs.upload_file(b"async signed url content", f"{ASYNC_SIGNED_URL_DIR}/download-test.txt")


async def test_async_download_url_returns_signed_url(async_sandbox):
    url = await async_sandbox.download_url(f"{ASYNC_SIGNED_URL_DIR}/download-test.txt")
    assert isinstance(url, str) and len(url) > 0, "download_url should return a non-empty string"
    assert "signature=" in url, f"URL should contain signature param: {url}"
    assert "expires=" in url, f"URL should contain expires param: {url}"


async def test_async_download_url_serves_correct_content(async_sandbox):
    import aiohttp

    url = await async_sandbox.download_url(f"{ASYNC_SIGNED_URL_DIR}/download-test.txt")
    async with aiohttp.ClientSession() as session:
        async with session.get(url) as resp:
            assert resp.status == 200, f"Expected 200, got {resp.status}"
            body = await resp.read()
            assert body == b"async signed url content", f"Expected exact content, got {body!r}"


async def test_async_upload_url_accepts_file(async_sandbox):
    import aiohttp

    url = await async_sandbox.upload_url(f"{ASYNC_SIGNED_URL_DIR}/uploaded-via-url.txt")
    assert isinstance(url, str) and "signature=" in url, "upload_url should return a signed URL"

    form = aiohttp.FormData()
    form.add_field("file", b"async uploaded via signed url", filename="uploaded-via-url.txt")
    async with aiohttp.ClientSession() as session:
        async with session.post(url, data=form) as resp:
            assert resp.status == 200, f"POST to upload URL failed: {resp.status}"

    content = await async_sandbox.fs.download_file(f"{ASYNC_SIGNED_URL_DIR}/uploaded-via-url.txt")
    assert content == b"async uploaded via signed url", f"Uploaded content mismatch: {content!r}"


async def test_async_rotate_signing_key_and_new_urls_work(async_sandbox):
    import aiohttp

    await async_sandbox.rotate_signing_key()
    url = await async_sandbox.download_url(f"{ASYNC_SIGNED_URL_DIR}/download-test.txt")
    async with aiohttp.ClientSession() as session:
        async with session.get(url) as resp:
            assert resp.status == 200, f"New URL after rotation should work, got {resp.status}"
            body = await resp.read()
            assert body == b"async signed url content"
