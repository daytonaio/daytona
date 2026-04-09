# frozen_string_literal: true

module Daytona
  ChartElement = DaytonaToolboxApiClient::ChartElement

  module ChartType
    LINE = 'line'
    SCATTER = 'scatter'
    BAR = 'bar'
    PIE = 'pie'
    BOX_AND_WHISKER = 'box_and_whisker'
    COMPOSITE_CHART = 'composite_chart'
    UNKNOWN = 'unknown'
  end

  PointData = Struct.new(:label, :points, keyword_init: true)
  BarData = Struct.new(:label, :value, :group, keyword_init: true)
  PieData = Struct.new(:label, :angle, :radius, keyword_init: true)
  BoxAndWhiskerData = Struct.new(:label, :min, :first_quartile, :median, :third_quartile, :max, :outliers,
                                 keyword_init: true)

  Chart = Struct.new(:type, :title, :png, :elements, keyword_init: true) do
    def initialize(type: nil, title: nil, png: nil, elements: [])
      super
    end
  end

  Chart2D = Struct.new(:type, :title, :png, :elements, :x_label, :y_label, keyword_init: true) do
    def initialize(type: nil, title: nil, png: nil, elements: [], x_label: nil, y_label: nil)
      super
    end
  end

  PointChart = Struct.new(:type, :title, :png, :elements, :x_label, :y_label,
                          :x_ticks, :y_ticks, :x_tick_labels, :y_tick_labels,
                          :x_scale, :y_scale, keyword_init: true) do
    def initialize(type: nil, title: nil, png: nil, elements: [], x_label: nil, y_label: nil,
                   x_ticks: nil, y_ticks: nil, x_tick_labels: nil, y_tick_labels: nil,
                   x_scale: nil, y_scale: nil)
      super
    end
  end

  class LineChart < PointChart
  end

  class ScatterChart < PointChart
  end

  class BarChart < Chart2D
  end

  class PieChart < Chart
  end

  class BoxAndWhiskerChart < Chart2D
  end

  class CompositeChart < Chart
  end

  module Charts
    Chart = Daytona::Chart
    ChartElement = Daytona::ChartElement
    ChartType = Daytona::ChartType
    LineChart = Daytona::LineChart
    ScatterChart = Daytona::ScatterChart
    BarChart = Daytona::BarChart
    PieChart = Daytona::PieChart
    BoxAndWhiskerChart = Daytona::BoxAndWhiskerChart
    CompositeChart = Daytona::CompositeChart
    PointData = Daytona::PointData
    BarData = Daytona::BarData
    PieData = Daytona::PieData
    BoxAndWhiskerData = Daytona::BoxAndWhiskerData

    def self.parse_chart(chart)
      type = chart.type || ChartType::UNKNOWN
      elements = (chart.elements || []).map { |el| map_element(el, type) }
      common = { type: chart.type, title: chart.title, png: chart.png, elements: elements }

      case type
      when ChartType::LINE
        LineChart.new(x_label: chart.x_label, y_label: chart.y_label, x_ticks: chart.x_ticks, y_ticks: chart.y_ticks,
                      x_tick_labels: chart.x_tick_labels, y_tick_labels: chart.y_tick_labels,
                      x_scale: chart.x_scale, y_scale: chart.y_scale, **common)
      when ChartType::SCATTER
        ScatterChart.new(x_label: chart.x_label, y_label: chart.y_label, x_ticks: chart.x_ticks, y_ticks: chart.y_ticks,
                         x_tick_labels: chart.x_tick_labels, y_tick_labels: chart.y_tick_labels,
                         x_scale: chart.x_scale, y_scale: chart.y_scale, **common)
      when ChartType::BAR
        BarChart.new(x_label: chart.x_label, y_label: chart.y_label, **common)
      when ChartType::PIE
        PieChart.new(**common)
      when ChartType::BOX_AND_WHISKER
        BoxAndWhiskerChart.new(x_label: chart.x_label, y_label: chart.y_label, **common)
      when ChartType::COMPOSITE_CHART
        CompositeChart.new(**common)
      else
        Chart.new(**common)
      end
    end

    def self.map_element(el, chart_type)
      case chart_type
      when ChartType::LINE, ChartType::SCATTER
        PointData.new(label: el.label, points: el.points)
      when ChartType::BAR
        BarData.new(label: el.label, value: el.value, group: el.group)
      when ChartType::PIE
        PieData.new(label: el.label, angle: el.angle, radius: el.radius)
      when ChartType::BOX_AND_WHISKER
        BoxAndWhiskerData.new(label: el.label, min: el.min, first_quartile: el.first_quartile,
                              median: el.median, third_quartile: el.third_quartile,
                              max: el.max, outliers: el.outliers)
      else
        el
      end
    end
  end
end
