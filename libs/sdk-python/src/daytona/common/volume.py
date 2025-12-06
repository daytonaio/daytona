# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from daytona_api_client import SandboxVolume as ApiVolumeMount
from daytona_api_client import VolumeDto
from daytona_api_client_async import SandboxVolume as AsyncApiVolumeMount
from daytona_api_client_async import VolumeDto as AsyncVolumeDto


class VolumeMount(ApiVolumeMount, AsyncApiVolumeMount):
    """Represents a Volume mount configuration for a Sandbox.

    Attributes:
        volume_id (str): ID of the volume to mount.
        mount_path (str): Path where the volume will be mounted in the sandbox.
        subpath (str, optional): Optional S3 subpath/prefix within the volume to mount.
            When specified, only this prefix will be accessible. When omitted,
            the entire volume is mounted.
    """


class Volume(VolumeDto):
    """Represents a Daytona Volume which is a shared storage volume for Sandboxes.

    Attributes:
        id (str): Unique identifier for the Volume.
        name (str): Name of the Volume.
        organization_id (str): Organization ID of the Volume.
        state (str): State of the Volume.
        created_at (str): Date and time when the Volume was created.
        updated_at (str): Date and time when the Volume was last updated.
        last_used_at (str): Date and time when the Volume was last used.
    """

    @classmethod
    def from_dto(cls, dto: VolumeDto | AsyncVolumeDto) -> "Volume":
        return cls.model_validate(dto.model_dump())
