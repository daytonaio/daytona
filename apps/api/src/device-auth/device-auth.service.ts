/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, NotFoundException, BadRequestException, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository, LessThan } from 'typeorm'
import * as crypto from 'crypto'
import { DeviceAuthorizationRequest, DeviceAuthStatus } from './device-auth.entity'
import { ApiKeyService } from '../api-key/api-key.service'
import { OrganizationService } from '../organization/services/organization.service'
import { OrganizationResourcePermission } from '../organization/enums/organization-resource-permission.enum'
import { Cron, CronExpression } from '@nestjs/schedule'

@Injectable()
export class DeviceAuthService {
  private readonly logger = new Logger(DeviceAuthService.name)
  private readonly DEVICE_CODE_EXPIRY_SECONDS = 900 // 15 minutes
  private readonly POLLING_INTERVAL_SECONDS = 5
  private readonly MIN_POLL_INTERVAL_MS = 4000 // 4 seconds (slightly less than 5 for timing tolerance)

  constructor(
    @InjectRepository(DeviceAuthorizationRequest)
    private readonly deviceAuthRepository: Repository<DeviceAuthorizationRequest>,
    private readonly apiKeyService: ApiKeyService,
    private readonly organizationService: OrganizationService,
  ) {}

  private generateDeviceCode(): string {
    return crypto.randomBytes(32).toString('base64url')
  }

  private generateUserCode(): string {
    // Generate a 8-character code like "WDJB-MJHT"
    const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ' // Excluding I, O to avoid confusion
    let code = ''
    for (let i = 0; i < 8; i++) {
      if (i === 4) code += '-'
      code += chars.charAt(Math.floor(Math.random() * chars.length))
    }
    return code
  }

  async createDeviceAuthorizationRequest(
    clientId: string,
    scope?: string,
  ): Promise<{
    deviceCode: string
    userCode: string
    expiresIn: number
    interval: number
  }> {
    const deviceCode = this.generateDeviceCode()
    let userCode = this.generateUserCode()

    // Ensure userCode is unique
    let attempts = 0
    while (await this.deviceAuthRepository.findOne({ where: { userCode } })) {
      userCode = this.generateUserCode()
      attempts++
      if (attempts > 10) {
        throw new BadRequestException('Failed to generate unique user code')
      }
    }

    const expiresAt = new Date(Date.now() + this.DEVICE_CODE_EXPIRY_SECONDS * 1000)

    const request = this.deviceAuthRepository.create({
      deviceCode,
      userCode,
      clientId,
      scope,
      status: DeviceAuthStatus.PENDING,
      expiresAt,
    })

    await this.deviceAuthRepository.save(request)

    return {
      deviceCode,
      userCode,
      expiresIn: this.DEVICE_CODE_EXPIRY_SECONDS,
      interval: this.POLLING_INTERVAL_SECONDS,
    }
  }

  async getDeviceAuthorizationByUserCode(userCode: string): Promise<DeviceAuthorizationRequest | null> {
    const normalizedCode = userCode.toUpperCase().trim()
    return this.deviceAuthRepository.findOne({
      where: { userCode: normalizedCode },
    })
  }

  async pollForToken(
    deviceCode: string,
    clientId: string,
  ): Promise<{
    status: 'pending' | 'approved' | 'denied' | 'expired' | 'slow_down'
    accessToken?: string
    organizationId?: string
    organizationName?: string
    scope?: string
  }> {
    const request = await this.deviceAuthRepository.findOne({
      where: { deviceCode, clientId },
    })

    if (!request) {
      throw new NotFoundException('Device authorization request not found')
    }

    // Check if expired
    if (new Date() > request.expiresAt) {
      if (request.status !== DeviceAuthStatus.EXPIRED) {
        request.status = DeviceAuthStatus.EXPIRED
        await this.deviceAuthRepository.save(request)
      }
      return { status: 'expired' }
    }

    // Check polling rate
    if (request.lastPolledAt) {
      const timeSinceLastPoll = Date.now() - request.lastPolledAt.getTime()
      if (timeSinceLastPoll < this.MIN_POLL_INTERVAL_MS) {
        return { status: 'slow_down' }
      }
    }

    // Update last polled time
    request.lastPolledAt = new Date()
    await this.deviceAuthRepository.save(request)

    if (request.status === DeviceAuthStatus.PENDING) {
      return { status: 'pending' }
    }

    if (request.status === DeviceAuthStatus.DENIED) {
      return { status: 'denied' }
    }

    if (request.status === DeviceAuthStatus.APPROVED && request.accessToken && request.organizationId) {
      // Get organization name
      let organizationName = 'Unknown'
      try {
        const org = await this.organizationService.findOne(request.organizationId)
        organizationName = org.name
      } catch {
        // Ignore error, use default name
      }

      return {
        status: 'approved',
        accessToken: request.accessToken,
        organizationId: request.organizationId,
        organizationName,
        scope: request.scope || undefined,
      }
    }

    return { status: 'pending' }
  }

  async approveDeviceAuthorization(
    userCode: string,
    userId: string,
    organizationId: string,
  ): Promise<{ success: boolean; message: string }> {
    const normalizedCode = userCode.toUpperCase().trim()
    const request = await this.deviceAuthRepository.findOne({
      where: { userCode: normalizedCode },
    })

    if (!request) {
      throw new NotFoundException('Device authorization request not found')
    }

    if (new Date() > request.expiresAt) {
      request.status = DeviceAuthStatus.EXPIRED
      await this.deviceAuthRepository.save(request)
      throw new BadRequestException('Device authorization request has expired')
    }

    if (request.status !== DeviceAuthStatus.PENDING) {
      throw new BadRequestException('Device authorization request is no longer pending')
    }

    // Get all available permissions for a full-access API key
    const permissions = Object.values(OrganizationResourcePermission)

    // Generate API key using the existing API key service
    const { value: accessToken } = await this.apiKeyService.createApiKey(
      organizationId,
      userId,
      `cli-device-${Date.now()}`,
      permissions,
    )

    request.status = DeviceAuthStatus.APPROVED
    request.userId = userId
    request.organizationId = organizationId
    request.accessToken = accessToken
    request.approvedAt = new Date()

    await this.deviceAuthRepository.save(request)

    return { success: true, message: 'Device authorization approved' }
  }

  async denyDeviceAuthorization(userCode: string): Promise<{ success: boolean; message: string }> {
    const normalizedCode = userCode.toUpperCase().trim()
    const request = await this.deviceAuthRepository.findOne({
      where: { userCode: normalizedCode },
    })

    if (!request) {
      throw new NotFoundException('Device authorization request not found')
    }

    if (new Date() > request.expiresAt) {
      request.status = DeviceAuthStatus.EXPIRED
      await this.deviceAuthRepository.save(request)
      throw new BadRequestException('Device authorization request has expired')
    }

    if (request.status !== DeviceAuthStatus.PENDING) {
      throw new BadRequestException('Device authorization request is no longer pending')
    }

    request.status = DeviceAuthStatus.DENIED
    await this.deviceAuthRepository.save(request)

    return { success: true, message: 'Device authorization denied' }
  }

  @Cron(CronExpression.EVERY_HOUR)
  async cleanupExpiredRequests(): Promise<void> {
    const oneHourAgo = new Date(Date.now() - 60 * 60 * 1000)

    const result = await this.deviceAuthRepository.delete({
      expiresAt: LessThan(oneHourAgo),
    })

    if (result.affected && result.affected > 0) {
      this.logger.log(`Cleaned up ${result.affected} expired device authorization requests`)
    }
  }
}
