# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from daytona_api_client import BuildInfo
from daytona_api_client import PaginatedSnapshots as PaginatedSnapshotsDto
from daytona_api_client import SnapshotDto as SyncSnapshotDto
from daytona_api_client_async import BuildInfo as AsyncBuildInfo
from daytona_api_client_async import SnapshotDto as AsyncSnapshotDto
from pydantic import BaseModel

from .image import Image
from .sandbox import Resources


class Snapshot(SyncSnapshotDto):
    """Represents a Daytona Snapshot which is a pre-configured sandbox.

    Attributes:
        id (str): Unique identifier for the Snapshot.
        organization_id (str | None): Organization ID of the Snapshot.
        general (bool): Whether the Snapshot is general.
        name (str): Name of the Snapshot.
        image_name (str): Name of the Image of the Snapshot.
        state (str): State of the Snapshot.
        size (float | int | None): Size of the Snapshot.
        entrypoint (list[str] | None): Entrypoint of the Snapshot.
        cpu (float | int): CPU of the Snapshot.
        gpu (float | int): GPU of the Snapshot.
        mem (float | int): Memory of the Snapshot in GiB.
        disk (float | int): Disk of the Snapshot in GiB.
        error_reason (str | None): Error reason of the Snapshot.
        created_at (str): Timestamp when the Snapshot was created.
        updated_at (str): Timestamp when the Snapshot was last updated.
        last_used_at (str): Timestamp when the Snapshot was last used.
    """

    @classmethod
    def from_dto(cls, dto: SyncSnapshotDto | AsyncSnapshotDto) -> "Snapshot":
        data = dto.model_dump(by_alias=False)
        build_info = data.get("build_info")
        if isinstance(build_info, AsyncBuildInfo):
            data["build_info"] = BuildInfo.model_validate(build_info.model_dump(by_alias=False))
        elif isinstance(build_info, dict) and not isinstance(build_info, BuildInfo):
            data["build_info"] = BuildInfo.model_validate(build_info)
        return cls.model_validate(data)


class PaginatedSnapshots(PaginatedSnapshotsDto):
    """Represents a paginated list of Daytona Snapshots.

    Attributes:
        items (list[Snapshot]): List of Snapshot instances in the current page.
        total (int): Total number of Snapshots across all pages.
        page (int): Current page number.
        total_pages (int): Total number of pages available.
    """


class CreateSnapshotParams(BaseModel):
    """Parameters for creating a new snapshot.

    Attributes:
        name (str): Name of the snapshot.
        image (str | Image): Image of the snapshot. If a string is provided,
            it should be available on some registry. If an Image instance is provided,
            it will be used to create a new image in Daytona.
        resources (Resources | None): Resources of the snapshot.
        entrypoint (list[str] | None): Entrypoint of the snapshot.
        skip_validation (bool | None): Whether to skip validation for the snapshot.
    """

    name: str
    image: str | Image
    resources: Resources | None = None
    entrypoint: list[str] | None = None
    skip_validation: bool | None = None
