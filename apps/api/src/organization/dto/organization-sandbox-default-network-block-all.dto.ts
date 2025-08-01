/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'OrganizationSandboxDefaultNetworkBlockAll' })
export class OrganizationSandboxDefaultNetworkBlockAllDto {
  @ApiProperty({
    description: 'Sandbox default network block all',
  })
  sandboxDefaultNetworkBlockAll: boolean
}
