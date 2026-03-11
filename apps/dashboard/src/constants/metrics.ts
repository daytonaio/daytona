/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export const METRIC_DISPLAY_NAMES: Record<string, string> = {
  'daytona.sandbox.cpu.utilization': 'CPU Utilization',
  'daytona.sandbox.cpu.limit': 'CPU Limit',
  'daytona.sandbox.memory.utilization': 'Memory Utilization',
  'daytona.sandbox.memory.usage': 'Memory Usage',
  'daytona.sandbox.memory.limit': 'Memory Limit',
  'daytona.sandbox.filesystem.utilization': 'Disk Utilization',
  'daytona.sandbox.filesystem.usage': 'Disk Usage',
  'daytona.sandbox.filesystem.total': 'Disk Total',
  'daytona.sandbox.filesystem.available': 'Disk Available',
  'system.memory.utilization': 'System Memory Utilization',
}

export function getMetricDisplayName(metricName: string): string {
  return METRIC_DISPLAY_NAMES[metricName] ?? metricName.replace(/^daytona\.sandbox\./, '')
}
