# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# import subprocess
import time

from daytona import Daytona, FilesystemEventType, WatchOptions


def basic_file_watching(sandbox):
    """Demonstrate basic file watching functionality."""
    print("Starting basic file watching...")

    # Watch a directory for all changes
    handle = sandbox.fs.watch_dir("/workspace", lambda event: print(f"{event.type}: {event.name}"))

    # Create some files to trigger events
    sandbox.fs.upload_file(b"Hello World!", "/workspace/test.txt")
    sandbox.fs.upload_file(b"Another file", "/workspace/another.txt")
    sandbox.fs.delete_file("/workspace/test.txt")

    # Wait a bit for events to be processed
    time.sleep(2)

    # Stop watching
    handle.close()
    print("Basic file watching completed")


def recursive_file_watching(sandbox):
    """Demonstrate recursive file watching with filtering."""
    print("Starting recursive file watching...")

    # Watch recursively with filtering
    def on_event(event):
        # Only log Python file changes
        if event.type == FilesystemEventType.WRITE and event.name.endswith(".py"):
            print(f"Python file changed: {event.name}")
        # Log all directory creation events
        if event.type == FilesystemEventType.CREATE and event.is_dir:
            print(f"Directory created: {event.name}")

    handle = sandbox.fs.watch_dir("/workspace", on_event, WatchOptions(recursive=True))

    # Create nested directory structure
    sandbox.fs.create_folder("/workspace/src", "755")
    sandbox.fs.upload_file(b'print("Hello from app!")', "/workspace/src/app.py")
    sandbox.fs.create_folder("/workspace/src/components", "755")
    sandbox.fs.upload_file(b"class Button:\n    pass", "/workspace/src/components/button.py")

    # Wait for events to be processed
    time.sleep(2)

    # Stop watching
    handle.close()
    print("Recursive file watching completed")


def file_watching_with_error_handling(sandbox):
    """Demonstrate file watching with error handling."""
    print("Starting file watching with error handling...")

    try:
        handle = sandbox.fs.watch_dir(
            "/workspace",
            lambda event: print(f"Event: {event.type} - {event.name} ({'dir' if event.is_dir else 'file'})"),
        )

        # Create files to trigger events
        sandbox.fs.upload_file(b"Testing error handling", "/workspace/error-test.txt")
        sandbox.fs.set_file_permissions("/workspace/error-test.txt", mode="644")
        sandbox.fs.delete_file("/workspace/error-test.txt")

        # Wait for events
        time.sleep(2)

        handle.close()
        print("File watching with error handling completed")
    except Exception as error:
        print(f"File watching error: {error}")


def file_watching_with_sync_callback(sandbox):
    """Demonstrate file watching with sync callback."""
    print("Starting file watching with sync callback...")

    event_count = 0

    def sync_callback(event):
        nonlocal event_count
        event_count += 1
        print(f"Event {event_count}: {event.type} - {event.name}")

        # Simulate sync processing
        time.sleep(0.1)

        # Log event details
        print(f"  Timestamp: {event.timestamp}")
        print(f"  Is Directory: {event.is_dir}")

    handle = sandbox.fs.watch_dir("/workspace", sync_callback)

    # Create multiple files quickly
    for i in range(3):
        sandbox.fs.upload_file(f"Content {i}".encode(), f"/workspace/sync-test-{i}.txt")

    # Wait for all events to be processed
    time.sleep(3)

    handle.close()
    print(f"Sync file watching completed. Processed {event_count} events.")


def main():
    """Main function demonstrating file watching capabilities."""
    daytona = Daytona()
    sandbox = daytona.create()

    # Local Hack for DNS resolution
    # subprocess.run(["hack/file-watching/dns_fix.sh"], check=True)

    try:
        basic_file_watching(sandbox)
        recursive_file_watching(sandbox)
        file_watching_with_error_handling(sandbox)
        file_watching_with_sync_callback(sandbox)
    except Exception as error:
        print(f"Error with file watching: {error}")
    finally:
        # Cleanup
        daytona.delete(sandbox)
        print("File watching demo completed. Sandbox cleaned up.")


if __name__ == "__main__":
    # Set up environment variables before running this example:
    # export DAYTONA_API_KEY="your-api-key"
    # export DAYTONA_API_URL="https://your-api-url"
    # export DAYTONA_TARGET="your-target"
    # export DAYTONA_ORGANIZATION_ID="your-organization-id"

    main()
