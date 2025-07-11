/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  ForbiddenException,
  Injectable,
  NotFoundException,
  Logger,
  OnModuleInit,
  BadRequestException,
} from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { EntityManager, In, Not, Repository } from 'typeorm'
import { CreateOrganizationDto } from '../dto/create-organization.dto'
import { OrganizationUsageOverviewDto } from '../dto/organization-usage-overview.dto'
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
import { ConfigService } from '@nestjs/config'
import { UserEmailVerifiedEvent } from '../../user/events/user-email-verified.event'
import { Volume } from '../../sandbox/entities/volume.entity'
import { Cron, CronExpression } from '@nestjs/schedule'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { RedisLockProvider } from '../../sandbox/common/redis-lock.provider'
import { OrganizationSuspendedSandboxStoppedEvent } from '../events/organization-suspended-sandbox-stopped.event'
import { SandboxDesiredState } from '../../sandbox/enums/sandbox-desired-state.enum'
import { SystemRole } from '../../user/enums/system-role.enum'
import { SnapshotState } from '../../sandbox/enums/snapshot-state.enum'
import { OrganizationSuspendedSnapshotDeactivatedEvent } from '../events/organization-suspended-snapshot-deactivated.event'
import { SandboxUsageOverviewInternalDto, SandboxUsageOverviewSchema } from '../dto/sandbox-usage-overview-internal.dto'
import {
  SnapshotUsageOverviewInternalDto,
  SnapshotUsageOverviewSchema,
} from '../dto/snapshot-usage-overview-internal.dto'
import { VolumeState } from '../../sandbox/enums/volume-state.enum'
import { VolumeUsageOverviewInternalDto, VolumeUsageOverviewSchema } from '../dto/volume-usage-overview-internal.dto'
import { SANDBOX_USAGE_OVERVIEW_IGNORED_STATES } from '../constants/sandbox-usage-overview-ignored-states.constant'
import { SANDBOX_USAGE_OVERVIEW_INACTIVE_STATES } from '../constants/sandbox-usage-overview-inactive-states.constant'

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
    @InjectRepository(Volume)
    private readonly volumeRepository: Repository<Volume>,
    private readonly eventEmitter: EventEmitter2,
    private readonly configService: ConfigService,
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

  async getSandboxUsageOverview(
    organizationId: string,
    organization?: Organization,
    excludeSandboxId?: string,
  ): Promise<SandboxUsageOverviewInternalDto> {
    if (organization && organization.id !== organizationId) {
      throw new BadRequestException('Organization ID mismatch')
    }

    if (!organization) {
      organization = await this.organizationRepository.findOne({ where: { id: organizationId } })
    }

    if (!organization) {
      throw new NotFoundException(`Organization with ID ${organizationId} not found`)
    }

    let sandboxUsageOverview: SandboxUsageOverviewInternalDto | null = null

    // check cache first
    const cacheKey = `sandbox-usage-${organization.id}`
    const cachedData = await this.redis.get(cacheKey)

    if (cachedData) {
      try {
        const parsed = JSON.parse(cachedData)
        sandboxUsageOverview = SandboxUsageOverviewSchema.parse(parsed) as SandboxUsageOverviewInternalDto
      } catch {
        this.logger.warn(`Failed to parse cached sandbox usage overview for organization ${organizationId}`)
        this.redis.del(cacheKey)
      }
    }

    // cache hit
    if (sandboxUsageOverview) {
      const oneHourAgo = new Date(Date.now() - 1000 * 60 * 60)

      if (sandboxUsageOverview._fetchedAt >= oneHourAgo) {
        if (excludeSandboxId) {
          return this.excludeSandboxFromUsageOverview(sandboxUsageOverview, excludeSandboxId)
        }

        return sandboxUsageOverview
      }

      // cache expired (fetched from db more than 1 hour ago), invalidate it
      await this.redis.del(cacheKey)
    }

    // cache miss
    const sandboxUsageMetrics: {
      used_disk: number
      used_cpu: number
      used_mem: number
    } = await this.sandboxRepository
      .createQueryBuilder('sandbox')
      .select([
        'SUM(CASE WHEN sandbox.state NOT IN (:...ignoredStates) THEN sandbox.disk ELSE 0 END) as used_disk',
        'SUM(CASE WHEN sandbox.state NOT IN (:...inactiveStates) THEN sandbox.cpu ELSE 0 END) as used_cpu',
        'SUM(CASE WHEN sandbox.state NOT IN (:...inactiveStates) THEN sandbox.mem ELSE 0 END) as used_mem',
      ])
      .where('sandbox.organizationId = :organizationId', { organizationId })
      .setParameter('ignoredStates', SANDBOX_USAGE_OVERVIEW_IGNORED_STATES)
      .setParameter('inactiveStates', SANDBOX_USAGE_OVERVIEW_INACTIVE_STATES)
      .getRawOne()

    const currentDiskUsage = Number(sandboxUsageMetrics.used_disk) || 0
    const currentCpuUsage = Number(sandboxUsageMetrics.used_cpu) || 0
    const currentMemoryUsage = Number(sandboxUsageMetrics.used_mem) || 0

    sandboxUsageOverview = {
      totalCpuQuota: organization.totalCpuQuota,
      totalMemoryQuota: organization.totalMemoryQuota,
      totalDiskQuota: organization.totalDiskQuota,
      currentCpuUsage,
      currentMemoryUsage,
      currentDiskUsage,
      _fetchedAt: new Date(),
    }

    // cache the result
    await this.redis.setex(cacheKey, 10, JSON.stringify(sandboxUsageOverview))

    if (excludeSandboxId) {
      return await this.excludeSandboxFromUsageOverview(sandboxUsageOverview, excludeSandboxId)
    }

    return sandboxUsageOverview
  }

  private async excludeSandboxFromUsageOverview(
    usageOverview: SandboxUsageOverviewInternalDto,
    excludeSandboxId: string,
  ): Promise<SandboxUsageOverviewInternalDto> {
    const excludedSandbox = await this.sandboxRepository.findOne({
      where: { id: excludeSandboxId },
    })

    if (!excludedSandbox) {
      return usageOverview
    }

    let cpuToSubtract = 0
    let memToSubtract = 0
    let diskToSubtract = 0

    if (!SANDBOX_USAGE_OVERVIEW_IGNORED_STATES.includes(excludedSandbox.state)) {
      diskToSubtract = excludedSandbox.disk
    }

    if (!SANDBOX_USAGE_OVERVIEW_INACTIVE_STATES.includes(excludedSandbox.state)) {
      cpuToSubtract = excludedSandbox.cpu
      memToSubtract = excludedSandbox.mem
    }

    return {
      ...usageOverview,
      currentCpuUsage: Math.max(0, usageOverview.currentCpuUsage - cpuToSubtract),
      currentMemoryUsage: Math.max(0, usageOverview.currentMemoryUsage - memToSubtract),
      currentDiskUsage: Math.max(0, usageOverview.currentDiskUsage - diskToSubtract),
    }
  }

  async getSnapshotUsageOverview(
    organizationId: string,
    organization?: Organization,
  ): Promise<SnapshotUsageOverviewInternalDto> {
    if (organization && organization.id !== organizationId) {
      throw new BadRequestException('Organization ID mismatch')
    }

    if (!organization) {
      organization = await this.organizationRepository.findOne({ where: { id: organizationId } })
    }

    if (!organization) {
      throw new NotFoundException(`Organization with ID ${organizationId} not found`)
    }

    let snapshotUsageOverview: SnapshotUsageOverviewInternalDto | null = null

    // check cache first
    const cacheKey = `snapshot-usage-${organizationId}`
    const cachedData = await this.redis.get(cacheKey)

    if (cachedData) {
      try {
        const parsed = JSON.parse(cachedData)
        return SnapshotUsageOverviewSchema.parse(parsed) as SnapshotUsageOverviewInternalDto
      } catch {
        this.logger.warn(`Failed to parse cached snapshot usage overview for organization ${organizationId}`)
        this.redis.del(cacheKey)
      }
    }

    // cache hit
    if (snapshotUsageOverview) {
      const oneHourAgo = new Date(Date.now() - 1000 * 60 * 60)

      if (snapshotUsageOverview._fetchedAt >= oneHourAgo) {
        return snapshotUsageOverview
      }

      // cache expired (fetched from db more than 1 hour ago), invalidate it
      await this.redis.del(cacheKey)
    }

    // cache miss
    const currentSnapshotUsage = await this.snapshotRepository.count({
      where: {
        organizationId,
        state: Not(In([SnapshotState.ERROR, SnapshotState.BUILD_FAILED, SnapshotState.INACTIVE])),
      },
    })

    snapshotUsageOverview = {
      totalSnapshotQuota: organization.snapshotQuota,
      currentSnapshotUsage,
      _fetchedAt: new Date(),
    }

    // cache the result
    await this.redis.setex(cacheKey, 10, JSON.stringify(snapshotUsageOverview))

    return snapshotUsageOverview
  }

  async getVolumeUsageOverview(
    organizationId: string,
    organization?: Organization,
  ): Promise<VolumeUsageOverviewInternalDto> {
    if (organization && organization.id !== organizationId) {
      throw new BadRequestException('Organization ID mismatch')
    }

    if (!organization) {
      organization = await this.organizationRepository.findOne({ where: { id: organizationId } })
    }

    if (!organization) {
      throw new NotFoundException(`Organization with ID ${organizationId} not found`)
    }

    let volumeUsageOverview: VolumeUsageOverviewInternalDto | null = null

    // check cache first
    const cacheKey = `volume-usage-${organizationId}`
    const cachedData = await this.redis.get(cacheKey)

    if (cachedData) {
      try {
        const parsed = JSON.parse(cachedData)
        return VolumeUsageOverviewSchema.parse(parsed) as VolumeUsageOverviewInternalDto
      } catch {
        this.logger.warn(`Failed to parse cached volume usage overview for organization ${organizationId}`)
        this.redis.del(cacheKey)
      }
    }

    // cache hit
    if (volumeUsageOverview) {
      const oneHourAgo = new Date(Date.now() - 1000 * 60 * 60)

      if (volumeUsageOverview._fetchedAt >= oneHourAgo) {
        return volumeUsageOverview
      }

      // cache expired (fetched from db more than 1 hour ago), invalidate it
      await this.redis.del(cacheKey)
    }

    // cache miss
    const currentVolumeUsage = await this.volumeRepository.count({
      where: {
        organizationId,
        state: Not(In([VolumeState.DELETED, VolumeState.ERROR])),
      },
    })

    volumeUsageOverview = {
      totalVolumeQuota: organization.volumeQuota,
      currentVolumeUsage,
      _fetchedAt: new Date(),
    }

    // cache the result
    await this.redis.setex(cacheKey, 10, JSON.stringify(volumeUsageOverview))

    return volumeUsageOverview
  }

  async getUsageOverview(organizationId: string): Promise<OrganizationUsageOverviewDto> {
    const organization = await this.organizationRepository.findOne({ where: { id: organizationId } })
    if (!organization) {
      throw new NotFoundException(`Organization with ID ${organizationId} not found`)
    }

    const sandboxUsageOverview = await this.getSandboxUsageOverview(organizationId, organization)
    const snapshotUsageOverview = await this.getSnapshotUsageOverview(organizationId, organization)
    const volumeUsageOverview = await this.getVolumeUsageOverview(organizationId, organization)

    return {
      ...sandboxUsageOverview,
      ...snapshotUsageOverview,
      ...volumeUsageOverview,
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
