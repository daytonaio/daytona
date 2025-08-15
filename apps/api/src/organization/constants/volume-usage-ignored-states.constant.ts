/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { VolumeState } from '../../sandbox/enums/volume-state.enum'

export const VOLUME_USAGE_IGNORED_STATES: VolumeState[] = [VolumeState.DELETED, VolumeState.ERROR]
