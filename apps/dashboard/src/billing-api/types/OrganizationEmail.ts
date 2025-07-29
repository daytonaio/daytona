/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export type OrganizationEmail = {
  email: string
  verified: boolean
  owner: boolean
  business: boolean
  verifiedAt?: Date
}
