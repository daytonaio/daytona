/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { AdminDockerRegistryController } from './docker-registry.controller'
import { SystemRole } from '../../user/enums/system-role.enum'
import { getRequiredSystemRole } from '../../test/helpers/controller-metadata.helper'

describe('[AUTH] AdminDockerRegistryController', () => {
  it('requires admin role', () => {
    expect(getRequiredSystemRole(AdminDockerRegistryController)).toBe(SystemRole.ADMIN)
  })
})
