# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from typing import Optional


class ScreenshotRegion:
    """Region coordinates for screenshot operations.

    Attributes:
        x (int): X coordinate of the region.
        y (int): Y coordinate of the region.
        width (int): Width of the region.
        height (int): Height of the region.
    """

    def __init__(self, x: int, y: int, width: int, height: int):
        self.x = x
        self.y = y
        self.width = width
        self.height = height


class ScreenshotOptions:
    """Options for screenshot compression and display.

    Attributes:
        show_cursor (bool): Whether to show the cursor in the screenshot.
        fmt (str): Image format (e.g., 'png', 'jpeg', 'webp').
        quality (int): Compression quality (0-100).
        scale (float): Scale factor for the screenshot.
    """

    def __init__(
        self,
        show_cursor: Optional[bool] = None,
        fmt: Optional[str] = None,
        quality: Optional[int] = None,
        scale: Optional[float] = None,
    ):
        self.show_cursor = show_cursor
        self.fmt = fmt
        self.quality = quality
        self.scale = scale
