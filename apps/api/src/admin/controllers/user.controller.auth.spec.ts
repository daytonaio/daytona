/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AdminUserController } from './user.controller'
import { SystemRole } from '../../user/enums/system-role.enum'
import { getRequiredSystemRole } from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] AdminUserController', () => {
  it('requires admin role', () => {
    expect(getRequiredSystemRole(AdminUserController)).toBe(SystemRole.ADMIN)
  })
})
