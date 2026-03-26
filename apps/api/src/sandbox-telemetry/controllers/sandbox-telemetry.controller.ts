/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Get, Param, Query, UseGuards } from '@nestjs/common'
import { ApiOAuth2, ApiResponse, ApiOperation, ApiParam, ApiTags, ApiHeader, ApiBearerAuth } from '@nestjs/swagger'
import { CombinedAuthGuard } from '../../auth/combined-auth.guard'
import { OrganizationResourceActionGuard } from '../../organization/guards/organization-resource-action.guard'
import { AuthenticatedRateLimitGuard } from '../../common/guards/authenticated-rate-limit.guard'
import { SandboxAccessGuard } from '../../sandbox/guards/sandbox-access.guard'
import { CustomHeaders } from '../../common/constants/header.constants'
import { SandboxTelemetryService } from '../services/sandbox-telemetry.service'
import { LogsQueryParamsDto, TelemetryQueryParamsDto, MetricsQueryParamsDto } from '../dto/telemetry-query-params.dto'
import { PaginatedLogsDto } from '../dto/paginated-logs.dto'
import { PaginatedTracesDto } from '../dto/paginated-traces.dto'
import { TraceSpanDto } from '../dto/trace-span.dto'
import { MetricsResponseDto } from '../dto/metrics-response.dto'
import { RequireFlagsEnabled } from '@openfeature/nestjs-sdk'
import { AnalyticsApiDisabledGuard } from '../guards/analytics-api-disabled.guard'

@ApiTags('sandbox')
@Controller('sandbox')
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@UseGuards(CombinedAuthGuard, OrganizationResourceActionGuard, AuthenticatedRateLimitGuard, AnalyticsApiDisabledGuard)
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
export class SandboxTelemetryController {
  constructor(private readonly sandboxTelemetryService: SandboxTelemetryService) {}

  @Get(':sandboxId/telemetry/logs')
  @ApiOperation({
    summary: 'Get sandbox logs',
    operationId: 'getSandboxLogs',
    description: 'Retrieve OTEL logs for a sandbox within a time range',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Paginated list of log entries',
    type: PaginatedLogsDto,
  })
  @UseGuards(SandboxAccessGuard)
  @RequireFlagsEnabled({ flags: [{ flagKey: 'organization_experiments', defaultValue: true }] })
  async getSandboxLogs(
    @Param('sandboxId') sandboxId: string,
    @Query() queryParams: LogsQueryParamsDto,
  ): Promise<PaginatedLogsDto> {
    return this.sandboxTelemetryService.getLogs(
      sandboxId,
      queryParams.from,
      queryParams.to,
      queryParams.page ?? 1,
      queryParams.limit ?? 100,
      queryParams.severities,
      queryParams.search,
    )
  }

  @Get(':sandboxId/telemetry/traces')
  @ApiOperation({
    summary: 'Get sandbox traces',
    operationId: 'getSandboxTraces',
    description: 'Retrieve OTEL traces for a sandbox within a time range',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Paginated list of trace summaries',
    type: PaginatedTracesDto,
  })
  @UseGuards(SandboxAccessGuard)
  @RequireFlagsEnabled({ flags: [{ flagKey: 'organization_experiments', defaultValue: true }] })
  async getSandboxTraces(
    @Param('sandboxId') sandboxId: string,
    @Query() queryParams: TelemetryQueryParamsDto,
  ): Promise<PaginatedTracesDto> {
    return this.sandboxTelemetryService.getTraces(
      sandboxId,
      queryParams.from,
      queryParams.to,
      queryParams.page ?? 1,
      queryParams.limit ?? 100,
    )
  }

  @Get(':sandboxId/telemetry/traces/:traceId')
  @ApiOperation({
    summary: 'Get trace spans',
    operationId: 'getSandboxTraceSpans',
    description: 'Retrieve all spans for a specific trace',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
    type: 'string',
  })
  @ApiParam({
    name: 'traceId',
    description: 'ID of the trace',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'List of spans in the trace',
    type: [TraceSpanDto],
  })
  @UseGuards(SandboxAccessGuard)
  @RequireFlagsEnabled({ flags: [{ flagKey: 'organization_experiments', defaultValue: true }] })
  async getSandboxTraceSpans(
    @Param('sandboxId') sandboxId: string,
    @Param('traceId') traceId: string,
  ): Promise<TraceSpanDto[]> {
    return this.sandboxTelemetryService.getTraceSpans(sandboxId, traceId)
  }

  @Get(':sandboxId/telemetry/metrics')
  @ApiOperation({
    summary: 'Get sandbox metrics',
    operationId: 'getSandboxMetrics',
    description: 'Retrieve OTEL metrics for a sandbox within a time range',
  })
  @ApiParam({
    name: 'sandboxId',
    description: 'ID of the sandbox',
    type: 'string',
  })
  @ApiResponse({
    status: 200,
    description: 'Metrics time series data',
    type: MetricsResponseDto,
  })
  @UseGuards(SandboxAccessGuard)
  @RequireFlagsEnabled({ flags: [{ flagKey: 'organization_experiments', defaultValue: true }] })
  async getSandboxMetrics(
    @Param('sandboxId') sandboxId: string,
    @Query() queryParams: MetricsQueryParamsDto,
  ): Promise<MetricsResponseDto> {
    return this.sandboxTelemetryService.getMetrics(sandboxId, queryParams.from, queryParams.to, queryParams.metricNames)
  }
}
