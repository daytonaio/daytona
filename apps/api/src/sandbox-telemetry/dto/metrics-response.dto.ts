/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'MetricDataPoint' })
export class MetricDataPointDto {
  @ApiProperty({ description: 'Timestamp of the data point' })
  timestamp: string

  @ApiProperty({ description: 'Value at this timestamp' })
  value: number
}

@ApiSchema({ name: 'MetricSeries' })
export class MetricSeriesDto {
  @ApiProperty({ description: 'Name of the metric' })
  metricName: string

  @ApiProperty({ type: [MetricDataPointDto], description: 'Data points for this metric' })
  dataPoints: MetricDataPointDto[]
}

@ApiSchema({ name: 'MetricsResponse' })
export class MetricsResponseDto {
  @ApiProperty({ type: [MetricSeriesDto], description: 'List of metric series' })
  series: MetricSeriesDto[]
}
