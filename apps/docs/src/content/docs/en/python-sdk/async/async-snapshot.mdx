---
title: "AsyncSnapshot"
hideTitleOnPage: true
---

## Snapshot

```python
class Snapshot(SyncSnapshotDto)
```

Represents a Daytona Snapshot which is a pre-configured sandbox.

**Attributes**:

- `id` _str_ - Unique identifier for the Snapshot.
- `organization_id` _str | None_ - Organization ID of the Snapshot.
- `general` _bool_ - Whether the Snapshot is general.
- `name` _str_ - Name of the Snapshot.
- `image_name` _str_ - Name of the Image of the Snapshot.
- `state` _str_ - State of the Snapshot.
- `size` _float | int | None_ - Size of the Snapshot.
- `entrypoint` _list[str] | None_ - Entrypoint of the Snapshot.
- `cpu` _float | int_ - CPU of the Snapshot.
- `gpu` _float | int_ - GPU of the Snapshot.
- `mem` _float | int_ - Memory of the Snapshot in GiB.
- `disk` _float | int_ - Disk of the Snapshot in GiB.
- `error_reason` _str | None_ - Error reason of the Snapshot.
- `created_at` _str_ - Timestamp when the Snapshot was created.
- `updated_at` _str_ - Timestamp when the Snapshot was last updated.
- `last_used_at` _str_ - Timestamp when the Snapshot was last used.


## AsyncSnapshotService

```python
class AsyncSnapshotService()
```

Service for managing Daytona Snapshots. Can be used to list, get, create and delete Snapshots.

#### AsyncSnapshotService.list

```python
@intercept_errors(message_prefix="Failed to list snapshots: ")
@with_instrumentation()
async def list(page: int | None = None,
               limit: int | None = None) -> PaginatedSnapshots
```

Returns paginated list of Snapshots.

**Arguments**:

- `page` _int | None_ - Page number for pagination (starting from 1).
- `limit` _int | None_ - Maximum number of items per page.
  

**Returns**:

- `PaginatedSnapshots` - Paginated list of Snapshots.
  

**Example**:

```python
async with AsyncDaytona() as daytona:
    result = await daytona.snapshot.list(page=2, limit=10)
    for snapshot in result.items:
        print(f"{snapshot.name} ({snapshot.image_name})")
```

#### AsyncSnapshotService.delete

```python
@intercept_errors(message_prefix="Failed to delete snapshot: ")
@with_instrumentation()
async def delete(snapshot: Snapshot) -> None
```

Delete a Snapshot.

**Arguments**:

- `snapshot` _Snapshot_ - Snapshot to delete.
  

**Example**:

```python
async with AsyncDaytona() as daytona:
    snapshot = await daytona.snapshot.get("test-snapshot")
    await daytona.snapshot.delete(snapshot)
    print("Snapshot deleted")
```

#### AsyncSnapshotService.get

```python
@intercept_errors(message_prefix="Failed to get snapshot: ")
@with_instrumentation()
async def get(name: str) -> Snapshot
```

Get a Snapshot by name.

**Arguments**:

- `name` _str_ - Name of the Snapshot to get.
  

**Returns**:

- `Snapshot` - The Snapshot object.
  

**Example**:

```python
async with AsyncDaytona() as daytona:
    snapshot = await daytona.snapshot.get("test-snapshot-name")
    print(f"{snapshot.name} ({snapshot.image_name})")
```

#### AsyncSnapshotService.create

```python
@intercept_errors(message_prefix="Failed to create snapshot: ")
@with_timeout()
@with_instrumentation()
async def create(params: CreateSnapshotParams,
                 *,
                 on_logs: Callable[[str], None] | None = None,
                 timeout: float | None = 0) -> Snapshot
```

Creates and registers a new snapshot from the given Image definition.

**Arguments**:

- `params` _CreateSnapshotParams_ - Parameters for snapshot creation.
- `on_logs` _Callable[[str], None]_ - This callback function handles snapshot creation logs.
- `timeout` _float | None_ - Default is no timeout. Timeout in seconds (0 means no timeout).

**Example**:

```python
image = Image.debianSlim('3.12').pipInstall('numpy')
daytona.snapshot.create(
    CreateSnapshotParams(name='my-snapshot', image=image),
    on_logs=lambda chunk: print(chunk, end=""),
)
```

#### AsyncSnapshotService.activate

```python
@with_instrumentation()
async def activate(snapshot: Snapshot) -> Snapshot
```

Activate a snapshot.

**Arguments**:

- `snapshot` _Snapshot_ - The Snapshot instance.

**Returns**:

- `Snapshot` - The activated Snapshot instance.

#### AsyncSnapshotService.process\_image\_context

```python
@staticmethod
@with_instrumentation()
async def process_image_context(object_storage_api: ObjectStorageApi,
                                image: Image) -> list[str]
```

Processes the image context by uploading it to object storage.

**Arguments**:

- `image` _Image_ - The Image instance.

**Returns**:

- `List[str]` - List of context hashes stored in object storage.

## PaginatedSnapshots

```python
class PaginatedSnapshots(PaginatedSnapshotsDto)
```

Represents a paginated list of Daytona Snapshots.

**Attributes**:

- `items` _list[Snapshot]_ - List of Snapshot instances in the current page.
- `total` _int_ - Total number of Snapshots across all pages.
- `page` _int_ - Current page number.
- `total_pages` _int_ - Total number of pages available.

## CreateSnapshotParams

```python
class CreateSnapshotParams(BaseModel)
```

Parameters for creating a new snapshot.

**Attributes**:

- `name` _str_ - Name of the snapshot.
- `image` _str | Image_ - Image of the snapshot. If a string is provided,
  it should be available on some registry. If an Image instance is provided,
  it will be used to create a new image in Daytona.
- `resources` _Resources | None_ - Resources of the snapshot.
- `entrypoint` _list[str] | None_ - Entrypoint of the snapshot.
- `region_id` _str | None_ - ID of the region where the snapshot will be available.
  Defaults to organization default region if not specified.

