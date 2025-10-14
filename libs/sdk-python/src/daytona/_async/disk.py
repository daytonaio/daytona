# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from typing import List

from daytona_api_client_async import CreateDiskDto, DisksApi

from ..common.disk import Disk


class AsyncDiskService:
    """Service for managing Daytona Disks. Can be used to list, get, create and delete Disks."""

    def __init__(self, disks_api: DisksApi):
        self.__disks_api = disks_api

    async def list(self) -> List[Disk]:
        """List all Disks.

        Returns:
            List[Disk]: List of all Disks.

        Example:
            ```python
            daytona = AsyncDaytona()
            disks = await daytona.disk.list()
            for disk in disks:
                print(f"{disk.name} ({disk.id}) - {disk.size}GB")
            ```
        """
        disks = await self.__disks_api.list_disks()
        return [Disk.from_dto(disk) for disk in disks]

    async def get(self, disk_id: str) -> Disk:
        """Get a Disk by ID.

        Args:
            disk_id (str): ID of the Disk to get.

        Returns:
            Disk: The Disk object.

        Example:
            ```python
            daytona = AsyncDaytona()
            disk = await daytona.disk.get("disk-id")
            print(f"{disk.name} ({disk.id}) - {disk.size}GB")
            ```
        """
        disk_dto = await self.__disks_api.get_disk(disk_id)
        return Disk.from_dto(disk_dto)

    async def create(self, name: str, size: int) -> Disk:
        """Create a new Disk.

        Args:
            name (str): Name of the Disk to create.
            size (int): Size of the Disk in GB.

        Returns:
            Disk: The Disk object.

        Example:
            ```python
            daytona = AsyncDaytona()
            disk = await daytona.disk.create("test-disk", 50)
            print(f"{disk.name} ({disk.id}); state: {disk.state}; size: {disk.size}GB")
            ```
        """
        disk_dto = await self.__disks_api.create_disk(CreateDiskDto(name=name, size=size))
        return Disk.from_dto(disk_dto)

    async def delete(self, disk: Disk) -> None:
        """Delete a Disk.

        Args:
            disk (Disk): Disk to delete.

        Example:
            ```python
            daytona = AsyncDaytona()
            disk = await daytona.disk.get("test-disk")
            await daytona.disk.delete(disk)
            print("Disk deleted")
            ```
        """
        await self.__disks_api.delete_disk(disk.id)
