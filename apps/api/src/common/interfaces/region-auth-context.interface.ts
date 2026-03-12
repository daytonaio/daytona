/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BaseAuthContext } from './auth-context.interface'

export interface RegionAuthContext extends BaseAuthContext {
  regionId: string
}

export function isRegionAuthContext(user: BaseAuthContext): user is RegionAuthContext {
  return 'regionId' in user
}
