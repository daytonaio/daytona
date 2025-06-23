/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { OnEvent } from '@nestjs/event-emitter'
import { SnapshotRunner } from '../entities/snapshot-runner.entity'
import { SnapshotRunnerState } from '../enums/snapshot-runner-state.enum'
import { OrganizationEvents } from '../../organization/constants/organization-events.constant'
import { OrganizationSuspendedSnapshotRunnerRemovedEvent } from '../../organization/events/organization-suspended-snapshot-runner-removed'

@Injectable()
export class SnapshotRunnerService {
  private readonly logger = new Logger(SnapshotRunnerService.name)

  constructor(
    @InjectRepository(SnapshotRunner)
    private readonly snapshotRunnerRepository: Repository<SnapshotRunner>,
  ) {}

  async remove(snapshotRunnerId: string): Promise<void> {
    const snapshotRunner = await this.snapshotRunnerRepository.findOne({
      where: {
        id: snapshotRunnerId,
      },
    })

    if (!snapshotRunner) {
      throw new NotFoundException()
    }

    snapshotRunner.state = SnapshotRunnerState.REMOVING
    await this.snapshotRunnerRepository.save(snapshotRunner)
  }

  @OnEvent(OrganizationEvents.SUSPENDED_SNAPSHOT_RUNNER_REMOVED)
  async handleSuspendedSnapshotRunnerRemoved(event: OrganizationSuspendedSnapshotRunnerRemovedEvent) {
    await this.remove(event.snapshotRunnerId).catch((error) => {
      //  log the error for now, but don't throw it as it will be retried
      this.logger.error(
        `Error removing snapshot runner from suspended organization. SnapshotRunnerId: ${event.snapshotRunnerId}: `,
        error,
      )
    })
  }
}
