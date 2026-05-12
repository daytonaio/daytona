# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import httpx

from daytona.internal.http_client import build_async_http_client, build_sync_http_client


class TestBuildHttpClient:
    def test_sync_returns_client(self):
        client = build_sync_http_client(pool_size=42)
        try:
            assert isinstance(client, httpx.Client)
        finally:
            client.close()

    def test_sync_accepts_none(self):
        client = build_sync_http_client(pool_size=None)
        try:
            assert isinstance(client, httpx.Client)
        finally:
            client.close()

    def test_async_returns_client(self):
        client = build_async_http_client(pool_size=17)
        assert isinstance(client, httpx.AsyncClient)
