/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  BadRequestException,
  ConflictException,
  ForbiddenException,
  Injectable,
  NotFoundException,
} from '@nestjs/common'
import { ConfigService } from '@nestjs/config'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { InjectRepository } from '@nestjs/typeorm'
import { DataSource, MoreThan, Repository } from 'typeorm'
import { OrganizationRoleService } from './organization-role.service'
import { OrganizationUserService } from './organization-user.service'
import { OrganizationService } from './organization.service'
import { OrganizationEvents } from '../constants/organization-events.constant'
import { CreateOrganizationInvitationDto } from '../dto/create-organization-invitation.dto'
import { UpdateOrganizationInvitationDto } from '../dto/update-organization-invitation.dto'
import { OrganizationInvitation } from '../entities/organization-invitation.entity'
import { OrganizationInvitationStatus } from '../enums/organization-invitation-status.enum'
import { OrganizationInvitationAcceptedEvent } from '../events/organization-invitation-accepted.event'
import { OrganizationInvitationCreatedEvent } from '../events/organization-invitation-created.event'
import { UserService } from '../../user/user.service'
import { EmailUtils } from '../../common/utils/email.util'

@Injectable()
export class OrganizationInvitationService {
  constructor(
    @InjectRepository(OrganizationInvitation)
    private readonly organizationInvitationRepository: Repository<OrganizationInvitation>,
    private readonly organizationService: OrganizationService,
    private readonly organizationUserService: OrganizationUserService,
    private readonly organizationRoleService: OrganizationRoleService,
    private readonly userService: UserService,
    private readonly eventEmitter: EventEmitter2,
    private readonly dataSource: DataSource,
    private readonly configService: ConfigService,
  ) {}

  async create(
    organizationId: string,
    createOrganizationInvitationDto: CreateOrganizationInvitationDto,
    invitedBy: string,
  ): Promise<OrganizationInvitation> {
    const organization = await this.organizationService.findOne(organizationId)
    if (!organization) {
      throw new NotFoundException(`Organization with ID ${organizationId} not found`)
    }
    if (organization.personal) {
      throw new ForbiddenException('Cannot invite users to personal organization')
    }

    const normalizedEmail = EmailUtils.normalize(createOrganizationInvitationDto.email)

    const existingUser = await this.userService.findOneByEmail(normalizedEmail, true)
    if (existingUser) {
      const organizationUser = await this.organizationUserService.findOne(organizationId, existingUser.id)
      if (organizationUser) {
        throw new ConflictException(`User with email ${normalizedEmail} is already associated with this organization`)
      }
    }

    const existingInvitation = await this.organizationInvitationRepository.findOne({
      where: {
        organizationId,
        email: normalizedEmail,
        status: OrganizationInvitationStatus.PENDING,
        expiresAt: MoreThan(new Date()),
      },
    })
    if (existingInvitation) {
      throw new ConflictException(`User with email "${normalizedEmail}" already invited to this organization`)
    }

    let invitation = new OrganizationInvitation()
    invitation.organizationId = organizationId
    invitation.organization = organization
    invitation.email = normalizedEmail
    invitation.expiresAt = createOrganizationInvitationDto.expiresAt || new Date(Date.now() + 7 * 24 * 60 * 60 * 1000)
    invitation.role = createOrganizationInvitationDto.role
    invitation.invitedBy = invitedBy

    const assignedRoles = await this.organizationRoleService.findByIds(createOrganizationInvitationDto.assignedRoleIds)
    if (assignedRoles.length !== createOrganizationInvitationDto.assignedRoleIds.length) {
      throw new BadRequestException('One or more role IDs are invalid')
    }
    invitation.assignedRoles = assignedRoles

    invitation = await this.organizationInvitationRepository.save(invitation)

    this.eventEmitter.emit(
      OrganizationEvents.INVITATION_CREATED,
      new OrganizationInvitationCreatedEvent(
        invitation.organization.name,
        invitation.invitedBy,
        invitation.email,
        invitation.id,
        invitation.expiresAt,
      ),
    )

    return invitation
  }

  async update(
    invitationId: string,
    updateOrganizationInvitationDto: UpdateOrganizationInvitationDto,
  ): Promise<OrganizationInvitation> {
    const invitation = await this.organizationInvitationRepository.findOne({
      where: { id: invitationId },
      relations: {
        organization: true,
        assignedRoles: true,
      },
    })

    if (!invitation) {
      throw new NotFoundException(`Invitation with ID ${invitationId} not found`)
    }

    if (invitation.expiresAt && invitation.expiresAt < new Date()) {
      throw new ForbiddenException(`Invitation with ID ${invitationId} is expired`)
    }

    if (invitation.status !== OrganizationInvitationStatus.PENDING) {
      throw new ForbiddenException(`Invitation with ID ${invitationId} is already ${invitation.status}`)
    }

    if (updateOrganizationInvitationDto.expiresAt) {
      invitation.expiresAt = updateOrganizationInvitationDto.expiresAt
    }
    invitation.role = updateOrganizationInvitationDto.role

    const assignedRoles = await this.organizationRoleService.findByIds(updateOrganizationInvitationDto.assignedRoleIds)
    if (assignedRoles.length !== updateOrganizationInvitationDto.assignedRoleIds.length) {
      throw new BadRequestException('One or more role IDs are invalid')
    }
    invitation.assignedRoles = assignedRoles

    return this.organizationInvitationRepository.save(invitation)
  }

  async findPending(organizationId: string): Promise<OrganizationInvitation[]> {
    return this.organizationInvitationRepository.find({
      where: {
        organizationId,
        status: OrganizationInvitationStatus.PENDING,
        expiresAt: MoreThan(new Date()),
      },
      relations: {
        organization: true,
        assignedRoles: true,
      },
    })
  }

  async findByUser(userId: string): Promise<OrganizationInvitation[]> {
    const user = await this.userService.findOne(userId)

    if (!user) {
      throw new NotFoundException(`User with ID ${userId} not found`)
    }

    return this.organizationInvitationRepository.find({
      where: {
        email: EmailUtils.normalize(user.email),
        status: OrganizationInvitationStatus.PENDING,
        expiresAt: MoreThan(new Date()),
      },
      relations: {
        organization: true,
        assignedRoles: true,
      },
    })
  }

  async getCountByUser(userId: string): Promise<number> {
    const user = await this.userService.findOne(userId)

    if (!user) {
      throw new NotFoundException(`User with ID ${userId} not found`)
    }

    return this.organizationInvitationRepository.count({
      where: {
        email: EmailUtils.normalize(user.email),
        status: OrganizationInvitationStatus.PENDING,
        expiresAt: MoreThan(new Date()),
      },
    })
  }

  async findOneOrFail(invitationId: string): Promise<OrganizationInvitation> {
    return this.organizationInvitationRepository.findOneOrFail({
      where: { id: invitationId },
      relations: {
        organization: true,
        assignedRoles: true,
      },
    })
  }

  async accept(invitationId: string, userId: string): Promise<void> {
    const invitation = await this.prepareStatusUpdate(invitationId, OrganizationInvitationStatus.ACCEPTED)

    await this.dataSource.transaction(async (em) => {
      await em.save(invitation)
      await this.eventEmitter.emitAsync(
        OrganizationEvents.INVITATION_ACCEPTED,
        new OrganizationInvitationAcceptedEvent(
          em,
          invitation.organizationId,
          userId,
          invitation.role,
          invitation.assignedRoles,
        ),
      )
    })
  }

  async decline(invitationId: string): Promise<void> {
    const invitation = await this.prepareStatusUpdate(invitationId, OrganizationInvitationStatus.DECLINED)
    await this.organizationInvitationRepository.save(invitation)
  }

  async cancel(invitationId: string): Promise<void> {
    const invitation = await this.prepareStatusUpdate(invitationId, OrganizationInvitationStatus.CANCELLED)
    await this.organizationInvitationRepository.save(invitation)
  }

  private async prepareStatusUpdate(
    invitationId: string,
    newStatus: OrganizationInvitationStatus,
  ): Promise<OrganizationInvitation> {
    const invitation = await this.organizationInvitationRepository.findOne({
      where: { id: invitationId },
      relations: {
        organization: true,
        assignedRoles: true,
      },
    })

    if (!invitation) {
      throw new NotFoundException(`Invitation with ID ${invitationId} not found`)
    }

    if (invitation.expiresAt && invitation.expiresAt < new Date()) {
      throw new ForbiddenException(`Invitation with ID ${invitationId} is expired`)
    }

    if (invitation.status !== OrganizationInvitationStatus.PENDING) {
      throw new ForbiddenException(`Invitation with ID ${invitationId} is already ${invitation.status}`)
    }

    invitation.status = newStatus
    return invitation
  }
}
