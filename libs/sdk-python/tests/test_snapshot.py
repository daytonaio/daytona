from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest

from daytona.common.errors import DaytonaError
from daytona.common.image import Image
from daytona.common.snapshot import CreateSnapshotParams, Snapshot


class TestSyncSnapshotService:
    def _make_service(self):
        from daytona._sync.snapshot import SnapshotService

        mock_snapshots_api = MagicMock()
        mock_object_storage_api = MagicMock()
        return SnapshotService(mock_snapshots_api, mock_object_storage_api, "us"), mock_snapshots_api

    def _make_snapshot_dto(self, name="test-snapshot"):
        dto = MagicMock()
        dto.model_dump.return_value = {
            "id": "snap-123",
            "organization_id": "org-1",
            "general": False,
            "name": name,
            "image_name": "python:3.12",
            "ref": None,
            "state": "active",
            "size": None,
            "entrypoint": None,
            "cpu": 4,
            "gpu": 0,
            "mem": 8,
            "disk": 30,
            "build_info": None,
            "error_reason": None,
            "created_at": "2025-01-01T00:00:00Z",
            "updated_at": "2025-01-01T00:00:00Z",
            "last_used_at": "2025-01-01T00:00:00Z",
        }
        return dto

    def test_list(self):
        service, api = self._make_service()
        mock_response = MagicMock()
        mock_response.items = [self._make_snapshot_dto()]
        mock_response.total = 1
        mock_response.page = 1
        mock_response.total_pages = 1
        api.get_all_snapshots.return_value = mock_response
        result = service.list()
        assert result.total == 1
        assert len(result.items) == 1

    def test_list_invalid_page_raises(self):
        service, api = self._make_service()
        with pytest.raises(DaytonaError, match="page must be a positive integer"):
            service.list(page=0)

    def test_list_invalid_limit_raises(self):
        service, api = self._make_service()
        with pytest.raises(DaytonaError, match="limit must be a positive integer"):
            service.list(limit=0)

    def test_get(self):
        service, api = self._make_service()
        api.get_snapshot.return_value = self._make_snapshot_dto()
        result = service.get("test-snapshot")
        assert isinstance(result, Snapshot)

    def test_delete(self):
        service, api = self._make_service()
        api.remove_snapshot.return_value = None
        snap = Snapshot.model_validate({
            "id": "snap-123",
            "organization_id": "org-1",
            "general": False,
            "name": "test-snapshot",
            "image_name": "python:3.12",
            "ref": None,
            "state": "active",
            "size": None,
            "entrypoint": None,
            "cpu": 4,
            "gpu": 0,
            "mem": 8,
            "disk": 30,
            "build_info": None,
            "error_reason": None,
            "created_at": "2025-01-01T00:00:00Z",
            "updated_at": "2025-01-01T00:00:00Z",
            "last_used_at": "2025-01-01T00:00:00Z",
        })
        service.delete(snap)
        api.remove_snapshot.assert_called_once_with("snap-123")


class TestAsyncSnapshotService:
    def _make_service(self):
        from daytona._async.snapshot import AsyncSnapshotService

        mock_snapshots_api = AsyncMock()
        mock_object_storage_api = AsyncMock()
        return AsyncSnapshotService(mock_snapshots_api, mock_object_storage_api, "us"), mock_snapshots_api

    def _make_snapshot_dto(self, name="test-snapshot"):
        dto = MagicMock()
        dto.model_dump.return_value = {
            "id": "snap-123",
            "organization_id": "org-1",
            "general": False,
            "name": name,
            "image_name": "python:3.12",
            "ref": None,
            "state": "active",
            "size": None,
            "entrypoint": None,
            "cpu": 4,
            "gpu": 0,
            "mem": 8,
            "disk": 30,
            "build_info": None,
            "error_reason": None,
            "created_at": "2025-01-01T00:00:00Z",
            "updated_at": "2025-01-01T00:00:00Z",
            "last_used_at": "2025-01-01T00:00:00Z",
        }
        return dto

    @pytest.mark.asyncio
    async def test_list(self):
        service, api = self._make_service()
        mock_response = MagicMock()
        mock_response.items = [self._make_snapshot_dto()]
        mock_response.total = 1
        mock_response.page = 1
        mock_response.total_pages = 1
        api.get_all_snapshots.return_value = mock_response
        result = await service.list()
        assert result.total == 1

    @pytest.mark.asyncio
    async def test_list_invalid_page_raises(self):
        service, api = self._make_service()
        with pytest.raises(DaytonaError, match="page must be a positive integer"):
            await service.list(page=0)

    @pytest.mark.asyncio
    async def test_get(self):
        service, api = self._make_service()
        api.get_snapshot.return_value = self._make_snapshot_dto()
        result = await service.get("test-snapshot")
        assert isinstance(result, Snapshot)

    @pytest.mark.asyncio
    async def test_delete(self):
        service, api = self._make_service()
        snap = Snapshot.model_validate({
            "id": "snap-123",
            "organization_id": "org-1",
            "general": False,
            "name": "test-snapshot",
            "image_name": "python:3.12",
            "ref": None,
            "state": "active",
            "size": None,
            "entrypoint": None,
            "cpu": 4,
            "gpu": 0,
            "mem": 8,
            "disk": 30,
            "build_info": None,
            "error_reason": None,
            "created_at": "2025-01-01T00:00:00Z",
            "updated_at": "2025-01-01T00:00:00Z",
            "last_used_at": "2025-01-01T00:00:00Z",
        })
        await service.delete(snap)
        api.remove_snapshot.assert_called_once_with("snap-123")
