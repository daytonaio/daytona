# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from dataclasses import dataclass
from typing import Union


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
