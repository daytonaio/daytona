/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AdminAuditController } from './audit.controller'
import { SystemRole } from '../../user/enums/system-role.enum'
import { getRequiredSystemRole } from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] AdminAuditController', () => {
  it('requires admin role', () => {
    expect(getRequiredSystemRole(AdminAuditController)).toBe(SystemRole.ADMIN)
  })
})
