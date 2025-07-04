# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from dataclasses import dataclass
from typing import Optional, Union


@dataclass
class FileUpload:
    """Represents a file to be uploaded to the Sandbox.

    Attributes:
        source (Union[bytes, str]): File contents as a bytes object or a local file path. If a bytes object is provided,
        make sure it fits into memory, otherwise use the local file path which content will be streamed to the Sandbox.
        destination (str): Absolute destination path in the Sandbox. Relative paths are resolved based on the user's
        root directory.
    """

    source: Union[bytes, str]
    destination: str


@dataclass
class FileDownloadRequest:
    """Represents a file to be downloaded from the Sandbox.

    Attributes:
        source (str): Absolute source path in the Sandbox. Relative paths are resolved based on the user's
        root directory.
        destination (Optional[str]): Destination path in the local filesystem where the file content will be streamed to.
        If not provided, the file will be downloaded in the bytes buffer (make sure it fits into memory).
    """

    source: str
    destination: Optional[str] = None


@dataclass
class FileDownloadResponse:
    source: str
    result: Union[str, bytes]
