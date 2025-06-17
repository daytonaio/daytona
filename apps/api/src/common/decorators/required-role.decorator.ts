/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Reflector } from '@nestjs/core'
import { SystemRole } from '../../user/enums/system-role.enum'
import { ApiRole } from '../interfaces/auth-context.interface'

export const RequiredSystemRole = Reflector.createDecorator<SystemRole | SystemRole[]>()
export const RequiredApiRole = Reflector.createDecorator<ApiRole | ApiRole[]>()
