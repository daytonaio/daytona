/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Reflector } from '@nestjs/core'
import { OrganizationResourcePermission } from '../enums/organization-resource-permission.enum'

export const RequiredOrganizationResourcePermissions = Reflector.createDecorator<OrganizationResourcePermission[]>()
