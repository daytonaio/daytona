/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxState } from '@daytonaio/api-client'
import { CheckCircle, Circle, AlertTriangle, Timer, Archive } from 'lucide-react'
import { FacetedFilterOption } from './types'

const STATE_PRIORITY_ORDER_ARRAY = [
  SandboxState.STARTED,
  SandboxState.BUILDING_SNAPSHOT,
  SandboxState.PENDING_BUILD,
  SandboxState.RESTORING,
  SandboxState.ERROR,
  SandboxState.BUILD_FAILED,
  SandboxState.STOPPED,
  SandboxState.ARCHIVING,
  SandboxState.ARCHIVED,
  SandboxState.CREATING,
  SandboxState.STARTING,
  SandboxState.STOPPING,
  SandboxState.DESTROYING,
  SandboxState.DESTROYED,
  SandboxState.PULLING_SNAPSHOT,
  SandboxState.UNKNOWN,
] as const

const STATE_COLOR_MAPPING = {
  [SandboxState.STARTED]: 'text-green-500',
  [SandboxState.STOPPED]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.ERROR]: 'text-red-500',
  [SandboxState.BUILD_FAILED]: 'text-red-500',
  [SandboxState.BUILDING_SNAPSHOT]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.PENDING_BUILD]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.RESTORING]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.ARCHIVING]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.ARCHIVED]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.CREATING]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.STARTING]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.STOPPING]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.DESTROYING]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.DESTROYED]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.PULLING_SNAPSHOT]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.UNKNOWN]: 'text-gray-800 dark:text-gray-200',
} as const

const STATE_LABEL_MAPPING: Record<SandboxState, string> = {
  [SandboxState.STARTED]: 'Started',
  [SandboxState.STOPPED]: 'Stopped',
  [SandboxState.ERROR]: 'Error',
  [SandboxState.BUILD_FAILED]: 'Build Failed',
  [SandboxState.BUILDING_SNAPSHOT]: 'Building Snapshot',
  [SandboxState.PENDING_BUILD]: 'Pending Build',
  [SandboxState.RESTORING]: 'Restoring',
  [SandboxState.ARCHIVING]: 'Archiving',
  [SandboxState.ARCHIVED]: 'Archived',
  [SandboxState.CREATING]: 'Creating',
  [SandboxState.STARTING]: 'Starting',
  [SandboxState.STOPPING]: 'Stopping',
  [SandboxState.DESTROYING]: 'Deleting',
  [SandboxState.DESTROYED]: 'Destroyed',
  [SandboxState.PULLING_SNAPSHOT]: 'Pulling Snapshot',
  [SandboxState.UNKNOWN]: 'Unknown',
}

export const STATE_PRIORITY_ORDER: Record<SandboxState, number> = Object.fromEntries(
  STATE_PRIORITY_ORDER_ARRAY.map((state, index) => [state, index + 1]),
) as Record<SandboxState, number>

export const STATE_COLORS: Record<SandboxState, string> = STATE_COLOR_MAPPING

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
  { label: getStateLabel(SandboxState.ARCHIVING), value: SandboxState.ARCHIVING, icon: Timer },
  { label: getStateLabel(SandboxState.ARCHIVED), value: SandboxState.ARCHIVED, icon: Archive },
]

export function getStateLabel(state?: SandboxState): string {
  if (!state) {
    return 'Unknown'
  }
  return STATE_LABEL_MAPPING[state]
}
