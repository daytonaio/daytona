# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import asyncio
import threading
import time
from typing import Callable, cast

from daytona_api_client import (
    CreateBuildInfo,
    CreateSnapshot,
    ObjectStorageApi,
    SnapshotDto,
    SnapshotsApi,
    SnapshotState,
)

from .._utils.errors import intercept_errors
from .._utils.otel_decorator import with_instrumentation
from .._utils.stream import process_streaming_response
from .._utils.timeout import with_timeout
from ..common.errors import DaytonaError
from ..common.image import Image
from ..common.snapshot import CreateSnapshotParams, PaginatedSnapshots, Snapshot
from .object_storage import ObjectStorage


class SnapshotService:
    """Service for managing Daytona Snapshots. Can be used to list, get, create and delete Snapshots."""

    def __init__(
        self, snapshots_api: SnapshotsApi, object_storage_api: ObjectStorageApi, default_region_id: str | None = None
    ):
        self.__snapshots_api = snapshots_api
        self.__object_storage_api = object_storage_api
        self.__default_region_id = default_region_id

    @intercept_errors(message_prefix="Failed to list snapshots: ")
    @with_instrumentation()
    def list(self, page: int | None = None, limit: int | None = None) -> PaginatedSnapshots:
        """Returns paginated list of Snapshots.

        Args:
            page (int | None): Page number for pagination (starting from 1).
            limit (int | None): Maximum number of items per page.

        Returns:
            PaginatedSnapshots: Paginated list of Snapshots.

        Example:
            ```python
            daytona = Daytona()
            result = daytona.snapshot.list(page=2, limit=10)
            for snapshot in result.items:
                print(f"{snapshot.name} ({snapshot.image_name})")
            ```
        """
        if page is not None and page < 1:
            raise DaytonaError("page must be a positive integer")

        if limit is not None and limit < 1:
            raise DaytonaError("limit must be a positive integer")

        response = self.__snapshots_api.get_all_snapshots(limit=limit, page=page)
        return PaginatedSnapshots(
            items=[Snapshot.from_dto(snapshot) for snapshot in response.items],
            total=response.total,
            page=response.page,
            total_pages=response.total_pages,
        )

    @intercept_errors(message_prefix="Failed to delete snapshot: ")
    @with_instrumentation()
    def delete(self, snapshot: Snapshot) -> None:
        """Delete a Snapshot.

        Args:
            snapshot (Snapshot): Snapshot to delete.

        Example:
            ```python
            daytona = Daytona()
            snapshot = daytona.snapshot.get("test-snapshot")
            daytona.snapshot.delete(snapshot)
            print("Snapshot deleted")
            ```
        """
        self.__snapshots_api.remove_snapshot(snapshot.id)

    @intercept_errors(message_prefix="Failed to get snapshot: ")
    @with_instrumentation()
    def get(self, name: str) -> Snapshot:
        """Get a Snapshot by name.

        Args:
            name (str): Name of the Snapshot to get.

        Returns:
            Snapshot: The Snapshot object.

        Example:
            ```python
            daytona = Daytona()
            snapshot = daytona.snapshot.get("test-snapshot-name")
            print(f"{snapshot.name} ({snapshot.image_name})")
            ```
        """
        return Snapshot.from_dto(self.__snapshots_api.get_snapshot(name))

    @intercept_errors(message_prefix="Failed to create snapshot: ")
    @with_timeout()
    @with_instrumentation()
    def create(
        self,
        params: CreateSnapshotParams,
        *,
        on_logs: Callable[[str], None] | None = None,
        timeout: float | None = 0,  # pylint: disable=unused-argument # pyright: ignore[reportUnusedParameter]
    ) -> Snapshot:
        """Creates and registers a new snapshot from the given Image definition.
        Args:
            params (CreateSnapshotParams): Parameters for snapshot creation.
            on_logs (Callable[[str], None]): This callback function handles snapshot creation logs.
            timeout (float | None): Default is no timeout. Timeout in seconds (0 means no timeout).
        Example:
            ```python
            image = Image.debianSlim('3.12').pipInstall('numpy')
            daytona.snapshot.create(
                CreateSnapshotParams(name='my-snapshot', image=image),
                on_logs=lambda chunk: print(chunk, end=""),
            )
            ```
        """
        create_snapshot_req = CreateSnapshot(
            name=params.name,
        )

        if isinstance(params.image, str):
            create_snapshot_req.image_name = params.image
            create_snapshot_req.entrypoint = params.entrypoint
        else:
            context_hashes = SnapshotService.process_image_context(self.__object_storage_api, params.image)
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

        create_snapshot_req.region_id = params.region_id or self.__default_region_id

        created_snapshot: SnapshotDto = self.__snapshots_api.create_snapshot(create_snapshot_req)

        terminal_states = [SnapshotState.ACTIVE, SnapshotState.ERROR, SnapshotState.BUILD_FAILED]

        def start_log_streaming():
            build_logs_url = (self.__snapshots_api.get_snapshot_build_logs_url(created_snapshot.id)).url

            def should_terminate():
                latest_snapshot = self.__snapshots_api.get_snapshot(created_snapshot.id)
                return latest_snapshot.state in terminal_states

            asyncio.run(
                process_streaming_response(
                    url=build_logs_url + "?follow=true",
                    headers=cast(dict[str, str], self.__snapshots_api.api_client.default_headers),
                    on_chunk=lambda chunk: on_logs(chunk.rstrip()) if on_logs else None,
                    should_terminate=should_terminate,
                )
            )

        log_task = None
        if on_logs:
            on_logs(f"Creating snapshot {created_snapshot.name} ({created_snapshot.state})")
            if (
                create_snapshot_req.build_info
                and created_snapshot.state != SnapshotState.PENDING
                and created_snapshot.state not in terminal_states
            ):
                log_task = threading.Thread(target=start_log_streaming)
                log_task.start()

        previous_state = created_snapshot.state
        while created_snapshot.state not in terminal_states:
            if on_logs and previous_state != created_snapshot.state:
                if create_snapshot_req.build_info and created_snapshot.state != SnapshotState.PENDING and not log_task:
                    log_task = threading.Thread(target=start_log_streaming)
                    log_task.start()
                on_logs(f"Creating snapshot {created_snapshot.name} ({created_snapshot.state})")
                previous_state = created_snapshot.state
            time.sleep(1)
            created_snapshot = self.__snapshots_api.get_snapshot(created_snapshot.id)

        if on_logs:
            if log_task:
                log_task.join()
            if created_snapshot.state == SnapshotState.ACTIVE:
                on_logs(f"Created snapshot {created_snapshot.name} ({created_snapshot.state})")

        if created_snapshot.state in (SnapshotState.ERROR, SnapshotState.BUILD_FAILED):
            raise DaytonaError(
                f"Failed to create snapshot {created_snapshot.name}, reason: {created_snapshot.error_reason}"
            )

        return created_snapshot if isinstance(created_snapshot, Snapshot) else Snapshot.from_dto(created_snapshot)

    @with_instrumentation()
    def activate(self, snapshot: Snapshot) -> Snapshot:
        """Activate a snapshot.
        Args:
            snapshot (Snapshot): The Snapshot instance.
        Returns:
            Snapshot: The activated Snapshot instance.
        """
        return Snapshot.from_dto(self.__snapshots_api.activate_snapshot(snapshot.id))

    @staticmethod
    @with_instrumentation()
    def process_image_context(object_storage_api: ObjectStorageApi, image: Image) -> list[str]:
        """Processes the image context by uploading it to object storage.
        Args:
            image (Image): The Image instance.
        Returns:
            list[str]: List of context hashes stored in object storage.
        """
        if not image._context_list:
            return []

        push_access_creds = object_storage_api.get_push_access()

        object_storage = ObjectStorage(
            push_access_creds.storage_url,
            push_access_creds.access_key,
            push_access_creds.secret,
            push_access_creds.session_token,
            push_access_creds.bucket,
        )
        context_hashes: list[str] = []
        for context in image._context_list:
            context_hash = object_storage.upload(
                context.source_path, push_access_creds.organization_id, context.archive_path
            )
            context_hashes.append(context_hash)

        return context_hashes
