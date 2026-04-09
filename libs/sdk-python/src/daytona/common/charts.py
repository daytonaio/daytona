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
    LINE = "line"
    SCATTER = "scatter"
    BAR = "bar"
    PIE = "pie"
    BOX_AND_WHISKER = "box_and_whisker"
    COMPOSITE_CHART = "composite_chart"
    UNKNOWN = "unknown"


class PointData(GeneratedChartElement):
    pass


class BarData(GeneratedChartElement):
    pass


class PieData(GeneratedChartElement):
    pass


class BoxAndWhiskerData(GeneratedChartElement):
    pass


class Chart(GeneratedChart):
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
    pass


class PointChart(Chart2D):
    pass


class LineChart(PointChart):
    pass


class ScatterChart(PointChart):
    pass


class BarChart(Chart2D):
    pass


class PieChart(Chart):
    pass


class BoxAndWhiskerChart(Chart2D):
    pass


class CompositeChart(Chart):
    pass


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
