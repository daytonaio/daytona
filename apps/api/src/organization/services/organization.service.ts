/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ForbiddenException, Injectable, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { EntityManager, In, IsNull, LessThan, MoreThan, Not, Or, Repository } from 'typeorm'
import { CreateOrganizationDto } from '../dto/create-organization.dto'
import { OverviewDto } from '../dto/overview.dto'
import { UpdateOrganizationQuotaDto } from '../dto/update-organization-quota.dto'
import { Organization } from '../entities/organization.entity'
import { OrganizationUser } from '../entities/organization-user.entity'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'
import { OnAsyncEvent } from '../../common/decorators/on-async-event.decorator'
import { UserEvents } from '../../user/constants/user-events.constant'
import { UserCreatedEvent } from '../../user/events/user-created.event'
import { UserDeletedEvent } from '../../user/events/user-deleted.event'
import { Workspace } from '../../workspace/entities/workspace.entity'
import { Image } from '../../workspace/entities/image.entity'
import { WorkspaceState } from '../../workspace/enums/workspace-state.enum'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { OrganizationEvents } from '../constants/organization-events.constant'
import { CreateOrganizationQuotaDto } from '../dto/create-organization-quota.dto'
import { DEFAULT_ORGANIZATION_QUOTA } from '../../common/constants/default-organization-quota'
import { ConfigService } from '@nestjs/config'
import { UserEmailVerifiedEvent } from '../../user/events/user-email-verified.event'
import { Volume } from '../../workspace/entities/volume.entity'
import { VolumeState } from '../../workspace/enums/volume-state.enum'

@Injectable()
export class OrganizationService {
  constructor(
    @InjectRepository(Organization)
    private readonly organizationRepository: Repository<Organization>,
    @InjectRepository(Workspace)
    private readonly workspaceRepository: Repository<Workspace>,
    @InjectRepository(Image)
    private readonly imageRepository: Repository<Image>,
    @InjectRepository(Volume)
    private readonly volumeRepository: Repository<Volume>,
    private readonly eventEmitter: EventEmitter2,
    private readonly configService: ConfigService,
  ) {}

  async create(
    createOrganizationDto: CreateOrganizationDto,
    createdBy: string,
    personal = false,
    creatorEmailVerified = false,
  ): Promise<Organization> {
    return this.createWithEntityManager(
      this.organizationRepository.manager,
      createOrganizationDto,
      createdBy,
      creatorEmailVerified,
      personal,
    )
  }

  async findByUser(userId: string): Promise<Organization[]> {
    return this.organizationRepository.find({
      where: {
        users: {
          userId,
        },
      },
    })
  }

  async findOne(organizationId: string): Promise<Organization | null> {
    return this.organizationRepository.findOne({
      where: { id: organizationId },
    })
  }

  async findSuspended(suspendedBefore?: Date): Promise<Organization[]> {
    return this.organizationRepository.find({
      where: {
        suspended: true,
        suspendedUntil: Or(IsNull(), MoreThan(new Date())),
        ...(suspendedBefore ? { suspendedAt: LessThan(suspendedBefore) } : {}),
      },
    })
  }

  async findPersonal(userId: string): Promise<Organization> {
    return this.findPersonalWithEntityManager(this.organizationRepository.manager, userId)
  }

  async delete(organizationId: string): Promise<void> {
    const organization = await this.organizationRepository.findOne({ where: { id: organizationId } })

    if (!organization) {
      throw new NotFoundException(`Organization with ID ${organizationId} not found`)
    }

    return this.removeWithEntityManager(this.organizationRepository.manager, organization)
  }

  async getUsageOverview(organizationId: string): Promise<OverviewDto> {
    const organization = await this.organizationRepository.findOne({ where: { id: organizationId } })
    if (!organization) {
      throw new NotFoundException(`Organization with ID ${organizationId} not found`)
    }

    // Get all workspaces for the organization, excluding destroyed and error ones
    const workspaces = await this.workspaceRepository.find({
      where: {
        organizationId,
        state: Not(In([WorkspaceState.DESTROYED, WorkspaceState.ERROR, WorkspaceState.ARCHIVED])),
      },
    })

    // Get running workspaces
    const runningWorkspaces = workspaces.filter((w) => w.state === WorkspaceState.STARTED)

    // Calculate current usage
    const currentCpuUsage = runningWorkspaces.reduce((sum, w) => sum + w.cpu, 0)
    const currentMemoryUsage = runningWorkspaces.reduce((sum, w) => sum + w.mem, 0)
    const currentDiskUsage = workspaces.reduce((sum, w) => sum + w.disk, 0)

    const currentImageNumber =
      (await this.imageRepository.count({
        where: {
          organizationId,
        },
      })) || 0
    const totalImageSizeUsed =
      (await this.imageRepository.sum('size', {
        organizationId,
      })) || 0

    const activeVolumesCount = await this.volumeRepository.count({
      where: {
        organizationId,
        state: Not(In([VolumeState.DELETED, VolumeState.ERROR])),
      },
    })

    return {
      totalCpuQuota: organization.totalCpuQuota,
      totalGpuQuota: 0,
      totalMemoryQuota: organization.totalMemoryQuota,
      totalDiskQuota: organization.totalDiskQuota,
      totalWorkspaceQuota: organization.workspaceQuota,
      concurrentWorkspaceQuota: organization.maxConcurrentWorkspaces,
      currentCpuUsage,
      currentMemoryUsage,
      currentDiskUsage,
      currentWorkspaces: workspaces.length,
      concurrentWorkspaces: runningWorkspaces.length,
      currentImageNumber,
      imageQuota: organization.imageQuota,
      totalImageSizeQuota: organization.totalImageSize,
      totalImageSizeUsed,
      maxVolumes: organization.volumeQuota,
      usedVolumes: activeVolumesCount,
    }
  }

  async updateQuota(
    organizationId: string,
    updateOrganizationQuotaDto: UpdateOrganizationQuotaDto,
  ): Promise<Organization> {
    const organization = await this.organizationRepository.findOne({ where: { id: organizationId } })
    if (!organization) {
      throw new NotFoundException(`Organization with ID ${organizationId} not found`)
    }

    organization.totalCpuQuota = updateOrganizationQuotaDto.totalCpuQuota
    organization.totalMemoryQuota = updateOrganizationQuotaDto.totalMemoryQuota
    organization.totalDiskQuota = updateOrganizationQuotaDto.totalDiskQuota
    organization.maxCpuPerWorkspace = updateOrganizationQuotaDto.maxCpuPerWorkspace
    organization.maxMemoryPerWorkspace = updateOrganizationQuotaDto.maxMemoryPerWorkspace
    organization.maxDiskPerWorkspace = updateOrganizationQuotaDto.maxDiskPerWorkspace
    organization.maxConcurrentWorkspaces = updateOrganizationQuotaDto.maxConcurrentWorkspaces
    organization.workspaceQuota = updateOrganizationQuotaDto.workspaceQuota
    organization.imageQuota = updateOrganizationQuotaDto.imageQuota
    organization.maxImageSize = updateOrganizationQuotaDto.maxImageSize
    organization.totalImageSize = updateOrganizationQuotaDto.totalImageSize
    organization.volumeQuota = updateOrganizationQuotaDto.volumeQuota

    return this.organizationRepository.save(organization)
  }

  async suspend(organizationId: string, suspensionReason?: string, suspendedUntil?: Date): Promise<void> {
    const organization = await this.organizationRepository.findOne({ where: { id: organizationId } })
    if (!organization) {
      throw new NotFoundException(`Organization with ID ${organizationId} not found`)
    }

    organization.suspended = true
    organization.suspensionReason = suspensionReason || null
    organization.suspendedUntil = suspendedUntil || null
    organization.suspendedAt = new Date()
    await this.organizationRepository.save(organization)
  }

  async unsuspend(organizationId: string): Promise<void> {
    const organization = await this.organizationRepository.findOne({ where: { id: organizationId } })
    if (!organization) {
      throw new NotFoundException(`Organization with ID ${organizationId} not found`)
    }

    organization.suspended = false
    organization.suspensionReason = null
    organization.suspendedUntil = null
    organization.suspendedAt = null

    await this.organizationRepository.save(organization)
  }

  private async createWithEntityManager(
    entityManager: EntityManager,
    createOrganizationDto: CreateOrganizationDto,
    createdBy: string,
    creatorEmailVerified: boolean,
    personal = false,
    quota: CreateOrganizationQuotaDto = DEFAULT_ORGANIZATION_QUOTA,
  ): Promise<Organization> {
    if (personal) {
      const count = await entityManager.count(Organization, {
        where: { createdBy, personal: true },
      })
      if (count > 0) {
        throw new ForbiddenException('Personal organization already exists')
      }
    }

    // set some limit to the number of created organizations
    const createdCount = await entityManager.count(Organization, {
      where: { createdBy },
    })
    if (createdCount >= 10) {
      throw new ForbiddenException('You have reached the maximum number of created organizations')
    }

    let organization = new Organization()

    organization.name = createOrganizationDto.name
    organization.createdBy = createdBy
    organization.personal = personal

    organization.totalCpuQuota = quota.totalCpuQuota
    organization.totalMemoryQuota = quota.totalMemoryQuota
    organization.totalDiskQuota = quota.totalDiskQuota
    organization.maxCpuPerWorkspace = quota.maxCpuPerWorkspace
    organization.maxMemoryPerWorkspace = quota.maxMemoryPerWorkspace
    organization.maxDiskPerWorkspace = quota.maxDiskPerWorkspace
    organization.maxConcurrentWorkspaces = quota.maxConcurrentWorkspaces
    organization.workspaceQuota = quota.workspaceQuota
    organization.imageQuota = quota.imageQuota
    organization.maxImageSize = quota.maxImageSize
    organization.totalImageSize = quota.totalImageSize
    organization.volumeQuota = quota.volumeQuota

    if (!creatorEmailVerified) {
      organization.suspended = true
      organization.suspendedAt = new Date()
      organization.suspensionReason = 'Please verify your email address'
    } else if (this.configService.get<boolean>('BILLING_ENABLED') && !personal) {
      organization.suspended = true
      organization.suspendedAt = new Date()
      organization.suspensionReason = 'Payment method required'
    }

    const owner = new OrganizationUser()
    owner.userId = createdBy
    owner.role = OrganizationMemberRole.OWNER

    organization.users = [owner]

    await entityManager.transaction(async (em) => {
      organization = await em.save(organization)
      await this.eventEmitter.emitAsync(OrganizationEvents.CREATED, organization)
    })

    return organization
  }

  private async removeWithEntityManager(
    entityManager: EntityManager,
    organization: Organization,
    force = false,
  ): Promise<void> {
    if (!force) {
      if (organization.personal) {
        throw new ForbiddenException('Cannot delete personal organization')
      }
    }
    await entityManager.remove(organization)
  }

  private async unsuspendPersonalWithEntityManager(entityManager: EntityManager, userId: string): Promise<void> {
    const organization = await this.findPersonalWithEntityManager(entityManager, userId)

    organization.suspended = false
    organization.suspendedAt = null
    organization.suspensionReason = null
    organization.suspendedUntil = null
    await entityManager.save(organization)
  }

  private async findPersonalWithEntityManager(entityManager: EntityManager, userId: string): Promise<Organization> {
    const organization = await entityManager.findOne(Organization, {
      where: { createdBy: userId, personal: true },
    })

    if (!organization) {
      throw new NotFoundException(`Personal organization for user ${userId} not found`)
    }

    return organization
  }

  @OnAsyncEvent({
    event: UserEvents.CREATED,
  })
  async handleUserCreatedEvent(payload: UserCreatedEvent): Promise<Organization> {
    return this.createWithEntityManager(
      payload.entityManager,
      {
        name: 'Personal',
      },
      payload.userId,
      payload.emailVerified || false,
      true,
      payload.personalOrganizationQuota,
    )
  }

  @OnAsyncEvent({
    event: UserEvents.EMAIL_VERIFIED,
  })
  async handleUserEmailVerifiedEvent(payload: UserEmailVerifiedEvent): Promise<void> {
    await this.unsuspendPersonalWithEntityManager(payload.entityManager, payload.userId)
  }

  @OnAsyncEvent({
    event: UserEvents.DELETED,
  })
  async handleUserDeletedEvent(payload: UserDeletedEvent): Promise<void> {
    const organization = await this.findPersonalWithEntityManager(payload.entityManager, payload.userId)

    await this.removeWithEntityManager(payload.entityManager, organization, true)
  }
}
