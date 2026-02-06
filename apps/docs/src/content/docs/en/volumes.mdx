---
title: Volumes
---

import { TabItem, Tabs } from '@astrojs/starlight/components'
import Label from '@components/Label.astro'

Volumes are FUSE-based mounts that provide shared file access across Daytona Sandboxes. They enable sandboxes to read from large files instantly - no need to upload files manually to each sandbox. Volume data is stored in an S3-compatible object store.

- multiple volumes can be mounted to a single sandbox
- a single volume can be mounted to multiple sandboxes

## Create Volumes

Daytona provides volumes as a shared storage solution for sandboxes. To create a volume:

1. Navigate to [Daytona Volumes ↗](https://app.daytona.io/dashboard/volumes)
2. Click the **Create Volume** button
3. Enter the volume name

The following snippets demonstrate how to create a volume using the Daytona SDK:

<Tabs syncKey="language">
  <TabItem label="Python" icon="seti:python">

    ```python
    daytona = Daytona()
    volume = daytona.volume.create("my-awesome-volume")
    ```

  </TabItem>
  <TabItem label="TypeScript" icon="seti:typescript">

    ```typescript
    const daytona = new Daytona();
    const volume = await daytona.volume.create("my-awesome-volume");
    ```

  </TabItem>
  <TabItem label="Ruby" icon="seti:ruby">
    ```ruby
    daytona = Daytona::Daytona.new
    volume = daytona.volume.create("my-awesome-volume")
    ```
  </TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), and [Ruby SDK](/docs/en/ruby-sdk/) references:

> [**volume.get (Python SDK)**](/docs/python-sdk/sync/volume/#volumeserviceget)
>
> [**volume.get (TypeScript SDK)**](/docs/typescript-sdk/volume/#get)
>
> [**volume.get (Ruby SDK)**](/docs/ruby-sdk/volume/#get)

## Mount Volumes

Daytona provides an option to mount a volume to a sandbox. Once a volume is created, it can be mounted to a sandbox by specifying it in the `CreateSandboxFromSnapshotParams` object. Volume mount paths must meet the following requirements:

- **Must be absolute paths**: Mount paths must start with `/` (e.g., `/home/daytona/volume`)
- **Cannot be root directory**: Cannot mount to `/` or `//`
- **No relative path components**: Cannot contain `/../`, `/./`, or end with `/..` or `/.`
- **No consecutive slashes**: Cannot contain multiple consecutive slashes like `//` (except at the beginning)
- **Cannot mount to system directories**: The following system directories are prohibited: `/proc`, `/sys`, `/dev`, `/boot`, `/etc`, `/bin`, `/sbin`, `/lib`, `/lib64`

The following snippets demonstrate how to mount a volume to a sandbox:

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
import os
from daytona import CreateSandboxFromSnapshotParams, Daytona, VolumeMount

daytona = Daytona()

# Create a new volume or get an existing one
volume = daytona.volume.get("my-volume", create=True)

# Mount the volume to the sandbox
mount_dir_1 = "/home/daytona/volume"

params = CreateSandboxFromSnapshotParams(
    language="python",
    volumes=[VolumeMount(volume_id=volume.id, mount_path=mount_dir_1)],
)
sandbox = daytona.create(params)

# Mount a specific subpath within the volume
# This is useful for isolating data or implementing multi-tenancy
params = CreateSandboxFromSnapshotParams(
    language="python",
    volumes=[VolumeMount(volume_id=volume.id, mount_path=mount_dir_1, subpath="users/alice")],
)
sandbox2 = daytona.create(params)
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
import { Daytona } from '@daytonaio/sdk'
import path from 'path'

const daytona = new Daytona()

//  Create a new volume or get an existing one
const volume = await daytona.volume.get('my-volume', true)

// Mount the volume to the sandbox
const mountDir1 = '/home/daytona/volume'

const sandbox1 = await daytona.create({
  language: 'typescript',
  volumes: [{ volumeId: volume.id, mountPath: mountDir1 }],
})

// Mount a specific subpath within the volume
// This is useful for isolating data or implementing multi-tenancy
const sandbox2 = await daytona.create({
  language: 'typescript',
  volumes: [
    { volumeId: volume.id, mountPath: mountDir, subpath: 'users/alice' },
  ],
})
```

</TabItem>

<TabItem label="Ruby" icon="seti:ruby">
```ruby
require 'daytona'

daytona = Daytona::Daytona.new

# Create a new volume or get an existing one
volume = daytona.volume.get('my-volume', create: true)

# Mount the volume to the sandbox
mount_dir = '/home/daytona/volume'

params = Daytona::CreateSandboxFromSnapshotParams.new(
  language: Daytona::CodeLanguage::PYTHON,
  volumes: [DaytonaApiClient::SandboxVolume.new(volume_id: volume.id, mount_path: mount_dir)]
)
sandbox = daytona.create(params)

# Mount a specific subpath within the volume
# This is useful for isolating data or implementing multi-tenancy
params2 = Daytona::CreateSandboxFromSnapshotParams.new(
  language: Daytona::CodeLanguage::PYTHON,
  volumes: [DaytonaApiClient::SandboxVolume.new(
    volume_id: volume.id,
    mount_path: mount_dir,
    subpath: 'users/alice'
  )]
)
sandbox2 = daytona.create(params2)
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/) and [Ruby SDK](/docs/en/ruby-sdk/) references:

> [**CreateSandboxFromSnapshotParams (Python SDK)**](/docs/python-sdk/sync/daytona#createSandboxBaseParams)
>
> [**CreateSandboxFromSnapshotParams (TypeScript SDK)**](/docs/typescript-sdk/daytona#createSandboxBaseParams)
>
> [**CreateSandboxFromSnapshotParams (Ruby SDK)**](/docs/ruby-sdk/daytona#createsandboxfromsnapshotparams)

## Work with Volumes

Daytona provides an option to read from and write to the volume just like any other directory in the sandbox file system. Files written to the volume persist beyond the lifecycle of any individual sandbox.

The following snippet demonstrate how to read from and write to a volume:

<Tabs syncKey="language">
  <TabItem label="Python" icon="seti:python">

    ```python
    # Write to a file in the mounted volume
    with open("/home/daytona/volume/example.txt", "w") as f:
        f.write("Hello from Daytona volume!")

    # When you're done with the sandbox, you can remove it
    # The volume will persist even after the sandbox is removed
    sandbox.delete()
    ```

  </TabItem>
  <TabItem label="TypeScript" icon="seti:typescript">

    ```typescript
    import fs from 'fs/promises'

    // Write to a file in the mounted volume
    await fs.writeFile('/home/daytona/volume/example.txt', 'Hello from Daytona volume!')

    // When you're done with the sandbox, you can remove it
    // The volume will persist even after the sandbox is removed
    await daytona.delete(sandbox1)
    ```

  </TabItem>

  <TabItem label="Ruby" icon="seti:ruby">
    ```ruby
    # Write to a file in the mounted volume using the Sandbox file system API
    sandbox.fs.upload_file('Hello from Daytona volume!', '/home/daytona/volume/example.txt')

    # When you're done with the sandbox, you can remove it
    # The volume will persist even after the sandbox is removed
    daytona.delete(sandbox)
    ```
  </TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), and [Ruby SDK](/docs/en/ruby-sdk/) references.

## Get a Volume by name

Daytona provides an option to get a volume by its name.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
daytona = Daytona()
volume = daytona.volume.get("my-awesome-volume", create=True)
print(f"{volume.name} ({volume.id})")
```

</TabItem>

<TabItem label="TypeScript" icon="seti:typescript">

```typescript
const daytona = new Daytona()
const volume = await daytona.volume.get('my-awesome-volume', true)
console.log(`Volume ${volume.name} is in state ${volume.state}`)
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), and [Ruby SDK](/docs/en/ruby-sdk/) references:

> [**volume.get (Python SDK)**](/docs/en/python-sdk/sync/volume#volumeserviceget)
>
> [**volume.get (TypeScript SDK)**](/docs/en/typescript-sdk/volume#get)
>
> [**volume.get (Ruby SDK)**](/docs/en/ruby-sdk/volume#get)

## List Volumes

Daytona provides an option to list all volumes.

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
daytona = Daytona()
volumes = daytona.volume.list()
for volume in volumes:
    print(f"{volume.name} ({volume.id})")
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
const daytona = new Daytona()
const volumes = await daytona.volume.list()
console.log(`Found ${volumes.length} volumes`)
volumes.forEach(vol => console.log(`${vol.name} (${vol.id})`))
```

</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), and [Ruby SDK](/docs/en/ruby-sdk/) references:

> [**volume.list (Python SDK)**](/docs/en/python-sdk/sync/volume#volumeservicelist)
>
> [**volume.list (TypeScript SDK)**](/docs/en/typescript-sdk/volume#list)
>
> [**volume.list (Ruby SDK)**](/docs/en/ruby-sdk/volume#list)

## Delete Volumes

Daytona provides an option to delete a volume. Deleted volumes cannot be recovered.

The following snippet demonstrate how to delete a volume:

<Tabs syncKey="language">
<TabItem label="Python" icon="seti:python">

```python
volume = daytona.volume.get("my-volume", create=True)
daytona.volume.delete(volume)
```

</TabItem>
<TabItem label="TypeScript" icon="seti:typescript">

```typescript
const volume = await daytona.volume.get('my-volume', true)
await daytona.volume.delete(volume)
```

</TabItem>
<TabItem label="Ruby" icon="seti:ruby">
```ruby
volume = daytona.volume.get('my-volume', create: true)
daytona.volume.delete(volume)
```
</TabItem>
</Tabs>

For more information, see the [Python SDK](/docs/en/python-sdk/), [TypeScript SDK](/docs/en/typescript-sdk/), and [Ruby SDK](/docs/en/ruby-sdk/) references:

> [**volume.delete (Python SDK)**](/docs/python-sdk/sync/volume/#volumeservicedelete)
>
> [**volume.delete (TypeScript SDK)**](/docs/typescript-sdk/volume/#delete)
>
> [**volume.delete (Ruby SDK)**](/docs/ruby-sdk/volume/#delete)

## Limitations

Since volumes are FUSE-based mounts, they can not be used for applications that require block storage access (like database tables).
Volumes are generally slower for both read and write operations compared to the local sandbox file system.

## Pricing & Limits

Daytona Volumes are included at no additional cost. Each organization can create up to 100 volumes, and volume data does not count against your storage quota.

You can view your current volume usage in the [Daytona Dashboard ↗](https://app.daytona.io/dashboard/volumes).
