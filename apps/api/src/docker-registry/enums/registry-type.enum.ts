/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

/**
 * Types of registries available in the system.
 * @enum {string}
 */
export enum RegistryType {
  /**
   * Registry for storing snapshots that can be used for creating sandboxes
   */
  SNAPSHOT = 'snapshot',

  /**
   * Registry that is used as a source for pulling private images, before they are pushed to the snapshot registry
   */
  SOURCE = 'source',

  /**
   * Registry for storing sandbox backups
   */
  BACKUP = 'backup',

  /**
   * Registry used for pushing local images, before they are pushed to the snapshot registry
   */
  TRANSIENT = 'transient',
}
