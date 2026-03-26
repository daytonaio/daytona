# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from dataclasses import dataclass
from enum import Enum
from typing import Literal

from typing_extensions import override


class LspLanguageId(str, Enum):
    """Language IDs for Language Server Protocol (LSP).

    **Enum Members**:
        - `PYTHON` ("python")
        - `TYPESCRIPT` ("typescript")
        - `JAVASCRIPT` ("javascript")
    """

    PYTHON = "python"
    TYPESCRIPT = "typescript"
    JAVASCRIPT = "javascript"

    @override
    def __str__(self):
        return self.value


LspLanguageIdLiteral = Literal["python", "typescript", "javascript"]


@dataclass
class LspCompletionPosition:
    """Represents a zero-based completion position in a text document,
    specified by line number and character offset.

    Attributes:
        line (int): Zero-based line number in the document.
        character (int): Zero-based character offset on the line.
    """

    line: int
    character: int
