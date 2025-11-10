# frozen_string_literal: true

module Daytona
  module Charts
    # @param data [Hash<Symbol, Object>]
    # @return [Daytona::Charts::Chart]
    def self.parse(data)
      case data.fetch(:type, ChartType::UNKNOWN)
      when ChartType::LINE then LineChart.new(data)
      when ChartType::SCATTER then ScatterChart.new(data)
      when ChartType::BAR then BarChart.new(data)
      when ChartType::PIE then PieChart.new(data)
      when ChartType::BOX_AND_WHISKER then BoxAndWhiskerChart.new(data)
      when ChartType::COMPOSITE_CHART then CompositeChart.new(data)
      else
        Chart.new(data)
      end
    end

    class Chart
      # @return [String, Nil] The type of chart
      attr_reader :type

      # @return [String, Nil] The title of the chart
      attr_reader :title

      # @return [Array<Object>] The elements of the chart
      attr_reader :elements

      # @return [String, Nil] The PNG representation of the chart encoded in base64
      attr_reader :png

      # @param data [Hash<Symbol, Object>]
      def initialize(data)
        @type = data.fetch(:type, nil)
        @title = data.fetch(:title, nil)
        @elements = data.fetch(:elements, [])
        @png = data.fetch(:png, nil)
      end

      # @return [Hash<Symbol, Object>] original metadata
      def to_h = { type:, title:, elements:, png: }
    end

    class Chart2D < Chart
      # @return [String, Nil] The label of the x-axis
      attr_reader :x_label

      # @return [String, Nil] The label of the y-axis
      attr_reader :y_label

      # @param data [Hash<Symbol, Object>]
      def initialize(data)
        super
        @x_label = data.fetch(:x_label, nil)
        @y_label = data.fetch(:y_label, nil)
      end

      # @return [Hash<Symbol, Object>] original metadata
      def to_h = super.merge(x_label:, y_label:)
    end

    class PointData
      # @return [String] The label of the point series
      attr_reader :label

      # @return [Array<Array<Object>>] Array of [x, y] points
      attr_reader :points

      # @param data [Hash<Symbol, Object>]
      def initialize(data)
        @label = data.fetch(:label)
        @points = data.fetch(:points)
      end

      # @return [Hash<Symbol, Object>] original data representation
      def to_h = { label:, points: }
    end

    class PointChart < Chart2D
      # @return [Array<Object>] The ticks of the x-axis
      attr_reader :x_ticks

      # @return [Array<String>] The labels of the x-axis
      attr_reader :x_tick_labels

      # @return [String, Nil] The scale of the x-axis
      attr_reader :x_scale

      # @return [Array<Object>] The ticks of the y-axis
      attr_reader :y_ticks

      # @return [Array<String>] The labels of the y-axis
      attr_reader :y_tick_labels

      # @return [String, Nil] The scale of the y-axis
      attr_reader :y_scale

      # @return [Array<Daytona::Charts::PointData>] The points of the chart
      attr_reader :elements

      # @param data [Hash<Symbol, Object>]
      def initialize(data)
        super
        @x_scale = data.fetch(:x_scale, nil)
        @x_ticks = data.fetch(:x_ticks, nil)
        @x_tick_labels = data.fetch(:x_tick_labels, nil)

        @y_scale = data.fetch(:y_scale, nil)
        @y_ticks = data.fetch(:y_ticks, nil)
        @y_tick_labels = data.fetch(:y_tick_labels, nil)

        @elements = data.fetch(:elements, []).map { |e| PointData.new(e) }
      end

      # @return [Hash<Symbol, Object>] original metadata
      def to_h
        super.merge(
          x_scale:,
          x_ticks:,
          x_tick_labels:,
          y_scale:,
          y_ticks:,
          y_tick_labels:,
          elements: elements.map(&:to_h)
        )
      end
    end

    class LineChart < PointChart
      def initialize(data)
        super
        @type = ChartType::LINE
      end
    end

    class ScatterChart < PointChart
      def initialize(data)
        super
        @type = ChartType::SCATTER
      end
    end

    class BarData
      # @return [String] The label of the bar
      attr_reader :label

      # @return [String] The value of the bar
      attr_reader :value

      # @return [String] The group of the bar
      attr_reader :group

      # @param data [Hash<Symbol, Object>]
      def initialize(data)
        @label = data.fetch(:label)
        @value = data.fetch(:value)
        @group = data.fetch(:group)
      end

      # @return [Hash<Symbol, Object>]
      def to_h = { label:, value:, group: }
    end

    class BarChart < Chart2D
      # @return [Array<Daytona::Charts::BarData>] The bars of the chart
      attr_reader :elements

      # @param data [Hash<Symbol, Object>]
      def initialize(data)
        super
        @type = ChartType::BAR
        @elements = data.fetch(:elements, []).map { |e| BarData.new(e) }
      end

      # @return [Hash<Symbol, Object>]
      def to_h = super.merge(elements: elements.map(&:to_h))
    end

    class PieData
      # @return [String] The label of the pie slice
      attr_reader :label

      # @return [Float] The angle of the pie slice
      attr_reader :angle

      # @return [Float] The radius of the pie slice
      attr_reader :radius

      # @return [Float] The autopct value of the pie slice
      attr_reader :autopct

      # @param data [Hash<Symbol, Object>]
      def initialize(data)
        @label = data.fetch(:label)
        @angle = data.fetch(:angle)
        @radius = data.fetch(:radius)
        @autopct = data.fetch(:autopct)
      end

      # @return [Hash<Symbol, Object>]
      def to_h = { label:, angle:, radius:, autopct: }
    end

    class PieChart < Chart
      # @return [Array<Daytona::Charts::PieData>] The pie slices of the chart
      attr_reader :elements

      # @param data [Hash<Symbol, Object>]
      def initialize(data)
        super
        @type = ChartType::PIE
        @elements = data.fetch(:elements, []).map { |e| PieData.new(e) }
      end

      # @return [Hash<Symbol, Object>]
      def to_h = super.merge(elements: elements.map(&:to_h))
    end

    class BoxAndWhiskerData
      # @return [String] The label of the box and whisker
      attr_reader :label

      # @return [Float] The minimum value of the box and whisker
      attr_reader :min

      # @return [Float] The first quartile of the box and whisker
      attr_reader :first_quartile

      # @return [Float] The median of the box and whisker
      attr_reader :median

      # @return [Float] The third quartile of the box and whisker
      attr_reader :third_quartile

      # @return [Float] The maximum value of the box and whisker
      attr_reader :max

      # @return [Array<Float>] The outliers of the box and whisker
      attr_reader :outliers

      # @param data [Hash<Symbol, Object>]
      def initialize(data)
        @label = data.fetch(:label)
        @min = data.fetch(:min)
        @first_quartile = data.fetch(:first_quartile)
        @median = data.fetch(:median)
        @third_quartile = data.fetch(:third_quartile)
        @max = data.fetch(:max)
        @outliers = data.fetch(:outliers, [])
      end

      # @return [Hash<Symbol, Object>]
      def to_h = { label:, min:, first_quartile:, median:, third_quartile:, max:, outliers: }
    end

    class BoxAndWhiskerChart < Chart2D
      # @return [Array<Daytona::Charts::BoxAndWhiskerData>] The box and whiskers of the chart
      attr_reader :elements

      # @param data [Hash<Symbol, Object>]
      def initialize(data)
        super
        @type = ChartType::BOX_AND_WHISKER
        @elements = data.fetch(:elements, []).map { |e| BoxAndWhiskerData.new(e) }
      end

      # @return [Hash<Symbol, Object>]
      def to_h = super.merge(elements: elements.map(&:to_h))
    end

    class CompositeChart < Chart
      # @return [Array<Daytona::Charts::Chart>] The charts of the composite chart
      attr_reader :elements

      # @param data [Hash<Symbol, Object>]
      def initialize(data)
        super
        @type = ChartType::COMPOSITE_CHART
        @elements = data.fetch(:elements, []).map { |e| Charts.parse(e) }
      end

      def to_h = super.merge(elements: elements.map(&:to_h))
    end

    module ChartType
      ALL = [
        LINE = 'line',
        SCATTER = 'scatter',
        BAR = 'bar',
        PIE = 'pie',
        BOX_AND_WHISKER = 'box_and_whisker',
        COMPOSITE_CHART = 'composite_chart',
        UNKNOWN = 'unknown'
      ].freeze
    end
  end
end
