/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

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
}

export type CreateRunnerV2InternalDto = {
  apiKey?: string
  regionId: string
  name: string
  apiVersion: '2'
  appVersion?: string
}

export type CreateRunnerInternalDto = CreateRunnerV0InternalDto | CreateRunnerV2InternalDto
