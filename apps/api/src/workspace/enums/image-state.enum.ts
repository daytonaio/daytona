/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export enum ImageState {
  BUILD_PENDING = 'build_pending',
  BUILDING = 'building',
  PENDING = 'pending',
  PULLING_IMAGE = 'pulling_image',
  PENDING_VALIDATION = 'pending_validation',
  VALIDATING = 'validating',
  ACTIVE = 'active',
  ERROR = 'error',
  REMOVING = 'removing',
}
