/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxListItem, SandboxState } from '@daytona/api-client'

export function isStartable(sandbox: SandboxListItem): boolean {
  return (
    sandbox.state === SandboxState.STOPPED ||
    sandbox.state === SandboxState.ARCHIVED ||
    sandbox.state === SandboxState.PAUSED
  )
}

export function isStoppable(sandbox: SandboxListItem): boolean {
  return sandbox.state === SandboxState.STARTED || sandbox.state === SandboxState.PAUSED
}

export function isArchivable(sandbox: SandboxListItem): boolean {
  return sandbox.state === SandboxState.STOPPED
}

export function isPausable(sandbox: SandboxListItem): boolean {
  return sandbox.state === SandboxState.STARTED
}

export function isRecoverable(sandbox: SandboxListItem): boolean {
  return sandbox.state === SandboxState.ERROR && sandbox.recoverable === true
}

export function isDeletable(_sandbox: SandboxListItem): boolean {
  return true
}

export function isTransitioning(sandbox: SandboxListItem): boolean {
  return (
    sandbox.state === SandboxState.CREATING ||
    sandbox.state === SandboxState.STARTING ||
    sandbox.state === SandboxState.STOPPING ||
    sandbox.state === SandboxState.DESTROYING ||
    sandbox.state === SandboxState.ARCHIVING ||
    sandbox.state === SandboxState.RESTORING ||
    sandbox.state === SandboxState.BUILDING_SNAPSHOT ||
    sandbox.state === SandboxState.PULLING_SNAPSHOT ||
    sandbox.state === SandboxState.PAUSING ||
    sandbox.state === SandboxState.RESUMING
  )
}

export function getSandboxDisplayLabel(sandbox: SandboxListItem): string {
  return sandbox.name ? `${sandbox.name} (${sandbox.id})` : sandbox.id
}

export function filterStartable<T extends SandboxListItem>(sandboxes: T[]): T[] {
  return sandboxes.filter(isStartable)
}

export function filterStoppable<T extends SandboxListItem>(sandboxes: T[]): T[] {
  return sandboxes.filter(isStoppable)
}

export function filterArchivable<T extends SandboxListItem>(sandboxes: T[]): T[] {
  return sandboxes.filter(isArchivable)
}

export function filterDeletable<T extends SandboxListItem>(sandboxes: T[]): T[] {
  return sandboxes.filter(isDeletable)
}

export interface BulkActionCounts {
  startable: number
  stoppable: number
  archivable: number
  deletable: number
}

export function getBulkActionCounts(sandboxes: SandboxListItem[]): BulkActionCounts {
  return {
    startable: filterStartable(sandboxes).length,
    stoppable: filterStoppable(sandboxes).length,
    archivable: filterArchivable(sandboxes).length,
    deletable: filterDeletable(sandboxes).length,
  }
}
