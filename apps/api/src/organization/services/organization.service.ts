/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ForbiddenException, Injectable, NotFoundException, Logger, OnModuleInit } from '@nestjs/common'
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
import { Sandbox } from '../../sandbox/entities/sandbox.entity'
import { Snapshot } from '../../sandbox/entities/snapshot.entity'
import { SandboxState } from '../../sandbox/enums/sandbox-state.enum'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { OrganizationEvents } from '../constants/organization-events.constant'
import { CreateOrganizationQuotaDto } from '../dto/create-organization-quota.dto'
import { DEFAULT_ORGANIZATION_QUOTA } from '../../common/constants/default-organization-quota'
import { UserEmailVerifiedEvent } from '../../user/events/user-email-verified.event'
import { Cron, CronExpression } from '@nestjs/schedule'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { RedisLockProvider } from '../../sandbox/common/redis-lock.provider'
import { OrganizationSuspendedSandboxStoppedEvent } from '../events/organization-suspended-sandbox-stopped.event'
import { SandboxDesiredState } from '../../sandbox/enums/sandbox-desired-state.enum'
import { SystemRole } from '../../user/enums/system-role.enum'
import { SnapshotState } from '../../sandbox/enums/snapshot-state.enum'
import { OrganizationSuspendedSnapshotDeactivatedEvent } from '../events/organization-suspended-snapshot-deactivated.event'
import { TypedConfigService } from '../../config/typed-config.service'

@Injectable()
export class OrganizationService implements OnModuleInit {
  private readonly logger = new Logger(OrganizationService.name)

  constructor(
    @InjectRedis() private readonly redis: Redis,
    @InjectRepository(Organization)
    private readonly organizationRepository: Repository<Organization>,
    @InjectRepository(Sandbox)
    private readonly sandboxRepository: Repository<Sandbox>,
    @InjectRepository(Snapshot)
    private readonly snapshotRepository: Repository<Snapshot>,
    private readonly eventEmitter: EventEmitter2,
    private readonly configService: TypedConfigService,
    private readonly redisLockProvider: RedisLockProvider,
  ) {}

  async onModuleInit(): Promise<void> {
    await this.stopSuspendedOrganizationSandboxes()
  }

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

    // Get all sandboxes for the organization, excluding destroyed and error ones
    const sandboxes = await this.sandboxRepository.find({
      where: {
        organizationId,
        state: Not(In([SandboxState.DESTROYED, SandboxState.ERROR, SandboxState.BUILD_FAILED, SandboxState.ARCHIVED])),
      },
    })

    // Get running sandboxes
    const runningSandboxes = sandboxes.filter((s) => s.state === SandboxState.STARTED)

    // Calculate current usage
    const currentCpuUsage = runningSandboxes.reduce((sum, s) => sum + s.cpu, 0)
    const currentMemoryUsage = runningSandboxes.reduce((sum, s) => sum + s.mem, 0)
    const currentDiskUsage = sandboxes.reduce((sum, s) => sum + s.disk, 0)

    return {
      totalCpuQuota: organization.totalCpuQuota,
      totalGpuQuota: 0,
      totalMemoryQuota: organization.totalMemoryQuota,
      totalDiskQuota: organization.totalDiskQuota,
      currentCpuUsage,
      currentMemoryUsage,
      currentDiskUsage,
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

    organization.totalCpuQuota = updateOrganizationQuotaDto.totalCpuQuota ?? organization.totalCpuQuota
    organization.totalMemoryQuota = updateOrganizationQuotaDto.totalMemoryQuota ?? organization.totalMemoryQuota
    organization.totalDiskQuota = updateOrganizationQuotaDto.totalDiskQuota ?? organization.totalDiskQuota
    organization.maxCpuPerSandbox = updateOrganizationQuotaDto.maxCpuPerSandbox ?? organization.maxCpuPerSandbox
    organization.maxMemoryPerSandbox =
      updateOrganizationQuotaDto.maxMemoryPerSandbox ?? organization.maxMemoryPerSandbox
    organization.maxDiskPerSandbox = updateOrganizationQuotaDto.maxDiskPerSandbox ?? organization.maxDiskPerSandbox
    organization.maxSnapshotSize = updateOrganizationQuotaDto.maxSnapshotSize ?? organization.maxSnapshotSize
    organization.volumeQuota = updateOrganizationQuotaDto.volumeQuota ?? organization.volumeQuota
    organization.snapshotQuota = updateOrganizationQuotaDto.snapshotQuota ?? organization.snapshotQuota
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
    organization.maxCpuPerSandbox = quota.maxCpuPerSandbox
    organization.maxMemoryPerSandbox = quota.maxMemoryPerSandbox
    organization.maxDiskPerSandbox = quota.maxDiskPerSandbox
    organization.snapshotQuota = quota.snapshotQuota
    organization.maxSnapshotSize = quota.maxSnapshotSize
    organization.volumeQuota = quota.volumeQuota

    if (!creatorEmailVerified) {
      organization.suspended = true
      organization.suspendedAt = new Date()
      organization.suspensionReason = 'Please verify your email address'
    } else if (this.configService.get('billingApiUrl') && !personal) {
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

  @Cron(CronExpression.EVERY_MINUTE, { name: 'stop-suspended-organization-sandboxes' })
  async stopSuspendedOrganizationSandboxes(): Promise<void> {
    //  lock the sync to only run one instance at a time
    const lockKey = 'stop-suspended-organization-sandboxes'
    if (!(await this.redisLockProvider.lock(lockKey, 60))) {
      return
    }

    const queryResult = await this.organizationRepository
      .createQueryBuilder('organization')
      .select('id')
      .where('suspended = true')
      .andWhere(`"suspendedAt" < NOW() - INTERVAL '1 day'`)
      .andWhere(`"suspendedAt" > NOW() - INTERVAL '7 day'`)
      .andWhereExists(
        this.sandboxRepository
          .createQueryBuilder('sandbox')
          .select('1')
          .where(
            `"sandbox"."organizationId" = "organization"."id" AND "sandbox"."desiredState" = '${SandboxDesiredState.STARTED}' and "sandbox"."state" NOT IN ('${SandboxState.ERROR}', '${SandboxState.BUILD_FAILED}')`,
          ),
      )
      .take(100)
      .getRawMany()

    const suspendedOrganizationIds = queryResult.map((result) => result.id)

    // Skip if no suspended organizations found to avoid empty IN clause
    if (suspendedOrganizationIds.length === 0) {
      await this.redisLockProvider.unlock(lockKey)
      return
    }

    const sandboxes = await this.sandboxRepository.find({
      where: {
        organizationId: In(suspendedOrganizationIds),
        desiredState: SandboxDesiredState.STARTED,
        state: Not(In([SandboxState.ERROR, SandboxState.BUILD_FAILED])),
      },
    })

    sandboxes.map((sandbox) =>
      this.eventEmitter.emitAsync(
        OrganizationEvents.SUSPENDED_SANDBOX_STOPPED,
        new OrganizationSuspendedSandboxStoppedEvent(sandbox.id),
      ),
    )

    await this.redisLockProvider.unlock(lockKey)
  }

  @Cron(CronExpression.EVERY_MINUTE, { name: 'deactivate-suspended-organization-snapshots' })
  async deactivateSuspendedOrganizationSnapshots(): Promise<void> {
    //  lock the sync to only run one instance at a time
    const lockKey = 'deactivate-suspended-organization-snapshots'
    if (!(await this.redisLockProvider.lock(lockKey, 60))) {
      return
    }

    const queryResult = await this.organizationRepository
      .createQueryBuilder('organization')
      .select('id')
      .where('suspended = true')
      .andWhere(`"suspendedAt" < NOW() - INTERVAL '1 day'`)
      .andWhere(`"suspendedAt" > NOW() - INTERVAL '7 day'`)
      .andWhereExists(
        this.snapshotRepository
          .createQueryBuilder('snapshot')
          .select('1')
          .where('snapshot.organizationId = organization.id')
          .andWhere(`snapshot.state = '${SnapshotState.ACTIVE}'`)
          .andWhere(`snapshot.general = false`),
      )
      .take(100)
      .getRawMany()

    const suspendedOrganizationIds = queryResult.map((result) => result.id)

    // Skip if no suspended organizations found to avoid empty IN clause
    if (suspendedOrganizationIds.length === 0) {
      await this.redisLockProvider.unlock(lockKey)
      return
    }

    const snapshotQueryResult = await this.snapshotRepository
      .createQueryBuilder('snapshot')
      .select('id')
      .where('snapshot.organizationId IN (:...suspendedOrgIds)', { suspendedOrgIds: suspendedOrganizationIds })
      .andWhere(`snapshot.state = '${SnapshotState.ACTIVE}'`)
      .andWhere(`snapshot.general = false`)
      .take(100)
      .getRawMany()

    const snapshotIds = snapshotQueryResult.map((result) => result.id)

    snapshotIds.map((id) =>
      this.eventEmitter.emitAsync(
        OrganizationEvents.SUSPENDED_SNAPSHOT_DEACTIVATED,
        new OrganizationSuspendedSnapshotDeactivatedEvent(id),
      ),
    )

    await this.redisLockProvider.unlock(lockKey)
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
      payload.user.id,
      payload.user.role === SystemRole.ADMIN ? true : payload.user.emailVerified,
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

  assertOrganizationIsNotSuspended(organization: Organization): void {
    if (!organization.suspended) {
      return
    }

    if (organization.suspendedUntil ? organization.suspendedUntil > new Date() : true) {
      if (organization.suspensionReason) {
        throw new ForbiddenException(`Organization is suspended: ${organization.suspensionReason}`)
      } else {
        throw new ForbiddenException('Organization is suspended')
      }
    }
  }
}
