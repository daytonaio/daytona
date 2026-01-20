# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from dataclasses import dataclass


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
    """

    source: str
    result: str | bytes | None = None
    error: str | None = None
