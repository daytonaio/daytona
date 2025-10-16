import sys
import time

from daytona import CreateSandboxFromImageParams, Daytona


def main():
    daytona = Daytona()

    print("🚀 Starting Disk Management Example")
    print("=====================================")

    try:
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

        # List files in the sandbox
        print("\n📋 Listing files in the sandbox...")
        files = sandbox.fs.list_files("/workspace")
        print(f"Found {len(files)} files in the sandbox:")
        for file in files:
            print(f"  - {file.name} - {file.size} - {file.is_dir}")

        sandbox.delete()

        # Wait a moment before deletion
        print("\n⏳ Waiting 5 seconds before cleanup...")
        time.sleep(5)

        # Delete the disk
        print("\n🗑️  Deleting the disk...")
        daytona.disk.delete(disk)
        print(f"✅ Deleted disk: {disk.name}")

        # Final list to confirm deletion
        print("\n📋 Final disk list...")
        final_disks = daytona.disk.list()
        print(f"Found {len(final_disks)} disks after cleanup")

        print("\n🎉 Disk management example completed successfully!")

    except Exception as error:
        print(f"❌ Error during disk management: {error}")
        sys.exit(1)


if __name__ == "__main__":
    main()
