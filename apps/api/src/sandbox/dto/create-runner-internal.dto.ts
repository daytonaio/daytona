/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxClass } from '../enums/sandbox-class.enum'

export class CreateRunnerInternalDto {
  domain: string
  apiUrl: string
  proxyUrl: string
  token?: string
  cpu: number
  memoryGiB: number
  diskGiB: number
  gpu: number
  gpuType: string
  class: SandboxClass
  regionId: string
  version: string
}
