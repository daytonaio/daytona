# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest

from daytona.common.volume import Volume
from daytona_api_client import VolumeDto


def _make_volume_dto(name="test-vol", vol_id="vol-123"):
    return VolumeDto(
        id=vol_id,
        name=name,
        organization_id="org-1",
        state="ready",
        error_reason=None,
        created_at="2025-01-01T00:00:00Z",
        updated_at="2025-01-01T00:00:00Z",
        last_used_at="2025-01-01T00:00:00Z",
    )


def _make_volume(name="test-vol", vol_id="vol-123"):
    return Volume(
        id=vol_id,
        name=name,
        organization_id="org-1",
        state="ready",
        error_reason=None,
        created_at="2025-01-01T00:00:00Z",
        updated_at="2025-01-01T00:00:00Z",
        last_used_at="2025-01-01T00:00:00Z",
    )


class TestSyncVolumeService:
    def _make_service(self):
        from daytona._sync.volume import VolumeService

        mock_api = MagicMock()
        return VolumeService(mock_api), mock_api

    def test_list(self):
        service, api = self._make_service()
        api.list_volumes.return_value = [_make_volume_dto()]
        result = service.list()
        assert len(result) == 1
        assert isinstance(result[0], Volume)

    def test_get(self):
        service, api = self._make_service()
        api.get_volume_by_name.return_value = _make_volume_dto()
        result = service.get("test-vol")
        assert isinstance(result, Volume)

    def test_get_with_create(self):
        from daytona_api_client.exceptions import NotFoundException

        service, api = self._make_service()
        api.get_volume_by_name.side_effect = NotFoundException(status=404, reason="Not found")
        api.create_volume.return_value = _make_volume_dto(name="new-vol")
        result = service.get("new-vol", create=True)
        assert isinstance(result, Volume)
        api.create_volume.assert_called_once()

    def test_get_not_found_raises(self):
        from daytona_api_client.exceptions import NotFoundException

        service, api = self._make_service()
        api.get_volume_by_name.side_effect = NotFoundException(status=404, reason="Not found")
        with pytest.raises(NotFoundException):
            service.get("nonexistent")

    def test_create(self):
        service, api = self._make_service()
        api.create_volume.return_value = _make_volume_dto(name="new-vol")
        result = service.create("new-vol")
        assert isinstance(result, Volume)

    def test_delete(self):
        service, api = self._make_service()
        api.delete_volume.return_value = None
        vol = _make_volume()
        service.delete(vol)
        api.delete_volume.assert_called_once_with("vol-123")


class TestAsyncVolumeService:
    def _make_service(self):
        from daytona._async.volume import AsyncVolumeService

        mock_api = AsyncMock()
        return AsyncVolumeService(mock_api), mock_api

    @pytest.mark.asyncio
    async def test_list(self):
        service, api = self._make_service()
        api.list_volumes.return_value = [_make_volume_dto()]
        result = await service.list()
        assert len(result) == 1
        assert isinstance(result[0], Volume)

    @pytest.mark.asyncio
    async def test_get(self):
        service, api = self._make_service()
        api.get_volume_by_name.return_value = _make_volume_dto()
        result = await service.get("test-vol")
        assert isinstance(result, Volume)

    @pytest.mark.asyncio
    async def test_create(self):
        service, api = self._make_service()
        api.create_volume.return_value = _make_volume_dto(name="new-vol")
        result = await service.create("new-vol")
        assert isinstance(result, Volume)

    @pytest.mark.asyncio
    async def test_delete(self):
        service, api = self._make_service()
        vol = _make_volume()
        await service.delete(vol)
        api.delete_volume.assert_called_once_with("vol-123")

    @pytest.mark.asyncio
    async def test_get_with_create(self):
        from daytona_api_client_async.exceptions import NotFoundException

        service, api = self._make_service()
        api.get_volume_by_name.side_effect = NotFoundException(status=404, reason="Not found")
        api.create_volume.return_value = _make_volume_dto(name="new-vol")

        result = await service.get("new-vol", create=True)

        assert isinstance(result, Volume)
        api.create_volume.assert_called_once()
