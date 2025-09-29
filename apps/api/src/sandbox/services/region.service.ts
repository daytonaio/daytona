/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, NotFoundException, ConflictException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { Region } from '../entities/region.entity'
import { CreateRegionDto } from '../dto/create-region.dto'
import { Organization } from '../../organization/entities/organization.entity'

@Injectable()
export class RegionService {
  private readonly logger = new Logger(RegionService.name)

  constructor(
    @InjectRepository(Region)
    private readonly regionRepository: Repository<Region>,
  ) {}

  async create(createRegionDto: CreateRegionDto, organization: Organization): Promise<Region> {
    const existingRegion = await this.regionRepository.findOne({
      where: {
        organizationId: organization.id,
        name: createRegionDto.name,
      },
    })

    if (existingRegion) {
      throw new ConflictException(`Region with name ${createRegionDto.name} already exists in this organization`)
    }

    const region = new Region(createRegionDto.name, organization.id)

    return await this.regionRepository.save(region)
  }

  async findOne(id: string): Promise<Region> {
    const region = await this.regionRepository.findOne({
      where: { id },
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

  async findAll(organizationId: string): Promise<Region[]> {
    return this.regionRepository.find({
      where: {
        organizationId,
      },
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
