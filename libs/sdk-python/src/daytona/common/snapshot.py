# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from typing import List, Optional, Union

from daytona_api_client import BuildInfo, SnapshotDto
from daytona_api_client_async import BuildInfo as AsyncBuildInfo
from pydantic import BaseModel

from .daytona import Resources
from .image import Image


class Snapshot(SnapshotDto):
    """Represents a Daytona Snapshot which is a pre-configured sandbox.

    Attributes:
        id (StrictStr): Unique identifier for the Snapshot.
        organization_id (Optional[StrictStr]): Organization ID of the Snapshot.
        general (Optional[bool]): Whether the Snapshot is general.
        name (StrictStr): Name of the Snapshot.
        image_name (StrictStr): Name of the Image of the Snapshot.
        enabled (StrictBool): Whether the Snapshot is enabled.
        state (StrictStr): State of the Snapshot.
        size (Optional[Union[StrictFloat, StrictInt]]): Size of the Snapshot.
        entrypoint (Optional[List[str]]): Entrypoint of the Snapshot.
        cpu (Union[StrictFloat, StrictInt]): CPU of the Snapshot.
        gpu (Union[StrictFloat, StrictInt]): GPU of the Snapshot.
        mem (Union[StrictFloat, StrictInt]): Memory of the Snapshot in GiB.
        disk (Union[StrictFloat, StrictInt]): Disk of the Snapshot in GiB.
        error_reason (Optional[StrictStr]): Error reason of the Snapshot.
        created_at (StrictStr): Timestamp when the Snapshot was created.
        updated_at (StrictStr): Timestamp when the Snapshot was last updated.
        last_used_at (StrictStr): Timestamp when the Snapshot was last used.
    """

    build_info: Optional[Union[BuildInfo, AsyncBuildInfo]] = None

    @classmethod
    def from_dto(cls, dto: SnapshotDto) -> "Snapshot":
        return cls(**dto.__dict__)


class CreateSnapshotParams(BaseModel):
    """Parameters for creating a new snapshot.

    Attributes:
        name (Optional[str]): Name of the snapshot.
        image (Union[str, Image]): Image of the snapshot. If a string is provided,
            it should be available on some registry. If an Image instance is provided,
            it will be used to create a new image in Daytona.
        resources (Optional[Resources]): Resources of the snapshot.
        entrypoint (Optional[List[str]]): Entrypoint of the snapshot.
    """

    name: str
    image: Union[str, Image]
    resources: Optional[Resources] = None
    entrypoint: Optional[List[str]] = None
