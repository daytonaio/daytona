# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from __future__ import annotations

from enum import Enum
from typing import Union

from typing_extensions import TypeAlias

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
    pass


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


def parse_chart(chart: GeneratedChartLike) -> Chart:
    chart_type = chart.type or ChartType.UNKNOWN.value
    chart_class: type[Chart]

    if chart_type == ChartType.LINE.value:
        chart_class = LineChart
    elif chart_type == ChartType.SCATTER.value:
        chart_class = ScatterChart
    elif chart_type == ChartType.BAR.value:
        chart_class = BarChart
    elif chart_type == ChartType.PIE.value:
        chart_class = PieChart
    elif chart_type == ChartType.BOX_AND_WHISKER.value:
        chart_class = BoxAndWhiskerChart
    elif chart_type == ChartType.COMPOSITE_CHART.value:
        chart_class = CompositeChart
    else:
        chart_class = Chart

    return chart_class.model_validate(chart.model_dump(exclude_none=True))
