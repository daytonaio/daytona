/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export enum RegionType {
  /**
   * Shared by all organizations.
   */
  SHARED = 'shared',
  /**
   * Dedicated to specific organizations.
   */
  DEDICATED = 'dedicated',
  /**
   * Created by and owned by a specific organization.
   */
  CUSTOM = 'custom',
}
