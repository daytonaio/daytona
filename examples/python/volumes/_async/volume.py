import asyncio
import os

from daytona import AsyncDaytona, CreateSandboxFromSnapshotParams, VolumeMount


async def main():
    async with AsyncDaytona() as daytona:
        # Create a new volume or get an existing one
        volume = await daytona.volume.get("my-volume", create=True)

        # Mount the volume to the sandbox
        mount_dir_1 = "/home/daytona/volume"

        params = CreateSandboxFromSnapshotParams(
            language="python",
            volumes=[VolumeMount(volume_id=volume.id, mount_path=mount_dir_1)],
        )
        sandbox = await daytona.create(params)

        # Create a new directory in the mount directory
        new_dir = os.path.join(mount_dir_1, "new-dir")
        await sandbox.fs.create_folder(new_dir, "755")

        # Create a new file in the mount directory
        new_file = os.path.join(mount_dir_1, "new-file.txt")
        await sandbox.fs.upload_file(b"Hello, World!", new_file)

        # Create a new sandbox with the same volume
        # and mount it to the different path
        mount_dir_2 = "/home/daytona/my-files"

        params = CreateSandboxFromSnapshotParams(
            language="python",
            volumes=[VolumeMount(volume_id=volume.id, mount_path=mount_dir_2)],
        )
        sandbox2 = await daytona.create(params)

        # List files in the mount directory
        files = await sandbox2.fs.list_files(mount_dir_2)
        print("Files:", files)

        # Get the file from the mount directory
        file = await sandbox2.fs.download_file(os.path.join(mount_dir_2, "new-file.txt"))
        print("File:", file)

        # Mount a specific subpath within the volume
        # This is useful for isolating data or implementing multi-tenancy
        mount_dir_3 = "/home/daytona/subpath"

        params = CreateSandboxFromSnapshotParams(
            language="python",
            volumes=[VolumeMount(volume_id=volume.id, mount_path=mount_dir_3, subpath="users/alice")],
        )
        sandbox3 = await daytona.create(params)

        # This sandbox will only see files within the 'users/alice' subpath
        # Create a file in the subpath
        subpath_file = os.path.join(mount_dir_3, "alice-file.txt")
        await sandbox3.fs.upload_file(b"Hello from Alice's subpath!", subpath_file)

        # The file is stored at: volume-root/users/alice/alice-file.txt
        # but appears at: /home/daytona/subpath/alice-file.txt in the sandbox

        # Cleanup
        await daytona.delete(sandbox)
        await daytona.delete(sandbox2)
        await daytona.delete(sandbox3)
        # daytona.volume.delete(volume)


if __name__ == "__main__":
    asyncio.run(main())
