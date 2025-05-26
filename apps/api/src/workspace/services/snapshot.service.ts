/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Injectable,
  NotFoundException,
  ConflictException,
  ForbiddenException,
  BadRequestException,
} from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { Snapshot } from '../entities/snapshot.entity'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { CreateSnapshotDto } from '../dto/create-snapshot.dto'
import { BuildInfo } from '../entities/build-info.entity'
import { CreateBuildInfoDto } from '../dto/create-build-info.dto'
import { generateBuildInfoHash as generateBuildSnapshotRef } from '../entities/build-info.entity'
import { OnEvent } from '@nestjs/event-emitter'
import { WorkspaceEvents } from '../constants/workspace-events.constants'
import { WorkspaceCreatedEvent } from '../events/workspace-create.event'
import { Organization } from '../../organization/entities/organization.entity'

@Injectable()
export class SnapshotService {
  constructor(
    @InjectRepository(Snapshot)
    private readonly snapshotRepository: Repository<Snapshot>,
    @InjectRepository(BuildInfo)
    private readonly buildInfoRepository: Repository<BuildInfo>,
  ) {}

  private validateSnapshotName(name: string): string | null {
    if (!name.includes(':') || name.endsWith(':') || /:\s*$/.test(name)) {
      return 'Snapshot name must include a tag (e.g., ubuntu:22.04)'
    }

    if (name.endsWith(':latest')) {
      return 'Snapshots with tag ":latest" are not allowed'
    }

    // Basic format check
    const snapshotNameRegex =
      /^[a-z0-9]+(?:[._-][a-z0-9]+)*(?:\/[a-z0-9]+(?:[._-][a-z0-9]+)*)*:[a-z0-9]+(?:[._-][a-z0-9]+)*$/

    if (!snapshotNameRegex.test(name)) {
      return 'Invalid snapshot name format. Must be lowercase, may contain digits, dots, dashes, and single slashes between components'
    }

    return null
  }

  async createSnapshot(
    organization: Organization,
    createSnapshotDto: CreateSnapshotDto,
    buildInfo?: CreateBuildInfoDto,
    general = false,
  ) {
    const validationError = this.validateSnapshotName(createSnapshotDto.name)
    if (validationError) {
      throw new BadRequestException(validationError)
    }

    // check if the organization has reached the snapshot quota
    const snapshots = await this.snapshotRepository.find({
      where: { organizationId: organization.id },
    })

    if (snapshots.length >= organization.snapshotQuota) {
      throw new ForbiddenException('Reached the maximum number of snapshots in the organization')
    }

    try {
      const snapshot = this.snapshotRepository.create({
        organizationId: organization.id,
        ...createSnapshotDto,
        state: buildInfo ? SnapshotState.BUILD_PENDING : SnapshotState.PENDING,
        general,
      })

      if (buildInfo) {
        const buildSnapshotRef = generateBuildSnapshotRef(buildInfo.dockerfileContent, buildInfo.contextHashes)

        // Check if buildInfo with the same snapshotRef already exists
        const existingBuildInfo = await this.buildInfoRepository.findOne({
          where: { snapshotRef: buildSnapshotRef },
        })

        if (existingBuildInfo) {
          snapshot.buildInfo = existingBuildInfo
        } else {
          const buildInfoEntity = this.buildInfoRepository.create({
            ...buildInfo,
          })
          await this.buildInfoRepository.save(buildInfoEntity)
          snapshot.buildInfo = buildInfoEntity
        }
      }

      return await this.snapshotRepository.save(snapshot)
    } catch (error) {
      if (error.code === '23505') {
        // PostgreSQL unique violation error code
        throw new ConflictException(
          `Snapshot with name "${createSnapshotDto.name}" already exists for this organization`,
        )
      }
      throw error
    }
  }

  async toggleSnapshotState(snapshotId: string, enabled: boolean) {
    const snapshot = await this.snapshotRepository.findOne({
      where: { id: snapshotId },
    })

    if (!snapshot) {
      throw new NotFoundException(`Snapshot with ID ${snapshotId} not found`)
    }

    snapshot.enabled = enabled
    return await this.snapshotRepository.save(snapshot)
  }

  async removeSnapshot(snapshotId: string) {
    const snapshot = await this.snapshotRepository.findOne({
      where: { id: snapshotId },
    })

    if (!snapshot) {
      throw new NotFoundException(`Snapshot with ID ${snapshotId} not found`)
    }
    if (snapshot.general) {
      throw new ForbiddenException('You cannot delete a general snapshot')
    }
    snapshot.state = SnapshotState.REMOVING
    await this.snapshotRepository.save(snapshot)
  }

  async getAllSnapshots(organizationId: string, page = 1, limit = 10) {
    const pageNum = Number(page)
    const limitNum = Number(limit)

    const [items, total] = await this.snapshotRepository.findAndCount({
      where: { organizationId },
      order: {
        lastUsedAt: {
          direction: 'DESC',
          nulls: 'LAST',
        },
        createdAt: 'DESC',
      },
      skip: (pageNum - 1) * limitNum,
      take: limitNum,
    })

    return {
      items,
      total,
      page: pageNum,
      totalPages: Math.ceil(total / limit),
    }
  }

  async getSnapshot(snapshotId: string): Promise<Snapshot> {
    const snapshot = await this.snapshotRepository.findOne({
      where: { id: snapshotId },
    })

    if (!snapshot) {
      throw new NotFoundException(`Snapshot with ID ${snapshotId} not found`)
    }

    return snapshot
  }

  async getSnapshotName(snapshotName: string, organizationId: string): Promise<Snapshot> {
    const snapshot = await this.snapshotRepository.findOne({
      where: { name: snapshotName, organizationId },
    })

    if (!snapshot) {
      //  check if the snapshot is general
      const generalSnapshot = await this.snapshotRepository.findOne({
        where: { name: snapshotName, general: true },
      })
      if (generalSnapshot) {
        return generalSnapshot
      }

      throw new NotFoundException(`Snapshot with name ${snapshotName} not found`)
    }

    return snapshot
  }

  async setSnapshotGeneralStatus(snapshotId: string, general: boolean) {
    const snapshot = await this.snapshotRepository.findOne({
      where: { id: snapshotId },
    })

    if (!snapshot) {
      throw new NotFoundException(`Snapshot with ID ${snapshotId} not found`)
    }

    snapshot.general = general
    return await this.snapshotRepository.save(snapshot)
  }

  @OnEvent(WorkspaceEvents.CREATED)
  private async handleWorkspaceCreatedEvent(event: WorkspaceCreatedEvent) {
    if (!event.workspace.snapshot) {
      return
    }

    const snapshot = await this.getSnapshotName(event.workspace.snapshot, event.workspace.organizationId)
    snapshot.lastUsedAt = event.workspace.createdAt
    await this.snapshotRepository.save(snapshot)
  }
}
