# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

RSpec.describe Daytona::Charts do
  def chart_double(type:, elements: [], **attrs)
    double('ChartDto', { type: type, title: 'Chart', png: 'png', elements: elements, x_label: 'x', y_label: 'y',
                         x_ticks: [1], y_ticks: [2], x_tick_labels: ['one'], y_tick_labels: ['two'],
                         x_scale: 'linear', y_scale: 'log' }.merge(attrs))
  end

  it 'parses line charts into LineChart objects' do
    element = double('Element', label: 'series', points: [[1, 2]])
    chart = described_class.parse_chart(chart_double(type: Daytona::ChartType::LINE, elements: [element]))

    expect(chart).to be_a(Daytona::LineChart)
    expect(chart.elements.first).to eq(Daytona::PointData.new(label: 'series', points: [[1, 2]]))
  end

  it 'parses scatter charts into ScatterChart objects' do
    element = double('Element', label: 'series', points: [[1, 2]])
    chart = described_class.parse_chart(chart_double(type: Daytona::ChartType::SCATTER, elements: [element]))

    expect(chart).to be_a(Daytona::ScatterChart)
  end

  it 'parses bar charts into BarChart objects' do
    element = double('Element', label: 'sales', value: 12, group: 'q1')
    chart = described_class.parse_chart(chart_double(type: Daytona::ChartType::BAR, elements: [element]))

    expect(chart).to be_a(Daytona::BarChart)
    expect(chart.elements.first).to eq(Daytona::BarData.new(label: 'sales', value: 12, group: 'q1'))
  end

  it 'parses pie charts into PieChart objects' do
    element = double('Element', label: 'slice', angle: 90, radius: 10)
    chart = described_class.parse_chart(chart_double(type: Daytona::ChartType::PIE, elements: [element]))

    expect(chart).to be_a(Daytona::PieChart)
    expect(chart.elements.first).to eq(Daytona::PieData.new(label: 'slice', angle: 90, radius: 10))
  end

  it 'parses box-and-whisker charts into BoxAndWhiskerChart objects' do
    element = double('Element', label: 'latency', min: 1, first_quartile: 2, median: 3, third_quartile: 4, max: 5,
                                outliers: [10])
    chart = described_class.parse_chart(chart_double(type: Daytona::ChartType::BOX_AND_WHISKER, elements: [element]))

    expect(chart).to be_a(Daytona::BoxAndWhiskerChart)
    expect(chart.elements.first).to eq(
      Daytona::BoxAndWhiskerData.new(label: 'latency', min: 1, first_quartile: 2, median: 3, third_quartile: 4,
                                     max: 5, outliers: [10])
    )
  end

  it 'parses composite charts into CompositeChart objects' do
    chart = described_class.parse_chart(chart_double(type: Daytona::ChartType::COMPOSITE_CHART))

    expect(chart).to be_a(Daytona::CompositeChart)
  end

  it 'falls back to Chart for unknown types' do
    chart = described_class.parse_chart(chart_double(type: 'mystery'))

    expect(chart).to be_a(Daytona::Chart)
  end

  it 'treats nil chart types as unknown' do
    chart = described_class.parse_chart(chart_double(type: nil))

    expect(chart).to be_a(Daytona::Chart)
  end

  it 'returns the raw element for unknown chart mappings' do
    element = double('Element')

    expect(described_class.map_element(element, Daytona::ChartType::UNKNOWN)).to eq(element)
  end

  it 'uses default empty arrays in chart structs' do
    chart = Daytona::Chart.new
    point_chart = Daytona::PointChart.new

    expect(chart.elements).to eq([])
    expect(point_chart.elements).to eq([])
  end
end
