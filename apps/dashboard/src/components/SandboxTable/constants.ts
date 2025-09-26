/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxState } from '@daytonaio/api-client'
import { CheckCircle, Circle, AlertTriangle, Timer, Archive } from 'lucide-react'
import { FacetedFilterOption } from './types'

const STATE_LABEL_MAPPING: Record<SandboxState, string> = {
  [SandboxState.STARTED]: 'Started',
  [SandboxState.STOPPED]: 'Stopped',
  [SandboxState.ERROR]: 'Error',
  [SandboxState.BUILD_FAILED]: 'Build Failed',
  [SandboxState.BUILDING_SNAPSHOT]: 'Building Snapshot',
  [SandboxState.PENDING_BUILD]: 'Pending Build',
  [SandboxState.RESTORING]: 'Restoring',
  [SandboxState.ARCHIVED]: 'Archived',
  [SandboxState.CREATING]: 'Creating',
  [SandboxState.STARTING]: 'Starting',
  [SandboxState.STOPPING]: 'Stopping',
  [SandboxState.DESTROYING]: 'Deleting',
  [SandboxState.DESTROYED]: 'Deleted',
  [SandboxState.PULLING_SNAPSHOT]: 'Pulling Snapshot',
  [SandboxState.UNKNOWN]: 'Unknown',
  [SandboxState.ARCHIVING]: 'Archiving',
}

export const STATUSES: FacetedFilterOption[] = [
  {
    label: getStateLabel(SandboxState.STARTED),
    value: SandboxState.STARTED,
    icon: CheckCircle,
  },
  { label: getStateLabel(SandboxState.STOPPED), value: SandboxState.STOPPED, icon: Circle },
  { label: getStateLabel(SandboxState.ERROR), value: SandboxState.ERROR, icon: AlertTriangle },
  { label: getStateLabel(SandboxState.BUILD_FAILED), value: SandboxState.BUILD_FAILED, icon: AlertTriangle },
  { label: getStateLabel(SandboxState.STARTING), value: SandboxState.STARTING, icon: Timer },
  { label: getStateLabel(SandboxState.STOPPING), value: SandboxState.STOPPING, icon: Timer },
  { label: getStateLabel(SandboxState.DESTROYING), value: SandboxState.DESTROYING, icon: Timer },
  { label: getStateLabel(SandboxState.ARCHIVED), value: SandboxState.ARCHIVED, icon: Archive },
  { label: getStateLabel(SandboxState.ARCHIVING), value: SandboxState.ARCHIVING, icon: Timer },
]

export function getStateLabel(state?: SandboxState): string {
  if (!state) {
    return 'Unknown'
  }
  return STATE_LABEL_MAPPING[state]
}
