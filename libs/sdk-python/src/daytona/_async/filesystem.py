# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

import io
import os
from contextlib import ExitStack
from typing import List, Union, overload

import aiofiles
import aiofiles.os
import httpx
from daytona_api_client_async import (
    DownloadFiles,
    FileInfo,
    Match,
    ReplaceRequest,
    ReplaceResult,
    SearchFilesResponse,
    ToolboxApi,
)
from multipart import MultipartSegment, PushMultipartParser, parse_options_header

from .._utils.errors import DaytonaError, intercept_errors
from ..common.filesystem import FileDownloadRequest, FileDownloadResponse, FileUpload


class AsyncFileSystem:
    """Provides file system operations within a Sandbox.

    This class implements a high-level interface to file system operations that can
    be performed within a Daytona Sandbox.
    """

    def __init__(
        self,
        sandbox_id: str,
        toolbox_api: ToolboxApi,
    ):
        """Initializes a new FileSystem instance.

        Args:
            sandbox_id (str): The Sandbox ID.
            toolbox_api (ToolboxApi): API client for Sandbox operations.
        """
        self._sandbox_id = sandbox_id
        self._toolbox_api = toolbox_api

    @intercept_errors(message_prefix="Failed to create folder: ")
    async def create_folder(self, path: str, mode: str) -> None:
        """Creates a new directory in the Sandbox at the specified path with the given
        permissions.

        Args:
            path (str): Path where the folder should be created. Relative paths are resolved based
            on the sandbox working directory.
            mode (str): Folder permissions in octal format (e.g., "755" for rwxr-xr-x).

        Example:
            ```python
            # Create a directory with standard permissions
            await sandbox.fs.create_folder("workspace/data", "755")

            # Create a private directory
            await sandbox.fs.create_folder("workspace/secrets", "700")
            ```
        """
        print(f"Creating folder {path} with mode {mode}")
        await self._toolbox_api.create_folder(
            self._sandbox_id,
            path=path,
            mode=mode,
        )

    @intercept_errors(message_prefix="Failed to delete file: ")
    async def delete_file(self, path: str, recursive: bool = False) -> None:
        """Deletes a file from the Sandbox.

        Args:
            path (str): Path to the file to delete. Relative paths are resolved based on the sandbox working directory.
            recursive (bool): If the file is a directory, this must be true to delete it.

        Example:
            ```python
            # Delete a file
            await sandbox.fs.delete_file("workspace/data/old_file.txt")
            ```
        """
        await self._toolbox_api.delete_file(self._sandbox_id, path=path, recursive=recursive)

    @overload
    async def download_file(self, remote_path: str, timeout: int = 30 * 60) -> bytes:
        """Downloads a file from the Sandbox. Returns the file contents as a bytes object.
        This method is useful when you want to load the file into memory without saving it to disk.
        It can only be used for smaller files.

        Args:
            remote_path (str): Path to the file in the Sandbox. Relative paths are resolved based
            on the sandbox working directory.
            timeout (int): Timeout for the download operation in seconds. 0 means no timeout. Default is 30 minutes.

        Returns:
            bytes: The file contents as a bytes object.

        Example:
            ```python
            # Download and save a file locally
            content = await sandbox.fs.download_file("workspace/data/file.txt")
            with open("local_copy.txt", "wb") as f:
                f.write(content)

            # Download and process text content
            content = await sandbox.fs.download_file("workspace/data/config.json")
            config = json.loads(content.decode('utf-8'))
            ```
        """

    @overload
    async def download_file(self, remote_path: str, local_path: str, timeout: int = 30 * 60) -> None:
        """Downloads a file from the Sandbox and saves it to a local file using stream.
        This method is useful when you want to download larger files that may not fit into memory.

        Args:
            remote_path (str): Path to the file in the Sandbox. Relative paths are resolved based
            on the sandbox working directory.
            local_path (str): Path to save the file locally.
            timeout (int): Timeout for the download operation in seconds. 0 means no timeout. Default is 30 minutes.

        Example:
            ```python
            local_path = "local_copy.txt"
            await sandbox.fs.download_file("tmp/large_file.txt", local_path)
            size_mb = os.path.getsize(local_path) / 1024 / 1024
            print(f"Size of the downloaded file {local_path}: {size_mb} MB")
            ```
        """

    @intercept_errors(message_prefix="Failed to download file: ")
    async def download_file(self, *args: str) -> Union[bytes, None]:
        if len(args) == 1 or (len(args) == 2 and isinstance(args[1], int)):
            remote_path = args[0]
            timeout = args[1] if len(args) == 2 else 30 * 60
            response = (await self.download_files([FileDownloadRequest(source=remote_path)], timeout=timeout))[0]
            if response.error:
                raise DaytonaError(response.error)
            return response.result

        remote_path = args[0]
        local_path = args[1]
        timeout = args[2] if len(args) == 3 else 30 * 60
        # pylint: disable=protected-access
        response = await self.download_files(
            [FileDownloadRequest(source=remote_path, destination=local_path)], timeout=timeout
        )[0]
        if response.error:
            raise DaytonaError(response.error)
        return None

    @intercept_errors(message_prefix="Failed to download files: ")
    async def download_files(
        self, files: List[FileDownloadRequest], timeout: int = 30 * 60
    ) -> List[FileDownloadResponse]:
        """Downloads multiple files from the Sandbox. If the files already exist locally, they will be overwritten.

        Args:
            files (List[FileDownloadRequest]): List of files to download.
            timeout (int): Timeout for the download operation in seconds. 0 means no timeout. Default is 30 minutes.

        Returns:
            List[FileDownloadResponse]: List of download results.

        Raises:
            Exception: Only if the request itself fails (network issues, invalid request/response, etc.). Individual
            file download errors are returned in the `FileDownloadResponse.error` field.

        Example:
            ```python
            # Download multiple files
            results = await sandbox.fs.download_files([
                FileDownloadRequest(source="tmp/data.json"),
                FileDownloadRequest(source="tmp/config.json", destination="local_config.json")
            ])
            for result in results:
                if result.error:
                    print(f"Error downloading {result.source}: {result.error}")
                elif result.result:
                    print(f"Downloaded {result.source} to {result.result}")
            ```
        """
        if not files:
            return []

        class FileMeta:
            def __init__(self, dst: str | None):
                self.dst = dst
                self.error: str | None = None
                self.result: str | bytes | None = None

        src_file_meta_dict = {}
        file_writers = []
        for f in files:
            src_file_meta_dict[f.source] = FileMeta(dst=f.destination)

        # pylint: disable=protected-access
        method, url, headers, body, *_ = self._toolbox_api._download_files_serialize(
            self._sandbox_id,
            download_files=DownloadFiles(paths=list(src_file_meta_dict.keys())),
            x_daytona_organization_id=None,
            _request_auth=None,
            _content_type=None,
            _headers=None,
            _host_index=None,
        )

        try:
            async with httpx.AsyncClient(timeout=timeout) as client:
                async with client.stream(
                    method,
                    url,
                    json=body,
                    headers=headers,
                ) as resp:
                    resp.raise_for_status()

                    content_type, options = parse_options_header(resp.headers.get("Content-Type", ""))
                    if not (content_type == "multipart/form-data" and "boundary" in options):
                        raise DaytonaError(f"Unexpected Content-Type: {content_type}")
                    boundary = options["boundary"]

                    with PushMultipartParser(boundary) as parser:
                        writer = None
                        mode = None  # "file" or "error"
                        source = None

                        async for chunk in resp.aiter_bytes(64 * 1024):
                            if parser.closed:
                                raise DaytonaError("Unexpected end of multipart data")

                            for result in parser.parse(chunk):
                                if isinstance(result, MultipartSegment):  # New part starting
                                    writer = None
                                    mode = None
                                    source = result.filename
                                    if not source:
                                        raise DaytonaError(f"No source path found for this file {result.filename}")

                                    if result.name == "error":
                                        mode = "error"
                                    elif result.name == "file":
                                        mode = "file"
                                        meta = src_file_meta_dict[source]
                                        if meta.dst:
                                            parent = os.path.dirname(meta.dst)
                                            if parent:
                                                await aiofiles.os.makedirs(parent, exist_ok=True)
                                            # pylint: disable=consider-using-with
                                            writer = await aiofiles.open(meta.dst, mode="wb")
                                            file_writers.append(writer)
                                            meta.result = meta.dst
                                        else:
                                            writer = io.BytesIO()
                                            meta.result = writer

                                elif result:  # Non-empty bytearray with content
                                    if mode == "error":
                                        error_text = bytes(result).decode("utf-8", errors="ignore").strip()
                                        src_file_meta_dict[source].error = error_text
                                    elif mode == "file":
                                        try:
                                            if isinstance(writer, io.BytesIO):
                                                writer.write(bytes(result))
                                            else:
                                                await writer.write(bytes(result))
                                        except Exception as e:
                                            src_file_meta_dict[source].error = f"Write failed: {e}"
                                            mode = None

                                else:  # None - end of current part
                                    if writer and not isinstance(writer, io.BytesIO):
                                        await writer.close()
                                    writer = None
                                    mode = None
                                    source = None
        finally:
            for writer in file_writers:
                await writer.close()

        # Build results for all requested files
        results = []
        for f in files:
            meta = src_file_meta_dict[f.source]
            # see if there's an explicit error; if not, but no data, set a default error
            err = meta.error
            if not err and not meta.result:
                err = "No data received for this file"
            # only fetch the value if there was no error
            res = None
            if err is None:
                res = meta.result
                if isinstance(res, io.BytesIO):
                    res = res.getvalue()
            results.append(
                FileDownloadResponse(
                    source=f.source,
                    result=res,
                    error=err,
                )
            )

        return results

    @intercept_errors(message_prefix="Failed to find files: ")
    async def find_files(self, path: str, pattern: str) -> List[Match]:
        """Searches for files containing a pattern, similar to
        the grep command.

        Args:
            path (str): Path to the file or directory to search. If the path is a directory,
                the search will be performed recursively. Relative paths are resolved based
                on the sandbox working directory.
            pattern (str): Search pattern to match against file contents.

        Returns:
            List[Match]: List of matches found in files. Each Match object includes:
                - file: Path to the file containing the match
                - line: The line number where the match was found
                - content: The matching line content

        Example:
            ```python
            # Search for TODOs in Python files
            matches = await sandbox.fs.find_files("workspace/src", "TODO:")
            for match in matches:
                print(f"{match.file}:{match.line}: {match.content.strip()}")
            ```
        """
        return await self._toolbox_api.find_in_files(
            self._sandbox_id,
            path=path,
            pattern=pattern,
        )

    @intercept_errors(message_prefix="Failed to get file info: ")
    async def get_file_info(self, path: str) -> FileInfo:
        """Gets detailed information about a file or directory, including its
        size, permissions, and timestamps.

        Args:
            path (str): Path to the file or directory. Relative paths are resolved based
            on the sandbox working directory.

        Returns:
            FileInfo: Detailed file information including:
                - name: File name
                - is_dir: Whether the path is a directory
                - size: File size in bytes
                - mode: File permissions
                - mod_time: Last modification timestamp
                - permissions: File permissions in octal format
                - owner: File owner
                - group: File group

        Example:
            ```python
            # Get file metadata
            info = await sandbox.fs.get_file_info("workspace/data/file.txt")
            print(f"Size: {info.size} bytes")
            print(f"Modified: {info.mod_time}")
            print(f"Mode: {info.mode}")

            # Check if path is a directory
            info = await sandbox.fs.get_file_info("workspace/data")
            if info.is_dir:
                print("Path is a directory")
            ```
        """
        return await self._toolbox_api.get_file_info(self._sandbox_id, path=path)

    @intercept_errors(message_prefix="Failed to list files: ")
    async def list_files(self, path: str) -> List[FileInfo]:
        """Lists files and directories in a given path and returns their information, similar to the ls -l command.

        Args:
            path (str): Path to the directory to list contents from. Relative paths are resolved
            based on the sandbox working directory.

        Returns:
            List[FileInfo]: List of file and directory information. Each FileInfo
            object includes the same fields as described in get_file_info().

        Example:
            ```python
            # List directory contents
            files = await sandbox.fs.list_files("workspace/data")

            # Print files and their sizes
            for file in files:
                if not file.is_dir:
                    print(f"{file.name}: {file.size} bytes")

            # List only directories
            dirs = [f for f in files if f.is_dir]
            print("Subdirectories:", ", ".join(d.name for d in dirs))
            ```
        """
        return await self._toolbox_api.list_files(self._sandbox_id, path=path)

    @intercept_errors(message_prefix="Failed to move files: ")
    async def move_files(self, source: str, destination: str) -> None:
        """Moves or renames a file or directory. The parent directory of the destination must exist.

        Args:
            source (str): Path to the source file or directory. Relative paths are resolved
            based on the sandbox working directory.
            destination (str): Path to the destination. Relative paths are resolved based on
            the sandbox working directory.

        Example:
            ```python
            # Rename a file
            await sandbox.fs.move_files(
                "workspace/data/old_name.txt",
                "workspace/data/new_name.txt"
            )

            # Move a file to a different directory
            await sandbox.fs.move_files(
                "workspace/data/file.txt",
                "workspace/archive/file.txt"
            )

            # Move a directory
            await sandbox.fs.move_files(
                "workspace/old_dir",
                "workspace/new_dir"
            )
            ```
        """
        await self._toolbox_api.move_file(
            self._sandbox_id,
            source=source,
            destination=destination,
        )

    @intercept_errors(message_prefix="Failed to replace in files: ")
    async def replace_in_files(self, files: List[str], pattern: str, new_value: str) -> List[ReplaceResult]:
        """Performs search and replace operations across multiple files.

        Args:
            files (List[str]): List of file paths to perform replacements in. Relative paths are
            resolved based on the sandbox working directory.
            pattern (str): Pattern to search for.
            new_value (str): Text to replace matches with.

        Returns:
            List[ReplaceResult]: List of results indicating replacements made in
                each file. Each ReplaceResult includes:
                - file: Path to the modified file
                - success: Whether the operation was successful
                - error: Error message if the operation failed

        Example:
            ```python
            # Replace in specific files
            results = await sandbox.fs.replace_in_files(
                files=["workspace/src/file1.py", "workspace/src/file2.py"],
                pattern="old_function",
                new_value="new_function"
            )

            # Print results
            for result in results:
                if result.success:
                    print(f"{result.file}: {result.success}")
                else:
                    print(f"{result.file}: {result.error}")
            ```
        """
        for i, file in enumerate(files):
            files[i] = file

        replace_request = ReplaceRequest(files=files, new_value=new_value, pattern=pattern)

        return await self._toolbox_api.replace_in_files(self._sandbox_id, replace_request=replace_request)

    @intercept_errors(message_prefix="Failed to search files: ")
    async def search_files(self, path: str, pattern: str) -> SearchFilesResponse:
        """Searches for files and directories whose names match the
        specified pattern. The pattern can be a simple string or a glob pattern.

        Args:
            path (str): Path to the root directory to start search from. Relative paths are resolved
            based on the sandbox working directory.
            pattern (str): Pattern to match against file names. Supports glob
                patterns (e.g., "*.py" for Python files).

        Returns:
            SearchFilesResponse: Search results containing:
                - files: List of matching file and directory paths

        Example:
            ```python
            # Find all Python files
            result = await sandbox.fs.search_files("workspace", "*.py")
            for file in result.files:
                print(file)

            # Find files with specific prefix
            result = await sandbox.fs.search_files("workspace/data", "test_*")
            print(f"Found {len(result.files)} test files")
            ```
        """
        return await self._toolbox_api.search_files(
            self._sandbox_id,
            path=path,
            pattern=pattern,
        )

    @intercept_errors(message_prefix="Failed to set file permissions: ")
    async def set_file_permissions(self, path: str, mode: str = None, owner: str = None, group: str = None) -> None:
        """Sets permissions and ownership for a file or directory. Any of the parameters can be None
        to leave that attribute unchanged.

        Args:
            path (str): Path to the file or directory. Relative paths are resolved based on
            the sandbox working directory.
            mode (Optional[str]): File mode/permissions in octal format
                (e.g., "644" for rw-r--r--).
            owner (Optional[str]): User owner of the file.
            group (Optional[str]): Group owner of the file.

        Example:
            ```python
            # Make a file executable
            await sandbox.fs.set_file_permissions(
                path="workspace/scripts/run.sh",
                mode="755"  # rwxr-xr-x
            )

            # Change file owner
            await sandbox.fs.set_file_permissions(
                path="workspace/data/file.txt",
                owner="daytona",
                group="daytona"
            )
            ```
        """
        await self._toolbox_api.set_file_permissions(
            self._sandbox_id,
            path=path,
            mode=mode,
            owner=owner,
            group=group,
        )

    @overload
    async def upload_file(self, file: bytes, remote_path: str, timeout: int = 30 * 60) -> None:
        """Uploads a file to the specified path in the Sandbox. If a file already exists at
        the destination path, it will be overwritten. This method is useful when you want to upload
        small files that fit into memory.

        Args:
            file (bytes): File contents as a bytes object.
            remote_path (str): Path to the destination file. Relative paths are resolved based on
            the sandbox working directory.
            timeout (int): Timeout for the upload operation in seconds. 0 means no timeout. Default is 30 minutes.

        Example:
            ```python
            # Upload a text file
            content = b"Hello, World!"
            await sandbox.fs.upload_file(content, "tmp/hello.txt")

            # Upload a local file
            with open("local_file.txt", "rb") as f:
                content = f.read()
            await sandbox.fs.upload_file(content, "tmp/file.txt")

            # Upload binary data
            import json
            data = {"key": "value"}
            content = json.dumps(data).encode('utf-8')
            await sandbox.fs.upload_file(content, "tmp/config.json")
            ```
        """

    @overload
    async def upload_file(self, local_path: str, remote_path: str, timeout: int = 30 * 60) -> None:
        """Uploads a file from the local file system to the specified path in the Sandbox.
        If a file already exists at the destination path, it will be overwritten. This method uses
        streaming to upload the file, so it is useful when you want to upload larger files that may
        not fit into memory.

        Args:
            local_path (str): Path to the local file to upload.
            remote_path (str): Path to the destination file in the Sandbox. Relative paths are
            resolved based on the sandbox working directory.
            timeout (int): Timeout for the upload operation in seconds. 0 means no timeout. Default is 30 minutes.

        Example:
            ```python
            await sandbox.fs.upload_file("local_file.txt", "tmp/large_file.txt")
            ```
        """

    async def upload_file(self, src: Union[str, bytes], dst: str, timeout: int = 30 * 60) -> None:
        await self.upload_files([FileUpload(src, dst)], timeout)

    @intercept_errors(message_prefix="Failed to upload files: ")
    async def upload_files(self, files: List[FileUpload], timeout: int = 30 * 60) -> None:
        """Uploads multiple files to the Sandbox. If files already exist at the destination paths,
        they will be overwritten.

        Args:
            files (List[FileUpload]): List of files to upload.
            timeout (int): Timeout for the upload operation in seconds. 0 means no timeout. Default is 30 minutes.
        Example:
            ```python
            # Upload multiple text files
            files = [
                FileUpload(
                    source=b"Content of file 1",
                    destination="/tmp/file1.txt"
                ),
                FileUpload(
                    source="workspace/data/file2.txt",
                    destination="/tmp/file2.txt"
                ),
                FileUpload(
                    source=b'{"key": "value"}',
                    destination="/tmp/config.json"
                )
            ]
            await sandbox.fs.upload_files(files)
            ```
        """
        data_fields: dict[str, str] = {}
        file_fields: dict[str, tuple[str, any]] = {}

        with ExitStack() as stack:
            for i, f in enumerate(files):
                data_fields[f"files[{i}].path"] = f.destination

                if isinstance(f.source, (bytes, bytearray)):
                    stream = io.BytesIO(f.source)
                    filename = f.destination
                else:
                    stream = stack.enter_context(open(f.source, "rb"))
                    filename = f.destination

                # HTTPX will stream this file object in 64 KiB chunks :contentReference[oaicite:1]{index=1}
                file_fields[f"files[{i}].file"] = (filename, stream)

            # pylint: disable=protected-access
            _, url, headers, *_ = self._toolbox_api._upload_files_serialize(
                self._sandbox_id, None, None, None, None, None
            )
            # strip any prior Content-Type so HTTPX can set its own multipart header
            headers.pop("Content-Type", None)

            async with httpx.AsyncClient(timeout=timeout or None) as client:
                response = await client.post(
                    url, data=data_fields, files=file_fields, headers=headers  # any non-file form fields
                )
                response.raise_for_status()
