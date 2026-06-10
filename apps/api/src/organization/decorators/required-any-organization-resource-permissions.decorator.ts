/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Reflector } from '@nestjs/core'
import { OrganizationResourcePermission } from '../enums/organization-resource-permission.enum'

/**
 * Marks a controller or handler as requiring at least one of the specified resource permissions.
 *
 * Evaluated by `OrganizationAuthContextGuard`.
 */
export const RequiredAnyOrganizationResourcePermissions = Reflector.createDecorator<OrganizationResourcePermission[]>()
