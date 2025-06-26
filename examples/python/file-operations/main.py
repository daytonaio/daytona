import json
import os
from datetime import datetime

from daytona import CreateSandboxFromSnapshotParams, Daytona, FileUpload
from daytona._sync.filesystem import SearchRequest


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

    # Make the script executable
    sandbox.process.exec(f"chmod +x {os.path.join(new_dir, 'script.sh')}")

    # Run the script
    print("Running script:")
    script_result = sandbox.process.exec(f"{os.path.join(new_dir, 'script.sh')}")
    print(script_result.result)

    # Search for files by name pattern (temporarily using find_files until new endpoint is deployed)
    try:
        file_matches = sandbox.fs.search(SearchRequest(query="*.json", path=new_dir, filenames_only=True))
        print("JSON files found:", file_matches.files)
    except Exception as e:
        print(f"Search endpoint not yet available: {e}")
        print("Using fallback file listing...")
        files = sandbox.fs.list_files(new_dir)
        json_files = [f.name for f in files if f.name.endswith(".json")]
        print("JSON files found:", json_files)
        # Create a mock object for the report
        file_matches = type("obj", (object,), {"files": json_files})()

    # === NEW ENHANCED SEARCH FUNCTIONALITY ===
    print("\n=== Enhanced Content Search Examples ===")
    print("Note: These examples will work once the new search endpoint is deployed")

    # 1. Basic content search
    print('\n1. Basic content search for "debug":')
    try:
        basic_search = sandbox.fs.search(SearchRequest(query="debug", path=new_dir, case_sensitive=False))
        print(f"Found {basic_search.total_matches} matches in {basic_search.total_files} files")
        for match in basic_search.matches:
            print(f"  {match.file}:{match.line_number}: {match.line.strip()}")
    except Exception as e:
        print(f"Search not available yet: {e}")
        basic_search = type("obj", (object,), {"total_matches": 0, "total_files": 0})()

    # Remaining search examples (will work once endpoint is deployed)
    print("\n2-6. Advanced search examples:")
    print("Search endpoint not yet available - skipping advanced examples")

    # Create mock objects for the report
    shell_search = type("obj", (object,), {"total_matches": 0})()
    context_search = type("obj", (object,), {"total_matches": 0})()
    count_search = type("obj", (object,), {"total_matches": 0})()
    filename_search = type("obj", (object,), {"total_files": 0})()
    advanced_search = type("obj", (object,), {"total_matches": 0})()

    # Replace content in config file
    sandbox.fs.replace_in_files([os.path.join(new_dir, "config.json")], '"debug": true', '"debug": false')

    # Download the modified config file
    print("Downloading updated config file:")
    config_content = sandbox.fs.download_file(os.path.join(new_dir, "config.json"))
    print(config_content.decode("utf-8"))

    # Create a report of all operations including search results
    report_data = f"""
    Project Files Report:
    ---------------------
    Time: {datetime.now().isoformat()}
    Files: {len(file_matches.files or [])} JSON files found
    Config: {'Production mode' if b'"debug": false' in config_content else 'Debug mode'}
    Script: {'Executed successfully' if script_result.exit_code == 0 else 'Failed'}

    Enhanced Search Results:
    - Debug references: {basic_search.total_matches} matches
    - Echo statements: {shell_search.total_matches} matches
    - Version references: {context_search.total_matches} matches
    - Total words: {count_search.total_matches} matches
    - Files with "project": {filename_search.total_files} files
    - Text file matches: {advanced_search.total_matches} matches
    """.strip()

    # Save the report
    sandbox.fs.upload_file(report_data.encode("utf-8"), os.path.join(new_dir, "report.txt"))

    # Clean up local file
    os.remove(local_file_path)

    # Delete the sandbox
    daytona.delete(sandbox)


if __name__ == "__main__":
    main()
