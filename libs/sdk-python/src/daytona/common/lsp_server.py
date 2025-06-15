# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from enum import Enum


class LspLanguageId(Enum):
    """Language IDs for Language Server Protocol (LSP).

    **Enum Members**:
        - `PYTHON` ("python")
        - `TYPESCRIPT` ("typescript")
        - `JAVASCRIPT` ("javascript")
    """

    PYTHON = "python"
    TYPESCRIPT = "typescript"
    JAVASCRIPT = "javascript"

    def __str__(self):
        return self.value

    def __eq__(self, other):
        if isinstance(other, str):
            return self.value == other
        return super().__eq__(other)


class Position:
    """Represents a zero-based position in a text document,
    specified by line number and character offset.

    Attributes:
        line (int): Zero-based line number in the document.
        character (int): Zero-based character offset on the line.
    """

    def __init__(self, line: int, character: int):
        """Initialize a new Position instance.

        Args:
            line (int): Zero-based line number in the document.
            character (int): Zero-based character offset on the line.
        """
        self.line = line
        self.character = character
