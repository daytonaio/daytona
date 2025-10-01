# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from dataclasses import dataclass
from typing import Optional


@dataclass
class PtySize:
    """PTY terminal size configuration.

    Used to specify the dimensions of a PTY terminal when creating or resizing sessions.
    Maximum size is 1000x1000.

    Attributes:
        rows: Number of terminal rows (height)
        cols: Number of terminal columns (width)

    Example:
        ```python
        # Create a 120x30 terminal
        pty_size = PtySize(rows=30, cols=120)

        # Create PTY session with specific size
        pty_handle = sandbox.process.create_pty_session(
            id="my-session",
            pty_size=pty_size
        )
        ```
    """

    rows: int
    cols: int


@dataclass
class PtyResult:
    """PTY session result containing exit information.

    Contains the final state of a PTY session after it has terminated, including
    the exit code and any error information.

    Attributes:
        exit_code: Exit code of the PTY process (0 for success, non-zero for errors).
                  None if the process hasn't exited yet or exit code couldn't be determined.
        error: Error message if the PTY failed or was terminated abnormally.
               None if no error occurred.

    Example:
        ```python
        # Wait for PTY to complete and get result
        result = pty_handle.wait()

        if result.exit_code == 0:
            print("PTY session completed successfully")
        else:
            print(f"PTY session failed with code {result.exit_code}")
            if result.error:
                print(f"Error: {result.error}")
        ```
    """

    exit_code: Optional[int] = None
    error: Optional[str] = None
