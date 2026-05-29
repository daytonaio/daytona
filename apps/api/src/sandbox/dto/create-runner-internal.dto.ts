/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxClass } from '../enums/sandbox-class.enum'

export type CreateRunnerV0InternalDto = {
  domain: string
  apiUrl: string
  proxyUrl: string
  cpu: number
  memoryGiB: number
  diskGiB: number
  regionId: string
  name: string
  apiKey?: string
  apiVersion: '0'
  appVersion?: string
  tags?: string[]
  sandboxClass?: SandboxClass
}

export type CreateRunnerV2InternalDto = {
  apiKey?: string
  regionId: string
  name: string
  apiVersion: '2'
  appVersion?: string
  tags?: string[]
  sandboxClass?: SandboxClass
}

export type CreateRunnerInternalDto = CreateRunnerV0InternalDto | CreateRunnerV2InternalDto
