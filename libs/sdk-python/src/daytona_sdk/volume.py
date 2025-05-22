# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

import re
from typing import List

from daytona_api_client import CreateVolume, VolumeDto, VolumesApi
from daytona_api_client.exceptions import NotFoundException


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


class VolumeService:
    """Service for managing Daytona Volumes. Can be used to list, get, create and delete Volumes."""

    def __init__(self, volumes_api: VolumesApi):
        self.__volumes_api = volumes_api

    def list(self) -> List[Volume]:
        """List all Volumes.

        Returns:
            List[Volume]: List of all Volumes.

        Example:
            ```python
            daytona = Daytona()
            volumes = daytona.volume.list()
            for volume in volumes:
                print(f"{volume.name} ({volume.id})")
            ```
        """
        return [Volume.from_dto(volume) for volume in self.__volumes_api.list_volumes()]

    def get(self, name: str, create: bool = False) -> Volume:
        """Get a Volume by name.

        Args:
            name (str): Name of the Volume to get.
            create (bool): If True, create a new Volume if it doesn't exist.

        Returns:
            Volume: The Volume object.

        Example:
            ```python
            daytona = Daytona()
            volume = daytona.volume.get("test-volume-name", create=True)
            print(f"{volume.name} ({volume.id})")
            ```
        """
        try:
            return Volume.from_dto(self.__volumes_api.get_volume_by_name(name))
        except NotFoundException as e:
            if create and re.search(r"Volume with name ([\w\-]+) not found", str(e)):
                return self.create(name)
            raise e

    def create(self, name: str) -> Volume:
        """Create a new Volume.

        Args:
            name (str): Name of the Volume to create.

        Returns:
            Volume: The Volume object.

        Example:
            ```python
            daytona = Daytona()
            volume = daytona.volume.create("test-volume")
            print(f"{volume.name} ({volume.id}); state: {volume.state}")
            ```
        """
        return Volume.from_dto(self.__volumes_api.create_volume(CreateVolume(name=name)))

    def delete(self, volume: Volume) -> None:
        """Delete a Volume.

        Args:
            volume (Volume): Volume to delete.

        Example:
            ```python
            daytona = Daytona()
            volume = daytona.volume.get("test-volume")
            daytona.volume.delete(volume)
            print("Volume deleted")
            ```
        """
        self.__volumes_api.delete_volume(volume.id)
