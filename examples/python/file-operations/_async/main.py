import asyncio
import json
import os
from datetime import datetime

from daytona import AsyncDaytona, CreateSandboxFromSnapshotParams, FileUpload
from daytona._async.filesystem import SearchRequest


async def main():
    async with AsyncDaytona() as daytona:
        params = CreateSandboxFromSnapshotParams(
            language="python",
        )

        # First, create a sandbox
        sandbox = await daytona.create(params)
        print(f"Created sandbox with ID: {sandbox.id}")

        # List files in the sandbox
        files = await sandbox.fs.list_files("~")
        print("Initial files:", files)

        # Create a new directory in the sandbox
        new_dir = "~/project-files"
        await sandbox.fs.create_folder(new_dir, "755")

        # Create a local file for demonstration
        local_file_path = "local-example.txt"
        with open(local_file_path, "w", encoding="utf-8") as f:
            f.write("This is a local file created for demonstration purposes")

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
        await sandbox.process.exec(f"chmod +x {os.path.join(new_dir, 'script.sh')}")

        # Run the script
        print("Running script:")
        script_result = await sandbox.process.exec(f"{os.path.join(new_dir, 'script.sh')}")
        print(script_result.result)

        # Search for files by name pattern (using new enhanced search)
        try:
            file_matches = await sandbox.fs.search(SearchRequest(query="*.json", path=new_dir, filenames_only=True))
            print("JSON files found:", file_matches.files)
        except Exception as e:
            print(f"Search endpoint not yet available: {e}")
            print("Using fallback file listing...")
            files = await sandbox.fs.list_files(new_dir)
            json_files = [f.name for f in files if f.name.endswith(".json")]
            print("JSON files found:", json_files)
            # Create a mock object for the report
            file_matches = type("obj", (object,), {"files": json_files})()

        # === NEW ENHANCED SEARCH FUNCTIONALITY ===
        print("\n=== Enhanced Content Search Examples (Async) ===")

        # 1. Basic content search
        print('\n1. Basic content search for "debug":')
        basic_search = await sandbox.fs.search(SearchRequest(query="debug", path=new_dir, case_sensitive=False))
        print(f"Found {basic_search.total_matches} matches in {basic_search.total_files} files")
        for match in basic_search.matches:
            print(f"  {match.file}:{match.line_number}: {match.line.strip()}")

        # 2. Search with file type filtering
        print('\n2. Search for "echo" in shell scripts only:')
        shell_search = await sandbox.fs.search(
            SearchRequest(query="echo", path=new_dir, file_types=["sh"], max_results=5)
        )
        print(f"Found {shell_search.total_matches} echo statements in shell files")
        for match in shell_search.matches:
            print(f"  {match.file}:{match.line_number}: {match.match}")

        # 3. Count-only search for performance
        print("\n3. Count-only search for all words:")
        count_search = await sandbox.fs.search(
            SearchRequest(query=r"\w+", path=new_dir, count_only=True)  # Regex for words
        )
        print(f"Total word matches: {count_search.total_matches} in {count_search.total_files} files")

        # 4. Advanced search with include/exclude patterns
        print("\n4. Search in text files only, excluding scripts:")
        advanced_search = await sandbox.fs.search(
            SearchRequest(
                query="file",
                path=new_dir,
                include_globs=["*.txt", "*.json"],
                exclude_globs=["*.sh"],
                case_sensitive=False,
                max_results=10,
            )
        )
        print(f"Found {advanced_search.total_matches} matches in text files")
        for match in advanced_search.matches:
            print(f"  {match.file}: {match.match}")

        # Replace content in config file
        await sandbox.fs.replace_in_files([os.path.join(new_dir, "config.json")], '"debug": true', '"debug": false')

        # Download the modified config file
        print("Downloading updated config file:")
        config_content = await sandbox.fs.download_file(os.path.join(new_dir, "config.json"))
        print(config_content.decode("utf-8"))

        # Create a report of all operations including search results
        report_data = f"""
        Project Files Report (Async):
        -----------------------------
        Time: {datetime.now().isoformat()}
        Files: {len(file_matches.files or [])} JSON files found
        Config: {'Production mode' if b'"debug": false' in config_content else 'Debug mode'}
        Script: {'Executed successfully' if script_result.exit_code == 0 else 'Failed'}

        Enhanced Search Results:
        - Debug references: {basic_search.total_matches} matches
        - Echo statements: {shell_search.total_matches} matches
        - Total words: {count_search.total_matches} matches
        - Text file matches: {advanced_search.total_matches} matches
        """.strip()

        # Save the report
        await sandbox.fs.upload_file(report_data.encode("utf-8"), os.path.join(new_dir, "report.txt"))

        # Clean up local file
        os.remove(local_file_path)

        # Delete the sandbox
        await daytona.delete(sandbox)


if __name__ == "__main__":
    asyncio.run(main())
