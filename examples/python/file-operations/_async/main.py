import asyncio
import json
import os
from datetime import datetime

from daytona import AsyncDaytona, CreateSandboxFromSnapshotParams, FileDownloadRequest, FileUpload


async def main():
    async with AsyncDaytona() as daytona:
        params = CreateSandboxFromSnapshotParams(
            language="python",
        )

        # First, create a sandbox
        sandbox = await daytona.create(params)
        print(f"Created sandbox with ID: {sandbox.id}")

        # List files in the sandbox
        files = await sandbox.fs.list_files(".")
        print("Initial files:", files)

        # Create a new directory in the sandbox
        new_dir = "project-files"
        await sandbox.fs.create_folder(new_dir, "755")

        # Create a local file for demonstration
        local_file_path = "local-example.txt"
        with open(local_file_path, "w", encoding="utf-8") as f:
            _ = f.write("This is a local file created for demonstration purposes")

        # Create a configuration file with JSON data
        config_data = json.dumps(
            {"name": "project-config", "version": "1.0.0", "settings": {"debug": True, "maxConnections": 10}}, indent=2
        )

        # Upload multiple files at once - both from local path and from bytes
        await sandbox.fs.upload_files(
            [
                FileUpload(source=local_file_path, destination=os.path.join(new_dir, "example.txt")),
                FileUpload(source=config_data.encode("utf-8"), destination=os.path.join(new_dir, "config.json")),
                FileUpload(
                    source=b'#!/bin/bash\necho "Hello from script!"\nexit 0',
                    destination=os.path.join(new_dir, "script.sh"),
                ),
            ]
        )

        # Execute commands on the sandbox to verify files and make them executable
        print("Verifying uploaded files:")
        ls_result = await sandbox.process.exec(f"ls -la {new_dir}")
        print(ls_result.result)

        # Make the script executable
        _ = await sandbox.process.exec(f"chmod +x {os.path.join(new_dir, 'script.sh')}")

        # Run the script
        print("Running script:")
        script_result = await sandbox.process.exec(f"{os.path.join(new_dir, 'script.sh')}")
        print(script_result.result)

        # Search for files in the project
        matches = await sandbox.fs.search_files(new_dir, "*.json")
        print("JSON files found:", matches)

        # Replace content in config file
        _ = await sandbox.fs.replace_in_files([os.path.join(new_dir, "config.json")], '"debug": true', '"debug": false')

        # Download multiple files - mix of local file and memory download
        print("Downloading multiple files:")
        download_results = await sandbox.fs.download_files(
            [
                FileDownloadRequest(source=os.path.join(new_dir, "config.json"), destination="local-config.json"),
                FileDownloadRequest(source=os.path.join(new_dir, "example.txt")),
                FileDownloadRequest(source=os.path.join(new_dir, "script.sh"), destination="local-script.sh"),
            ]
        )

        for result in download_results:
            if result.error:
                print(f"Error downloading {result.source}: {result.error}")
            elif isinstance(result.result, str):
                print(f"Downloaded {result.source} to {result.result}")
            elif result.result:
                print(f"Downloaded {result.source} to memory ({len(result.result)} bytes)")
            else:
                print(f"Downloaded {result.source} to None (unknown result type)")

        # Single file download example
        print("Single file download example:")
        config_content = await sandbox.fs.download_file(os.path.join(new_dir, "config.json"))
        print("Config content:", config_content.decode("utf-8"))

        # Create a report of all operations
        report_data = f"""
        Project Files Report:
        ---------------------
        Time: {datetime.now().isoformat()}
        Files: {len(matches.files)} JSON files found
        Config: {'Production mode' if b'"debug": false' in config_content else 'Debug mode'}
        Script: {'Executed successfully' if script_result.exit_code == 0 else 'Failed'}
        """.strip()

        # Save the report
        await sandbox.fs.upload_file(report_data.encode("utf-8"), os.path.join(new_dir, "report.txt"))

        # Clean up local file
        os.remove(local_file_path)
        if os.path.exists("local-config.json"):
            os.remove("local-config.json")
        if os.path.exists("local-script.sh"):
            os.remove("local-script.sh")

        # Delete the sandbox
        await daytona.delete(sandbox)


if __name__ == "__main__":
    asyncio.run(main())
