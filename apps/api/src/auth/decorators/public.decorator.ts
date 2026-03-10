/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SetMetadata } from '@nestjs/common'

export const IS_PUBLIC_KEY = 'isPublic'

/**
 * Marks a controller or handler as publicly accessible, bypassing authentication.
 *
 * Evaluated by `GlobalAuthGuard`.
 */
export const Public = () => SetMetadata(IS_PUBLIC_KEY, true)
