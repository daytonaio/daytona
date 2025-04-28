/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export enum ImageNodeState {
  PULLING_IMAGE = 'pulling_image',
  BUILDING_IMAGE = 'building_image',
  READY = 'ready',
  ERROR = 'error',
  REMOVING = 'removing',
}
