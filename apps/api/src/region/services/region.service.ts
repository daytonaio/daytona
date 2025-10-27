/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, NotFoundException, ConflictException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { FindOptionsWhere, In, Repository } from 'typeorm'
import { GLOBAL_REGION_ORGANIZATION_ID } from '../constants/region.constants'
import { CreateRegionDto } from '../dto/create-region.dto'
import { Region } from '../entities/region.entity'
import { Organization } from '../../organization/entities/organization.entity'
import { OrganizationService } from '../../organization/services/organization.service'

@Injectable()
export class RegionService {
  private readonly logger = new Logger(RegionService.name)

  constructor(
    @InjectRepository(Region)
    private readonly regionRepository: Repository<Region>,
    private readonly organizationService: OrganizationService,
  ) {}

  async create(createRegionDto: CreateRegionDto, organization: Organization): Promise<Region> {
    try {
      const region = new Region(createRegionDto.name, organization.id)
      return await this.regionRepository.save(region)
    } catch (error) {
      if (error.code === '23505') {
        throw new ConflictException(`Region with name ${createRegionDto.name} already exists in this organization`)
      }
      throw error
    }
  }

  async findOne(id: string, organizationId?: string): Promise<Region> {
    const region = await this.regionRepository.findOne({
      where: { id },
      ...(organizationId ? { organizationId } : {}),
    })

    if (!region) {
      throw new NotFoundException('Region not found')
    }

    return region
  }

  async findOneByNameAndOrganization(name: string, organizationId: string): Promise<Region> {
    const region = await this.regionRepository.findOne({
      where: {
        name,
        organizationId,
      },
    })

    if (!region) {
      throw new NotFoundException('Region not found')
    }

    return region
  }

  async getOrganizationId(regionId: string): Promise<string> {
    const region = await this.regionRepository.findOne({
      where: {
        id: regionId,
      },
      select: ['organizationId'],
      loadEagerRelations: false,
    })

    if (!region || !region.organizationId) {
      throw new NotFoundException('Region not found')
    }

    return region.organizationId
  }

  async findAll(organizationIdOrEntity?: string | Organization, includeGlobal = false): Promise<Region[]> {
    const organizationId =
      organizationIdOrEntity === undefined
        ? undefined
        : typeof organizationIdOrEntity === 'string'
          ? organizationIdOrEntity
          : organizationIdOrEntity.id

    if (organizationId && includeGlobal) {
      const organization =
        typeof organizationIdOrEntity === 'string'
          ? await this.organizationService.findOne(organizationIdOrEntity)
          : organizationIdOrEntity

      if (!organization) {
        throw new NotFoundException('Organization not found')
      }

      if (organization.blockSharedInfrastructure) {
        includeGlobal = false
      }
    }

    const where: FindOptionsWhere<Region> = {}

    if (organizationId && includeGlobal) {
      where.organizationId = In([organizationId, GLOBAL_REGION_ORGANIZATION_ID])
    } else if (organizationId) {
      where.organizationId = organizationId
    }

    return this.regionRepository.find({
      where,
      order: {
        name: 'ASC',
      },
    })
  }

  async delete(id: string): Promise<void> {
    const region = await this.findOne(id)
    await this.regionRepository.remove(region)
  }
}
