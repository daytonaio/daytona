# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import io
import os
from collections.abc import Callable, Iterator
from contextlib import ExitStack
from typing import Any, cast, final, overload

import httpx
from python_multipart.multipart import MultipartParser, parse_options_header
from typing_extensions import override

from daytona_toolbox_api_client import (
    FileInfo,
    FilesDownloadRequest,
    FileSystemApi,
    Match,
    ReplaceRequest,
    ReplaceResult,
    SearchFilesResponse,
)

from .._utils.errors import intercept_errors
from .._utils.otel_decorator import with_instrumentation
from ..common.errors import DaytonaError
from ..common.file_transfer import create_multipart_parser, parse_content_type_boundary, serialize_download_request
from ..common.filesystem import (
    CancelEvent,
    DownloadProgress,
    FileDownloadErrorDetails,
    FileDownloadRequest,
    FileDownloadResponse,
    FileUpload,
    UploadProgress,
    create_file_download_error,
    parse_file_download_error_payload,
    raise_if_stream_error,
)


class FileSystem:
    """Provides file system operations within a Sandbox.

    This class implements a high-level interface to file system operations that can
    be performed within a Daytona Sandbox.
    """

    def __init__(
        self,
        api_client: FileSystemApi,
    ):
        """Initializes a new FileSystem instance.

        Args:
            api_client (FileSystemApi): API client for Sandbox file system operations.
        """
        self._api_client: FileSystemApi = api_client

    @intercept_errors(message_prefix="Failed to create folder: ")
    @with_instrumentation()
    def create_folder(self, path: str, mode: str) -> None:
        """Creates a new directory in the Sandbox at the specified path with the given
        permissions.

        Args:
            path (str): Path where the folder should be created. Relative paths are resolved based
            on the sandbox working directory.
            mode (str): Folder permissions in octal format (e.g., "755" for rwxr-xr-x).

        Example:
            ```python
            # Create a directory with standard permissions
            sandbox.fs.create_folder("workspace/data", "755")

            # Create a private directory
            sandbox.fs.create_folder("workspace/secrets", "700")
            ```
        """
        self._api_client.create_folder(
            path=path,
            mode=mode,
        )

    @intercept_errors(message_prefix="Failed to delete file: ")
    @with_instrumentation()
    def delete_file(self, path: str, recursive: bool = False) -> None:
        """Deletes a file from the Sandbox.

        Args:
            path (str): Path to the file to delete. Relative paths are resolved based on the sandbox working directory.
            recursive (bool): If the file is a directory, this must be true to delete it.

        Example:
            ```python
            # Delete a file
            sandbox.fs.delete_file("workspace/data/old_file.txt")
            ```
        """
        self._api_client.delete_file(path=path, recursive=recursive)

    @overload
    def download_file(self, remote_path: str, timeout: int = 30 * 60) -> bytes:
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
            content = sandbox.fs.download_file("workspace/data/file.txt")
            with open("local_copy.txt", "wb") as f:
                f.write(content)

            # Download and process text content
            content = sandbox.fs.download_file("workspace/data/config.json")
            config = json.loads(content.decode('utf-8'))
            ```
        """

    @overload
    def download_file(self, remote_path: str, local_path: str, timeout: int = 30 * 60) -> None:
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
            sandbox.fs.download_file("tmp/large_file.txt", local_path)
            size_mb = os.path.getsize(local_path) / 1024 / 1024
            print(f"Size of the downloaded file {local_path}: {size_mb} MB")
            ```
        """

    @intercept_errors(message_prefix="Failed to download file: ")
    @with_instrumentation()
    def download_file(self, *args: str) -> bytes | None:  # pyright: ignore[reportInconsistentOverload]
        if len(args) == 1 or (len(args) == 2 and isinstance(args[1], int)):
            remote_path = args[0]
            timeout = int(args[1]) if len(args) == 2 else 30 * 60
            response = (self.download_files([FileDownloadRequest(source=remote_path)], timeout=timeout))[0]
            if response.error:
                raise create_file_download_error(response)
            result = response.result
            if isinstance(result, str):
                result = result.encode("utf-8")
            return result

        remote_path = args[0]
        local_path = args[1]
        timeout = int(args[2]) if len(args) == 3 else 30 * 60
        response = (
            self.download_files([FileDownloadRequest(source=remote_path, destination=local_path)], timeout=timeout)
        )[0]
        if response.error:
            raise create_file_download_error(response)
        return None

    @intercept_errors(message_prefix="Failed to download file: ")
    @with_instrumentation()
    def download_file_stream(
        self,
        remote_path: str,
        timeout: int = 30 * 60,
        on_progress: Callable[[DownloadProgress], None] | None = None,
        cancel_event: CancelEvent | None = None,
    ) -> Iterator[bytes]:
        """Downloads a single file from the Sandbox as a stream without buffering the entire file
        into memory. Returns an iterator that yields file content in chunks, which can be piped
        directly to an HTTP response, written to a file incrementally, or processed on the fly.

        Args:
            remote_path (str): Path to the file in the Sandbox. Relative paths are resolved based
                on the sandbox working directory.
            timeout (int): Timeout for the download operation in seconds. 0 means no timeout.
                Default is 30 minutes.
            on_progress (Callable[[DownloadProgress], None] | None): Optional callback invoked with
                cumulative bytes received and total bytes, when known, as the download progresses.
                Default is None.
            cancel_event (CancelEvent | None): Optional ``threading.Event``-compatible token. When
                set during streaming, the next chunk raises ``DaytonaError`` and the underlying
                HTTP connection is closed.

        Returns:
            Iterator[bytes]: An iterator yielding chunks of file content as bytes.

        Raises:
            DaytonaError: If the file does not exist, access is denied, or the download is
                cancelled via ``cancel_event``.

        Example:
            ```python
            # Stream to a local file without loading into memory
            with open("local_copy.bin", "wb") as f:
                for chunk in sandbox.fs.download_file_stream("workspace/large-file.bin"):
                    f.write(chunk)

            # Cancel an in-progress download from another thread
            import threading
            cancel = threading.Event()
            threading.Timer(5.0, cancel.set).start()
            for chunk in sandbox.fs.download_file_stream("workspace/big.bin", cancel_event=cancel):
                process(chunk)
            ```
        """

        def stream_generator() -> Iterator[bytes]:
            method, url, headers, body = serialize_download_request(self._api_client, remote_path)

            mode: str | None = None
            part_content_type: str | None = None
            header_field = bytearray()
            header_value = bytearray()
            part_headers: dict[str, str] = {}
            error_buffer = bytearray()
            pending_chunks: list[bytes] = []
            error_text: str | None = None
            error_details: FileDownloadErrorDetails | None = None
            received_file_data = False
            bytes_received = 0
            total_bytes: int | None = None

            def on_part_begin() -> None:
                nonlocal total_bytes
                part_headers.clear()
                header_field.clear()
                header_value.clear()
                error_buffer.clear()
                total_bytes = None

            def on_header_field(data: bytes, start: int, end: int) -> None:
                header_field.extend(data[start:end])

            def on_header_value(data: bytes, start: int, end: int) -> None:
                header_value.extend(data[start:end])

            def on_header_end() -> None:
                field = bytes(header_field).decode("utf-8", errors="ignore").lower()
                value = bytes(header_value).decode("utf-8", errors="ignore")
                part_headers[field] = value
                header_field.clear()
                header_value.clear()

            def on_headers_finished() -> None:
                nonlocal mode, part_content_type, total_bytes
                cd = part_headers.get("content-disposition", "")
                _, cd_params = parse_options_header(cd)
                name = cd_params.get(b"name", b"").decode("utf-8", errors="ignore")
                if not cd_params.get(b"filename"):
                    raise DaytonaError("No source path found for this file")
                part_content_type = part_headers.get("content-type")
                cl = part_headers.get("content-length")
                if cl is not None:
                    try:
                        total_bytes = int(cl)
                    except (TypeError, ValueError):
                        total_bytes = None
                else:
                    total_bytes = None
                mode = name if name in ("file", "error") else None

            def on_part_data(data: bytes, start: int, end: int) -> None:
                if mode == "error":
                    error_buffer.extend(data[start:end])
                elif mode == "file":
                    pending_chunks.append(bytes(data[start:end]))

            def on_part_end() -> None:
                nonlocal mode, part_content_type, error_text, error_details
                if mode == "error" and error_buffer:
                    error_text, error_details = parse_file_download_error_payload(
                        bytes(error_buffer),
                        part_content_type,
                    )
                    error_buffer.clear()
                mode = None
                part_content_type = None

            def drain() -> Iterator[bytes]:
                nonlocal received_file_data, bytes_received
                if not pending_chunks:
                    return
                if cancel_event is not None and cancel_event.is_set():
                    raise DaytonaError(f"Download cancelled: {remote_path}")
                emitted = pending_chunks.copy()
                pending_chunks.clear()
                received_file_data = True
                for piece in emitted:
                    bytes_received += len(piece)
                    if on_progress is not None:
                        on_progress(DownloadProgress(bytes_received=bytes_received, total_bytes=total_bytes))
                    yield piece

            httpx_timeout = None if timeout == 0 else timeout
            with httpx.Client(timeout=httpx_timeout) as client:
                with client.stream(method, url, json=body, headers=headers) as resp:
                    _ = resp.raise_for_status()

                    boundary = parse_content_type_boundary(resp.headers)
                    parser = create_multipart_parser(
                        boundary,
                        on_part_begin,
                        on_header_field,
                        on_header_value,
                        on_header_end,
                        on_headers_finished,
                        on_part_data,
                        on_part_end,
                    )

                    for chunk in resp.iter_bytes(64 * 1024):
                        _ = parser.write(chunk)
                        yield from drain()

                    parser.finalize()
                    yield from drain()

            raise_if_stream_error(remote_path, error_text, error_details, received_file_data)

        return stream_generator()

    @intercept_errors(message_prefix="Failed to download files: ")
    @with_instrumentation()
    def download_files(self, files: list[FileDownloadRequest], timeout: int = 30 * 60) -> list[FileDownloadResponse]:
        """Downloads multiple files from the Sandbox. If the files already exist locally, they will be overwritten.

        Args:
            files (list[FileDownloadRequest]): List of files to download.
            timeout (int): Timeout for the download operation in seconds. 0 means no timeout. Default is 30 minutes.

        Returns:
            list[FileDownloadResponse]: List of download results.

        Raises:
            Exception: Only if the request itself fails (network issues, invalid request/response, etc.). Individual
            file download errors are returned in `FileDownloadResponse.error`. When the daemon provides structured
            per-file metadata, it is also available in `FileDownloadResponse.error_details`.

        Example:
            ```python
            # Download multiple files
            results = sandbox.fs.download_files([
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
                self.dst: str | None = dst
                self.error: str | None = None
                self.error_details: FileDownloadErrorDetails | None = None
                self.result: str | bytes | io.BytesIO | None = None

        src_file_meta_dict: dict[str, FileMeta] = {}
        file_writers: list[io.BufferedIOBase] = []
        for f in files:
            src_file_meta_dict[f.source] = FileMeta(dst=f.destination)

        method, url, headers, body, *_ = self._api_client._download_files_serialize(
            download_files=FilesDownloadRequest(paths=list(src_file_meta_dict.keys())),
            _request_auth=None,
            _content_type=None,
            _headers=None,
            _host_index=None,
        )

        try:
            with httpx.Client(timeout=timeout) as client:
                with client.stream(
                    method,
                    url,
                    json=body,
                    headers=headers,
                ) as resp:
                    _ = resp.raise_for_status()

                    content_type_raw, options = parse_options_header(resp.headers.get("Content-Type", ""))
                    if not (content_type_raw == b"multipart/form-data" and b"boundary" in options):
                        raise DaytonaError(f"Unexpected Content-Type: {content_type_raw!r}")
                    boundary = options[b"boundary"]

                    writer: io.BytesIO | io.BufferedIOBase | None = None
                    mode: str | None = None
                    part_content_type: str | None = None
                    source: str | None = None
                    header_field = bytearray()
                    header_value = bytearray()
                    part_headers: dict[str, str] = {}
                    error_buffer = bytearray()

                    def on_part_begin() -> None:
                        nonlocal writer, mode, part_content_type, source
                        part_headers.clear()
                        error_buffer.clear()
                        writer = None
                        mode = None
                        part_content_type = None
                        source = None

                    def on_header_field(data: bytes, start: int, end: int) -> None:
                        header_field.extend(data[start:end])

                    def on_header_value(data: bytes, start: int, end: int) -> None:
                        header_value.extend(data[start:end])

                    def on_header_end() -> None:
                        field = bytes(header_field).decode("utf-8", errors="ignore").lower()
                        value = bytes(header_value).decode("utf-8", errors="ignore")
                        part_headers[field] = value
                        header_field.clear()
                        header_value.clear()

                    def on_headers_finished() -> None:
                        nonlocal writer, mode, part_content_type, source
                        cd = part_headers.get("content-disposition", "")
                        _, cd_params = parse_options_header(cd)
                        name = cd_params.get(b"name", b"").decode("utf-8", errors="ignore")
                        source = cd_params.get(b"filename", b"").decode("utf-8", errors="ignore") or None
                        if not source:
                            raise DaytonaError("No source path found for this file")
                        part_content_type = part_headers.get("content-type")

                        if name == "error":
                            mode = "error"
                        elif name == "file":
                            mode = "file"
                            meta = src_file_meta_dict[source]
                            if meta.dst:
                                parent = os.path.dirname(meta.dst)
                                if parent:
                                    os.makedirs(parent, exist_ok=True)
                                # pylint: disable=consider-using-with
                                writer = open(meta.dst, mode="wb")
                                file_writers.append(writer)
                                meta.result = meta.dst
                            else:
                                writer = io.BytesIO()
                                meta.result = writer

                    def on_part_data(data: bytes, start: int, end: int) -> None:
                        nonlocal mode
                        part_data = data[start:end]
                        if mode == "error":
                            error_buffer.extend(part_data)
                        elif mode == "file":
                            try:
                                if writer:
                                    _ = writer.write(part_data)
                            except Exception as e:
                                if source:
                                    src_file_meta_dict[source].error = f"Write failed: {e}"
                                else:
                                    raise DaytonaError(f"Write failed for unknown file with error {e}") from e
                                mode = None

                    def on_part_end() -> None:
                        nonlocal writer, mode, part_content_type, source
                        if mode == "error" and error_buffer:
                            error_text, error_details = parse_file_download_error_payload(
                                bytes(error_buffer),
                                part_content_type,
                            )
                            if source:
                                src_file_meta_dict[source].error = error_text
                                src_file_meta_dict[source].error_details = error_details
                            else:
                                raise DaytonaError(f"Error happened for unknown file with error {error_text}")
                            error_buffer.clear()
                        if writer and not isinstance(writer, io.BytesIO):
                            writer.close()
                        writer = None
                        mode = None
                        part_content_type = None
                        source = None

                    parser = MultipartParser(
                        boundary,
                        callbacks={
                            "on_part_begin": on_part_begin,
                            "on_header_field": on_header_field,
                            "on_header_value": on_header_value,
                            "on_header_end": on_header_end,
                            "on_headers_finished": on_headers_finished,
                            "on_part_data": on_part_data,
                            "on_part_end": on_part_end,
                        },
                    )

                    for chunk in resp.iter_bytes(64 * 1024):
                        _ = parser.write(chunk)
                    parser.finalize()
        finally:
            for writer in file_writers:
                writer.close()

        # Build results for all requested files
        results: list[FileDownloadResponse] = []
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
                    error_details=meta.error_details,
                )
            )

        return results

    @intercept_errors(message_prefix="Failed to find files: ")
    @with_instrumentation()
    def find_files(self, path: str, pattern: str) -> list[Match]:
        """Searches for files containing a pattern, similar to
        the grep command.

        Args:
            path (str): Path to the file or directory to search. If the path is a directory,
                the search will be performed recursively. Relative paths are resolved based
                on the sandbox working directory.
            pattern (str): Search pattern to match against file contents.

        Returns:
            list[Match]: List of matches found in files. Each Match object includes:
                - file: Path to the file containing the match
                - line: The line number where the match was found
                - content: The matching line content

        Example:
            ```python
            # Search for TODOs in Python files
            matches = sandbox.fs.find_files("workspace/src", "TODO:")
            for match in matches:
                print(f"{match.file}:{match.line}: {match.content.strip()}")
            ```
        """
        return self._api_client.find_in_files(
            path=path,
            pattern=pattern,
        )

    @intercept_errors(message_prefix="Failed to get file info: ")
    @with_instrumentation()
    def get_file_info(self, path: str) -> FileInfo:
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
            info = sandbox.fs.get_file_info("workspace/data/file.txt")
            print(f"Size: {info.size} bytes")
            print(f"Modified: {info.mod_time}")
            print(f"Mode: {info.mode}")

            # Check if path is a directory
            info = sandbox.fs.get_file_info("workspace/data")
            if info.is_dir:
                print("Path is a directory")
            ```
        """
        return self._api_client.get_file_info(path=path)

    @intercept_errors(message_prefix="Failed to list files: ")
    @with_instrumentation()
    def list_files(self, path: str) -> list[FileInfo]:
        """Lists files and directories in a given path and returns their information, similar to the ls -l command.

        Args:
            path (str): Path to the directory to list contents from. Relative paths are resolved
            based on the sandbox working directory.

        Returns:
            list[FileInfo]: List of file and directory information. Each FileInfo
            object includes the same fields as described in get_file_info().

        Example:
            ```python
            # List directory contents
            files = sandbox.fs.list_files("workspace/data")

            # Print files and their sizes
            for file in files:
                if not file.is_dir:
                    print(f"{file.name}: {file.size} bytes")

            # List only directories
            dirs = [f for f in files if f.is_dir]
            print("Subdirectories:", ", ".join(d.name for d in dirs))
            ```
        """
        return self._api_client.list_files(path=path)

    @intercept_errors(message_prefix="Failed to move files: ")
    @with_instrumentation()
    def move_files(self, source: str, destination: str) -> None:
        """Moves or renames a file or directory. The parent directory of the destination must exist.

        Args:
            source (str): Path to the source file or directory. Relative paths are resolved
            based on the sandbox working directory.
            destination (str): Path to the destination. Relative paths are resolved based on
            the sandbox working directory.

        Example:
            ```python
            # Rename a file
            sandbox.fs.move_files(
                "workspace/data/old_name.txt",
                "workspace/data/new_name.txt"
            )

            # Move a file to a different directory
            sandbox.fs.move_files(
                "workspace/data/file.txt",
                "workspace/archive/file.txt"
            )

            # Move a directory
            sandbox.fs.move_files(
                "workspace/old_dir",
                "workspace/new_dir"
            )
            ```
        """
        self._api_client.move_file(
            source=source,
            destination=destination,
        )

    @intercept_errors(message_prefix="Failed to replace in files: ")
    @with_instrumentation()
    def replace_in_files(self, files: list[str], pattern: str, new_value: str) -> list[ReplaceResult]:
        """Performs search and replace operations across multiple files.

        Args:
            files (list[str]): List of file paths to perform replacements in. Relative paths are
            resolved based on the sandbox working directory.
            pattern (str): Pattern to search for.
            new_value (str): Text to replace matches with.

        Returns:
            list[ReplaceResult]: List of results indicating replacements made in
                each file. Each ReplaceResult includes:
                - file: Path to the modified file
                - success: Whether the operation was successful
                - error: Error message if the operation failed

        Example:
            ```python
            # Replace in specific files
            results = sandbox.fs.replace_in_files(
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

        return self._api_client.replace_in_files(request=replace_request)

    @intercept_errors(message_prefix="Failed to search files: ")
    @with_instrumentation()
    def search_files(self, path: str, pattern: str) -> SearchFilesResponse:
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
            result = sandbox.fs.search_files("workspace", "*.py")
            for file in result.files:
                print(file)

            # Find files with specific prefix
            result = sandbox.fs.search_files("workspace/data", "test_*")
            print(f"Found {len(result.files)} test files")
            ```
        """
        return self._api_client.search_files(
            path=path,
            pattern=pattern,
        )

    @intercept_errors(message_prefix="Failed to set file permissions: ")
    @with_instrumentation()
    def set_file_permissions(
        self, path: str, mode: str | None = None, owner: str | None = None, group: str | None = None
    ) -> None:
        """Sets permissions and ownership for a file or directory. Any of the parameters can be None
        to leave that attribute unchanged.

        Args:
            path (str): Path to the file or directory. Relative paths are resolved based on
            the sandbox working directory.
            mode (str | None): File mode/permissions in octal format
                (e.g., "644" for rw-r--r--).
            owner (str | None): User owner of the file.
            group (str | None): Group owner of the file.

        Example:
            ```python
            # Make a file executable
            sandbox.fs.set_file_permissions(
                path="workspace/scripts/run.sh",
                mode="755"  # rwxr-xr-x
            )

            # Change file owner
            sandbox.fs.set_file_permissions(
                path="workspace/data/file.txt",
                owner="daytona",
                group="daytona"
            )
            ```
        """
        self._api_client.set_file_permissions(
            path=path,
            mode=mode,
            owner=owner,
            group=group,
        )

    @overload
    def upload_file(self, file: bytes, remote_path: str, timeout: int = 30 * 60) -> None:
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
            sandbox.fs.upload_file(content, "tmp/hello.txt")

            # Upload a local file
            with open("local_file.txt", "rb") as f:
                content = f.read()
            sandbox.fs.upload_file(content, "tmp/file.txt")

            # Upload binary data
            import json
            data = {"key": "value"}
            content = json.dumps(data).encode('utf-8')
            sandbox.fs.upload_file(content, "tmp/config.json")
            ```
        """

    @overload
    def upload_file(self, local_path: str, remote_path: str, timeout: int = 30 * 60) -> None:
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
            sandbox.fs.upload_file("local_file.txt", "tmp/large_file.txt")
            ```
        """

    @with_instrumentation()
    def upload_file(  # pyright: ignore[reportInconsistentOverload]
        self, src: str | bytes, dst: str, timeout: int = 30 * 60
    ) -> None:
        self.upload_files([FileUpload(src, dst)], timeout)

    @intercept_errors(message_prefix="Failed to upload files: ")
    @with_instrumentation()
    def upload_files(self, files: list[FileUpload], timeout: int = 30 * 60) -> None:
        """Uploads multiple files to the Sandbox. If files already exist at the destination paths,
        they will be overwritten.

        Args:
            files (list[FileUpload]): List of files to upload.
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
            sandbox.fs.upload_files(files)
            ```
        """
        data_fields: dict[str, str] = {}
        file_fields: dict[str, tuple[str, io.BytesIO | io.BufferedReader]] = {}

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

            _, url, headers, *_ = self._api_client._upload_files_serialize(None, None, None, None)
            # strip any prior Content-Type so HTTPX can set its own multipart header
            _ = headers.pop("Content-Type", None)

            with httpx.Client(timeout=timeout or None) as client:
                response = client.post(
                    url, data=data_fields, files=file_fields, headers=headers  # any non-file form fields
                )

                if not response.is_success:
                    try:
                        detail = ", ".join(response.json()["errors"])
                    except Exception:
                        detail = response.text
                    raise DaytonaError(
                        f"{response.status_code}: {detail}",
                        status_code=response.status_code,
                    )

    @intercept_errors(message_prefix="Failed to upload file: ")
    @with_instrumentation()
    def upload_file_stream(
        self,
        source: bytes | bytearray | str | io.IOBase,
        remote_path: str,
        timeout: int = 30 * 60,
        on_progress: Callable[[UploadProgress], None] | None = None,
        cancel_event: CancelEvent | None = None,
    ) -> None:
        """Uploads a single file to the Sandbox using true streaming, with optional progress
        tracking and cancellation. Memory usage stays flat regardless of source size. The
        HTTP layer uses chunked transfer encoding, so the source's natural EOF terminates
        the upload — no advance size is needed.

        Args:
            source (bytes | bytearray | str | IOBase): Data to upload. ``bytes`` is uploaded
                from memory; ``str`` is treated as a local file path and read in chunks; a
                file-like object (anything implementing ``.read()``) is streamed as-is.
            remote_path (str): Destination path in the Sandbox.
            timeout (int): Timeout in seconds. 0 means no timeout. Default is 30 minutes.
            on_progress (Callable[[UploadProgress], None] | None): Optional callback invoked
                with cumulative bytes sent.
            cancel_event (CancelEvent | None): Optional ``threading.Event``-compatible token.
                When set during streaming, the next chunk read raises ``DaytonaError`` and
                tears down the request.

        Raises:
            DaytonaError: If the upload fails or is cancelled via ``cancel_event``.

        Example:
            ```python
            import threading
            cancel = threading.Event()
            with open("large_dataset.csv", "rb") as f:
                sandbox.fs.upload_file_stream(
                    f,
                    "tmp/dataset.csv",
                    on_progress=lambda p: print(f"{p.bytes_sent} bytes sent"),
                    cancel_event=cancel,
                )
            ```
        """
        if cancel_event is not None and cancel_event.is_set():
            raise DaytonaError(f"Upload cancelled: {remote_path}")

        with ExitStack() as stack:
            stream = _open_upload_source(stack, source)
            wrapped = _CountingUploadReader(stream, on_progress, cancel_event, remote_path)

            data_fields = {"files[0].path": remote_path}
            # httpx accepts any IO[bytes]-shaped object as a multipart file; our
            # _CountingUploadReader is RawIOBase but pyright can't see that the
            # (filename, reader) tuple satisfies the structural FileTypes union,
            # so we drop into Any for the call.
            file_fields = cast(Any, {"files[0].file": (remote_path, wrapped)})

            _, url, headers, *_ = self._api_client._upload_files_serialize(None, None, None, None)
            _ = headers.pop("Content-Type", None)

            with httpx.Client(timeout=timeout or None) as client:
                response = client.post(url, data=data_fields, files=file_fields, headers=headers)

                if not response.is_success:
                    try:
                        detail = ", ".join(response.json()["errors"])
                    except Exception:
                        detail = response.text
                    raise DaytonaError(
                        f"{response.status_code}: {detail}",
                        status_code=response.status_code,
                    )


def _open_upload_source(stack: ExitStack, source: object) -> io.IOBase:
    """Coerces ``upload_file_stream`` source variants into a uniform read-ready stream.

    The stack owns closing any file we opened on the caller's behalf; file-like objects passed
    in by the caller are returned untouched (caller retains ownership and lifecycle).
    """
    if isinstance(source, (bytes, bytearray)):
        return io.BytesIO(bytes(source))
    if isinstance(source, str):
        return stack.enter_context(open(source, "rb"))
    if hasattr(source, "read"):
        # Caller-supplied IOBase (or duck-typed file-like). httpx will read
        # from it via .read(); the cast is structural since pyright can't
        # infer "anything with .read" implies io.IOBase.
        return cast(io.IOBase, source)
    raise DaytonaError(f"Unsupported upload source: {type(source).__name__}")


@final
class _CountingUploadReader(io.RawIOBase):
    """File-like wrapper that meters bytes flowing into httpx and honours cancellation
    between chunks. ``read()`` is what httpx calls during multipart streaming, so the
    cancellation check naturally interleaves with network progress."""

    def __init__(
        self,
        source: io.IOBase,
        on_progress: Callable[[UploadProgress], None] | None,
        cancel_event: CancelEvent | None,
        remote_path: str,
    ) -> None:
        super().__init__()
        self._source = source
        self._on_progress = on_progress
        self._cancel_event = cancel_event
        self._remote_path = remote_path
        self._sent = 0

    @override
    def readable(self) -> bool:
        return True

    @override
    def read(self, size: int = -1) -> bytes:
        if self._cancel_event is not None and self._cancel_event.is_set():
            raise DaytonaError(f"Upload cancelled: {self._remote_path}")
        chunk = self._source.read(size)
        if chunk:
            self._sent += len(chunk)
            if self._on_progress is not None:
                self._on_progress(UploadProgress(bytes_sent=self._sent))
        return chunk

    @override
    def readall(self) -> bytes:
        return self.read(-1)
