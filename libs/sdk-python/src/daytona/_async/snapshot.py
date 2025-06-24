# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import asyncio
from typing import Callable, List, Optional

from daytona_api_client_async import ObjectStorageApi, SnapshotsApi
from daytona_api_client_async.models.create_build_info import CreateBuildInfo
from daytona_api_client_async.models.create_snapshot import CreateSnapshot
from daytona_api_client_async.models.snapshot_state import SnapshotState

from .._utils.errors import intercept_errors
from .._utils.stream import process_streaming_response
from .._utils.timeout import with_timeout
from ..common.errors import DaytonaError
from ..common.image import Image
from ..common.snapshot import CreateSnapshotParams, Snapshot
from .object_storage import AsyncObjectStorage

SNAPSHOTS_FETCH_LIMIT = 200


class AsyncSnapshotService:
    """Service for managing Daytona Snapshots. Can be used to list, get, create and delete Snapshots."""

    def __init__(self, snapshots_api: SnapshotsApi, object_storage_api: ObjectStorageApi):
        self.__snapshots_api = snapshots_api
        self.__object_storage_api = object_storage_api

    @intercept_errors(message_prefix="Failed to list snapshots: ")
    async def list(self) -> List[Snapshot]:
        """List all Snapshots.

        Returns:
            List[Snapshot]: List of all Snapshots.

        Example:
            ```python
            async with AsyncDaytona() as daytona:
                snapshots = await daytona.snapshot.list()
                for snapshot in snapshots:
                    print(f"{snapshot.name} ({snapshot.image_name})")
            ```
        """
        response = await self.__snapshots_api.get_all_snapshots(limit=SNAPSHOTS_FETCH_LIMIT)
        if response.total > SNAPSHOTS_FETCH_LIMIT:
            response = await self.__snapshots_api.get_all_snapshots(limit=response.total)
        return [Snapshot.from_dto(snapshot) for snapshot in response.items]

    @intercept_errors(message_prefix="Failed to delete snapshot: ")
    async def delete(self, snapshot: Snapshot) -> None:
        """Delete a Snapshot.

        Args:
            snapshot (Snapshot): Snapshot to delete.

        Example:
            ```python
            async with AsyncDaytona() as daytona:
                snapshot = await daytona.snapshot.get("test-snapshot")
                await daytona.snapshot.delete(snapshot)
                print("Snapshot deleted")
            ```
        """
        await self.__snapshots_api.remove_snapshot(snapshot.id)

    @intercept_errors(message_prefix="Failed to get snapshot: ")
    async def get(self, name: str) -> Snapshot:
        """Get a Snapshot by name.

        Args:
            name (str): Name of the Snapshot to get.

        Returns:
            Snapshot: The Snapshot object.

        Example:
            ```python
            async with AsyncDaytona() as daytona:
                snapshot = await daytona.snapshot.get("test-snapshot-name")
                print(f"{snapshot.name} ({snapshot.image_name})")
            ```
        """
        return Snapshot.from_dto(await self.__snapshots_api.get_snapshot(name))

    @intercept_errors(message_prefix="Failed to create snapshot: ")
    @with_timeout(
        error_message=lambda self, timeout: (f"Failed to create snapshot within {timeout} seconds timeout period.")
    )
    async def create(
        self,
        params: CreateSnapshotParams,
        *,
        on_logs: Callable[[str], None] = None,
        timeout: Optional[float] = 0,  # pylint: disable=unused-argument
    ) -> Snapshot:
        """Creates and registers a new snapshot from the given Image definition.
        Args:
            params (CreateSnapshotParams): Parameters for snapshot creation.
            on_logs (Callable[[str], None]): This callback function handles snapshot creation logs.
            timeout (Optional[float]): Default is no timeout. Timeout in seconds (0 means no timeout).
        Example:
            ```python
            image = Image.debianSlim('3.12').pipInstall('numpy')
            daytona.snapshot.create(
                CreateSnapshotParams(name='my-snapshot', image=image),
                on_logs=lambda chunk: print(chunk, end=""),
            )
            ```
        """
        created_snapshot = None
        create_snapshot_req = CreateSnapshot(
            name=params.name,
        )

        if isinstance(params.image, str):
            create_snapshot_req.image_name = params.image
            create_snapshot_req.entrypoint = params.entrypoint
        else:
            context_hashes = await AsyncSnapshotService.process_image_context(self.__object_storage_api, params.image)
            create_snapshot_req.build_info = CreateBuildInfo(
                context_hashes=context_hashes,
                dockerfile_content=(
                    params.image.entrypoint(params.entrypoint).dockerfile()
                    if params.entrypoint
                    else params.image.dockerfile()
                ),
            )

        if params.resources:
            create_snapshot_req.cpu = params.resources.cpu
            create_snapshot_req.gpu = params.resources.gpu
            create_snapshot_req.memory = params.resources.memory
            create_snapshot_req.disk = params.resources.disk

        created_snapshot = await self.__snapshots_api.create_snapshot(create_snapshot_req)

        terminal_states = [SnapshotState.ACTIVE, SnapshotState.ERROR, SnapshotState.BUILD_FAILED]
        log_terminal_states = [*terminal_states, SnapshotState.PENDING_VALIDATION, SnapshotState.VALIDATING]

        async def start_log_streaming():
            _, url, *_ = self.__snapshots_api._get_snapshot_build_logs_serialize(  # pylint: disable=protected-access
                id=created_snapshot.id,
                follow=True,
                x_daytona_organization_id=None,
                _request_auth=None,
                _content_type=None,
                _headers=None,
                _host_index=None,
            )

            async def should_terminate():
                latest_snapshot = await self.__snapshots_api.get_snapshot(created_snapshot.id)
                return latest_snapshot.state in log_terminal_states

            await process_streaming_response(
                url=url,
                headers=self.__snapshots_api.api_client.default_headers,
                on_chunk=lambda chunk: on_logs(chunk.rstrip()),
                should_terminate=should_terminate,
            )

        log_task = None
        if on_logs:
            on_logs(f"Creating snapshot {created_snapshot.name} ({created_snapshot.state})")
            if created_snapshot.state != SnapshotState.BUILD_PENDING:
                log_task = asyncio.create_task(start_log_streaming())

        previous_state = created_snapshot.state
        while created_snapshot.state not in terminal_states:
            if on_logs and previous_state != created_snapshot.state:
                if created_snapshot.state != SnapshotState.BUILD_PENDING and not log_task:
                    log_task = asyncio.create_task(start_log_streaming())
                on_logs(f"Creating snapshot {created_snapshot.name} ({created_snapshot.state})")
                previous_state = created_snapshot.state
            await asyncio.sleep(1)
            created_snapshot = await self.__snapshots_api.get_snapshot(created_snapshot.id)

        if on_logs:
            await log_task
            if created_snapshot.state == SnapshotState.ACTIVE:
                on_logs(f"Created snapshot {created_snapshot.name} ({created_snapshot.state})")

        if created_snapshot.state in (SnapshotState.ERROR, SnapshotState.BUILD_FAILED):
            raise DaytonaError(
                f"Failed to create snapshot {created_snapshot.name}, reason: {created_snapshot.error_reason}"
            )

        return created_snapshot if isinstance(created_snapshot, Snapshot) else Snapshot.from_dto(created_snapshot)

    @staticmethod
    async def process_image_context(object_storage_api: ObjectStorageApi, image: Image) -> List[str]:
        """Processes the image context by uploading it to object storage.
        Args:
            image (Image): The Image instance.
        Returns:
            List[str]: List of context hashes stored in object storage.
        """
        if not image._context_list:  # pylint: disable=protected-access
            return []

        push_access_creds = await object_storage_api.get_push_access()

        async with AsyncObjectStorage(
            push_access_creds.storage_url,
            push_access_creds.access_key,
            push_access_creds.secret,
            push_access_creds.session_token,
            push_access_creds.bucket,
        ) as object_storage:
            context_hashes = []
            for context in image._context_list:  # pylint: disable=protected-access
                context_hash = await object_storage.upload(
                    context.source_path, push_access_creds.organization_id, context.archive_path
                )
                context_hashes.append(context_hash)

        return context_hashes
