# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import re
from typing import List

from daytona_api_client_async import CreateVolume, VolumesApi
from daytona_api_client_async.exceptions import NotFoundException

from ..common.volume import Volume


class AsyncVolumeService:
    """Service for managing Daytona Volumes. Can be used to list, get, create and delete Volumes."""

    def __init__(self, volumes_api: VolumesApi):
        self.__volumes_api = volumes_api

    async def list(self) -> List[Volume]:
        """List all Volumes.

        Returns:
            List[Volume]: List of all Volumes.

        Example:
            ```python
            async with AsyncDaytona() as daytona:
                volumes = await daytona.volume.list()
                for volume in volumes:
                    print(f"{volume.name} ({volume.id})")
            ```
        """
        return [Volume.from_dto(volume) for volume in await self.__volumes_api.list_volumes()]

    async def get(self, name: str, create: bool = False) -> Volume:
        """Get a Volume by name.

        Args:
            name (str): Name of the Volume to get.
            create (bool): If True, create a new Volume if it doesn't exist.

        Returns:
            Volume: The Volume object.

        Example:
            ```python
            async with AsyncDaytona() as daytona:
                volume = await daytona.volume.get("test-volume-name", create=True)
                print(f"{volume.name} ({volume.id})")
            ```
        """
        try:
            return Volume.from_dto(await self.__volumes_api.get_volume_by_name(name))
        except NotFoundException as e:
            if create and re.search(r"Volume with name ([\w\-]+) not found", str(e)):
                return await self.create(name)
            raise e

    async def create(self, name: str) -> Volume:
        """Create a new Volume.

        Args:
            name (str): Name of the Volume to create.

        Returns:
            Volume: The Volume object.

        Example:
            ```python
            async with AsyncDaytona() as daytona:
                volume = await daytona.volume.create("test-volume")
                print(f"{volume.name} ({volume.id}); state: {volume.state}")
            ```
        """
        return Volume.from_dto(await self.__volumes_api.create_volume(CreateVolume(name=name)))

    async def delete(self, volume: Volume) -> None:
        """Delete a Volume.

        Args:
            volume (Volume): Volume to delete.

        Example:
            ```python
            async with AsyncDaytona() as daytona:
                volume = await daytona.volume.get("test-volume")
                await daytona.volume.delete(volume)
                print("Volume deleted")
            ```
        """
        await self.__volumes_api.delete_volume(volume.id)
