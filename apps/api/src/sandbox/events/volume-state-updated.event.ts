/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Volume } from '../entities/volume.entity'
import { VolumeState } from '../enums/volume-state.enum'

export class VolumeStateUpdatedEvent {
  constructor(
    public readonly volume: Volume,
    public readonly oldState: VolumeState,
    public readonly newState: VolumeState,
  ) {}
}
