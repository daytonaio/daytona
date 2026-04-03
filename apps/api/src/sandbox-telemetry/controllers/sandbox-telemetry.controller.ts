/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Get, Param, Query, UseGuards } from '@nestjs/common'
import { ApiOAuth2, ApiResponse, ApiOperation, ApiParam, ApiTags, ApiHeader, ApiBearerAuth } from '@nestjs/swagger'
import { OrganizationAuthContextGuard } from '../../organization/guards/organization-auth-context.guard'
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
import { AuthStrategy } from '../../auth/decorators/auth-strategy.decorator'
import { AuthStrategyType } from '../../auth/enums/auth-strategy-type.enum'

@Controller('sandbox')
@ApiTags('sandbox')
@ApiOAuth2(['openid', 'profile', 'email'])
@ApiBearerAuth()
@ApiHeader(CustomHeaders.ORGANIZATION_ID)
@AuthStrategy([AuthStrategyType.API_KEY, AuthStrategyType.JWT])
@UseGuards(AuthenticatedRateLimitGuard)
@UseGuards(AnalyticsApiDisabledGuard)
@UseGuards(OrganizationAuthContextGuard, SandboxAccessGuard)
export class SandboxTelemetryController {
  constructor(private readonly sandboxTelemetryService: SandboxTelemetryService) {}

  @Get(':sandboxId/telemetry/logs')
  @RequireFlagsEnabled({ flags: [{ flagKey: 'organization_experiments', defaultValue: true }] })
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
  @RequireFlagsEnabled({ flags: [{ flagKey: 'organization_experiments', defaultValue: true }] })
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
  @RequireFlagsEnabled({ flags: [{ flagKey: 'organization_experiments', defaultValue: true }] })
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
  async getSandboxTraceSpans(
    @Param('sandboxId') sandboxId: string,
    @Param('traceId') traceId: string,
  ): Promise<TraceSpanDto[]> {
    return this.sandboxTelemetryService.getTraceSpans(sandboxId, traceId)
  }

  @Get(':sandboxId/telemetry/metrics')
  @RequireFlagsEnabled({ flags: [{ flagKey: 'organization_experiments', defaultValue: true }] })
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
  async getSandboxMetrics(
    @Param('sandboxId') sandboxId: string,
    @Query() queryParams: MetricsQueryParamsDto,
  ): Promise<MetricsResponseDto> {
    return this.sandboxTelemetryService.getMetrics(sandboxId, queryParams.from, queryParams.to, queryParams.metricNames)
  }
}
