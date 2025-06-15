import os

from daytona import CreateSandboxFromSnapshotParams, Daytona, VolumeMount


def main():
    daytona = Daytona()

    # Create a new volume or get an existing one
    volume = daytona.volume.get("my-volume", create=True)

    # Mount the volume to the sandbox
    mount_dir_1 = "/home/daytona/volume"

    params = CreateSandboxFromSnapshotParams(
        language="python",
        volumes=[VolumeMount(volumeId=volume.id, mountPath=mount_dir_1)],
    )
    sandbox = daytona.create(params)

    # Create a new directory in the mount directory
    new_dir = os.path.join(mount_dir_1, "new-dir")
    sandbox.fs.create_folder(new_dir, "755")

    # Create a new file in the mount directory
    new_file = os.path.join(mount_dir_1, "new-file.txt")
    sandbox.fs.upload_file(b"Hello, World!", new_file)

    # Create a new sandbox with the same volume
    # and mount it to the different path
    mount_dir_2 = "/home/daytona/my-files"

    params = CreateSandboxFromSnapshotParams(
        language="python",
        volumes=[VolumeMount(volumeId=volume.id, mountPath=mount_dir_2)],
    )
    sandbox2 = daytona.create(params)

    # List files in the mount directory
    files = sandbox2.fs.list_files(mount_dir_2)
    print("Files:", files)

    # Get the file from the mount directory
    file = sandbox2.fs.download_file(os.path.join(mount_dir_2, "new-file.txt"))
    print("File:", file)

    # Cleanup
    daytona.delete(sandbox)
    daytona.delete(sandbox2)
    # daytona.volume.delete(volume)


if __name__ == "__main__":
    main()
