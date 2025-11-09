import time

from daytona import CreateSandboxFromImageParams, Daytona, DiskState, SandboxState
from daytona.common.errors import DaytonaNotFoundError


def wait_for_disk_state(daytona, disk_id, target_states, timeout_seconds=60, poll_interval=2, verbose=True):
    """
    Wait for a disk to reach one of the target states.

    Args:
        daytona: The Daytona client
        disk_id: The disk ID to monitor
        target_states: List of acceptable target states (e.g., [DiskState.FRESH, DiskState.STORED])
        timeout_seconds: Maximum time to wait in seconds (default: 60)
        poll_interval: Time between state checks in seconds (default: 2)
        verbose: Whether to print status messages (default: True)

    Returns:
        bool: True if disk reached target state, False if timeout occurred
    """
    start_time = time.time()

    try:
        while time.time() - start_time < timeout_seconds:
            disk = daytona.disk.get(disk_id)
            if verbose:
                print(f"⏳ Disk {disk_id[:8]}... state: {disk.state}")
            if disk.state in target_states:
                if verbose:
                    print(f"✅ Disk reached state: {disk.state}")
                return True
            time.sleep(poll_interval)
        if verbose:
            print(f"⏰ Timeout waiting for disk to reach one of {target_states}")
        return False
    except Exception as e:
        if verbose:
            print(f"❌ Error waiting for disk state: {e}")
        return False


def wait_for_sandbox_state(sandbox, target_state, timeout_seconds=60, poll_interval=1, verbose=True):
    """
    Wait for a sandbox to reach a specific state.

    Args:
        sandbox: The sandbox object to monitor
        target_state: The target state to wait for (e.g., SandboxState.DESTROYED)
        timeout_seconds: Maximum time to wait in seconds (default: 60)
        poll_interval: Time between state checks in seconds (default: 1)
        verbose: Whether to print status messages (default: True)

    Returns:
        bool: True if sandbox reached target state, False if timeout or error occurred
    """
    if verbose:
        print(f"⏳ Waiting for sandbox {sandbox.name} to reach state: {target_state}")

    start_time = time.time()

    try:
        while sandbox.state != target_state:
            if time.time() - start_time > timeout_seconds:
                if verbose:
                    print(f"⏰ Timeout waiting for sandbox to reach {target_state}")
                return False

            sandbox.refresh_data()
            time.sleep(poll_interval)

    except Exception as e:
        # Check if it's a 404 error (sandbox not found, which means it's deleted)
        if isinstance(e, DaytonaNotFoundError):
            if verbose:
                print(f"✅ Sandbox not found - assuming {target_state}")
            return True
        if verbose:
            print(f"❌ Error refreshing sandbox data: {e}")
        return False

    if verbose:
        print(f"✅ Sandbox {sandbox.name} reached state: {sandbox.state}")

    return sandbox.state == target_state


def main():
    daytona = Daytona()

    print("🚀 Starting Disk Management Example")
    print("=====================================")

    # List all existing disks
    print("\n📋 Listing all disks...")
    existing_disks = daytona.disk.list()
    print(f"Found {len(existing_disks)} existing disks:")
    for disk in existing_disks:
        print(f"  - {disk.name} ({disk.id}) - {disk.size}GB - State: {disk.state}")

    # Create a new disk
    print("\n💾 Creating a new disk...")
    disk_name = f"example-disk-{int(time.time())}"
    disk_size = 20  # 20GB
    disk = daytona.disk.create(disk_name, disk_size)
    print(f"✅ Created disk: {disk.name} ({disk.id}) - {disk.size}GB - State: {disk.state}")

    # Get the disk by ID
    print("\n🔍 Getting disk details...")
    retrieved_disk = daytona.disk.get(disk.id)
    print(f"✅ Retrieved disk: {retrieved_disk.name} - {retrieved_disk.size}GB - State: {retrieved_disk.state}")

    # List disks again to see the new one
    print("\n📋 Listing disks after creation...")
    updated_disks = daytona.disk.list()
    print(f"Found {len(updated_disks)} disks:")
    for d in updated_disks:
        print(f"  - {d.name} ({d.id}) - {d.size}GB - State: {d.state}")

    # Create a sandbox with the disk attached
    print("\n🏗️ Creating a sandbox...")
    params = CreateSandboxFromImageParams(
        image="ubuntu:22.04",
        sandbox_name=f"example-sandbox-{int(time.time())}",
        disk_id=disk.id,
        language="python",
    )
    sandbox = daytona.create(params, timeout=150, on_snapshot_create_logs=print)
    print(f"✅ Created sandbox: {sandbox.name} ({sandbox.id}) - State: {sandbox.state}")

    # Create a new file in the sandbox
    print("\n📋 Creating a new file in the sandbox...")
    sandbox.fs.upload_file(b"Hello, World!", "/workspace/new-file.txt")
    print(f"✅ Created file: {sandbox.name} ({sandbox.id}) - State: {sandbox.state}")

    print("Sleeping for 3 seconds...")
    time.sleep(3)

    # List files in the sandbox
    print("\n📋 Listing files in the sandbox...")
    files = sandbox.fs.list_files("/workspace")
    print(f"Found {len(files)} files in the sandbox:")
    for file in files:
        print(f"  - {file.name} - {file.size} - {file.is_dir}")

    sandbox.stop()

    # Wait sandbox to be stopped
    print("\n⏳ Waiting for sandbox to be stopped...")

    if not wait_for_sandbox_state(sandbox, SandboxState.STOPPED, timeout_seconds=60):
        print("❌ Failed to wait for sandbox stop")
        return

    # Fork the disk
    forked_disk = daytona.disk.fork(disk, f"example-disk-{int(time.time())}-2")

    # Wait for disk fork to complete
    print("\n⏳ Waiting for forked disk to be ready...")
    # Note: We need to catch the disk quickly while it's in DETACHED state
    # before the background job pushes it to cloud storage (PENDING_PUSH)
    # DETACHED disks can be attached immediately
    if not wait_for_disk_state(
        daytona,
        forked_disk.id,
        [DiskState.FRESH, DiskState.STORED, DiskState.DETACHED, "pending_push"],
        timeout_seconds=30,
    ):
        print("❌ Failed to wait for forked disk to be ready")
        return

    print("\n🏗️ Creating a new sandbox with the forked disk...")
    params = CreateSandboxFromImageParams(
        image="ubuntu:22.04",
        sandbox_name=f"example-sandbox-{int(time.time())}-2",
        disk_id=forked_disk.id,
        language="python",
    )
    new_sandbox = daytona.create(params, timeout=150, on_snapshot_create_logs=print)
    print(f"✅ Created sandbox: {new_sandbox.name} ({new_sandbox.id}) - State: {new_sandbox.state}")

    # List files in the sandbox
    print("\n📋 Listing files in the new sandbox with the forked disk...")
    files = new_sandbox.fs.list_files("/workspace")
    print(f"Found {len(files)} files in the sandbox:")
    for file in files:
        print(f"  - {file.name} - {file.size} - {file.is_dir}")

    # add a new file to the new sandbox
    print("\n📋 Adding a new file to the new sandbox with the forked disk...")
    new_sandbox.fs.upload_file(b"Hello, World!", "/workspace/new-file-2.txt")
    print(f"✅ Added file: {new_sandbox.name} ({new_sandbox.id}) - State: {new_sandbox.state}")

    # stop the new sandbox
    print("\n📋 Stopping the new sandbox with the forked disk...")
    new_sandbox.stop()
    print(f"✅ Stopped sandbox: {new_sandbox.name} ({new_sandbox.id}) - State: {new_sandbox.state}")

    # wait for the new sandbox to be stopped
    print("\n⏳ Waiting for the new sandbox to be stopped...")
    if not wait_for_sandbox_state(new_sandbox, SandboxState.STOPPED, timeout_seconds=60):
        print("❌ Failed to wait for new sandbox stop")
        return

    # fork the forked disk
    print("\n📋 Forking the forked disk...")
    forked_disk_2 = daytona.disk.fork(forked_disk, f"example-disk-{int(time.time())}-3")
    print(f"✅ Forked disk: {forked_disk_2.name} ({forked_disk_2.id}) - State: {forked_disk_2.state}")

    # wait for the forked disk to be ready
    print("\n⏳ Waiting for the forked disk to be ready...")
    if not wait_for_disk_state(
        daytona,
        forked_disk_2.id,
        [DiskState.FRESH, DiskState.STORED, DiskState.DETACHED, "pending_push"],
        timeout_seconds=30,
    ):
        print("❌ Failed to wait for forked disk to be ready")
        return

    # create a new sandbox with the forked disk
    print("\n🏗️ Creating a new sandbox with the forked disk 2...")
    params = CreateSandboxFromImageParams(
        image="ubuntu:22.04",
        sandbox_name=f"example-sandbox-{int(time.time())}-3",
        disk_id=forked_disk_2.id,
        language="python",
    )
    new_sandbox_2 = daytona.create(params, timeout=150, on_snapshot_create_logs=print)
    print(f"✅ Created sandbox: {new_sandbox_2.name} ({new_sandbox_2.id}) - State: {new_sandbox_2.state}")

    # list files in the new sandbox
    print("\n📋 Listing files in the new sandbox with the forked disk 2...")
    files = new_sandbox_2.fs.list_files("/workspace")
    print(f"Found {len(files)} files in the sandbox:")
    for file in files:
        print(f"  - {file.name} - {file.size} - {file.is_dir}")

    sandbox.delete()

    # Wait for sandbox to be deleted
    if not wait_for_sandbox_state(sandbox, SandboxState.DESTROYED, timeout_seconds=60):
        print("❌ Failed to wait for sandbox deletion")
        return

    new_sandbox.delete()

    # Wait for the second sandbox to be deleted
    print("\n⏳ Waiting for the second sandbox to be deleted...")

    if not wait_for_sandbox_state(new_sandbox, SandboxState.DESTROYED, timeout_seconds=60):
        print("❌ Failed to wait for second sandbox deletion")
        return

    new_sandbox_2.delete()

    # Wait for the second sandbox 2 to be deleted
    print("\n⏳ Waiting for the second sandbox 2 to be deleted...")
    if not wait_for_sandbox_state(new_sandbox_2, SandboxState.DESTROYED, timeout_seconds=60):
        print("❌ Failed to wait for second sandbox 2 deletion")
        return

    # Wait for the disk to be detached after the second sandbox is deleted
    time.sleep(10)

    # Delete the disk
    print("\n🗑️  Deleting the original disk...")
    daytona.disk.delete(disk)
    print(f"✅ Deleted disk: {disk.name}")

    # Delete the forked disk
    print("\n🗑️  Deleting the forked disk...")
    daytona.disk.delete(forked_disk)
    print(f"✅ Forked disk: {forked_disk.name}")

    # Delete the forked disk 2
    print("\n🗑️  Deleting the forked disk 2...")
    daytona.disk.delete(forked_disk)
    print(f"✅ Forked disk 2: {forked_disk_2.name}")

    # Final list to confirm deletion
    print("\n📋 Final disk list...")
    final_disks = daytona.disk.list()
    print(f"Found {len(final_disks)} disks after cleanup")

    print("\n🎉 Disk management example completed successfully!")


if __name__ == "__main__":
    main()
