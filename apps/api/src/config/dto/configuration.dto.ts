/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiExtraModels, ApiProperty, ApiPropertyOptional, ApiSchema, getSchemaPath } from '@nestjs/swagger'
import { IsBoolean, IsNumber, IsOptional, IsString } from 'class-validator'
import { TypedConfigService } from '../typed-config.service'

@ApiSchema({ name: 'Announcement' })
export class Announcement {
  @ApiProperty({
    description: 'The announcement text',
    example: 'New feature available!',
  })
  @IsString()
  text: string

  @ApiPropertyOptional({
    description: 'URL to learn more about the announcement',
    example: 'https://example.com/learn-more',
  })
  @IsString()
  @IsOptional()
  learnMoreUrl?: string
}

@ApiSchema({ name: 'PosthogConfig' })
export class PosthogConfig {
  @ApiProperty({
    description: 'PostHog API key',
    example: 'phc_abc123',
  })
  @IsString()
  apiKey: string

  @ApiProperty({
    description: 'PostHog host URL',
    example: 'https://app.posthog.com',
  })
  @IsString()
  host: string
}

@ApiSchema({ name: 'RateLimitEntry' })
export class RateLimitEntry {
  @ApiPropertyOptional({
    description: 'Rate limit TTL in seconds',
    example: 60,
  })
  @IsNumber()
  @IsOptional()
  ttl?: number

  @ApiPropertyOptional({
    description: 'Rate limit max requests',
    example: 100,
  })
  @IsNumber()
  @IsOptional()
  limit?: number
}

@ApiSchema({ name: 'RateLimitConfig' })
export class RateLimitConfig {
  @ApiPropertyOptional({
    description: 'Failed authentication rate limit',
    type: RateLimitEntry,
  })
  @IsOptional()
  failedAuth?: RateLimitEntry

  @ApiPropertyOptional({
    description: 'Authenticated rate limit',
    type: RateLimitEntry,
  })
  @IsOptional()
  authenticated?: RateLimitEntry

  @ApiPropertyOptional({
    description: 'Sandbox create rate limit',
    type: RateLimitEntry,
  })
  @IsOptional()
  sandboxCreate?: RateLimitEntry

  @ApiPropertyOptional({
    description: 'Sandbox lifecycle rate limit',
    type: RateLimitEntry,
  })
  @IsOptional()
  sandboxLifecycle?: RateLimitEntry
}

@ApiSchema({ name: 'OidcConfig' })
export class OidcConfig {
  @ApiProperty({
    description: 'OIDC issuer',
    example: 'https://auth.example.com',
  })
  @IsString()
  issuer: string

  @ApiProperty({
    description: 'OIDC client ID',
    example: 'daytona-client',
  })
  @IsString()
  clientId: string

  @ApiProperty({
    description: 'OIDC audience',
    example: 'daytona-api',
  })
  @IsString()
  audience: string
}

@ApiExtraModels(Announcement)
@ApiSchema({ name: 'DaytonaConfiguration' })
export class ConfigurationDto {
  @ApiProperty({
    description: 'Daytona version',
    example: '0.0.1',
  })
  @IsString()
  version: string

  @ApiPropertyOptional({
    description: 'PostHog configuration',
    type: PosthogConfig,
  })
  posthog?: PosthogConfig

  @ApiProperty({
    description: 'OIDC configuration',
    type: OidcConfig,
  })
  oidc: OidcConfig

  @ApiProperty({
    description: 'Whether linked accounts are enabled',
    example: true,
  })
  @IsBoolean()
  linkedAccountsEnabled: boolean

  @ApiProperty({
    description: 'System announcements',
    type: 'object',
    additionalProperties: { $ref: getSchemaPath(Announcement) },
    example: { 'feature-update': { text: 'New feature available!', learnMoreUrl: 'https://example.com' } },
  })
  announcements: Record<string, Announcement>

  @ApiPropertyOptional({
    description: 'Pylon application ID',
    example: 'pylon-app-123',
  })
  @IsString()
  @IsOptional()
  pylonAppId?: string

  @ApiProperty({
    description: 'Proxy template URL',
    example: 'https://{{PORT}}-{{sandboxId}}.proxy.example.com',
  })
  @IsString()
  proxyTemplateUrl: string

  @ApiProperty({
    description: 'Toolbox template URL',
    example: 'https://proxy.example.com/toolbox',
  })
  @IsString()
  proxyToolboxUrl: string

  @ApiProperty({
    description: 'Default snapshot for sandboxes',
    example: 'ubuntu:22.04',
  })
  @IsString()
  defaultSnapshot: string

  @ApiProperty({
    description: 'Dashboard URL',
    example: 'https://dashboard.example.com',
  })
  @IsString()
  dashboardUrl: string

  @ApiProperty({
    description: 'Maximum auto-archive interval in minutes',
    example: 43200,
  })
  @IsNumber()
  maxAutoArchiveInterval: number

  @ApiProperty({
    description: 'Whether maintenance mode is enabled',
    example: false,
  })
  @IsBoolean()
  maintananceMode: boolean

  @ApiProperty({
    description: 'Current environment',
    example: 'production',
  })
  @IsString()
  environment: string

  @ApiPropertyOptional({
    description: 'Billing API URL',
    example: 'https://billing.example.com',
  })
  @IsString()
  @IsOptional()
  billingApiUrl?: string

  @ApiPropertyOptional({
    description: 'Analytics API URL',
    example: 'https://analytics.example.com',
  })
  @IsString()
  @IsOptional()
  analyticsApiUrl?: string

  @ApiPropertyOptional({
    description: 'SSH Gateway command',
    example: 'ssh -p 2222 {{TOKEN}}@localhost',
  })
  @IsOptional()
  @IsString()
  sshGatewayCommand?: string

  @ApiPropertyOptional({
    description: 'Base64 encoded SSH Gateway public key',
    example: 'ssh-gateway-public-key',
  })
  @IsOptional()
  @IsString()
  sshGatewayPublicKey?: string

  @ApiPropertyOptional({
    description: 'Rate limit configuration',
    type: RateLimitConfig,
  })
  @IsOptional()
  rateLimit?: RateLimitConfig

  constructor(configService: TypedConfigService) {
    this.version = configService.getOrThrow('version')

    this.oidc = {
      issuer: configService.get('oidc.publicIssuer') || configService.getOrThrow('oidc.issuer'),
      clientId: configService.getOrThrow('oidc.clientId'),
      audience: configService.getOrThrow('oidc.audience'),
    }
    this.linkedAccountsEnabled = configService.get('oidc.managementApi.enabled')
    this.proxyTemplateUrl = configService.getOrThrow('proxy.templateUrl')
    this.proxyToolboxUrl = configService.getOrThrow('proxy.toolboxUrl')
    this.defaultSnapshot = configService.getOrThrow('defaultSnapshot')
    this.dashboardUrl = configService.getOrThrow('dashboardUrl')
    this.maxAutoArchiveInterval = configService.getOrThrow('maxAutoArchiveInterval')
    this.maintananceMode = configService.getOrThrow('maintananceMode')
    this.environment = configService.getOrThrow('environment')

    this.sshGatewayCommand = configService.get('sshGateway.command')
    this.sshGatewayPublicKey = configService.get('sshGateway.publicKey')

    if (configService.get('billingApiUrl')) {
      this.billingApiUrl = configService.get('billingApiUrl')
    }

    if (configService.get('analyticsApiUrl')) {
      this.analyticsApiUrl = configService.get('analyticsApiUrl')
    }

    if (configService.get('posthog.apiKey')) {
      this.posthog = {
        apiKey: configService.get('posthog.apiKey'),
        host: configService.get('posthog.host'),
      }
    }
    if (configService.get('pylonAppId')) {
      this.pylonAppId = configService.get('pylonAppId')
    }
    // TODO: announcements
    this.announcements = {}

    this.rateLimit = {
      authenticated: {
        ttl: configService.get('rateLimit.authenticated.ttl'),
        limit: configService.get('rateLimit.authenticated.limit'),
      },
      sandboxCreate: {
        ttl: configService.get('rateLimit.sandboxCreate.ttl'),
        limit: configService.get('rateLimit.sandboxCreate.limit'),
      },
      sandboxLifecycle: {
        ttl: configService.get('rateLimit.sandboxLifecycle.ttl'),
        limit: configService.get('rateLimit.sandboxLifecycle.limit'),
      },
    }
  }
}
