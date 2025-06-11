/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IAuthContext } from './auth-context.interface'

export interface ProxyContext extends IAuthContext {
  proxy: boolean
}

export function isProxyContext(user: IAuthContext): user is ProxyContext {
  return 'proxy' in user
}
