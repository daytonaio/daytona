/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiSchema, OmitType } from '@nestjs/swagger'
import { ListSandboxesQueryDto } from './list-sandboxes-query.dto'

@ApiSchema({ name: 'GetSandboxesSummaryQuery' })
export class GetSandboxesSummaryQueryDto extends OmitType(ListSandboxesQueryDto, [
  'cursor',
  'limit',
  'sort',
  'order',
] as const) {}
