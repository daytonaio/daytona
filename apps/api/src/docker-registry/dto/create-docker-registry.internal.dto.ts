/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RegistryType } from './../../docker-registry/enums/registry-type.enum'

export class CreateDockerRegistryInternalDto {
  name: string
  url: string
  username: string
  password: string
  project?: string
  registryType: RegistryType
  isActive?: boolean
  isFallback?: boolean
}
