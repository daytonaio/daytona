# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import json
import threading
from dataclasses import dataclass
from typing import IO, Any, Callable, cast

from .errors import DaytonaError, DaytonaValidationError, UploadAbortedError, create_daytona_error

TransferProgressCallback = Callable[[int], None]
"""Progress callback invoked with the cumulative number of bytes transferred."""


@dataclass
class UploadFileOptions:
    """Options for uploading a file to the Sandbox via the options-bag overload.

    Attributes:
        source: bytes/str delegates to existing bulk path. IO[bytes] streams without buffering.
        destination: Absolute destination path in the Sandbox.
        signal: Set this event to abort the upload. Note: abort is cooperative — up to one
            in-flight chunk (~64KB) may be written after signal fires.
        on_progress: Progress callback. Only valid with IO[bytes] sources; raises DaytonaValidationError otherwise.
        timeout: Timeout in seconds. 0 means no timeout. Default 30 minutes.
        content_length: Content length for IO[bytes] sources. Currently ignored. Reserved for a future iteration.
        cleanup_on_abort: When True, attempt best-effort delete_file on abort. Default False.

    Example:
        ```python
        signal = threading.Event()
        sandbox.fs.upload_file(options=UploadFileOptions(
            source=open("large.bin", "rb"),
            destination="/workspace/large.bin",
            signal=signal,
            on_progress=lambda b: print(f"{b} bytes uploaded"),
            cleanup_on_abort=True,
        ))
        ```
    """

    source: str | bytes | IO[bytes]
    destination: str
    signal: threading.Event | None = None
    on_progress: TransferProgressCallback | None = None
    timeout: int = 30 * 60
    content_length: int | None = None
    cleanup_on_abort: bool = False


@dataclass
class DownloadFileOptions:
    """Options for downloading a file from the Sandbox via the options-bag overload.

    Attributes:
        remote_path: Path to the file in the Sandbox.
        destination: None returns bytes. str writes to local path. IO[bytes] streams without buffering.
        signal: Set this event to abort the download. Note: abort is cooperative — up to one
            in-flight chunk (~64KB) may be written after signal fires.
        on_progress: Progress callback. Only valid with IO[bytes] destinations; raises DaytonaValidationError otherwise.
        timeout: Timeout in seconds. 0 means no timeout. Default 30 minutes.

    Example:
        ```python
        signal = threading.Event()
        sandbox.fs.download_file(options=DownloadFileOptions(
            remote_path="/workspace/large.bin",
            destination=open("local.bin", "wb"),
            signal=signal,
            on_progress=lambda b: print(f"{b} bytes downloaded"),
        ))
        ```
    """

    remote_path: str
    destination: str | IO[bytes] | None = None
    signal: threading.Event | None = None
    on_progress: TransferProgressCallback | None = None
    timeout: int = 30 * 60


@dataclass
class FileUpload:
    """Represents a file to be uploaded to the Sandbox.

    Attributes:
        source (bytes | str | IO[bytes]): File contents as a bytes object, a local file path, or an IO[bytes]
        stream. If a bytes object is provided, make sure it fits into memory, otherwise use the local file path
        or an IO[bytes] stream which content will be streamed to the Sandbox.
        destination (str): Absolute destination path in the Sandbox. Relative paths are resolved based on
        the sandbox working directory.
    """

    source: bytes | str | IO[bytes]
    destination: str


@dataclass
class FileDownloadRequest:
    """Represents a request to download a single file from the Sandbox.

    Attributes:
        source (str): Source path in the Sandbox. Relative paths are resolved based on the user's
        root directory.
        destination (str | None): Destination path in the local filesystem where the file content will be
        streamed to. If not provided, the file will be downloaded in the bytes buffer
        (might cause memory issues if the file is large).
    """

    source: str
    destination: str | None = None


@dataclass
class FileDownloadResponse:
    """Represents the response to a single file download request.

    Attributes:
        source (str): The original source path requested for download.
        result (str | bytes | None): The download result - file path (if destination provided in the request)
            or bytes content (if no destination in the request), None if failed or no data received.
        error (str | None): Error message if the download failed, None if successful.
        error_details (FileDownloadErrorDetails | None): Structured error metadata when the server provides it.
    """

    source: str
    result: str | bytes | None = None
    error: str | None = None
    error_details: FileDownloadErrorDetails | None = None


@dataclass
class FileDownloadErrorDetails:
    """Structured error metadata for a failed bulk file download item."""

    message: str
    status_code: int | None = None
    error_code: str | None = None


def create_file_download_error(response: FileDownloadResponse) -> DaytonaError:
    """Create the appropriate Daytona exception for a failed file download response."""

    if response.error is None:
        raise DaytonaValidationError("response.error must not be None")

    if response.error_details is None:
        return DaytonaError(response.error)

    return create_daytona_error(
        response.error_details.message,
        status_code=response.error_details.status_code,
        error_code=response.error_details.error_code,
    )


def parse_file_download_error_payload(
    payload: bytes,
    content_type: str | None,
) -> tuple[str, FileDownloadErrorDetails | None]:
    """Parse a bulk-download error part into the legacy message and structured metadata."""

    message = payload.decode("utf-8", errors="ignore").strip()
    if not content_type or "application/json" not in content_type.lower():
        return message, None

    try:
        data = json.loads(message)
    except json.JSONDecodeError:
        return message, None

    if not isinstance(data, dict):
        return message, None

    payload_dict = cast(dict[str, Any], data)
    structured_message = payload_dict.get("message")
    status_code = payload_dict.get("statusCode")
    if status_code is None:
        status_code = payload_dict.get("status_code")
    error_code = payload_dict.get("code")
    if error_code is None:
        error_code = payload_dict.get("error_code")

    if isinstance(structured_message, str):
        message = structured_message

    details = FileDownloadErrorDetails(
        message=message,
        status_code=status_code if isinstance(status_code, int) else None,
        error_code=error_code if isinstance(error_code, str) else None,
    )

    return message, details


class ProgressCountingIO:
    """Internal: wraps IO[bytes] to count bytes read and check abort signal.

    Not part of the public API. Do not add to __all__ or _DYNAMIC_IMPORTS.
    """

    def __init__(
        self,
        inner: IO[bytes],
        on_progress: TransferProgressCallback | None,
        signal: threading.Event | None,
    ):
        self._inner: IO[bytes] = inner
        self._on_progress: TransferProgressCallback | None = on_progress
        self._signal: threading.Event | None = signal
        self._bytes_read: int = 0

    def read(self, size: int = -1) -> bytes:
        if self._signal and self._signal.is_set():
            raise UploadAbortedError()
        data = self._inner.read(size)
        self._bytes_read += len(data)
        if data and self._on_progress:
            self._on_progress(self._bytes_read)
        return data

    def seek(self, offset: int, whence: int = 0) -> int:
        return self._inner.seek(offset, whence)

    def tell(self) -> int:
        return self._inner.tell()

    def seekable(self) -> bool:
        return getattr(self._inner, "seekable", lambda: False)()

    def readable(self) -> bool:
        return True

    def __iter__(self) -> ProgressCountingIO:
        return self

    def __next__(self) -> bytes:
        data = self.read(65536)
        if not data:
            raise StopIteration
        return data
