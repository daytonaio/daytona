/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiSchema } from '@nestjs/swagger'
import { ListSandboxesQueryDto } from './list-sandboxes-query.dto'

@ApiSchema({ name: 'ListWorkspacesQuery' })
export class ListWorkspacesQueryDto extends ListSandboxesQueryDto {}
