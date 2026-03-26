/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Sandbox, SandboxState } from '@daytonaio/api-client'

export function isStartable(sandbox: Sandbox): boolean {
  return sandbox.state === SandboxState.STOPPED || sandbox.state === SandboxState.ARCHIVED
}

export function isStoppable(sandbox: Sandbox): boolean {
  return sandbox.state === SandboxState.STARTED
}

export function isArchivable(sandbox: Sandbox): boolean {
  return sandbox.state === SandboxState.STOPPED
}

export function isRecoverable(sandbox: Sandbox): boolean {
  return sandbox.state === SandboxState.ERROR && sandbox.recoverable === true
}

export function isDeletable(_sandbox: Sandbox): boolean {
  return true
}

export function isTransitioning(sandbox: Sandbox): boolean {
  return (
    sandbox.state === SandboxState.CREATING ||
    sandbox.state === SandboxState.STARTING ||
    sandbox.state === SandboxState.STOPPING ||
    sandbox.state === SandboxState.DESTROYING ||
    sandbox.state === SandboxState.ARCHIVING ||
    sandbox.state === SandboxState.RESTORING ||
    sandbox.state === SandboxState.BUILDING_SNAPSHOT ||
    sandbox.state === SandboxState.PULLING_SNAPSHOT
  )
}

export function getSandboxDisplayLabel(sandbox: Sandbox): string {
  return sandbox.name ? `${sandbox.name} (${sandbox.id})` : sandbox.id
}

export function filterStartable<T extends Sandbox>(sandboxes: T[]): T[] {
  return sandboxes.filter(isStartable)
}

export function filterStoppable<T extends Sandbox>(sandboxes: T[]): T[] {
  return sandboxes.filter(isStoppable)
}

export function filterArchivable<T extends Sandbox>(sandboxes: T[]): T[] {
  return sandboxes.filter(isArchivable)
}

export function filterDeletable<T extends Sandbox>(sandboxes: T[]): T[] {
  return sandboxes.filter(isDeletable)
}

export interface BulkActionCounts {
  startable: number
  stoppable: number
  archivable: number
  deletable: number
}

export function getBulkActionCounts(sandboxes: Sandbox[]): BulkActionCounts {
  return {
    startable: filterStartable(sandboxes).length,
    stoppable: filterStoppable(sandboxes).length,
    archivable: filterArchivable(sandboxes).length,
    deletable: filterDeletable(sandboxes).length,
  }
}
