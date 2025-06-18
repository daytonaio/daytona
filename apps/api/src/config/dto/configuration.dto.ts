/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiExtraModels, ApiProperty, ApiPropertyOptional, ApiSchema, getSchemaPath } from '@nestjs/swagger'
import { IsString, IsBoolean, IsNumber, IsOptional } from 'class-validator'
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

  constructor(configService: TypedConfigService) {
    this.oidc = {
      issuer: configService.getOrThrow('oidc.issuer'),
      clientId: configService.getOrThrow('oidc.clientId'),
      audience: configService.getOrThrow('oidc.audience'),
    }
    this.linkedAccountsEnabled = configService.get('oidc.managementApi.enabled')
    this.proxyTemplateUrl = configService.getOrThrow('proxy.templateUrl')
    this.defaultSnapshot = configService.getOrThrow('defaultSnapshot')
    this.dashboardUrl = configService.getOrThrow('dashboardUrl')
    this.maxAutoArchiveInterval = configService.getOrThrow('maxAutoArchiveInterval')
    this.maintananceMode = configService.getOrThrow('maintananceMode')
    this.environment = configService.getOrThrow('environment')

    if (configService.get('billingApiUrl')) {
      this.billingApiUrl = configService.get('billingApiUrl')
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
    // this.announcements = configService.get('announcements')
  }
}
