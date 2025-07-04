import json
import os
from datetime import datetime

from daytona import CreateSandboxFromSnapshotParams, Daytona, FileDownloadRequest, FileUpload


def main():
    daytona = Daytona()
    params = CreateSandboxFromSnapshotParams(
        language="python",
    )

    # First, create a sandbox
    sandbox = daytona.create(params)
    print(f"Created sandbox with ID: {sandbox.id}")

    # List files in the sandbox
    files = sandbox.fs.list_files("~")
    print("Initial files:", files)

    # Create a new directory in the sandbox
    new_dir = "~/project-files"
    sandbox.fs.create_folder(new_dir, "755")

    # Create a local file for demonstration
    local_file_path = "local-example.txt"
    with open(local_file_path, "w", encoding="utf-8") as f:
        f.write("This is a local file created for demonstration purposes")

    # Create a configuration file with JSON data
    config_data = json.dumps(
        {"name": "project-config", "version": "1.0.0", "settings": {"debug": True, "maxConnections": 10}}, indent=2
    )

    # Upload multiple files at once - both from local path and from bytes
    sandbox.fs.upload_files(
        [
            FileUpload(source=local_file_path, destination=os.path.join(new_dir, "example.txt")),
            FileUpload(source=config_data.encode("utf-8"), destination=os.path.join(new_dir, "config.json")),
            FileUpload(
                source=b'#!/bin/bash\necho "Hello from script!"\nexit 0', destination=os.path.join(new_dir, "script.sh")
            ),
        ]
    )

    # Execute commands on the sandbox to verify files and make them executable
    print("Verifying uploaded files:")
    ls_result = sandbox.process.exec(f"ls -la {new_dir}")
    print(ls_result.result)

    # Download the modified config file
    print("Downloading updated config file:")
    download_results = sandbox.fs.download_files(
        [
            FileDownloadRequest(
                source=os.path.join(new_dir, "config.json"),
                destination="/workspaces/daytona/config.json",
            ),
            FileDownloadRequest(
                source=os.path.join(new_dir, "exampleee.txt"),
                destination="/workspaces/daytona/example.txt",
            ),
            FileDownloadRequest(
                source=os.path.join(new_dir, "script.sh"),
            ),
        ]
    )
    for f in download_results:
        print(f.result)

    # Clean up local file
    os.remove(local_file_path)

    # Delete the sandbox
    daytona.delete(sandbox)


if __name__ == "__main__":
    main()
