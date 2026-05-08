/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AdminOrganizationController } from './organization.controller'
import { SystemRole } from '../../user/enums/system-role.enum'
import { getRequiredSystemRole } from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] AdminOrganizationController', () => {
  it('requires admin role', () => {
    expect(getRequiredSystemRole(AdminOrganizationController)).toBe(SystemRole.ADMIN)
  })
})
