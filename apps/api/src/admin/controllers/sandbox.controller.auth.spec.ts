/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AdminSandboxController } from './sandbox.controller'
import { SystemRole } from '../../user/enums/system-role.enum'
import { getRequiredSystemRole } from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] AdminSandboxController', () => {
  it('requires admin role', () => {
    expect(getRequiredSystemRole(AdminSandboxController)).toBe(SystemRole.ADMIN)
  })
})
