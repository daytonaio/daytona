/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BadRequestException, ForbiddenException, Injectable, Logger, NotFoundException } from '@nestjs/common'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { InjectRepository } from '@nestjs/typeorm'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { DataSource, EntityManager, Repository } from 'typeorm'
import { OrganizationRoleService } from './organization-role.service'
import { OrganizationEvents } from '../constants/organization-events.constant'
import { OrganizationUserDto } from '../dto/organization-user.dto'
import { OrganizationUser } from '../entities/organization-user.entity'
import { OrganizationRole } from '../entities/organization-role.entity'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { OrganizationResourcePermission } from '../enums/organization-resource-permission.enum'
import { OrganizationInvitationAcceptedEvent } from '../events/organization-invitation-accepted.event'
import { OrganizationResourcePermissionsUnassignedEvent } from '../events/organization-resource-permissions-unassigned.event'
import { OnAsyncEvent } from '../../common/decorators/on-async-event.decorator'
import { UserService } from '../../user/user.service'
import { UserEvents } from '../../user/constants/user-events.constant'
import { UserDeletedEvent } from '../../user/events/user-deleted.event'

@Injectable()
export class OrganizationUserService {
  private readonly logger = new Logger(OrganizationUserService.name)

  constructor(
    @InjectRepository(OrganizationUser)
    private readonly organizationUserRepository: Repository<OrganizationUser>,
    private readonly organizationRoleService: OrganizationRoleService,
    private readonly userService: UserService,
    private readonly eventEmitter: EventEmitter2,
    private readonly dataSource: DataSource,
    @InjectRedis() private readonly redis: Redis,
  ) {}

  /**
   * Evicts the cached organization-user authorization record so that role and permission
   * changes take effect immediately, rather than remaining stale for the duration of the
   * OrganizationAuthContextGuard cache TTL. Must be called after the mutation has committed.
   *
   * Cache eviction failures are logged and swallowed: they must not turn an already-committed
   * access change into a request failure, and the entry self-expires at the guard's TTL. Mirrors
   * ApiKeyService.invalidateApiKeyCache.
   */
  private async evictOrganizationUserCache(organizationId: string, userId: string): Promise<void> {
    try {
      await this.redis.del(`organization-user:${organizationId}:${userId}`)
    } catch (error) {
      this.logger.error('Failed to evict organization-user cache:', error)
    }
  }

  async findAll(organizationId: string): Promise<OrganizationUserDto[]> {
    const organizationUsers = await this.organizationUserRepository.find({
      where: { organizationId },
      relations: {
        assignedRoles: true,
      },
    })

    const userIds = organizationUsers.map((orgUser) => orgUser.userId)

    const users = await this.userService.findByIds(userIds)
    const userMap = new Map(users.map((user) => [user.id, user]))

    const dtos: OrganizationUserDto[] = organizationUsers.map((orgUser) => {
      const user = userMap.get(orgUser.userId)
      return OrganizationUserDto.fromEntities(orgUser, user)
    })

    return dtos
  }

  async findOne(organizationId: string, userId: string): Promise<OrganizationUser | null> {
    return this.organizationUserRepository.findOne({
      where: { organizationId, userId },
      relations: {
        assignedRoles: true,
      },
    })
  }

  async exists(organizationId: string, userId: string): Promise<boolean> {
    return this.organizationUserRepository.exists({
      where: { organizationId, userId },
    })
  }

  async updateAccess(
    organizationId: string,
    userId: string,
    role: OrganizationMemberRole,
    assignedRoleIds: string[],
  ): Promise<OrganizationUserDto> {
    let organizationUser = await this.organizationUserRepository.findOne({
      where: {
        organizationId,
        userId,
      },
      relations: {
        assignedRoles: true,
      },
    })

    if (!organizationUser) {
      throw new NotFoundException(`User with ID ${userId} not found in organization with ID ${organizationId}`)
    }

    // validate role
    if (organizationUser.role === OrganizationMemberRole.OWNER && role !== OrganizationMemberRole.OWNER) {
      const ownersCount = await this.organizationUserRepository.count({
        where: {
          organizationId,
          role: OrganizationMemberRole.OWNER,
        },
      })

      if (ownersCount === 1) {
        throw new ForbiddenException('The organization must have at least one owner')
      }
    }

    // validate assignments
    const assignedRoles = await this.organizationRoleService.findByIds(organizationId, assignedRoleIds)
    if (assignedRoles.length !== assignedRoleIds.length) {
      throw new BadRequestException('One or more role IDs are invalid')
    }

    // check if any previous permissions are not present in the new assignments, api keys with those permissions will be revoked
    let permissionsToRevoke: OrganizationResourcePermission[] = []
    if (role !== OrganizationMemberRole.OWNER) {
      const prevPermissions = this.getAssignedPermissions(organizationUser.role, organizationUser.assignedRoles)
      const newPermissions = this.getAssignedPermissions(role, assignedRoles)
      permissionsToRevoke = Array.from(prevPermissions).filter((permission) => !newPermissions.has(permission))
    }

    organizationUser.role = role
    organizationUser.assignedRoles = assignedRoles

    if (permissionsToRevoke.length > 0) {
      await this.dataSource.transaction(async (em) => {
        organizationUser = await em.save(organizationUser)
        await this.eventEmitter.emitAsync(
          OrganizationEvents.PERMISSIONS_UNASSIGNED,
          new OrganizationResourcePermissionsUnassignedEvent(em, organizationId, userId, permissionsToRevoke),
        )
      })
    } else {
      organizationUser = await this.organizationUserRepository.save(organizationUser)
    }

    await this.evictOrganizationUserCache(organizationId, userId)

    const user = await this.userService.findOne(userId)

    return OrganizationUserDto.fromEntities(organizationUser, user)
  }

  async delete(organizationId: string, userId: string): Promise<void> {
    const organizationUser = await this.organizationUserRepository.findOne({
      where: {
        organizationId,
        userId,
      },
    })

    if (!organizationUser) {
      throw new NotFoundException(`User with ID ${userId} not found in organization with ID ${organizationId}`)
    }

    await this.removeWithEntityManager(this.organizationUserRepository.manager, organizationUser)

    // Evict after the removal has committed (the repository manager autocommits). Done here rather
    // than inside removeWithEntityManager so eviction never runs inside a caller-owned transaction
    // (e.g. the user-deletion flow), where a concurrent guard read could re-cache the still-present
    // row before commit.
    await this.evictOrganizationUserCache(organizationId, userId)
  }

  private async removeWithEntityManager(
    entityManager: EntityManager,
    organizationUser: OrganizationUser,
    force = false,
  ): Promise<void> {
    if (!force) {
      if (organizationUser.role === OrganizationMemberRole.OWNER) {
        const ownersCount = await entityManager.count(OrganizationUser, {
          where: {
            organizationId: organizationUser.organizationId,
            role: OrganizationMemberRole.OWNER,
          },
        })

        if (ownersCount === 1) {
          throw new ForbiddenException(
            `Organization with ID ${organizationUser.organizationId} must have at least one owner`,
          )
        }
      }
    }

    await entityManager.remove(organizationUser)
  }

  private async createWithEntityManager(
    entityManager: EntityManager,
    organizationId: string,
    userId: string,
    role: OrganizationMemberRole,
    assignedRoles: OrganizationRole[],
  ): Promise<OrganizationUser> {
    const organizationUser = new OrganizationUser()
    organizationUser.organizationId = organizationId
    organizationUser.userId = userId
    organizationUser.role = role
    organizationUser.assignedRoles = assignedRoles
    return entityManager.save(organizationUser)
  }

  @OnAsyncEvent({
    event: OrganizationEvents.INVITATION_ACCEPTED,
  })
  async handleOrganizationInvitationAcceptedEvent(
    payload: OrganizationInvitationAcceptedEvent,
  ): Promise<OrganizationUser> {
    return this.createWithEntityManager(
      payload.entityManager,
      payload.organizationId,
      payload.userId,
      payload.role,
      payload.assignedRoles,
    )
  }

  @OnAsyncEvent({
    event: UserEvents.DELETED,
  })
  async handleUserDeletedEvent(payload: UserDeletedEvent): Promise<void> {
    const memberships = await payload.entityManager.find(OrganizationUser, {
      where: {
        userId: payload.userId,
        organization: {
          personal: false,
        },
      },
      relations: {
        organization: true,
      },
    })

    /*
    // TODO
    // user deletion will fail if the user is the only owner of some non-personal organization
    // potential improvements:
    //  - auto-delete the organization if there are no other members
    //  - auto-promote a new owner if there are other members
    */
    // Cache eviction is intentionally not done here. The DELETED event is emitted inside the
    // deletion transaction, so evicting now would race a concurrent read re-caching the
    // not-yet-removed row before commit; effective eviction has to run post-commit. Any residual
    // entry is bounded by the guard's short cache TTL. Handled as a separate change.
    await Promise.all(memberships.map((membership) => this.removeWithEntityManager(payload.entityManager, membership)))
  }

  private getAssignedPermissions(
    role: OrganizationMemberRole,
    assignedRoles: OrganizationRole[],
  ): Set<OrganizationResourcePermission> {
    if (role === OrganizationMemberRole.OWNER) {
      return new Set(Object.values(OrganizationResourcePermission))
    }

    return new Set(assignedRoles.flatMap((role) => role.permissions))
  }
}
