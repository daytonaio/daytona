/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Reflector } from '@nestjs/core'
import { OrganizationMemberRole } from '../enums/organization-member-role.enum'

/**
 * Marks a controller or handler as requiring the specified organization member role.
 *
 * Evaluated by `OrganizationAuthContextGuard`.
 */
export const RequiredOrganizationMemberRole = Reflector.createDecorator<OrganizationMemberRole>()
