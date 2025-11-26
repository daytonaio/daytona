/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export class CreateRegionInternalDto {
  id?: string
  name: string
  proxyUrl?: string | null
  sshGatewayUrl?: string | null
}
