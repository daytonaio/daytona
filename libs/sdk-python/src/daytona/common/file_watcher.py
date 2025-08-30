# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from enum import Enum
from typing import Callable


class FilesystemEventType(str, Enum):
    """Types of file system events that can be watched."""

    CREATE = "CREATE"
    WRITE = "WRITE"
    DELETE = "DELETE"
    RENAME = "RENAME"
    CHMOD = "CHMOD"


class FilesystemEvent:
    """Represents a file system event."""

    def __init__(
        self,
        event_type: FilesystemEventType,
        name: str,
        is_dir: bool,
        timestamp: int,
    ):
        """Initialize a new FilesystemEvent.

        Args:
            event_type: The type of file system event
            name: The name/path of the file or directory
            is_dir: Whether this event is for a directory
            timestamp: Unix timestamp of when the event occurred
        """
        self.type = event_type
        self.name = name
        self.is_dir = is_dir
        self.timestamp = timestamp


class WatchOptions:
    """Options for file watching."""

    def __init__(self, recursive: bool = False):
        """Initialize watch options.

        Args:
            recursive: Whether to watch directories recursively
        """
        self.recursive = recursive


class WatchHandle:
    """Handle for a file watching session."""

    def __init__(self, close_func: Callable[[], None]):
        """Initialize a new WatchHandle.

        Args:
            close_func: Function to call to stop watching
        """
        self._close_func = close_func

    async def close(self) -> None:
        """Stop watching the directory."""
        self._close_func()


class SyncWatchHandle:
    """Synchronous handle for a file watching session."""

    def __init__(self, close_func: Callable[[], None]):
        """Initialize a new SyncWatchHandle.

        Args:
            close_func: Function to call to stop watching
        """
        self._close_func = close_func

    def close(self) -> None:
        """Stop watching the directory."""
        self._close_func()


# Type alias for the callback function
FileWatchCallback = Callable[[FilesystemEvent], None]
