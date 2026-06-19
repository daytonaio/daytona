# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import base64
import hashlib
import hmac
import time
from urllib.parse import urlencode

from ..common.errors import DaytonaError

SIGNATURE_V1_PREFIX = "v1_"
DEFAULT_TTL_SECONDS = 3600


def compute_file_url_signature(signing_key: str, method: str, path: str, expires: int) -> str:
    canonical = f"v1:files:{method}:{path}:{expires}".encode()
    digest = hmac.new(signing_key.encode(), canonical, hashlib.sha256).digest()
    return SIGNATURE_V1_PREFIX + base64.urlsafe_b64encode(digest).rstrip(b"=").decode()


def resolve_expires(ttl_seconds: int | None) -> int:
    if ttl_seconds is None:
        return int(time.time()) + DEFAULT_TTL_SECONDS
    if ttl_seconds <= 0:
        return 0
    return int(time.time()) + ttl_seconds


def build_signed_file_url(
    toolbox_proxy_url: str,
    sandbox_id: str,
    operation_path: str,
    method: str,
    file_path: str,
    signing_key: str | None,
    ttl_seconds: int | None,
) -> str:
    if not signing_key:
        raise DaytonaError(
            "Sandbox signing key is not available. " + "Call refresh_data() or fetch the sandbox by ID to load it."
        )
    expires = resolve_expires(ttl_seconds)
    signature = compute_file_url_signature(signing_key, method, file_path, expires)
    query = urlencode({"path": file_path, "expires": str(expires), "signature": signature})
    return f"{toolbox_proxy_url}/{sandbox_id}{operation_path}?{query}"
