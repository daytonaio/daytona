/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CanActivate, ExecutionContext, Injectable, Logger } from '@nestjs/common'
import { Reflector } from '@nestjs/core'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { OrganizationService } from '../services/organization.service'
import { OrganizationUserService } from '../services/organization-user.service'
import { RequiredOrganizationMemberRole } from '../decorators/required-organization-member-role.decorator'
import { RequiredOrganizationResourcePermissions } from '../decorators/required-organization-resource-permissions.decorator'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { UserAuthContext, isUserAuthContext } from '../../common/interfaces/user-auth-context.interface'
import { OrganizationAuthContext } from '../../common/interfaces/organization-auth-context.interface'
import { getAuthContext } from '../../common/utils/get-auth-context'
import { Organization } from '../entities/organization.entity'
import { OrganizationUser } from '../entities/organization-user.entity'
import { InvalidAuthenticationContextException } from '../../common/exceptions/invalid-authentication-context.exception'
import { AccessDeniedException } from '../../common/exceptions/access-denied.exception'

/**
 * Guard that validates access to an organization, enforces role/permission requirements, and enriches the auth context with organization data.
 *
 * Enforces the `@RequiredOrganizationMemberRole` and `@RequiredOrganizationResourcePermissions` decorators.
 */
@Injectable()
export class OrganizationAuthContextGuard implements CanActivate {
  private readonly logger = new Logger(OrganizationAuthContextGuard.name)

  constructor(
    @InjectRedis() private readonly redis: Redis,
    private readonly organizationService: OrganizationService,
    private readonly organizationUserService: OrganizationUserService,
    private readonly reflector: Reflector,
  ) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest()
    const authContext = getAuthContext(context, isUserAuthContext)

    const organizationId = this.resolveOrganizationId(request, authContext)
    if (!organizationId) {
      throw new InvalidAuthenticationContextException()
    }

    const organization = await this.getOrganization(organizationId)
    if (!organization) {
      throw new InvalidAuthenticationContextException()
    }

    const organizationUser = await this.getOrganizationUser(organizationId, authContext.userId)
    if (!organizationUser) {
      throw new InvalidAuthenticationContextException()
    }

    if (!this.isAuthorized(context, authContext, organizationUser)) {
      throw new AccessDeniedException()
    }

    request.user = {
      ...authContext,
      organizationId,
      organization,
      organizationUser,
    } satisfies OrganizationAuthContext

    return true
  }

  /**
   * Resolves the organization ID from the request params or current auth context.
   *
   * Additionally, validates that the API key's organization matches the requested organization.
   */
  private resolveOrganizationId(request: any, authContext: UserAuthContext): string | null {
    const organizationIdParam = request.params.organizationId || request.params.orgId

    if (!organizationIdParam && !authContext.organizationId) {
      this.logger.warn('Organization ID missing from the request context.')
      return null
    }

    if (organizationIdParam && authContext.apiKey && authContext.apiKey.organizationId !== organizationIdParam) {
      this.logger.warn(
        `Organization ID mismatch in the request context. Expected: ${organizationIdParam}, Actual: ${authContext.apiKey.organizationId}`,
      )
      return null
    }

    return organizationIdParam || authContext.organizationId || null
  }

  /**
   * Fetches an organization by ID, using a Redis cache to reduce DB lookups.
   */
  private async getOrganization(organizationId: string): Promise<Organization | null> {
    try {
      const cachedOrganization = await this.redis.get(`organization:${organizationId}`)
      if (cachedOrganization) {
        return JSON.parse(cachedOrganization)
      }

      // cache miss - fetch from DB
      const organization = await this.organizationService.findOne(organizationId)
      if (organization) {
        await this.redis.set(`organization:${organizationId}`, JSON.stringify(organization), 'EX', 10)
        return organization
      }

      // not found
      return null
    } catch (error) {
      this.logger.error('Error getting organization:', error)
      return null
    }
  }

  /**
   * Fetches an organization user by org and user ID, using a Redis cache to reduce DB lookups.
   */
  private async getOrganizationUser(organizationId: string, userId: string): Promise<OrganizationUser | null> {
    try {
      const cachedOrganizationUser = await this.redis.get(`organization-user:${organizationId}:${userId}`)
      if (cachedOrganizationUser) {
        return JSON.parse(cachedOrganizationUser)
      }

      // cache miss - fetch from DB
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

      // not found
      return null
    } catch (ex) {
      this.logger.error('Error getting organization user:', ex)
      return null
    }
  }

  /**
   * Checks if the organization user has authorization for the required role and(or) permissions.
   *
   * Enforces the `RequiredOrganizationMemberRole` and `RequiredOrganizationResourcePermissions` decorators.
   */
  private isAuthorized(
    context: ExecutionContext,
    authContext: UserAuthContext,
    organizationUser: OrganizationUser,
  ): boolean {
    const requiredRole = this.reflector.getAllAndOverride(RequiredOrganizationMemberRole, [
      context.getHandler(),
      context.getClass(),
    ])

    if (requiredRole && requiredRole !== organizationUser.role) {
      return false
    }

    // Owner has full access unless a scoped API key is used.
    if (organizationUser.role === OrganizationMemberRole.OWNER && !authContext.apiKey) {
      return true
    }

    const requiredPermissions = this.reflector.getAllAndOverride(RequiredOrganizationResourcePermissions, [
      context.getHandler(),
      context.getClass(),
    ])

    if (!requiredPermissions) {
      return true
    }

    const assignedPermissions = authContext.apiKey
      ? new Set(authContext.apiKey.permissions)
      : new Set(organizationUser.assignedRoles.flatMap((role) => role.permissions))

    return requiredPermissions.every((permission) => assignedPermissions.has(permission))
  }
}
