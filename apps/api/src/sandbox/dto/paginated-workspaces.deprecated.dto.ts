/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { WorkspaceDto } from './workspace.deprecated.dto'

@ApiSchema({ name: 'PaginatedWorkspaces' })
export class PaginatedWorkspacesDto {
  @ApiProperty({ type: [WorkspaceDto] })
  items: WorkspaceDto[]

  @ApiProperty()
  total: number

  @ApiProperty()
  page: number

  @ApiProperty()
  totalPages: number
}
