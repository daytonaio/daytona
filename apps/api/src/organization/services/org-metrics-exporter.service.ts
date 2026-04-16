/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { Cron, CronExpression } from '@nestjs/schedule'
import { RedisLockProvider } from '../../sandbox/common/redis-lock.provider'
import { TypedConfigService } from '../../config/typed-config.service'
import { WithInstrumentation } from '../../common/decorators/otel.decorator'
import { OrganizationService } from './organization.service'
import { OrganizationUsageService } from './organization-usage.service'
import { Organization } from '../entities/organization.entity'

interface OtlpAttribute {
  key: string
  value: { stringValue: string }
}

interface OtlpDataPoint {
  attributes: OtlpAttribute[]
  asInt: string
  timeUnixNano: string
}

interface OtlpMetric {
  name: string
  description: string
  unit: string
  gauge: {
    dataPoints: OtlpDataPoint[]
  }
}

const MAX_CONCURRENT_EXPORTS = 5

@Injectable()
export class OrgMetricsExporterService {
  private readonly logger = new Logger(OrgMetricsExporterService.name)
  private readonly collectorUrl: string | undefined

  constructor(
    private readonly organizationService: OrganizationService,
    private readonly organizationUsageService: OrganizationUsageService,
    private readonly configService: TypedConfigService,
    private readonly redisLockProvider: RedisLockProvider,
  ) {
    this.collectorUrl = this.configService.get('otelCollector.endpointUrl')
  }

  @Cron(CronExpression.EVERY_MINUTE, { name: 'org-metrics-export' })
  @WithInstrumentation()
  async exportOrgMetrics(): Promise<void> {
    if (!this.collectorUrl) {
      return
    }

    const lockKey = 'org-metrics-export'
    const acquired = await this.redisLockProvider.lock(lockKey, 300)
    if (!acquired) {
      return
    }

    try {
      const organizations = await this.organizationService.findOrganizationsWithOtelConfig()
      if (organizations.length === 0) {
        return
      }

      for (let i = 0; i < organizations.length; i += MAX_CONCURRENT_EXPORTS) {
        const batch = organizations.slice(i, i + MAX_CONCURRENT_EXPORTS)
        const results = await Promise.allSettled(batch.map((org) => this.exportMetricsForOrganization(org)))

        for (let j = 0; j < results.length; j++) {
          if (results[j].status === 'rejected') {
            this.logger.warn(
              `Failed to export metrics for organization ${batch[j].id}`,
              (results[j] as PromiseRejectedResult).reason,
            )
          }
        }
      }
    } catch (error) {
      this.logger.warn('Failed to export organization metrics', error)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  private async exportMetricsForOrganization(organization: Organization): Promise<void> {
    const regionQuotas = await this.organizationService.getRegionQuotas(organization.id)
    if (regionQuotas.length === 0) {
      regionQuotas.push({
        organizationId: organization.id,
        regionId: organization.defaultRegionId || 'default',
        totalCpuQuota: -1,
        totalMemoryQuota: -1,
        totalDiskQuota: -1,
        maxCpuPerSandbox: null,
        maxMemoryPerSandbox: null,
        maxDiskPerSandbox: null,
        maxDiskPerNonEphemeralSandbox: null,
      })
    }

    const nowNano = `${Date.now() * 1_000_000}`
    const metrics: OtlpMetric[] = this.createMetricDefinitions()

    for (const rq of regionQuotas) {
      const usage = await this.organizationUsageService.getSandboxUsageOverview(organization.id, rq.regionId)

      const regionAttrs: OtlpAttribute[] = [{ key: 'region.id', value: { stringValue: rq.regionId } }]

      this.addDataPoint(metrics, 'daytona.sandbox.used_cpu', regionAttrs, usage.currentCpuUsage, nowNano)
      this.addDataPoint(metrics, 'daytona.sandbox.used_ram', regionAttrs, usage.currentMemoryUsage, nowNano)
      this.addDataPoint(metrics, 'daytona.sandbox.used_storage', regionAttrs, usage.currentDiskUsage, nowNano)
      if (rq.totalCpuQuota > 0) {
        this.addDataPoint(metrics, 'daytona.sandbox.total_cpu', regionAttrs, rq.totalCpuQuota, nowNano)
      }
      if (rq.totalMemoryQuota > 0) {
        this.addDataPoint(metrics, 'daytona.sandbox.total_ram', regionAttrs, rq.totalMemoryQuota, nowNano)
      }
      if (rq.totalDiskQuota > 0) {
        this.addDataPoint(metrics, 'daytona.sandbox.total_storage', regionAttrs, rq.totalDiskQuota, nowNano)
      }
    }

    const payload = {
      resourceMetrics: [
        {
          resource: {
            attributes: [{ key: 'organization.id', value: { stringValue: organization.id } }],
          },
          scopeMetrics: [
            {
              scope: { name: 'daytona.api.org_metrics', version: '1.0.0' },
              metrics,
            },
          ],
        },
      ],
    }

    await this.pushToCollector(organization.id, payload)
  }

  private createMetricDefinitions(): OtlpMetric[] {
    const defs = [
      { name: 'daytona.sandbox.used_cpu', description: 'Total CPU usage', unit: '{cpu}' },
      { name: 'daytona.sandbox.used_ram', description: 'Total memory usage', unit: 'GiBy' },
      { name: 'daytona.sandbox.used_storage', description: 'Total disk usage', unit: 'GiBy' },
      { name: 'daytona.sandbox.total_cpu', description: 'Total CPU quota', unit: '{cpu}' },
      { name: 'daytona.sandbox.total_ram', description: 'Total memory quota', unit: 'GiBy' },
      { name: 'daytona.sandbox.total_storage', description: 'Total disk quota', unit: 'GiBy' },
    ]

    return defs.map((d) => ({
      ...d,
      gauge: { dataPoints: [] },
    }))
  }

  private addDataPoint(
    metrics: OtlpMetric[],
    metricName: string,
    attributes: OtlpAttribute[],
    value: number,
    timeUnixNano: string,
  ): void {
    const metric = metrics.find((m) => m.name === metricName)
    if (!metric) {
      return
    }

    metric.gauge.dataPoints.push({
      attributes,
      asInt: `${value}`,
      timeUnixNano,
    })
  }

  private async pushToCollector(organizationId: string, payload: unknown): Promise<void> {
    const endpoint = `${this.collectorUrl}/v1/metrics`

    const response = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'organization-id': organizationId,
      },
      body: JSON.stringify(payload),
      signal: AbortSignal.timeout(10_000),
    })

    if (!response.ok) {
      const body = await response.text().catch(() => '')
      this.logger.warn(`Failed to push metrics for org ${organizationId}: HTTP ${response.status} - ${body}`)
    }
  }
}
