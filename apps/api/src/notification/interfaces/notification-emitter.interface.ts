/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxDto } from '../../sandbox/dto/sandbox.dto'
import { SandboxState } from '../../sandbox/enums/sandbox-state.enum'
import { SandboxDesiredState } from '../../sandbox/enums/sandbox-desired-state.enum'
import { SnapshotDto } from '../../sandbox/dto/snapshot.dto'
import { SnapshotState } from '../../sandbox/enums/snapshot-state.enum'
import { VolumeDto } from '../../sandbox/dto/volume.dto'
import { VolumeState } from '../../sandbox/enums/volume-state.enum'
import { RunnerDto } from '../../sandbox/dto/runner.dto'
import { RunnerState } from '../../sandbox/enums/runner-state.enum'

export abstract class NotificationEmitterInterface {
  abstract emitSandboxCreated(sandbox: SandboxDto): void
  abstract emitSandboxStateUpdated(sandbox: SandboxDto, oldState: SandboxState, newState: SandboxState): void
  abstract emitSandboxDesiredStateUpdated(
    sandbox: SandboxDto,
    oldDesiredState: SandboxDesiredState,
    newDesiredState: SandboxDesiredState,
  ): void
  abstract emitSnapshotCreated(snapshot: SnapshotDto): void
  abstract emitSnapshotStateUpdated(snapshot: SnapshotDto, oldState: SnapshotState, newState: SnapshotState): void
  abstract emitSnapshotRemoved(snapshot: SnapshotDto): void
  abstract emitVolumeCreated(volume: VolumeDto): void
  abstract emitVolumeStateUpdated(volume: VolumeDto, oldState: VolumeState, newState: VolumeState): void
  abstract emitVolumeLastUsedAtUpdated(volume: VolumeDto): void
  abstract emitRunnerCreated(runner: RunnerDto, organizationId: string | null): void
  abstract emitRunnerStateUpdated(
    runner: RunnerDto,
    organizationId: string | null,
    oldState: RunnerState,
    newState: RunnerState,
  ): void
  abstract emitRunnerUnschedulableUpdated(runner: RunnerDto, organizationId: string | null): void
}
