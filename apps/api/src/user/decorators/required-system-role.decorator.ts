/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Reflector } from '@nestjs/core'
import { SystemRole } from '../enums/system-role.enum'

/**
 * Marks a controller or handler as requiring one of the specified system roles.
 * When multiple roles are provided, access is granted if the user holds _any_ of them.
 *
 * Evaluated by `SystemActionGuard`.
 */
export const RequiredSystemRole = Reflector.createDecorator<SystemRole | SystemRole[]>()
