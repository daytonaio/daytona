# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from enum import Enum
from typing import Union

from pydantic import Field
from typing_extensions import TypeAlias, override

from daytona_toolbox_api_client import Chart as GeneratedChart
from daytona_toolbox_api_client import ChartElement as GeneratedChartElement
from daytona_toolbox_api_client_async import Chart as AsyncGeneratedChart

GeneratedChartLike: TypeAlias = Union[GeneratedChart, AsyncGeneratedChart]
ChartElement: TypeAlias = GeneratedChartElement
ChartElementLike: TypeAlias = GeneratedChartElement


class ChartType(str, Enum):
    """Supported chart types returned by the daemon's code-run endpoint."""

    LINE = "line"
    SCATTER = "scatter"
    BAR = "bar"
    PIE = "pie"
    BOX_AND_WHISKER = "box_and_whisker"
    COMPOSITE_CHART = "composite_chart"
    UNKNOWN = "unknown"


class PointData(GeneratedChartElement):
    """Data element for line and scatter charts. Fields: label, points."""


class BarData(GeneratedChartElement):
    """Data element for bar charts. Fields: label, value, group."""


class PieData(GeneratedChartElement):
    """Data element for pie charts. Fields: label, angle, radius."""


class BoxAndWhiskerData(GeneratedChartElement):
    """Data element for box-and-whisker charts.
    Fields: label, min, first_quartile, median, third_quartile, max, outliers.
    """


class Chart(GeneratedChart):
    """Base chart class. All chart types inherit from this. Fields are sourced from the daemon's typed response."""

    elements: list[GeneratedChartElement] = Field(  # pyright: ignore[reportIncompatibleVariableOverride]
        default_factory=list
    )

    @classmethod
    @override
    def model_validate(cls, obj: object, **kwargs: object) -> Chart:
        instance = super().model_validate(obj, **kwargs)  # pyright: ignore[reportArgumentType]
        if cls is Chart:
            return _resolve_chart_subclass(instance)
        return instance  # type: ignore[return-value]


class Chart2D(Chart):
    """Chart with x/y axes. Adds x_label, y_label fields."""


class PointChart(Chart2D):
    """Chart with axis ticks and scales. Adds x_ticks, y_ticks, x_scale, y_scale fields."""


class LineChart(PointChart):
    """Line chart. Elements are PointData."""


class ScatterChart(PointChart):
    """Scatter plot. Elements are PointData."""


class BarChart(Chart2D):
    """Bar chart. Elements are BarData."""


class PieChart(Chart):
    """Pie chart. Elements are PieData."""


class BoxAndWhiskerChart(Chart2D):
    """Box-and-whisker chart. Elements are BoxAndWhiskerData."""


class CompositeChart(Chart):
    """Composite chart containing multiple sub-charts as elements."""


_CHART_TYPE_MAP: dict[str, type[Chart]] = {
    ChartType.LINE.value: LineChart,
    ChartType.SCATTER.value: ScatterChart,
    ChartType.BAR.value: BarChart,
    ChartType.PIE.value: PieChart,
    ChartType.BOX_AND_WHISKER.value: BoxAndWhiskerChart,
    ChartType.COMPOSITE_CHART.value: CompositeChart,
}


def parse_chart(chart: GeneratedChartLike) -> Chart:
    chart_class = _CHART_TYPE_MAP.get(chart.type or "", Chart)
    return chart_class.model_validate(chart.model_dump(exclude_none=True))


def _resolve_chart_subclass(chart: Chart) -> Chart:
    chart_class = _CHART_TYPE_MAP.get(chart.type or "", Chart)
    if isinstance(chart, chart_class):
        return chart
    return chart_class.model_validate(chart.model_dump(exclude_none=True))
