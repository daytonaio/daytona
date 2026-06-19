/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxClass, SandboxState } from '@daytona/api-client'
import { AndroidLogoIcon, LinuxLogoIcon, WindowsLogoIcon } from '@phosphor-icons/react'
import { AlertTriangle, Archive, CheckCircle, Circle, Container, LucideIcon, Timer } from 'lucide-react'
import { FacetedFilterOption } from './types'

const STATE_PRIORITY_ORDER_ARRAY = [
  SandboxState.STARTED,
  SandboxState.BUILDING_SNAPSHOT,
  SandboxState.PENDING_BUILD,
  SandboxState.RESTORING,
  SandboxState.ERROR,
  SandboxState.BUILD_FAILED,
  SandboxState.STOPPED,
  SandboxState.ARCHIVED,
  SandboxState.CREATING,
  SandboxState.STARTING,
  SandboxState.STOPPING,
  SandboxState.DESTROYING,
  SandboxState.DESTROYED,
  SandboxState.PULLING_SNAPSHOT,
  SandboxState.PAUSING,
  SandboxState.PAUSED,
  SandboxState.RESUMING,
  SandboxState.UNKNOWN,
  SandboxState.UNKNOWN_DEFAULT_OPEN_API,
] as const

const STATE_COLOR_MAPPING = {
  [SandboxState.STARTED]: 'text-green-500',
  [SandboxState.STOPPED]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.ERROR]: 'text-red-500',
  [SandboxState.BUILD_FAILED]: 'text-red-500',
  [SandboxState.BUILDING_SNAPSHOT]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.PENDING_BUILD]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.RESTORING]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.ARCHIVED]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.CREATING]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.STARTING]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.STOPPING]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.DESTROYING]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.DESTROYED]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.PULLING_SNAPSHOT]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.UNKNOWN]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.ARCHIVING]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.RESIZING]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.SNAPSHOTTING]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.FORKING]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.PAUSING]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.PAUSED]: 'text-yellow-600 dark:text-yellow-400',
  [SandboxState.RESUMING]: 'text-gray-800 dark:text-gray-200',
  [SandboxState.UNKNOWN_DEFAULT_OPEN_API]: 'text-gray-800 dark:text-gray-200',
} as const

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
  [SandboxState.RESIZING]: 'Resizing',
  [SandboxState.SNAPSHOTTING]: 'Snapshotting',
  [SandboxState.FORKING]: 'Forking',
  [SandboxState.PAUSING]: 'Pausing',
  [SandboxState.PAUSED]: 'Paused',
  [SandboxState.RESUMING]: 'Resuming',
  [SandboxState.UNKNOWN_DEFAULT_OPEN_API]: 'Unknown',
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
  { label: getStateLabel(SandboxState.ARCHIVED), value: SandboxState.ARCHIVED, icon: Archive },
  { label: getStateLabel(SandboxState.ARCHIVING), value: SandboxState.ARCHIVING, icon: Timer },
  { label: getStateLabel(SandboxState.PAUSED), value: SandboxState.PAUSED, icon: Circle },
  { label: getStateLabel(SandboxState.PAUSING), value: SandboxState.PAUSING, icon: Timer },
  { label: getStateLabel(SandboxState.RESUMING), value: SandboxState.RESUMING, icon: Timer },
]

export function getStateLabel(state?: SandboxState): string {
  if (!state) {
    return 'Unknown'
  }
  return STATE_LABEL_MAPPING[state]
}

const SANDBOX_CLASS_LABEL_MAPPING: Record<SandboxClass, string> = {
  [SandboxClass.CONTAINER]: 'Container',
  [SandboxClass.LINUX_VM]: 'Linux VM',
  [SandboxClass.ANDROID]: 'Android',
  [SandboxClass.WINDOWS]: 'Windows',
  [SandboxClass.UNKNOWN_DEFAULT_OPEN_API]: 'Unknown',
}

const SANDBOX_CLASS_ICON_MAPPING: Record<SandboxClass, LucideIcon> = {
  [SandboxClass.CONTAINER]: Container,
  [SandboxClass.LINUX_VM]: LinuxLogoIcon,
  [SandboxClass.ANDROID]: AndroidLogoIcon,
  [SandboxClass.WINDOWS]: WindowsLogoIcon,
  [SandboxClass.UNKNOWN_DEFAULT_OPEN_API]: Container,
}

export const SANDBOX_CLASS_OPTIONS: FacetedFilterOption[] = [
  {
    label: SANDBOX_CLASS_LABEL_MAPPING[SandboxClass.CONTAINER],
    value: SandboxClass.CONTAINER,
    icon: SANDBOX_CLASS_ICON_MAPPING[SandboxClass.CONTAINER],
  },
  {
    label: SANDBOX_CLASS_LABEL_MAPPING[SandboxClass.LINUX_VM],
    value: SandboxClass.LINUX_VM,
    icon: SANDBOX_CLASS_ICON_MAPPING[SandboxClass.LINUX_VM],
  },
  {
    label: SANDBOX_CLASS_LABEL_MAPPING[SandboxClass.ANDROID],
    value: SandboxClass.ANDROID,
    icon: SANDBOX_CLASS_ICON_MAPPING[SandboxClass.ANDROID],
  },
  {
    label: SANDBOX_CLASS_LABEL_MAPPING[SandboxClass.WINDOWS],
    value: SandboxClass.WINDOWS,
    icon: SANDBOX_CLASS_ICON_MAPPING[SandboxClass.WINDOWS],
  },
]

export function getSandboxClassLabel(sandboxClass?: SandboxClass): string {
  if (!sandboxClass) {
    return SANDBOX_CLASS_LABEL_MAPPING[SandboxClass.CONTAINER]
  }
  return SANDBOX_CLASS_LABEL_MAPPING[sandboxClass]
}

export function getSandboxClassIcon(sandboxClass?: SandboxClass): LucideIcon {
  if (!sandboxClass) {
    return SANDBOX_CLASS_ICON_MAPPING[SandboxClass.CONTAINER]
  }
  return SANDBOX_CLASS_ICON_MAPPING[sandboxClass]
}
