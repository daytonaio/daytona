/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AdminSnapshotController } from './snapshot.controller'
import { SystemRole } from '../../user/enums/system-role.enum'
import { getRequiredSystemRole } from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] AdminSnapshotController', () => {
  it('requires admin role', () => {
    expect(getRequiredSystemRole(AdminSnapshotController)).toBe(SystemRole.ADMIN)
  })
})
