# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from daytona_api_client import DiskDto


class Disk(DiskDto):
    """Represents a Daytona Disk which is persistent storage for Sandboxes.

    Attributes:
        id (StrictStr): Unique identifier for the Disk.
        name (StrictStr): Name of the Disk.
        organization_id (StrictStr): Organization ID of the Disk.
        size (int): Disk size in GB.
        state (DiskState): State of the Disk.
        runner_id (Optional[StrictStr]): Runner ID where disk is located.
        error_reason (Optional[StrictStr]): Error reason if in error state.
        created_at (StrictStr): Date and time when the Disk was created.
        updated_at (StrictStr): Date and time when the Disk was last updated.
    """

    @classmethod
    def from_dto(cls, dto: DiskDto) -> "Disk":
        return cls(**dto.__dict__)
