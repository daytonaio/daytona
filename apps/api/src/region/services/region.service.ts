/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, NotFoundException, ConflictException, BadRequestException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { IsNull, Repository } from 'typeorm'
import { REGION_NAME_REGEX } from '../constants/region-name-regex.constant'
import { CreateRegionInternalDto } from '../dto/create-region.internal.dto'
import { Region } from '../entities/region.entity'

@Injectable()
export class RegionService {
  private readonly logger = new Logger(RegionService.name)

  constructor(
    @InjectRepository(Region)
    private readonly regionRepository: Repository<Region>,
  ) {}

  /**
   * @param createRegionDto - The region details.
   * @param organizationId - The ID of the organization, or null for non-organization regions.
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
      const region = new Region(createRegionDto.name, createRegionDto.enforceQuotas, createRegionDto.id, organizationId)
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
   * @param organizationId - The ID of the organization, or null for non-organization regions, or undefined to skip organization check.
   * @returns The region if found, or null otherwise.
   */
  async findOne(id: string, organizationId?: string | null): Promise<Region | null> {
    const region = await this.regionRepository.findOne({
      where: {
        id,
        ...(organizationId === undefined
          ? {}
          : organizationId === null
            ? { organizationId: IsNull() }
            : { organizationId }),
      },
    })

    if (!region) {
      return null
    }

    return region
  }

  /**
   * @param name - The name of the region.
   * @param organizationId - The organization ID, or null for non-organization regions.
   * @returns The region if found, or null otherwise.
   */
  async findOneByName(name: string, organizationId: string | null): Promise<Region | null> {
    const region = await this.regionRepository.findOne({
      where: [{ name, organizationId: organizationId ?? IsNull() }],
    })

    if (!region) {
      return null
    }

    return region
  }

  /**
   * @param regionId - The ID of the region.
   * @returns The ID of the organization or null for non-organization regions if the region is found, or undefined if the region is not found.
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
   * @param organizationId - The organization ID of the regions to find, or null for non-organization regions.
   * @returns The regions found ordered by name ascending.
   */
  async findAll(organizationId: string | null): Promise<Region[]> {
    return this.regionRepository.find({
      where: {
        organizationId: organizationId ?? IsNull(),
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
    const result = await this.regionRepository.delete(id)

    if (!result.affected) {
      throw new NotFoundException('Region not found')
    }
  }
}
