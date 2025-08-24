# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import asyncio

from daytona import AsyncDaytona, FilesystemEventType, WatchOptions

# import os
# import subprocess


async def basic_file_watching(sandbox):
    """Demonstrate basic file watching functionality."""
    print("Starting basic file watching...")

    # Watch a directory for all changes
    handle = await sandbox.fs.watch_dir("/workspace", lambda event: print(f"{event.type}: {event.name}"))

    # Create some files to trigger events
    await sandbox.fs.upload_file(b"Hello World!", "/workspace/test.txt")
    await sandbox.fs.upload_file(b"Another file", "/workspace/another.txt")
    await sandbox.fs.delete_file("/workspace/test.txt")

    # Wait a bit for events to be processed
    await asyncio.sleep(2)

    # Stop watching
    await handle.close()
    print("Basic file watching completed")


async def recursive_file_watching(sandbox):
    """Demonstrate recursive file watching with filtering."""
    print("Starting recursive file watching...")

    # Watch recursively with filtering
    def on_event(event):
        # Only log TypeScript file changes
        if event.type == FilesystemEventType.WRITE and event.name.endswith(".py"):
            print(f"Python file changed: {event.name}")
        # Log all directory creation events
        if event.type == FilesystemEventType.CREATE and event.is_dir:
            print(f"Directory created: {event.name}")

    handle = await sandbox.fs.watch_dir("/workspace", on_event, WatchOptions(recursive=True))

    # Create nested directory structure
    await sandbox.fs.create_folder("/workspace/src", "755")
    await sandbox.fs.upload_file(b'print("Hello from app!")', "/workspace/src/app.py")
    await sandbox.fs.create_folder("/workspace/src/components", "755")
    await sandbox.fs.upload_file(b"class Button:\n    pass", "/workspace/src/components/button.py")

    # Wait for events to be processed
    await asyncio.sleep(2)

    # Stop watching
    await handle.close()
    print("Recursive file watching completed")


async def file_watching_with_error_handling(sandbox):
    """Demonstrate file watching with error handling."""
    print("Starting file watching with error handling...")

    try:
        handle = await sandbox.fs.watch_dir(
            "/workspace",
            lambda event: print(f"Event: {event.type} - {event.name} ({'dir' if event.is_dir else 'file'})"),
        )

        # Create files to trigger events
        await sandbox.fs.upload_file(b"Testing error handling", "/workspace/error-test.txt")
        await sandbox.fs.set_file_permissions("/workspace/error-test.txt", mode="644")
        await sandbox.fs.delete_file("/workspace/error-test.txt")

        # Wait for events
        await asyncio.sleep(2)

        await handle.close()
        print("File watching with error handling completed")
    except Exception as error:
        print(f"File watching error: {error}")


async def file_watching_with_async_callback(sandbox):
    """Demonstrate file watching with async callback."""
    print("Starting file watching with async callback...")

    event_count = 0

    async def async_callback(event):
        nonlocal event_count
        event_count += 1
        print(f"Event {event_count}: {event.type} - {event.name}")

        # Simulate async processing
        await asyncio.sleep(0.1)

        # Log event details
        print(f"  Timestamp: {event.timestamp}")
        print(f"  Is Directory: {event.is_dir}")

    handle = await sandbox.fs.watch_dir("/workspace", async_callback)

    # Create multiple files quickly
    for i in range(3):
        await sandbox.fs.upload_file(f"Content {i}".encode(), f"/workspace/async-test-{i}.txt")

    # Wait for all events to be processed
    await asyncio.sleep(3)

    await handle.close()
    print(f"Async file watching completed. Processed {event_count} events.")


async def main():
    """Main function demonstrating file watching capabilities."""
    async with AsyncDaytona() as daytona:
        sandbox = await daytona.create()

        # Local Hack for DNS resolution - run after sandbox creation
        # subprocess.run(["hack/file-watching/dns_fix.sh"], check=True)

        try:
            await basic_file_watching(sandbox)
            await recursive_file_watching(sandbox)
            await file_watching_with_error_handling(sandbox)
            await file_watching_with_async_callback(sandbox)
        except Exception as error:
            print(f"Error with file watching: {error}")
        finally:
            # Cleanup
            await daytona.delete(sandbox)
            print("File watching demo completed. Sandbox cleaned up.")


if __name__ == "__main__":
    # Set up environment variables before running this example:
    # export DAYTONA_API_KEY="your-api-key"
    # export DAYTONA_API_URL="https://your-api-url"
    # export DAYTONA_TARGET="your-target"
    # export DAYTONA_ORGANIZATION_ID="your-organization-id"

    asyncio.run(main())
