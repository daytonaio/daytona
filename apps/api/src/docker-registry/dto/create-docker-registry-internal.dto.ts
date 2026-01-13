/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RegistryType } from '../enums/registry-type.enum'

export class CreateDockerRegistryInternalDto {
  name: string
  url: string
  username: string
  password: string
  project?: string
  registryType: RegistryType
  isDefault?: boolean
  regionId?: string | null
}
