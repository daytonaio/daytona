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

  async create(organization: Organization, createRegionDto: CreateRegionDto): Promise<Region> {
    // Check if region with same name already exists for organization
    const existingRegion = await this.regionRepository.findOne({
      where: {
        organizationId: organization.id,
        name: createRegionDto.name,
      },
    })

    if (existingRegion) {
      throw new ConflictException(`Region with name ${createRegionDto.name} already exists in this organization`)
    }

    const region = new Region()
    region.code = Region.generateCode()
    region.name = createRegionDto.name
    region.organizationId = organization.id

    const savedRegion = await this.regionRepository.save(region)
    this.logger.debug(`Created region ${savedRegion.code} for organization ${organization.id}`)
    return savedRegion
  }

  async findOne(code: string): Promise<Region> {
    const region = await this.regionRepository.findOne({
      where: { code },
    })

    if (!region) {
      throw new NotFoundException(`Region with code ${code} not found`)
    }

    return region
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

  async delete(code: string): Promise<void> {
    const region = await this.findOne(code)

    await this.regionRepository.remove(region)
    this.logger.debug(`Deleted region ${code}`)
  }

  async findByNameAndOrganization(name: string, organizationId: string): Promise<Region | null> {
    return this.regionRepository.findOne({
      where: {
        name,
        organizationId,
      },
    })
  }
}
