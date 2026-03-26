---
title: "Volume"
hideTitleOnPage: true
---

## Volume

```python
class Volume(VolumeDto)
```

Represents a Daytona Volume which is a shared storage volume for Sandboxes.

**Attributes**:

- `id` _str_ - Unique identifier for the Volume.
- `name` _str_ - Name of the Volume.
- `organization_id` _str_ - Organization ID of the Volume.
- `state` _str_ - State of the Volume.
- `created_at` _str_ - Date and time when the Volume was created.
- `updated_at` _str_ - Date and time when the Volume was last updated.
- `last_used_at` _str_ - Date and time when the Volume was last used.


## VolumeService

```python
class VolumeService()
```

Service for managing Daytona Volumes. Can be used to list, get, create and delete Volumes.

#### VolumeService.list

```python
def list() -> list[Volume]
```

List all Volumes.

**Returns**:

- `list[Volume]` - List of all Volumes.
  

**Example**:

```python
daytona = Daytona()
volumes = daytona.volume.list()
for volume in volumes:
    print(f"{volume.name} ({volume.id})")
```

#### VolumeService.get

```python
@with_instrumentation()
def get(name: str, create: bool = False) -> Volume
```

Get a Volume by name.

**Arguments**:

- `name` _str_ - Name of the Volume to get.
- `create` _bool_ - If True, create a new Volume if it doesn't exist.
  

**Returns**:

- `Volume` - The Volume object.
  

**Example**:

```python
daytona = Daytona()
volume = daytona.volume.get("test-volume-name", create=True)
print(f"{volume.name} ({volume.id})")
```

#### VolumeService.create

```python
@with_instrumentation()
def create(name: str) -> Volume
```

Create a new Volume.

**Arguments**:

- `name` _str_ - Name of the Volume to create.
  

**Returns**:

- `Volume` - The Volume object.
  

**Example**:

```python
daytona = Daytona()
volume = daytona.volume.create("test-volume")
print(f"{volume.name} ({volume.id}); state: {volume.state}")
```

#### VolumeService.delete

```python
@with_instrumentation()
def delete(volume: Volume) -> None
```

Delete a Volume.

**Arguments**:

- `volume` _Volume_ - Volume to delete.
  

**Example**:

```python
daytona = Daytona()
volume = daytona.volume.get("test-volume")
daytona.volume.delete(volume)
print("Volume deleted")
```

## VolumeMount

```python
class VolumeMount(ApiVolumeMount, AsyncApiVolumeMount)
```

Represents a Volume mount configuration for a Sandbox.

**Attributes**:

- `volume_id` _str_ - ID of the volume to mount.
- `mount_path` _str_ - Path where the volume will be mounted in the sandbox.
- `subpath` _str | None_ - Optional S3 subpath/prefix within the volume to mount.
  When specified, only this prefix will be accessible. When omitted,
  the entire volume is mounted.

