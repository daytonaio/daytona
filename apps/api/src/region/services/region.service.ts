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
import { REGION_NAME_REGEX } from '../constants/region-name-regex.constant'
import { CreateRegionInternalDto } from '../dto/create-region-internal.dto'
import { Region } from '../entities/region.entity'
import { Runner } from '../../sandbox/entities/runner.entity'
import { RegionType } from '../enums/region-type.enum'

@Injectable()
export class RegionService {
  private readonly logger = new Logger(RegionService.name)

  constructor(
    @InjectRepository(Region)
    private readonly regionRepository: Repository<Region>,
    @InjectRepository(Runner)
    private readonly runnerRepository: Repository<Runner>,
  ) {}

  /**
   * @param createRegionDto - The region details.
   * @param organizationId - The ID of the organization, or null for shared regions.
   * @throws {BadRequestException} If the region name is invalid.
   * @throws {ConflictException} If the region with the same ID already exists or region with the same name already exists in the organization.
   */
  async create(createRegionDto: CreateRegionInternalDto, organizationId: string | null): Promise<Region> {
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
      const region = new Region(
        createRegionDto.name,
        createRegionDto.enforceQuotas,
        createRegionDto.regionType,
        createRegionDto.id,
        organizationId,
      )
      return await this.regionRepository.save(region)
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
  async findOne(id: string): Promise<Region | null> {
    return await this.regionRepository.findOne({
      where: {
        id,
      },
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
   * @param type - The type of the regions to find.
   * @returns The regions found ordered by name ascending.
   */
  async findAllByRegionType(regionType: RegionType): Promise<Region[]> {
    return this.regionRepository.find({
      where: {
        regionType,
      },
      order: {
        name: 'ASC',
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

    const runners = await this.runnerRepository.find({
      where: {
        region: id,
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
