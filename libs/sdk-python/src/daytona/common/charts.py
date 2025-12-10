# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from collections.abc import Iterable, Mapping
from enum import Enum
from typing import Any, ClassVar

from pydantic import BaseModel, ConfigDict, field_validator


class ChartType(str, Enum):
    """
    Chart types

    **Enum Members**:
        - `LINE` ("line")
        - `SCATTER` ("scatter")
        - `BAR` ("bar")
        - `PIE` ("pie")
        - `BOX_AND_WHISKER` ("box_and_whisker")
        - `COMPOSITE_CHART` ("composite_chart")
        - `UNKNOWN` ("unknown")
    """

    LINE = "line"
    SCATTER = "scatter"
    BAR = "bar"
    PIE = "pie"
    BOX_AND_WHISKER = "box_and_whisker"
    COMPOSITE_CHART = "composite_chart"
    UNKNOWN = "unknown"


class Chart(BaseModel):
    """Represents a chart with metadata from matplotlib.

    Attributes:
        type (ChartType): The type of chart
        title (str): The title of the chart
        elements (list[Any]): The elements of the chart
        png (str | None): The PNG representation of the chart encoded in base64
    """

    model_config: ClassVar[ConfigDict] = ConfigDict(extra="allow")

    type: ChartType = ChartType.UNKNOWN
    title: str | None = None
    elements: list[Any] = []
    png: str | None = None

    def to_dict(self) -> dict[str, Any]:
        """Return the metadata dictionary used to create the chart."""
        return self.model_dump(exclude_none=True)


class Chart2D(Chart):
    """Represents a 2D chart with metadata.

    Attributes:
        x_label (str | None): The label of the x-axis
        y_label (str | None): The label of the y-axis
    """

    x_label: str | None = None
    y_label: str | None = None


class PointData(BaseModel):
    """Represents a point in a 2D chart.

    Attributes:
        label (str): The label of the point
        points (list[tuple[str | float, str | float]]): The points of the chart
    """

    model_config: ClassVar[ConfigDict] = ConfigDict(extra="allow")

    label: str
    points: list[tuple[str | float, str | float]] = []


class PointChart(Chart2D):
    """Represents a point chart with metadata.

    Attributes:
        x_ticks (list[str | float]): The ticks of the x-axis
        x_tick_labels (list[str]): The labels of the x-axis
        x_scale (str): The scale of the x-axis
        y_ticks (list[str | float]): The ticks of the y-axis
        y_tick_labels (list[str]): The labels of the y-axis
        y_scale (str): The scale of the y-axis
        elements (list[PointData]): The points of the chart
    """

    x_ticks: list[str | float] = []
    x_tick_labels: list[str] = []
    x_scale: str | None = None

    y_ticks: list[str | float] = []
    y_tick_labels: list[str] = []
    y_scale: str | None = None

    elements: list[PointData] = []

    @field_validator("elements", mode="before")
    @classmethod
    def _parse_elements(cls, value: list[Any]) -> list[PointData]:
        return [PointData.model_validate(element) for element in value]


class LineChart(PointChart):
    """Represents a line chart with metadata.

    Attributes:
        type (ChartType): The type of chart
    """

    type: ChartType = ChartType.LINE


class ScatterChart(PointChart):
    """Represents a scatter chart with metadata.

    Attributes:
        type (ChartType): The type of chart
    """

    type: ChartType = ChartType.SCATTER


class BarData(BaseModel):
    """Represents a bar in a bar chart.

    Attributes:
        label (str): The label of the bar
        group (str): The group of the bar
        value (str): The value of the bar
    """

    model_config: ClassVar[ConfigDict] = ConfigDict(extra="allow")

    label: str
    group: str | None = None
    value: str | float | int | None = None


class BarChart(Chart2D):
    """Represents a bar chart with metadata.

    Attributes:
        type (ChartType): The type of chart
        elements (list[BarData]): The bars of the chart
    """

    type: ChartType = ChartType.BAR

    elements: list[BarData] = []

    @field_validator("elements", mode="before")
    @classmethod
    def _parse_elements(cls, value: list[Any]) -> list[BarData]:
        return [BarData.model_validate(element) for element in value]


class PieData(BaseModel):
    """Represents a pie slice in a pie chart.

    Attributes:
        label (str): The label of the pie slice
        angle (float): The angle of the pie slice
        radius (float): The radius of the pie slice
        autopct (str | float): The autopct value of the pie slice
    """

    model_config: ClassVar[ConfigDict] = ConfigDict(extra="allow")

    label: str | None = None
    angle: float | None = None
    radius: float | None = None
    autopct: str | float | None = None


class PieChart(Chart):
    """Represents a pie chart with metadata.

    Attributes:
        type (ChartType): The type of chart
        elements (list[PieData]): The pie slices of the chart
    """

    type: ChartType = ChartType.PIE

    elements: list[PieData] = []

    @field_validator("elements", mode="before")
    @classmethod
    def _parse_elements(cls, value: list[Any]) -> list[PieData]:
        return [PieData.model_validate(element) for element in value]


class BoxAndWhiskerData(BaseModel):
    """Represents a box and whisker in a box and whisker chart.

    Attributes:
        label (str): The label of the box and whisker
        min (float): The minimum value of the box and whisker
        first_quartile (float): The first quartile of the box and whisker
        median (float): The median of the box and whisker
        third_quartile (float): The third quartile of the box and whisker
        max (float): The maximum value of the box and whisker
        outliers (list[float]): The outliers of the box and whisker
    """

    model_config: ClassVar[ConfigDict] = ConfigDict(extra="allow")

    label: str | None = None
    min: float | None = None
    first_quartile: float | None = None
    median: float | None = None
    third_quartile: float | None = None
    max: float | None = None
    outliers: list[float] = []


class BoxAndWhiskerChart(Chart2D):
    """Represents a box and whisker chart with metadata.

    Attributes:
        type (ChartType): The type of chart
        elements (list[BoxAndWhiskerData]): The box and whiskers of the chart
    """

    type: ChartType = ChartType.BOX_AND_WHISKER

    elements: list[BoxAndWhiskerData] = []

    @field_validator("elements", mode="before")
    @classmethod
    def _parse_elements(cls, value: list[Any]) -> list[BoxAndWhiskerData]:
        return [BoxAndWhiskerData.model_validate(element) for element in value]


class CompositeChart(Chart):
    """Represents a composite chart with metadata. A composite chart is a chart
    that contains multiple charts (subplots).

    Attributes:
        type (ChartType): The type of chart
        elements (list[Chart]): The charts (subplots) of the composite chart
    """

    type: ChartType = ChartType.COMPOSITE_CHART

    elements: list[Chart] = []

    @field_validator("elements", mode="before")
    @classmethod
    def _parse_elements(cls, value: list[Any]) -> list[Chart]:
        chart_list: Iterable[Chart | None] = [parse_chart(**element) for element in value]
        return [chart for chart in chart_list if chart is not None]


def parse_chart(**kwargs: Mapping[str, Any]) -> Chart | None:
    if not kwargs:
        return None

    chart_type = ChartType(kwargs.get("type", ChartType.UNKNOWN))
    chart_map = {
        ChartType.LINE: LineChart,
        ChartType.SCATTER: ScatterChart,
        ChartType.BAR: BarChart,
        ChartType.PIE: PieChart,
        ChartType.BOX_AND_WHISKER: BoxAndWhiskerChart,
        ChartType.COMPOSITE_CHART: CompositeChart,
    }

    model_cls = chart_map.get(chart_type, Chart)
    return model_cls.model_validate(kwargs)
