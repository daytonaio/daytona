/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext, isBaseAuthContext } from './base-auth-context.interface'

export interface RegionAuthContext extends BaseAuthContext {
  regionId: string
}

export function isRegionAuthContext(user: unknown): user is RegionAuthContext {
  return isBaseAuthContext(user) && 'regionId' in user
}
