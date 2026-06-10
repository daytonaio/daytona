/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */
import { Snapshot } from '../entities/snapshot.entity'

/** A "cold" snapshot is never auto-propagated to runners; sandboxes pull it on demand. */
export function isColdSnapshot(snapshot: Pick<Snapshot, 'propagationFactor'>): boolean {
  return snapshot.propagationFactor === 0
}
