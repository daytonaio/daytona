# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import time

import pytest

from daytona._utils.file_url_signing import build_signed_file_url, compute_file_url_signature, resolve_expires
from daytona.common.errors import DaytonaError

# Cross-language test vectors — the proxy's Go test suite asserts the exact
# same signatures (apps/proxy/pkg/proxy/file_url_auth_test.go). If one of
# these changes, both sides must change together.
CROSS_LANG_TEST_KEY = "testsigningkey1234567890abcdefgh"

CROSS_LANG_VECTORS = [
    ("GET", "/home/user/report.pdf", 1781234567, "v1_QXoy36mypac2FAv33L7jnN44GEUx8KrdwT0vuKgKiQg"),
    ("GET", "/home/user/report.pdf", 0, "v1_lpj67Q-1iHxBviass5MZhGs36X80uk3DgCaRjjmyPrk"),
    ("POST", "/tmp/incoming/data.bin", 1781234567, "v1_CziiRdFkC9asB7q1mi0-fDvvwkpTxcI7yR8N35ht9Vw"),
    ("GET", "/path with spaces/f.txt", 1900000000, "v1_GynMaKcifGfdmrBJHusa_ucAXowjZ_g4KP6lcMJ4WXE"),
]


@pytest.mark.parametrize("method,path,expires,expected", CROSS_LANG_VECTORS)
def test_cross_language_vectors(method: str, path: str, expires: int, expected: str):
    assert compute_file_url_signature(CROSS_LANG_TEST_KEY, method, path, expires) == expected


def test_resolve_expires_default_ttl():
    expires = resolve_expires(None)
    assert abs(expires - (int(time.time()) + 3600)) <= 2


def test_resolve_expires_negative_means_no_expiry():
    assert resolve_expires(-1) == 0


def test_resolve_expires_zero_means_no_expiry():
    assert resolve_expires(0) == 0


def test_build_signed_file_url_shape():
    url = build_signed_file_url(
        "https://proxy.example.com/toolbox",
        "sandbox-123",
        "/files/download",
        "GET",
        "/home/user/f.txt",
        CROSS_LANG_TEST_KEY,
        -1,
    )
    assert url.startswith("https://proxy.example.com/toolbox/sandbox-123/files/download?")
    assert "path=%2Fhome%2Fuser%2Ff.txt" in url
    assert "expires=0" in url
    assert "signature=v1_" in url


def test_build_signed_file_url_requires_key():
    with pytest.raises(DaytonaError):
        build_signed_file_url(
            "https://proxy.example.com/toolbox",
            "sandbox-123",
            "/files/download",
            "GET",
            "/f.txt",
            None,
            None,
        )
