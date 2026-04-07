/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CanActivate, ExecutionContext } from '@nestjs/common'

/**
 *  Base class for guards that validate that the current request is authenticated with an auth context type that has access to the requested resource.
 */
export abstract class ResourceAccessGuard implements CanActivate {
  abstract canActivate(context: ExecutionContext): Promise<boolean>
}
