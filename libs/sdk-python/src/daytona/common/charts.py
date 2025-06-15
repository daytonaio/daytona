# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

from enum import Enum
from typing import Any, List, Optional, Tuple, Union


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


class Chart:
    """Represents a chart with metadata from matplotlib.

    Attributes:
        type (ChartType): The type of chart
        title (str): The title of the chart
        elements (List[Any]): The elements of the chart
        png (Optional[str]): The PNG representation of the chart encoded in base64
    """

    type: ChartType
    title: str
    elements: List[Any]
    png: Optional[str] = None

    def __init__(self, **kwargs):
        super().__init__()
        self._metadata = kwargs
        self.type = kwargs.get("type")
        self.title = kwargs.get("title")
        self.elements = kwargs.get("elements", [])
        self.png = kwargs.get("png")

    def to_dict(self):
        return self._metadata


class Chart2D(Chart):
    """Represents a 2D chart with metadata.

    Attributes:
        x_label (Optional[str]): The label of the x-axis
        y_label (Optional[str]): The label of the y-axis
    """

    x_label: Optional[str]
    y_label: Optional[str]

    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        self.x_label = kwargs.get("x_label")
        self.y_label = kwargs.get("y_label")


class PointData:
    """Represents a point in a 2D chart.

    Attributes:
        label (str): The label of the point
        points (List[Tuple[Union[str, float], Union[str, float]]]): The points of the chart
    """

    label: str
    points: List[Tuple[Union[str, float], Union[str, float]]]

    def __init__(self, **kwargs):
        self.label = kwargs["label"]
        self.points = list(kwargs["points"])


class PointChart(Chart2D):
    """Represents a point chart with metadata.

    Attributes:
        x_ticks (List[Union[str, float]]): The ticks of the x-axis
        x_tick_labels (List[str]): The labels of the x-axis
        x_scale (str): The scale of the x-axis
        y_ticks (List[Union[str, float]]): The ticks of the y-axis
        y_tick_labels (List[str]): The labels of the y-axis
        y_scale (str): The scale of the y-axis
        elements (List[PointData]): The points of the chart
    """

    x_ticks: List[Union[str, float]]
    x_tick_labels: List[str]
    x_scale: str

    y_ticks: List[Union[str, float]]
    y_tick_labels: List[str]
    y_scale: str

    elements: List[PointData]

    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        self.x_label = kwargs.get("x_label")
        self.x_scale = kwargs.get("x_scale")
        self.x_ticks = kwargs.get("x_ticks")
        self.x_tick_labels = kwargs.get("x_tick_labels")

        self.y_label = kwargs.get("y_label")
        self.y_scale = kwargs.get("y_scale")
        self.y_ticks = kwargs.get("y_ticks")
        self.y_tick_labels = kwargs.get("y_tick_labels")

        self.elements = [PointData(**d) for d in kwargs.get("elements", [])]


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


class BarData:
    """Represents a bar in a bar chart.

    Attributes:
        label (str): The label of the bar
        group (str): The group of the bar
        value (str): The value of the bar
    """

    label: str
    group: str
    value: str

    def __init__(self, **kwargs):
        self.label = kwargs.get("label")
        self.value = kwargs.get("value")
        self.group = kwargs.get("group")


class BarChart(Chart2D):
    """Represents a bar chart with metadata.

    Attributes:
        type (ChartType): The type of chart
        elements (List[BarData]): The bars of the chart
    """

    type: ChartType = ChartType.BAR

    elements: List[BarData]

    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        self.elements = [BarData(**element) for element in kwargs.get("elements", [])]


class PieData:
    """Represents a pie slice in a pie chart.

    Attributes:
        label (str): The label of the pie slice
        angle (float): The angle of the pie slice
        radius (float): The radius of the pie slice
        autopct (float): The autopct value of the pie slice
    """

    label: str
    angle: float
    radius: float
    autopct: float

    def __init__(self, **kwargs):
        self.label = kwargs.get("label")
        self.angle = kwargs.get("angle")
        self.radius = kwargs.get("radius")
        self.autopct = kwargs.get("autopct")


class PieChart(Chart):
    """Represents a pie chart with metadata.

    Attributes:
        type (ChartType): The type of chart
        elements (List[PieData]): The pie slices of the chart
    """

    type: ChartType = ChartType.PIE

    elements: List[PieData]

    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        self.elements = [PieData(**element) for element in kwargs.get("elements", [])]


class BoxAndWhiskerData:
    """Represents a box and whisker in a box and whisker chart.

    Attributes:
        label (str): The label of the box and whisker
        min (float): The minimum value of the box and whisker
        first_quartile (float): The first quartile of the box and whisker
        median (float): The median of the box and whisker
        third_quartile (float): The third quartile of the box and whisker
        max (float): The maximum value of the box and whisker
        outliers (List[float]): The outliers of the box and whisker
    """

    label: str
    min: float
    first_quartile: float
    median: float
    third_quartile: float
    max: float
    outliers: List[float]

    def __init__(self, **kwargs):
        self.label = kwargs.get("label")
        self.min = kwargs.get("min")
        self.first_quartile = kwargs.get("first_quartile")
        self.median = kwargs.get("median")
        self.third_quartile = kwargs.get("third_quartile")
        self.max = kwargs.get("max")
        self.outliers = kwargs.get("outliers", [])


class BoxAndWhiskerChart(Chart2D):
    """Represents a box and whisker chart with metadata.

    Attributes:
        type (ChartType): The type of chart
        elements (List[BoxAndWhiskerData]): The box and whiskers of the chart
    """

    type: ChartType = ChartType.BOX_AND_WHISKER

    elements: List[BoxAndWhiskerData]

    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        self.elements = [BoxAndWhiskerData(**element) for element in kwargs.get("elements", [])]


class CompositeChart(Chart):
    """Represents a composite chart with metadata. A composite chart is a chart
    that contains multiple charts (subplots).

    Attributes:
        type (ChartType): The type of chart
        elements (List[Chart]): The charts (subplots) of the composite chart
    """

    type: ChartType = ChartType.COMPOSITE_CHART

    elements: List[Chart]

    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        self.elements = [parse_chart(**element) for element in kwargs.get("elements", [])]


def parse_chart(**kwargs) -> Optional[Chart]:
    if not kwargs:
        return None

    chart_type = ChartType(kwargs.get("type", ChartType.UNKNOWN))

    match chart_type:
        case ChartType.LINE:
            return LineChart(**kwargs)
        case ChartType.SCATTER:
            return ScatterChart(**kwargs)
        case ChartType.BAR:
            return BarChart(**kwargs)
        case ChartType.PIE:
            return PieChart(**kwargs)
        case ChartType.BOX_AND_WHISKER:
            return BoxAndWhiskerChart(**kwargs)
        case ChartType.COMPOSITE_CHART:
            return CompositeChart(**kwargs)
        case _:
            return Chart(**kwargs)
