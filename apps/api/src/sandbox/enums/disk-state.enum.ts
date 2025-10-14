/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export enum DiskState {
  FRESH = 'fresh',
  PULLING = 'pulling',
  READY = 'ready',
  ATTACHED = 'attached',
  DETACHED = 'detached',
  UPLOADING = 'uploading',
  STORED = 'stored',
}
