# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from typing import ClassVar

from pydantic import BaseModel, ConfigDict, Field


class ScreenshotRegion(BaseModel):
    """Region coordinates for screenshot operations.

    Attributes:
        x (int): X coordinate of the region.
        y (int): Y coordinate of the region.
        width (int): Width of the region.
        height (int): Height of the region.
    """

    model_config: ClassVar[ConfigDict] = ConfigDict(frozen=True)

    x: int
    y: int
    width: int
    height: int


class ScreenshotOptions(BaseModel):
    """Options for screenshot compression and display.

    Attributes:
        show_cursor (bool | None): Whether to show the cursor in the screenshot.
        fmt (str | None): Image format (e.g., 'png', 'jpeg', 'webp').
        quality (int | None): Compression quality (0-100).
        scale (float | None): Scale factor for the screenshot.
    """

    model_config: ClassVar[ConfigDict] = ConfigDict(extra="forbid")

    show_cursor: bool | None = Field(default=None, description="Whether to show the cursor in the screenshot.")
    fmt: str | None = Field(default=None, description="Image format (png, jpeg, webp).")
    quality: int | None = Field(default=None, ge=0, le=100, description="Compression quality.")
    scale: float | None = Field(default=None, gt=0, description="Scale factor for the screenshot.")
