/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Chart types
 */
export enum ChartType {
  LINE = 'line',
  SCATTER = 'scatter',
  BAR = 'bar',
  PIE = 'pie',
  BOX_AND_WHISKER = 'box_and_whisker',
  COMPOSITE_CHART = 'composite_chart',
  UNKNOWN = 'unknown',
}

/**
 * Represents a chart with metadata from matplotlib.
 */
export type Chart = {
  /** The type of chart */
  type: ChartType
  /** The title of the chart */
  title: string
  /** The elements of the chart */
  elements: any[]
  /** The PNG representation of the chart encoded in base64 */
  png?: string
}

/**
 * Represents a 2D chart with metadata.
 */
export type Chart2D = Chart & {
  /** The label of the x-axis */
  x_label?: string
  /** The label of the y-axis */
  y_label?: string
}

/**
 * Represents a point in a 2D chart.
 */
export type PointData = {
  /** The label of the point */
  label: string
  /** The points of the chart */
  points: [number | string, number | string][]
}

/**
 * Represents a point chart with metadata.
 */
export type PointChart = Chart2D & {
  /** The ticks of the x-axis */
  x_ticks: (number | string)[]
  /** The scale of the x-axis */
  x_scale: string
  /** The labels of the x-axis */
  x_tick_labels: string[]
  /** The ticks of the y-axis */
  y_ticks: (number | string)[]
  /** The scale of the y-axis */
  y_scale: string
  /** The labels of the y-axis */
  y_tick_labels: string[]
  /** The points of the chart */
  elements: PointData[]
}

/**
 * Represents a line chart with metadata.
 */
export type LineChart = PointChart & {
  /** The type of chart */
  type: ChartType.LINE
}

/**
 * Represents a scatter chart with metadata.
 */
export type ScatterChart = PointChart & {
  /** The type of chart */
  type: ChartType.SCATTER
}

/**
 * Represents a bar in a bar chart.
 */
export type BarData = {
  /** The label of the bar */
  label: string
  /** The value of the bar */
  value: string
  /** The group of the bar */
  group: string
}

/**
 * Represents a bar chart with metadata.
 */
export type BarChart = Chart2D & {
  /** The type of chart */
  type: ChartType.BAR
  /** The bars of the chart */
  elements: BarData[]
}

/**
 * Represents a pie slice in a pie chart.
 */
export type PieData = {
  /** The label of the pie slice */
  label: string
  /** The angle of the pie slice */
  angle: number
  /** The radius of the pie slice */
  radius: number
}

/**
 * Represents a pie chart with metadata.
 */
export type PieChart = Chart & {
  /** The type of chart */
  type: ChartType.PIE
  /** The pie slices of the chart */
  elements: PieData[]
}

/**
 * Represents a box and whisker in a box and whisker chart.
 */
export type BoxAndWhiskerData = {
  /** The label of the box and whisker */
  label: string
  /** The minimum value of the box and whisker */
  min: number
  /** The first quartile of the box and whisker */
  first_quartile: number
  /** The median of the box and whisker */
  median: number
  /** The third quartile of the box and whisker */
  max: number
  outliers: number[]
}

/**
 * Represents a box and whisker chart with metadata.
 */
export type BoxAndWhiskerChart = Chart2D & {
  /** The type of chart */
  type: ChartType.BOX_AND_WHISKER
  /** The box and whiskers of the chart */
  elements: BoxAndWhiskerData[]
}

/**
 * Represents a composite chart with metadata.
 */
export type CompositeChart = Chart & {
  /** The type of chart */
  type: ChartType.COMPOSITE_CHART
  /** The charts of the composite chart */
  elements: Chart[]
}

export function parseChart(data: any): Chart {
  switch (data.type) {
    case ChartType.LINE:
      return { ...data } as LineChart
    case ChartType.SCATTER:
      return { ...data } as ScatterChart
    case ChartType.BAR:
      return { ...data } as BarChart
    case ChartType.PIE:
      return { ...data } as PieChart
    case ChartType.BOX_AND_WHISKER:
      return { ...data } as BoxAndWhiskerChart
    case ChartType.COMPOSITE_CHART:
      // eslint-disable-next-line no-case-declarations
      const charts = data.elements.map((g: any) => parseChart(g))
      delete data.data
      return {
        ...data,
        data: charts,
      } as CompositeChart
    default:
      return { ...data, type: ChartType.UNKNOWN } as Chart
  }
}
