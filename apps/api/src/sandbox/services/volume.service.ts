/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ForbiddenException, Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository, Not, In } from 'typeorm'
import { Volume } from '../entities/volume.entity'
import { VolumeState } from '../enums/volume-state.enum'
import { CreateVolumeDto } from '../dto/create-volume.dto'
import { v4 as uuidv4 } from 'uuid'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { Organization } from '../../organization/entities/organization.entity'

@Injectable()
export class VolumeService {
  private readonly logger = new Logger(VolumeService.name)

  constructor(
    @InjectRepository(Volume)
    private readonly volumeRepository: Repository<Volume>,
  ) {}

  async create(organization: Organization, createVolumeDto: CreateVolumeDto): Promise<Volume> {
    // Validate quota
    const activeVolumeCount = await this.countActive(organization.id)

    if (activeVolumeCount >= organization.volumeQuota) {
      throw new ForbiddenException(`Volume quota limit (${organization.volumeQuota}) reached`)
    }

    const volume = new Volume()

    // Generate ID
    volume.id = uuidv4()

    // Set name from DTO or use ID as default
    volume.name = createVolumeDto.name || volume.id

    // Check if volume with same name already exists for organization
    const existingVolume = await this.volumeRepository.findOne({
      where: {
        organizationId: organization.id,
        name: volume.name,
        state: Not(VolumeState.DELETED),
      },
    })

    if (existingVolume) {
      throw new BadRequestError(`Volume with name ${volume.name} already exists`)
    }

    volume.organizationId = organization.id
    volume.state = VolumeState.PENDING_CREATE
    volume.lastUsedAt = new Date()

    const savedVolume = await this.volumeRepository.save(volume)
    this.logger.debug(`Created volume ${savedVolume.id} for organization ${organization.id}`)
    return savedVolume
  }

  async delete(volumeId: string): Promise<void> {
    const volume = await this.volumeRepository.findOne({
      where: {
        id: volumeId,
      },
    })

    if (!volume) {
      throw new NotFoundException(`Volume with ID ${volumeId} not found`)
    }

    if (volume.state !== VolumeState.READY) {
      throw new BadRequestError(`Volume must be in '${VolumeState.READY}' state in order to be deleted`)
    }

    // Update state to mark as deleting
    volume.state = VolumeState.PENDING_DELETE
    await this.volumeRepository.save(volume)
    this.logger.debug(`Marked volume ${volumeId} for deletion`)
  }

  async findOne(volumeId: string): Promise<Volume> {
    const volume = await this.volumeRepository.findOne({
      where: { id: volumeId },
    })

    if (!volume) {
      throw new NotFoundException(`Volume with ID ${volumeId} not found`)
    }

    return volume
  }

  async findAll(organizationId: string, includeDeleted = false): Promise<Volume[]> {
    return this.volumeRepository.find({
      where: {
        organizationId,
        ...(includeDeleted ? {} : { state: Not(VolumeState.DELETED) }),
      },
    })
  }

  async findByName(organizationId: string, name: string): Promise<Volume> {
    const volume = await this.volumeRepository.findOne({
      where: {
        organizationId,
        name,
        state: Not(VolumeState.DELETED),
      },
    })

    if (!volume) {
      throw new NotFoundException(`Volume with name ${name} not found`)
    }

    return volume
  }

  async countActive(organizationId: string): Promise<number> {
    return this.volumeRepository.count({
      where: {
        organizationId,
        state: Not(In([VolumeState.DELETED, VolumeState.ERROR])),
      },
    })
  }
}
