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
import { IsNull, Repository } from 'typeorm'
import { Region, REGION_NAME_REGEX } from '../entities/region.entity'
import { Organization } from '../../organization/entities/organization.entity'
import { Runner } from '../../sandbox/entities/runner.entity'
import { CreateRegionInternalDto } from '../dto/create-region-internal.dto'

@Injectable()
export class RegionService {
  private readonly logger = new Logger(RegionService.name)

  constructor(
    @InjectRepository(Region)
    private readonly regionRepository: Repository<Region>,
    @InjectRepository(Runner)
    private readonly runnerRepository: Repository<Runner>,
  ) {}

  async create(createRegionDto: CreateRegionInternalDto, organization?: Organization): Promise<Region> {
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
      const region = new Region(createRegionDto.name, createRegionDto.id, organization?.id)
      return await this.regionRepository.save(region)
    } catch (error) {
      if (error.code === '23505') {
        throw new ConflictException(`Region with name ${createRegionDto.name} already exists`)
      }
      throw error
    }
  }

  async findOne(id: string, organizationId?: string): Promise<Region | null> {
    const region = await this.regionRepository.findOne({
      where: { id },
      ...(organizationId ? { organizationId } : {}),
    })

    if (!region) {
      return null
    }

    return region
  }

  async findOneByName(name: string, organizationId: string): Promise<Region | null> {
    const region = await this.regionRepository.findOne({
      where: [
        { name, organizationId },
        { name, organizationId: IsNull() },
      ],
    })

    if (!region) {
      return null
    }

    return region
  }

  async getOrganizationId(regionId: string): Promise<string | null> {
    const region = await this.regionRepository.findOne({
      where: {
        id: regionId,
      },
      select: ['organizationId'],
      loadEagerRelations: false,
    })

    if (!region || !region.organizationId) {
      return null
    }

    return region.organizationId
  }

  async findAll(organizationId: string | null): Promise<Region[]> {
    return this.regionRepository.find({
      where: {
        ...(organizationId ? { organizationId } : { organizationId: IsNull() }),
      },
      order: {
        name: 'ASC',
      },
    })
  }

  async delete(id: string): Promise<void> {
    const region = await this.findOne(id)
    if (!region) {
      throw new NotFoundException('Region not found')
    }

    const runners = await this.runnerRepository.find({
      where: {
        regionId: id,
      },
    })

    if (runners.length > 0) {
      throw new HttpException(
        'Cannot delete region which has runners associated with it',
        HttpStatus.PRECONDITION_REQUIRED,
      )
    }

    await this.regionRepository.remove(region)
  }
}
