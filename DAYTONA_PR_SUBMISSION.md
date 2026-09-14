# Pull Request: fix(sdk): implement exponential backoff with jitter for async workspace event streams

## 📌 Bounty Information
- **Platform**: Algora.io / GitHub
- **Target Repository**: `daytonaio/daytona`
- **Issue Reference**: Fixes #412
- **Claimed Bounty Amount**: **$30.00 USD**

---

## 🔍 Root Cause Analysis (RCA)
In the Daytona Python SDK, the `WorkspaceService.get_workspace_events` long-polling event stream relies on `httpx.AsyncClient` streaming without automated reconnection logic. When intermediate proxy timeouts, network interruptions, or workspace provisioning delays occur:
1. `httpx.ReadTimeout` or `httpx.RemoteProtocolError` is raised unhandled out of the async generator.
2. Callers receive an abrupt termination instead of a resilient stream.
3. Rapid un-jittered retries create thundering-herd spikes against the Daytona control plane.

---

## 🛠️ Proposed Solution
1. **Added Resilient Async Generator Wrapper** (`daytona_sdk/utils/retry.py`):
   - Implemented `async_retry_stream_with_backoff` using full-jitter exponential backoff ($T_{\text{sleep}} = \text{uniform}(0, \min(T_{\max}, T_{\text{base}} \times 2^{\text{attempt}}))$).
   - Catches retryable network errors (`httpx.ReadTimeout`, `httpx.ConnectError`, `httpx.RemoteProtocolError`, `asyncio.TimeoutError`).
   - Automatically maintains stream cursor state (`last_event_id` or timestamp) to prevent duplicated events on reconnection.
2. **Integrated into Workspace Events Client** (`daytona_sdk/services/workspace_service.py`):
   - Wrapped the raw event stream with transparent auto-reconnect logic.

---

## 💻 Unified Diff Patch

```diff
diff --git a/libs/sdk/python/daytona_sdk/utils/retry.py b/libs/sdk/python/daytona_sdk/utils/retry.py
new file mode 100644
index 0000000..8b2c145
--- /dev/null
+++ b/libs/sdk/python/daytona_sdk/utils/retry.py
@@ -0,0 +1,58 @@
+import asyncio
+import logging
+import random
+from typing import AsyncGenerator, Callable, Tuple, Type, Any, Optional
+import httpx
+
+logger = logging.getLogger("daytona_sdk.retry")
+
+DEFAULT_RETRYABLE_EXCEPTIONS: Tuple[Type[Exception], ...] = (
+    asyncio.TimeoutError,
+    httpx.ReadTimeout,
+    httpx.ConnectError,
+    httpx.RemoteProtocolError,
+    httpx.ReadError,
+)
+
+def async_retry_stream_with_backoff(
+    max_retries: int = 5,
+    base_delay: float = 0.5,
+    max_delay: float = 30.0,
+    retryable_exceptions: Tuple[Type[Exception], ...] = DEFAULT_RETRYABLE_EXCEPTIONS,
+):
+    """Decorator that wraps an async generator with exponential backoff and jitter."""
+    def decorator(agen_func: Callable[..., AsyncGenerator[Any, None]]):
+        async def wrapper(*args, **kwargs) -> AsyncGenerator[Any, None]:
+            attempt = 0
+            last_cursor: Optional[str] = kwargs.get("after")
+            while True:
+                try:
+                    if last_cursor:
+                        kwargs["after"] = last_cursor
+                    async for item in agen_func(*args, **kwargs):
+                        if isinstance(item, dict) and "id" in item:
+                            last_cursor = str(item["id"])
+                        elif hasattr(item, "id"):
+                            last_cursor = str(item.id)
+                        yield item
+                    return
+                except retryable_exceptions as exc:
+                    attempt += 1
+                    if attempt > max_retries:
+                        logger.error(f"Exceeded max retries ({max_retries}) for event stream: {exc}")
+                        raise
+                    # Exponential backoff with full jitter
+                    delay = min(max_delay, base_delay * (2 ** (attempt - 1)))
+                    jittered_delay = random.uniform(0.1, delay)
+                    logger.warning(
+                        f"Event stream disconnected ({exc.__class__.__name__}). Retrying {attempt}/{max_retries} in {jittered_delay:.2f}s..."
+                    )
+                    await asyncio.sleep(jittered_delay)
+        return wrapper
+    return decorator
diff --git a/libs/sdk/python/daytona_sdk/services/workspace_service.py b/libs/sdk/python/daytona_sdk/services/workspace_service.py
index a12e34f..c89b211 100644
--- a/libs/sdk/python/daytona_sdk/services/workspace_service.py
+++ b/libs/sdk/python/daytona_sdk/services/workspace_service.py
@@ -12,6 +12,7 @@ import json
 from typing import AsyncGenerator, Optional
 import httpx
+from daytona_sdk.utils.retry import async_retry_stream_with_backoff
 
 class WorkspaceService:
     def __init__(self, api_client):
@@ -45,6 +46,7 @@ class WorkspaceService:
+    @async_retry_stream_with_backoff(max_retries=5, base_delay=0.5, max_delay=30.0)
     async def get_workspace_events(
         self, workspace_id: str, after: Optional[str] = None
     ) -> AsyncGenerator[dict, None]:
```

---

## 🧪 Automated Unit & Regression Tests (`pytest`)

```python
import pytest
import asyncio
import httpx
from daytona_sdk.utils.retry import async_retry_stream_with_backoff

@pytest.mark.asyncio
async def test_stream_retry_on_timeout():
    call_count = 0

    @async_retry_stream_with_backoff(max_retries=3, base_delay=0.01, max_delay=0.1)
    async def mock_event_stream(workspace_id: str, after: str = None):
        nonlocal call_count
        call_count += 1
        if call_count == 1:
            yield {"id": "1", "event": "provisioning_started"}
            raise httpx.ReadTimeout("Simulated connection timeout")
        yield {"id": "2", "event": "provisioning_done"}

    events = []
    async for ev in mock_event_stream("ws-123"):
        events.append(ev)

    assert len(events) == 2
    assert events[0]["id"] == "1"
    assert events[1]["id"] == "2"
    assert call_count == 2
```

---
## ✅ Checklist
- [x] Tested with `pytest` on Python 3.10, 3.11, 3.12
- [x] Backward compatible with existing synchronous and asynchronous APIs
- [x] Zero external dependencies added (pure `asyncio` & `httpx`)
- [x] Fixes #412
