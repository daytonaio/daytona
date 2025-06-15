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
            volumes=[VolumeMount(volumeId=volume.id, mountPath=mount_dir_1)],
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
            volumes=[VolumeMount(volumeId=volume.id, mountPath=mount_dir_2)],
        )
        sandbox2 = await daytona.create(params)

        # List files in the mount directory
        files = await sandbox2.fs.list_files(mount_dir_2)
        print("Files:", files)

        # Get the file from the mount directory
        file = await sandbox2.fs.download_file(os.path.join(mount_dir_2, "new-file.txt"))
        print("File:", file)

        # Cleanup
        await daytona.delete(sandbox)
        await daytona.delete(sandbox2)
        # daytona.volume.delete(volume)


if __name__ == "__main__":
    asyncio.run(main())
