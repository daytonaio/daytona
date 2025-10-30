/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, NotFoundException, ConflictException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { CreateRegionInternalDto } from '../dto/create-region.internal.dto'
import { Region } from '../entities/region.entity'

@Injectable()
export class RegionService {
  private readonly logger = new Logger(RegionService.name)

  constructor(
    @InjectRepository(Region)
    private readonly regionRepository: Repository<Region>,
  ) {}

  async create(createRegionDto: CreateRegionInternalDto, organizationId: string): Promise<Region> {
    try {
      const region = new Region(organizationId, createRegionDto.name, createRegionDto.enforceQuotas)
      return this.regionRepository.save(region)
    } catch (error) {
      if (error.code === '23505') {
        throw new ConflictException(`Region with name ${createRegionDto.name} already exists in this organization`)
      }
      throw error
    }
  }

  async findOneOrFail(id: string, organizationId?: string): Promise<Region> {
    const region = await this.regionRepository.findOne({
      where: { id },
      ...(organizationId ? { organizationId } : {}),
    })

    if (!region) {
      throw new NotFoundException('Region not found')
    }

    return region
  }

  async findOneByNameAndOrganization(name: string, organizationId: string): Promise<Region | null> {
    const region = await this.regionRepository.findOne({
      where: {
        name,
        organizationId,
      },
    })

    if (!region) {
      return null
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

  async delete(id: string): Promise<void> {
    const region = await this.findOneOrFail(id)
    await this.regionRepository.remove(region)
  }
}
