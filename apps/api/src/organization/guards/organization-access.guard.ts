/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CanActivate, ExecutionContext, Injectable, Logger } from '@nestjs/common'
import { OrganizationService } from '../services/organization.service'
import { OrganizationUserService } from '../services/organization-user.service'
import { isUserAuthContext } from '../../common/interfaces/user-auth-context.interface'
import { OrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { getAuthContext } from '../../common/utils/get-auth-context'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { Organization } from '../entities/organization.entity'
import { OrganizationUser } from '../entities/organization-user.entity'

@Injectable()
export class OrganizationAccessGuard implements CanActivate {
  protected readonly logger = new Logger(OrganizationAccessGuard.name)
  @InjectRedis() private readonly redis: Redis

  constructor(
    private readonly organizationService: OrganizationService,
    private readonly organizationUserService: OrganizationUserService,
  ) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const authContext = getAuthContext(context, isUserAuthContext)

    // note: semantic parameter names must be used (avoid :id)
    const organizationIdParam = request.params.organizationId || request.params.orgId

    if (
      authContext.role !== 'ssh-gateway' &&
      authContext.role !== 'runner' &&
      authContext.role !== 'proxy' &&
      authContext.role !== 'region-proxy' &&
      authContext.role !== 'region-ssh-gateway' &&
      !organizationIdParam &&
      !authContext.organizationId
    ) {
      this.logger.warn('Organization ID missing from the request context.')
      return false
    }

    if (
      organizationIdParam &&
      authContext.apiKey &&
      authContext.apiKey.organizationId !== organizationIdParam &&
      authContext.role !== 'ssh-gateway'
    ) {
      this.logger.warn(
        `Organization ID mismatch in the request context. Expected: ${organizationIdParam}, Actual: ${authContext.apiKey.organizationId}`,
      )
      return false
    }

    const organizationId = organizationIdParam || authContext.organizationId

    const organization = await this.getCachedOrganization(organizationId)

    if (!organization) {
      this.logger.warn(`Organization not found. Organization ID: ${organizationId}`)
      return false
    }

    const organizationUser = await this.getCachedOrganizationUser(organizationId, authContext.userId)

    if (!organizationUser) {
      this.logger.warn(
        `Organization user not found. User ID: ${authContext.userId}, Organization ID: ${organizationId}`,
      )
      return false
    }

    const organizationAuthContext: OrganizationAuthContext = {
      ...authContext,
      organizationId,
      organization,
      organizationUser,
    }
    request.user = organizationAuthContext

    return true
  }

  private async getCachedOrganization(organizationId: string): Promise<Organization | null> {
    try {
      const cachedOrganization = await this.redis.get(`organization:${organizationId}`)
      if (cachedOrganization) {
        return JSON.parse(cachedOrganization)
      }
      const organization = await this.organizationService.findOne(organizationId)
      if (organization) {
        await this.redis.set(`organization:${organizationId}`, JSON.stringify(organization), 'EX', 10)
        return organization
      }
      return null
    } catch (error) {
      this.logger.error('Error getting cached organization:', error)
      return null
    }
  }

  private async getCachedOrganizationUser(organizationId: string, userId: string): Promise<OrganizationUser | null> {
    try {
      const cachedOrganizationUser = await this.redis.get(`organization-user:${organizationId}:${userId}`)
      if (cachedOrganizationUser) {
        return JSON.parse(cachedOrganizationUser)
      }
      const organizationUser = await this.organizationUserService.findOne(organizationId, userId)
      if (organizationUser) {
        await this.redis.set(
          `organization-user:${organizationId}:${userId}`,
          JSON.stringify(organizationUser),
          'EX',
          10,
        )
        return organizationUser
      }
      return null
    } catch (ex) {
      this.logger.error('Error getting cached organization user:', ex)
      return null
    }
  }
}
