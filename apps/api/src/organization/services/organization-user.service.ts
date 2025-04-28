/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BadRequestException, ForbiddenException, Injectable, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { EntityManager, Repository } from 'typeorm'
import { OrganizationRoleService } from './organization-role.service'
import { OrganizationEvents } from '../constants/organization-events.constant'
import { OrganizationUserDto } from '../dto/organization-user.dto'
import { OrganizationUser } from '../entities/organization-user.entity'
import { OrganizationRole } from '../entities/organization-role.entity'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { OrganizationInvitationAcceptedEvent } from '../events/organization-invitation-accepted.event'
import { OnAsyncEvent } from '../../common/decorators/on-async-event.decorator'
import { UserService } from '../../user/user.service'
import { UserEvents } from '../../user/constants/user-events.constant'
import { UserDeletedEvent } from '../../user/events/user-deleted.event'

@Injectable()
export class OrganizationUserService {
  constructor(
    @InjectRepository(OrganizationUser)
    private readonly organizationUserRepository: Repository<OrganizationUser>,
    private readonly organizationRoleService: OrganizationRoleService,
    private readonly userService: UserService,
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

  async findOne(organizationId: string, userId: string): Promise<OrganizationUser | null> {
    return this.organizationUserRepository.findOne({
      where: { organizationId, userId },
      relations: {
        assignedRoles: true,
      },
    })
  }

  async updateRole(organizationId: string, userId: string, role: OrganizationMemberRole): Promise<OrganizationUserDto> {
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

    organizationUser.role = role
    organizationUser = await this.organizationUserRepository.save(organizationUser)

    const user = await this.userService.findOne(userId)

    return OrganizationUserDto.fromEntities(organizationUser, user)
  }

  async updateAssignedRoles(organizationId: string, userId: string, roleIds: string[]): Promise<OrganizationUserDto> {
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

    const roles = await this.organizationRoleService.findByIds(roleIds)

    if (roles.length !== roleIds.length) {
      throw new BadRequestException('One or more role IDs are invalid')
    }

    organizationUser.assignedRoles = roles
    organizationUser = await this.organizationUserRepository.save(organizationUser)

    const user = await this.userService.findOne(userId)

    return OrganizationUserDto.fromEntities(organizationUser, user)
  }

  async assignRoles(organizationId: string, userId: string, roleIds: string[]): Promise<OrganizationUserDto> {
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

    const newRoles = await this.organizationRoleService.findByIds(roleIds)

    if (newRoles.length !== roleIds.length) {
      throw new BadRequestException('One or more role IDs are invalid')
    }

    organizationUser.assignedRoles = [
      ...organizationUser.assignedRoles,
      ...newRoles.filter(
        (newRole) => !organizationUser.assignedRoles.some((existingRole) => existingRole.id === newRole.id),
      ),
    ]
    organizationUser = await this.organizationUserRepository.save(organizationUser)

    const user = await this.userService.findOne(userId)

    return OrganizationUserDto.fromEntities(organizationUser, user)
  }

  async unassignRoles(organizationId: string, userId: string, roleIds: string[]): Promise<OrganizationUserDto> {
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

    const removedRoles = await this.organizationRoleService.findByIds(roleIds)

    if (removedRoles.length !== roleIds.length) {
      throw new BadRequestException('One or more role IDs are invalid')
    }

    organizationUser.assignedRoles = organizationUser.assignedRoles.filter(
      (existingRole) => !roleIds.includes(existingRole.id),
    )
    organizationUser = await this.organizationUserRepository.save(organizationUser)

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
    await Promise.all(memberships.map((membership) => this.removeWithEntityManager(payload.entityManager, membership)))
  }
}
