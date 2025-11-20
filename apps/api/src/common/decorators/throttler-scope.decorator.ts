/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SetMetadata } from '@nestjs/common'

export const THROTTLER_SCOPE_KEY = 'throttler:scope'

/**
 * Marks a route or controller with specific throttler scopes.
 * Only the specified throttlers will be applied to this route.
 * The 'authenticated' throttler always applies to authenticated routes.
 *
 * @example
 * // Apply sandbox-create throttler
 * @ThrottlerScope('sandbox-create')
 * @Post()
 * createSandbox() {}
 *
 * @example
 * // Apply multiple throttlers
 * @ThrottlerScope('sandbox-create', 'sandbox-lifecycle')
 * @Post()
 * createAndStart() {}
 */
export const ThrottlerScope = (...scopes: string[]) => SetMetadata(THROTTLER_SCOPE_KEY, scopes)
