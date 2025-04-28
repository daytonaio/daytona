/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Reflector } from '@nestjs/core'
import { SystemRole } from '../../user/enums/system-role.enum'

export const RequiredSystemRole = Reflector.createDecorator<SystemRole>()
