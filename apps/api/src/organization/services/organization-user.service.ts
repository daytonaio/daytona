/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BadRequestException, ForbiddenException, Injectable, Logger, NotFoundException } from '@nestjs/common'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { InjectRepository } from '@nestjs/typeorm'
import { DataSource, EntityManager, Repository } from 'typeorm'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { OrganizationRoleService } from './organization-role.service'
import { OrganizationEvents } from '../constants/organization-events.constant'
import { OrganizationUserDto } from '../dto/organization-user.dto'
import { OrganizationUser } from '../entities/organization-user.entity'
import { OrganizationRole } from '../entities/organization-role.entity'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { OrganizationResourcePermission } from '../enums/organization-resource-permission.enum'
import { OrganizationInvitationAcceptedEvent } from '../events/organization-invitation-accepted.event'
import { OrganizationResourcePermissionsUnassignedEvent } from '../events/organization-resource-permissions-unassigned.event'
import { OrganizationDeletedEvent } from '../events/organization-deleted.event'
import { OnAsyncEvent } from '../../common/decorators/on-async-event.decorator'
import { UserService } from '../../user/user.service'
import { OrganizationAssertDeletableEvent } from '../events/organization-assert-deletable.event'
import { getOrganizationUserCacheKey } from '../constants/organization-cache-keys.constant'

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

  async countByUserId(userId: string): Promise<number> {
    return this.organizationUserRepository.count({ where: { userId } })
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
    const assignedRoles = await this.organizationRoleService.findByIds(assignedRoleIds)
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
    event: OrganizationEvents.ASSERT_NO_USERS,
  })
  async handleAssertNoUsers(event: OrganizationAssertDeletableEvent): Promise<void> {
    let count = 0

    try {
      count = await this.organizationUserRepository.count({
        where: { organizationId: event.organizationId },
      })
    } catch (error) {
      this.logger.error(
        `Failed to check if the organization ${event.organizationId} has users that must be removed`,
        error,
      )
      throw new Error('Failed to check if the organization has users that must be removed')
    }

    // not a single-user organization
    if (count > 1) {
      throw new Error(`Organization has ${count - 1} user(s) that must be removed from the organization`)
    }
  }

  @OnAsyncEvent({
    event: OrganizationEvents.DELETED,
  })
  async handleOrganizationDeletedEvent(payload: OrganizationDeletedEvent): Promise<void> {
    const { entityManager, organizationId } = payload

    // Get users before deletion to invalidate caches
    const users = await entityManager.find(OrganizationUser, {
      where: { organizationId },
      select: ['userId'],
    })

    await entityManager.delete(OrganizationUser, { organizationId })

    // Invalidate caches
    try {
      const cacheKeys = users.map((user) => getOrganizationUserCacheKey(organizationId, user.userId))
      if (cacheKeys.length > 0) {
        const BATCH_SIZE = 500
        for (let i = 0; i < cacheKeys.length; i += BATCH_SIZE) {
          const batch = cacheKeys.slice(i, i + BATCH_SIZE)
          await this.redis.del(...batch)
        }
      }
    } catch (error) {
      this.logger.error(`Failed to invalidate caches for organization ${organizationId}:`, error)
    }
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
