# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from typing import Any, Awaitable, Callable, Optional, TypeVar, Union

from pydantic import BaseModel


class OutputMessage(BaseModel):
    """Represents stdout or stderr output from code execution.

    Attributes:
        output: The output content.
    """

    output: str


class ExecutionError(BaseModel):
    """Represents an error that occurred during code execution.

    Attributes:
        name: The error type/class name (e.g., "ValueError", "SyntaxError").
        value: The error value.
        traceback: Full traceback of the error.
    """

    name: str
    value: str
    traceback: str = ""


class ExecutionResult(BaseModel):
    """Result of code execution.

    Attributes:
        stdout: Standard output from the code execution.
        stderr: Standard error output from the code execution.
        error: Error details if execution failed, None otherwise.
    """

    stdout: str = ""
    stderr: str = ""
    error: Optional[ExecutionError] = None


# Type aliases for callbacks
T = TypeVar("T")
OutputHandler = Union[
    Callable[[T], Any],
    Callable[[T], Awaitable[Any]],
]
