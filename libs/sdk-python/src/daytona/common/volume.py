# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from daytona_api_client import SandboxVolume as ApiVolumeMount
from daytona_api_client import VolumeDto
from daytona_api_client_async import SandboxVolume as AsyncApiVolumeMount


class VolumeMount(ApiVolumeMount, AsyncApiVolumeMount):
    ...


class Volume(VolumeDto):
    """Represents a Daytona Volume which is a shared storage volume for Sandboxes.

    Attributes:
        id (StrictStr): Unique identifier for the Volume.
        name (StrictStr): Name of the Volume.
        organization_id (StrictStr): Organization ID of the Volume.
        state (StrictStr): State of the Volume.
        created_at (StrictStr): Date and time when the Volume was created.
        updated_at (StrictStr): Date and time when the Volume was last updated.
        last_used_at (StrictStr): Date and time when the Volume was last used.
    """

    @classmethod
    def from_dto(cls, dto: VolumeDto) -> "Volume":
        return cls(**dto.__dict__)
