# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

import json
from dataclasses import dataclass
from typing import Any, Protocol, cast, runtime_checkable

from .errors import DaytonaError, DaytonaValidationError, create_daytona_error


@runtime_checkable
class CancelEvent(Protocol):
    """Duck-typed cancellation token. Compatible with ``threading.Event`` and
    ``asyncio.Event`` (both expose ``is_set()``). When supplied to a streaming
    download, the next chunk read after the event becomes set raises
    ``DaytonaError``, closing the underlying HTTP connection."""

    def is_set(self) -> bool:
        ...


@dataclass
class FileUpload:
    """Represents a file to be uploaded to the Sandbox.

    Attributes:
        source (bytes | str): File contents as a bytes object or a local file path. If a bytes object is provided,
        make sure it fits into memory, otherwise use the local file path which content will be streamed to the Sandbox.
        destination (str): Absolute destination path in the Sandbox. Relative paths are resolved based on
        the sandbox working directory.
    """

    source: bytes | str
    destination: str


@dataclass
class DownloadProgress:
    """Progress information for a streaming download.

    Attributes:
        bytes_received (int): Cumulative bytes received so far.
        total_bytes (int | None): Total bytes expected, if known.
    """

    bytes_received: int
    total_bytes: int | None = None


@dataclass
class UploadProgress:
    """Progress information for a streaming upload.

    Attributes:
        bytes_sent (int): Cumulative bytes sent so far.
    """

    bytes_sent: int


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


def raise_if_stream_error(
    remote_path: str,
    error_text: str | None,
    error_details: FileDownloadErrorDetails | None,
    received_file_data: bool,
) -> None:
    """Raise the appropriate error after streaming a single-file multipart download."""
    if error_text is not None:
        raise create_file_download_error(
            FileDownloadResponse(
                source=remote_path,
                error=error_text,
                error_details=error_details,
            )
        )
    if not received_file_data:
        raise DaytonaError(f"No file data received for: {remote_path}")


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
