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
  OnApplicationShutdown,
  ConflictException,
  HttpException,
  HttpStatus,
} from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { EntityManager, In, IsNull, Not, Repository } from 'typeorm'
import { CreateOrganizationInternalDto } from '../dto/create-organization.internal.dto'
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
import { UserEmailVerifiedEvent } from '../../user/events/user-email-verified.event'
import { Cron, CronExpression } from '@nestjs/schedule'
import { RedisLockProvider } from '../../sandbox/common/redis-lock.provider'
import { OrganizationSuspendedSandboxStoppedEvent } from '../events/organization-suspended-sandbox-stopped.event'
import { SandboxDesiredState } from '../../sandbox/enums/sandbox-desired-state.enum'
import { SystemRole } from '../../user/enums/system-role.enum'
import { SnapshotState } from '../../sandbox/enums/snapshot-state.enum'
import { OrganizationSuspendedSnapshotDeactivatedEvent } from '../events/organization-suspended-snapshot-deactivated.event'
import { TrackJobExecution } from '../../common/decorators/track-job-execution.decorator'
import { TrackableJobExecutions } from '../../common/interfaces/trackable-job-executions'
import { setTimeout } from 'timers/promises'
import { TypedConfigService } from '../../config/typed-config.service'
import { LogExecution } from '../../common/decorators/log-execution.decorator'
import { WithInstrumentation } from '../../common/decorators/otel.decorator'
import { RegionQuota } from '../entities/region-quota.entity'
import { UpdateOrganizationRegionQuotaDto } from '../dto/update-organization-region-quota.dto'
import { RegionService } from '../../region/services/region.service'
import {
  PAYMENT_METHOD_REQUIRED_SUSPENSION_REASON,
  VERIFY_EMAIL_SUSPENSION_REASON,
} from '../constants/suspension-reasons.constant'
import { Region } from '../../region/entities/region.entity'
import { RegionQuotaDto } from '../dto/region-quota.dto'
import { RegionType } from '../../region/enums/region-type.enum'
import { RegionDto } from '../../region/dto/region.dto'
import { OrganizationDeletedEvent } from '../events/organization-deleted.event'
import { OrganizationAssertDeletableEvent } from '../events/organization-assert-deletable.event'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { getOrganizationCacheKey } from '../constants/organization-cache-keys.constant'
import { OrganizationUserService } from './organization-user.service'

@Injectable()
export class OrganizationService implements OnModuleInit, TrackableJobExecutions, OnApplicationShutdown {
  activeJobs = new Set<string>()
  private readonly logger = new Logger(OrganizationService.name)
  private defaultOrganizationQuota: CreateOrganizationQuotaDto
  private defaultSandboxLimitedNetworkEgress: boolean

  constructor(
    @InjectRepository(Organization)
    private readonly organizationRepository: Repository<Organization>,
    @InjectRepository(Sandbox)
    private readonly sandboxRepository: Repository<Sandbox>,
    @InjectRepository(Snapshot)
    private readonly snapshotRepository: Repository<Snapshot>,
    private readonly eventEmitter: EventEmitter2,
    private readonly configService: TypedConfigService,
    private readonly redisLockProvider: RedisLockProvider,
    @InjectRepository(RegionQuota)
    private readonly regionQuotaRepository: Repository<RegionQuota>,
    @InjectRepository(Region)
    private readonly regionRepository: Repository<Region>,
    private readonly regionService: RegionService,
    @InjectRedis() private readonly redis: Redis,
    private readonly organizationUserService: OrganizationUserService,
  ) {
    this.defaultOrganizationQuota = this.configService.getOrThrow('defaultOrganizationQuota')
    this.defaultSandboxLimitedNetworkEgress = this.configService.getOrThrow(
      'organizationSandboxDefaultLimitedNetworkEgress',
    )
  }

  async onApplicationShutdown() {
    //  wait for all active jobs to finish
    while (this.activeJobs.size > 0) {
      this.logger.log(`Waiting for ${this.activeJobs.size} active jobs to finish`)
      await setTimeout(1000)
    }
  }

  async onModuleInit(): Promise<void> {
    await this.stopSuspendedOrganizationSandboxes()
  }

  async create(
    createOrganizationDto: CreateOrganizationInternalDto,
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
        deletedAt: IsNull(),
      },
    })
  }

  async findOne(organizationId: string): Promise<Organization | null> {
    return this.organizationRepository.findOne({
      where: { id: organizationId, deletedAt: IsNull() },
    })
  }

  async findBySandboxId(sandboxId: string): Promise<Organization | null> {
    const sandbox = await this.sandboxRepository.findOne({
      where: { id: sandboxId },
    })

    if (!sandbox) {
      return null
    }

    return this.organizationRepository.findOne({
      where: {
        id: sandbox.organizationId,
        deletedAt: IsNull(),
      },
    })
  }

  async findPersonal(userId: string): Promise<Organization> {
    return this.findPersonalWithEntityManager(this.organizationRepository.manager, userId)
  }

  async delete(params: {
    organizationId: string
    organization?: Organization
    allowPersonalDeletion?: boolean
    entityManager?: EntityManager
  }): Promise<void> {
    const { organizationId, allowPersonalDeletion = false, entityManager } = params

    if (!entityManager) {
      return this.organizationRepository.manager.transaction((em) => this.delete({ ...params, entityManager: em }))
    }

    let organization = params.organization
    if (!organization) {
      organization = await entityManager.findOne(Organization, {
        where: { id: organizationId, deletedAt: IsNull() },
      })
    }

    if (!organization) {
      throw new NotFoundException(`Organization with ID ${organizationId} not found`)
    }

    if (organization.personal && !allowPersonalDeletion) {
      throw new ForbiddenException('Cannot delete personal organization')
    }

    await this.assertOrganizationCanBeDeleted(organization)

    // Soft delete the organization
    organization.deletedAt = new Date()
    await entityManager.save(organization)

    // Emit event for side effects handled by other services
    await this.eventEmitter.emitAsync(
      OrganizationEvents.DELETED,
      new OrganizationDeletedEvent(entityManager, organization.id),
    )

    // Invalidate organization cache
    try {
      await this.redis.del(getOrganizationCacheKey(organizationId))
    } catch (error) {
      this.logger.error(`Failed to invalidate organization cache for ${organizationId}:`, error)
    }
  }

  async updateQuota(organizationId: string, updateDto: UpdateOrganizationQuotaDto): Promise<void> {
    const organization = await this.organizationRepository.findOne({ where: { id: organizationId } })
    if (!organization) {
      throw new NotFoundException(`Organization with ID ${organizationId} not found`)
    }

    organization.maxCpuPerSandbox = updateDto.maxCpuPerSandbox ?? organization.maxCpuPerSandbox
    organization.maxMemoryPerSandbox = updateDto.maxMemoryPerSandbox ?? organization.maxMemoryPerSandbox
    organization.maxDiskPerSandbox = updateDto.maxDiskPerSandbox ?? organization.maxDiskPerSandbox
    organization.maxSnapshotSize = updateDto.maxSnapshotSize ?? organization.maxSnapshotSize
    organization.volumeQuota = updateDto.volumeQuota ?? organization.volumeQuota
    organization.snapshotQuota = updateDto.snapshotQuota ?? organization.snapshotQuota
    organization.authenticatedRateLimit = updateDto.authenticatedRateLimit ?? organization.authenticatedRateLimit
    organization.sandboxCreateRateLimit = updateDto.sandboxCreateRateLimit ?? organization.sandboxCreateRateLimit
    organization.sandboxLifecycleRateLimit =
      updateDto.sandboxLifecycleRateLimit ?? organization.sandboxLifecycleRateLimit

    await this.organizationRepository.save(organization)
  }

  async updateRegionQuota(
    organizationId: string,
    regionId: string,
    updateDto: UpdateOrganizationRegionQuotaDto,
  ): Promise<void> {
    const regionQuota = await this.regionQuotaRepository.findOne({ where: { organizationId, regionId } })
    if (!regionQuota) {
      throw new NotFoundException('Region not found')
    }

    regionQuota.totalCpuQuota = updateDto.totalCpuQuota ?? regionQuota.totalCpuQuota
    regionQuota.totalMemoryQuota = updateDto.totalMemoryQuota ?? regionQuota.totalMemoryQuota
    regionQuota.totalDiskQuota = updateDto.totalDiskQuota ?? regionQuota.totalDiskQuota

    await this.regionQuotaRepository.save(regionQuota)
  }

  async getRegionQuotas(organizationId: string): Promise<RegionQuotaDto[]> {
    const regionQuotas = await this.regionQuotaRepository.find({ where: { organizationId } })
    return regionQuotas.map((regionQuota) => new RegionQuotaDto(regionQuota))
  }

  async getRegionQuota(organizationId: string, regionId: string): Promise<RegionQuotaDto | null> {
    const regionQuota = await this.regionQuotaRepository.findOne({ where: { organizationId, regionId } })
    if (!regionQuota) {
      return null
    }
    return new RegionQuotaDto(regionQuota)
  }

  async getRegionQuotaBySandboxId(sandboxId: string): Promise<RegionQuotaDto | null> {
    const sandbox = await this.sandboxRepository.findOne({
      where: { id: sandboxId },
    })
    if (!sandbox) {
      return null
    }
    return this.getRegionQuota(sandbox.organizationId, sandbox.region)
  }

  /**
   * Lists all available regions for the organization.
   *
   * A region is available for the organization if either:
   * - It is directly associated with the organization, or
   * - It is not associated with any organization, but the organization has quotas allocated for the region or quotas are not enforced for the region
   *
   * @param organizationId - The organization ID.
   * @returns The available regions
   */
  async listAvailableRegions(organizationId: string): Promise<RegionDto[]> {
    const regions = await this.regionRepository
      .createQueryBuilder('region')
      .where('region."regionType" = :customRegionType AND region."organizationId" = :organizationId', {
        customRegionType: RegionType.CUSTOM,
        organizationId,
      })
      .orWhere('region."regionType" IN (:...otherRegionTypes) AND region."enforceQuotas" = false', {
        otherRegionTypes: [RegionType.DEDICATED, RegionType.SHARED],
      })
      .orWhere(
        'region."regionType" IN (:...otherRegionTypes) AND region."enforceQuotas" = true AND EXISTS (SELECT 1 FROM region_quota rq WHERE rq."regionId" = region."id" AND rq."organizationId" = :organizationId)',
        {
          otherRegionTypes: [RegionType.DEDICATED, RegionType.SHARED],
          organizationId,
        },
      )
      .orderBy(
        `CASE region."regionType" 
          WHEN '${RegionType.CUSTOM}' THEN 1 
          WHEN '${RegionType.DEDICATED}' THEN 2 
          WHEN '${RegionType.SHARED}' THEN 3 
          ELSE 4 
        END`,
      )
      .getMany()

    return regions.map(RegionDto.fromRegion)
  }

  async suspend(
    organizationId: string,
    suspensionReason?: string,
    suspendedUntil?: Date,
    suspensionCleanupGracePeriodHours?: number,
  ): Promise<void> {
    const organization = await this.organizationRepository.findOne({ where: { id: organizationId } })
    if (!organization) {
      throw new NotFoundException(`Organization with ID ${organizationId} not found`)
    }

    organization.suspended = true
    organization.suspensionReason = suspensionReason || null
    organization.suspendedUntil = suspendedUntil || null
    organization.suspendedAt = new Date()
    if (suspensionCleanupGracePeriodHours) {
      organization.suspensionCleanupGracePeriodHours = suspensionCleanupGracePeriodHours
    }

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

  async updateSandboxDefaultLimitedNetworkEgress(
    organizationId: string,
    sandboxDefaultLimitedNetworkEgress: boolean,
  ): Promise<void> {
    const organization = await this.organizationRepository.findOne({ where: { id: organizationId } })
    if (!organization) {
      throw new NotFoundException(`Organization with ID ${organizationId} not found`)
    }
    organization.sandboxLimitedNetworkEgress = sandboxDefaultLimitedNetworkEgress

    await this.organizationRepository.save(organization)
  }

  /**
   * @param organizationId - The ID of the organization.
   * @param defaultRegionId - The ID of the region to set as the default region.
   * @throws {NotFoundException} If the organization is not found.
   * @throws {ConflictException} If the organization already has a default region set.
   */
  async setDefaultRegion(organizationId: string, defaultRegionId: string): Promise<void> {
    const organization = await this.organizationRepository.findOne({ where: { id: organizationId } })
    if (!organization) {
      throw new NotFoundException(`Organization with ID ${organizationId} not found`)
    }

    if (organization.defaultRegionId) {
      throw new ConflictException('Organization already has a default region set')
    }

    const defaultRegion = await this.validateOrganizationDefaultRegion(defaultRegionId)
    organization.defaultRegionId = defaultRegionId

    if (defaultRegion.enforceQuotas) {
      const regionQuota = new RegionQuota(
        organization.id,
        defaultRegionId,
        this.defaultOrganizationQuota.totalCpuQuota,
        this.defaultOrganizationQuota.totalMemoryQuota,
        this.defaultOrganizationQuota.totalDiskQuota,
      )
      if (organization.regionQuotas) {
        organization.regionQuotas = [...organization.regionQuotas, regionQuota]
      } else {
        organization.regionQuotas = [regionQuota]
      }
    }

    await this.organizationRepository.save(organization)
  }

  private async createWithEntityManager(
    entityManager: EntityManager,
    createOrganizationDto: CreateOrganizationInternalDto,
    createdBy: string,
    creatorEmailVerified: boolean,
    personal = false,
    quota: CreateOrganizationQuotaDto = this.defaultOrganizationQuota,
    sandboxLimitedNetworkEgress: boolean = this.defaultSandboxLimitedNetworkEgress,
  ): Promise<Organization> {
    if (personal) {
      const count = await entityManager.count(Organization, {
        where: { createdBy, deletedAt: IsNull(), personal: true },
      })
      if (count > 0) {
        throw new ForbiddenException('Personal organization already exists')
      }
    }

    // set some limit to the number of created organizations
    const createdCount = await entityManager.count(Organization, {
      where: { createdBy, deletedAt: IsNull() },
    })
    if (createdCount >= 10) {
      throw new ForbiddenException('You have reached the maximum number of created organizations')
    }

    let organization = new Organization(createOrganizationDto.defaultRegionId)

    organization.name = createOrganizationDto.name
    organization.createdBy = createdBy
    organization.personal = personal

    organization.maxCpuPerSandbox = quota.maxCpuPerSandbox
    organization.maxMemoryPerSandbox = quota.maxMemoryPerSandbox
    organization.maxDiskPerSandbox = quota.maxDiskPerSandbox
    organization.snapshotQuota = quota.snapshotQuota
    organization.maxSnapshotSize = quota.maxSnapshotSize
    organization.volumeQuota = quota.volumeQuota

    if (!creatorEmailVerified && !this.configService.get('skipUserEmailVerification')) {
      organization.suspended = true
      organization.suspendedAt = new Date()
      organization.suspensionReason = VERIFY_EMAIL_SUSPENSION_REASON
    } else if (this.configService.get('billingApiUrl') && !personal) {
      organization.suspended = true
      organization.suspendedAt = new Date()
      organization.suspensionReason = PAYMENT_METHOD_REQUIRED_SUSPENSION_REASON
    }

    organization.sandboxLimitedNetworkEgress = sandboxLimitedNetworkEgress

    const owner = new OrganizationUser()
    owner.userId = createdBy
    owner.role = OrganizationMemberRole.OWNER

    organization.users = [owner]

    if (createOrganizationDto.defaultRegionId) {
      const defaultRegion = await this.validateOrganizationDefaultRegion(createOrganizationDto.defaultRegionId)

      if (defaultRegion.enforceQuotas) {
        const regionQuota = new RegionQuota(
          organization.id,
          createOrganizationDto.defaultRegionId,
          quota.totalCpuQuota,
          quota.totalMemoryQuota,
          quota.totalDiskQuota,
        )
        organization.regionQuotas = [regionQuota]
      }
    }

    await entityManager.transaction(async (em) => {
      organization = await em.save(organization)
      await this.eventEmitter.emitAsync(OrganizationEvents.CREATED, organization)
    })

    return organization
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
      where: { createdBy: userId, deletedAt: IsNull(), personal: true },
    })

    if (!organization) {
      throw new NotFoundException(`Personal organization for user ${userId} not found`)
    }

    return organization
  }

  /**
   * @throws NotFoundException - If the region is not found or not available to the organization
   */
  async validateOrganizationDefaultRegion(defaultRegionId: string): Promise<Region> {
    const region = await this.regionService.findOne(defaultRegionId)
    if (!region || region.regionType !== RegionType.SHARED) {
      throw new NotFoundException('Region not found')
    }

    return region
  }

  @Cron(CronExpression.EVERY_MINUTE, { name: 'stop-suspended-organization-sandboxes' })
  @TrackJobExecution()
  @LogExecution('stop-suspended-organization-sandboxes')
  @WithInstrumentation()
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
      .andWhere(`"suspendedAt" < NOW() - INTERVAL '1 hour' * "suspensionCleanupGracePeriodHours"`)
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
  @TrackJobExecution()
  @LogExecution('deactivate-suspended-organization-snapshots')
  @WithInstrumentation()
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
      .andWhere(`"suspendedAt" < NOW() - INTERVAL '1 hour' * "suspensionCleanupGracePeriodHours"`)
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
  @TrackJobExecution()
  async handleUserCreatedEvent(payload: UserCreatedEvent): Promise<Organization> {
    return this.createWithEntityManager(
      payload.entityManager,
      {
        name: 'Personal',
        defaultRegionId: payload.personalOrganizationDefaultRegionId,
      },
      payload.user.id,
      payload.user.role === SystemRole.ADMIN ? true : payload.user.emailVerified,
      true,
      payload.personalOrganizationQuota,
      payload.user.role === SystemRole.ADMIN ? false : undefined,
    )
  }

  @OnAsyncEvent({
    event: UserEvents.EMAIL_VERIFIED,
  })
  @TrackJobExecution()
  async handleUserEmailVerifiedEvent(payload: UserEmailVerifiedEvent): Promise<void> {
    await this.unsuspendPersonalWithEntityManager(payload.entityManager, payload.userId)
  }

  @OnAsyncEvent({
    event: UserEvents.DELETED,
  })
  @TrackJobExecution()
  async handleUserDeletedEvent(payload: UserDeletedEvent): Promise<void> {
    const count = await this.organizationUserService.countByUserId(payload.userId)
    if (count > 1) {
      throw new HttpException(
        'You must leave or delete any non-personal organizations you are a member of before deleting your account',
        HttpStatus.PRECONDITION_REQUIRED,
      )
    }

    const organization = await this.findPersonalWithEntityManager(payload.entityManager, payload.userId)

    await this.delete({
      organizationId: organization.id,
      organization,
      allowPersonalDeletion: true,
      entityManager: payload.entityManager,
    })
  }

  @OnAsyncEvent({
    event: OrganizationEvents.DELETED,
  })
  async handleOrganizationDeletedEvent(payload: OrganizationDeletedEvent): Promise<void> {
    const { entityManager, organizationId } = payload

    await entityManager.delete(RegionQuota, { organizationId })
  }

  assertOrganizationIsNotSuspended(organization: Organization): void {
    if (!organization.isSuspended) {
      return
    }

    if (organization.suspensionReason) {
      throw new ForbiddenException(`Organization is suspended: ${organization.suspensionReason}`)
    } else {
      throw new ForbiddenException('Organization is suspended')
    }
  }

  /**
   * Asserts that the organization can be deleted.
   *
   * @param organization - The organization to check
   * @throws HttpException - If the organization cannot be deleted because of a suspension or resources that must be cleaned up first, with concatenated error messages and HTTP status 428
   */
  async assertOrganizationCanBeDeleted(organization: Organization): Promise<void> {
    const errors: string[] = []

    // assert organization is not suspended due to a reason that prohibits deletion
    if (
      organization.isSuspended &&
      organization.suspensionReason !== VERIFY_EMAIL_SUSPENSION_REASON &&
      organization.suspensionReason !== PAYMENT_METHOD_REQUIRED_SUSPENSION_REASON
    ) {
      errors.push('Organization is currently suspended. Please contact support to resolve the suspension first.')
    }

    // assert organization has no resources that must be cleaned up first
    const event = new OrganizationAssertDeletableEvent(organization.id)

    const results = await Promise.allSettled([
      this.eventEmitter.emitAsync(OrganizationEvents.ASSERT_NO_USERS, event),
      this.eventEmitter.emitAsync(OrganizationEvents.ASSERT_NO_SANDBOXES, event),
      this.eventEmitter.emitAsync(OrganizationEvents.ASSERT_NO_SNAPSHOTS, event),
      this.eventEmitter.emitAsync(OrganizationEvents.ASSERT_NO_VOLUMES, event),
      this.eventEmitter.emitAsync(OrganizationEvents.ASSERT_NO_RUNNERS, event),
    ])

    for (const result of results) {
      if (result.status === 'rejected') {
        const reason = result.reason
        if (reason instanceof Error) {
          errors.push(reason.message)
        } else {
          errors.push(String(reason))
        }
      }
    }

    if (errors.length > 0) {
      throw new HttpException(errors.join('; '), HttpStatus.PRECONDITION_REQUIRED)
    }
  }
}
