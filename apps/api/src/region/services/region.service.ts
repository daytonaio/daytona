/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */
import {
  Injectable,
  Logger,
  NotFoundException,
  ConflictException,
  BadRequestException,
  HttpException,
  HttpStatus,
} from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { DataSource, In, IsNull, Like, Repository } from 'typeorm'
import { REGION_NAME_REGEX } from '../constants/region-name-regex.constant'
import { CreateRegionInternalDto } from '../dto/create-region-internal.dto'
import { Region } from '../entities/region.entity'
import { Runner } from '../../sandbox/entities/runner.entity'
import { RegionType } from '../enums/region-type.enum'
import { CreateRegionResponseDto } from '../dto/create-region.dto'
import { generateApiKeyHash, generateApiKeyValue, generateRandomString } from '../../common/utils/api-key'
import { RegionDto } from '../dto/region.dto'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { RegionEvents } from '../constants/region-events.constant'
import { RegionCreatedEvent } from '../events/region-created.event'
import { RegionDeletedEvent } from '../events/region-deleted.event'
import { SnapshotManagerCredentialsDto } from '../dto/snapshot-manager-credentials.dto'
import {
  RegionSnapshotManagerCredsRegeneratedEvent,
  RegionSnapshotManagerUpdatedEvent,
} from '../events/region-snapshot-manager-creds.event'
import { UpdateRegionDto } from '../dto/update-region.dto'
import { Snapshot } from '../../sandbox/entities/snapshot.entity'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { toolboxProxyUrlCacheKey } from '../../sandbox/utils/sandbox-lookup-cache.util'

@Injectable()
export class RegionService {
  private readonly logger = new Logger(RegionService.name)

  constructor(
    @InjectRepository(Region)
    private readonly regionRepository: Repository<Region>,
    @InjectRepository(Runner)
    private readonly runnerRepository: Repository<Runner>,
    private readonly dataSource: DataSource,
    private readonly eventEmitter: EventEmitter2,
    @InjectRepository(Snapshot)
    private readonly snapshotRepository: Repository<Snapshot>,
    @InjectRedis() private readonly redis: Redis,
  ) {}

  /**
   * @param createRegionDto - The region details.
   * @param organizationId - The ID of the organization, or null for regions not associated with an organization.
   * @throws {BadRequestException} If the region name is invalid.
   * @throws {ConflictException} If the region with the same ID already exists or region with the same name already exists in the organization.
   */
  async create(
    createRegionDto: CreateRegionInternalDto,
    organizationId: string | null,
  ): Promise<CreateRegionResponseDto> {
    if (!REGION_NAME_REGEX.test(createRegionDto.name)) {
      throw new BadRequestException('Region name must contain only letters, numbers, underscores, periods, and hyphens')
    }
    if (createRegionDto.name.length < 2 || createRegionDto.name.length > 255) {
      throw new BadRequestException('Region name must be between 3 and 255 characters')
    }

    if (createRegionDto.id) {
      const existingRegion = await this.findOne(createRegionDto.id)
      if (existingRegion) {
        throw new ConflictException(`Region with id ${createRegionDto.id} already exists`)
      }
    }

    try {
      const proxyApiKey = createRegionDto.proxyUrl ? generateApiKeyValue() : undefined
      const sshGatewayApiKey = createRegionDto.sshGatewayUrl ? generateApiKeyValue() : undefined

      const snapshotManagerUsername = createRegionDto.snapshotManagerUrl ? 'daytona' : undefined
      const snapshotManagerPassword = createRegionDto.snapshotManagerUrl ? generateRandomString(16) : undefined

      const region = new Region({
        name: createRegionDto.name,
        enforceQuotas: createRegionDto.enforceQuotas,
        regionType: createRegionDto.regionType,
        id: createRegionDto.id,
        organizationId,
        proxyUrl: createRegionDto.proxyUrl,
        sshGatewayUrl: createRegionDto.sshGatewayUrl,
        proxyApiKeyHash: proxyApiKey ? generateApiKeyHash(proxyApiKey) : null,
        sshGatewayApiKeyHash: sshGatewayApiKey ? generateApiKeyHash(sshGatewayApiKey) : null,
        snapshotManagerUrl: createRegionDto.snapshotManagerUrl,
      })

      await this.dataSource.transaction(async (em) => {
        await em.save(region)
        await this.eventEmitter.emitAsync(
          RegionEvents.CREATED,
          new RegionCreatedEvent(em, region, organizationId, snapshotManagerUsername, snapshotManagerPassword),
        )
      })

      return new CreateRegionResponseDto({
        id: region.id,
        proxyApiKey,
        sshGatewayApiKey,
        snapshotManagerUsername,
        snapshotManagerPassword,
      })
    } catch (error) {
      if (error.code === '23505') {
        throw new ConflictException(`Region with name ${createRegionDto.name} already exists`)
      }
      throw error
    }
  }

  /**
   * @param id - The ID of the region.
   * @returns The region if found, or null otherwise.
   */
  async findOne(id: string, cache = false): Promise<Region | null> {
    return await this.regionRepository.findOne({
      where: {
        id,
      },
      cache: cache
        ? {
            id: `region:${id}`,
            milliseconds: 30000,
          }
        : undefined,
    })
  }

  /**
   * @param name - The name of the region.
   * @param organizationId - The organization ID, or null for regions not associated with an organization.
   * @returns The region if found, or null otherwise.
   */
  async findOneByName(name: string, organizationId: string | null): Promise<Region | null> {
    return await this.regionRepository.findOne({
      where: [{ name, organizationId: organizationId ?? IsNull() }],
    })
  }

  /**
   * @param proxyApiKey - The proxy API key.
   * @returns The region if found, or null otherwise.
   */
  async findOneByProxyApiKey(proxyApiKey: string): Promise<Region | null> {
    return await this.regionRepository.findOne({
      where: { proxyApiKeyHash: generateApiKeyHash(proxyApiKey) },
    })
  }

  /**
   * @param sshGatewayApiKey - The SSH gateway API key.
   * @returns The region if found, or null otherwise.
   */
  async findOneBySshGatewayApiKey(sshGatewayApiKey: string): Promise<Region | null> {
    return await this.regionRepository.findOne({
      where: { sshGatewayApiKeyHash: generateApiKeyHash(sshGatewayApiKey) },
    })
  }

  /**
   * @param regionId - The ID of the region.
   * @returns The organization ID or null for for regions not associated with an organization if the region is found, or undefined if the region is not found.
   */
  async getOrganizationId(regionId: string): Promise<string | null | undefined> {
    const region = await this.regionRepository.findOne({
      where: {
        id: regionId,
      },
      select: ['organizationId'],
      loadEagerRelations: false,
    })

    if (!region) {
      return undefined
    }

    return region.organizationId ?? null
  }

  /**
   * @param organizationId - The organization ID of the regions to find.
   * @param regionType - If provided, only return regions of the specified type.
   * @returns The regions found ordered by name ascending.
   */
  async findAllByOrganization(organizationId: string, regionType?: RegionType): Promise<Region[]> {
    return this.regionRepository.find({
      where: {
        organizationId,
        ...(regionType ? { regionType } : {}),
      },
      order: {
        name: 'ASC',
      },
    })
  }

  /**
   * @param regionTypes - Types of regions to find.
   * @returns The regions found ordered by name ascending.
   */
  async findAllByRegionTypes(regionTypes: RegionType[]): Promise<RegionDto[]> {
    if (regionTypes.length === 0) {
      return []
    }

    const regions = await this.regionRepository.find({
      where: {
        regionType: In(regionTypes),
      },
      order: {
        name: 'ASC',
      },
    })

    return regions.map(RegionDto.fromRegion)
  }

  /**
   * @param ids - The IDs of the regions to find.
   * @returns The regions found.
   */
  async findByIds(ids: string[]): Promise<Region[]> {
    if (ids.length === 0) {
      return []
    }

    return this.regionRepository.find({
      where: {
        id: In(ids),
      },
    })
  }

  /**
   * @param id - The ID of the region to delete.
   * @throws {NotFoundException} If the region is not found.
   */
  async delete(id: string): Promise<void> {
    const region = await this.findOne(id)

    if (!region) {
      throw new NotFoundException('Region not found')
    }

    const runnerCount = await this.runnerRepository.count({
      where: {
        region: id,
      },
    })

    if (runnerCount > 0) {
      throw new HttpException(
        'Cannot delete region which has runners associated with it',
        HttpStatus.PRECONDITION_REQUIRED,
      )
    }

    await this.dataSource.transaction(async (em) => {
      await this.eventEmitter.emitAsync(RegionEvents.DELETED, new RegionDeletedEvent(em, region))
      await em.remove(region)
    })

    this.redis.del(toolboxProxyUrlCacheKey(id)).catch((err) => {
      this.logger.warn(`Failed to invalidate toolbox proxy URL cache for region ${id}: ${err.message}`)
    })
  }

  async update(regionId: string, updateRegion: UpdateRegionDto): Promise<void> {
    const region = await this.findOne(regionId)

    if (!region) {
      throw new NotFoundException('Region not found')
    }

    await this.dataSource.transaction(async (em) => {
      if (updateRegion.proxyUrl !== undefined) {
        region.proxyUrl = updateRegion.proxyUrl ?? null
        region.toolboxProxyUrl = updateRegion.proxyUrl ?? null
      }

      if (updateRegion.sshGatewayUrl !== undefined) {
        region.sshGatewayUrl = updateRegion.sshGatewayUrl ?? null
      }

      if (updateRegion.snapshotManagerUrl !== undefined) {
        if (region.snapshotManagerUrl) {
          // If snapshots already exist, prevent changing the snapshot manager URL
          const exists = await this.snapshotRepository.exists({
            where: {
              ref: Like(`${region.snapshotManagerUrl.replace(/^https?:\/\//, '')}%`),
            },
          })
          if (exists) {
            throw new BadRequestException(
              'Cannot change snapshot manager URL for region with existing snapshots. Please delete existing snapshots first.',
            )
          }
        }

        const prevSnapshotManagerUrl = region.snapshotManagerUrl
        region.snapshotManagerUrl = updateRegion.snapshotManagerUrl ?? null

        let newUsername: string | undefined = undefined
        let newPassword: string | undefined = undefined

        // If the region did not have a snapshot manager, create new credentials
        if (!prevSnapshotManagerUrl) {
          newUsername = 'daytona'
          newPassword = generateRandomString(16)
        }

        await this.eventEmitter.emitAsync(
          RegionEvents.SNAPSHOT_MANAGER_UPDATED,
          new RegionSnapshotManagerUpdatedEvent(
            region,
            region.organizationId,
            region.snapshotManagerUrl,
            prevSnapshotManagerUrl,
            newUsername,
            newPassword,
            em,
          ),
        )
      }

      await em.save(region)
    })

    if (updateRegion.proxyUrl !== undefined) {
      this.redis.del(toolboxProxyUrlCacheKey(regionId)).catch((err) => {
        this.logger.warn(`Failed to invalidate toolbox proxy URL cache for region ${regionId}: ${err.message}`)
      })
    }
  }

  /**
   * @param regionId - The ID of the region.
   * @throws {NotFoundException} If the region is not found.
   * @throws {BadRequestException} If the region does not have a proxy URL configured.
   * @returns The newly generated proxy API key.
   */
  async regenerateProxyApiKey(regionId: string): Promise<string> {
    const region = await this.findOne(regionId)

    if (!region) {
      throw new NotFoundException('Region not found')
    }

    if (!region.proxyUrl) {
      throw new BadRequestException('Region does not have a proxy URL configured')
    }

    const newApiKey = generateApiKeyValue()
    region.proxyApiKeyHash = generateApiKeyHash(newApiKey)

    await this.regionRepository.save(region)

    return newApiKey
  }

  /**
   * @param regionId - The ID of the region.
   * @throws {NotFoundException} If the region is not found.
   * @throws {BadRequestException} If the region does not have an SSH gateway URL configured.
   * @returns The newly generated SSH gateway API key.
   */
  async regenerateSshGatewayApiKey(regionId: string): Promise<string> {
    const region = await this.findOne(regionId)

    if (!region) {
      throw new NotFoundException('Region not found')
    }

    if (!region.sshGatewayUrl) {
      throw new BadRequestException('Region does not have an SSH gateway URL configured')
    }

    const newApiKey = generateApiKeyValue()
    region.sshGatewayApiKeyHash = generateApiKeyHash(newApiKey)

    await this.regionRepository.save(region)

    return newApiKey
  }

  /**
   * @param regionId - The ID of the region.
   * @throws {NotFoundException} If the region is not found.
   * @throws {BadRequestException} If the region does not have a snapshot manager URL configured.
   * @returns The newly generated snapshot manager credentials.
   */
  async regenerateSnapshotManagerCredentials(regionId: string): Promise<SnapshotManagerCredentialsDto> {
    const region = await this.findOne(regionId)

    if (!region) {
      throw new NotFoundException('Region not found')
    }

    if (!region.snapshotManagerUrl) {
      throw new BadRequestException('Region does not have a snapshot manager URL configured')
    }

    const newUsername = 'daytona'
    const newPassword = generateRandomString(16)

    await this.eventEmitter.emitAsync(
      RegionEvents.SNAPSHOT_MANAGER_CREDENTIALS_REGENERATED,
      new RegionSnapshotManagerCredsRegeneratedEvent(regionId, region.snapshotManagerUrl, newUsername, newPassword),
    )

    return new SnapshotManagerCredentialsDto({
      username: newUsername,
      password: newPassword,
    })
  }
}
