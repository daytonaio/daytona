/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiSchema } from '@nestjs/swagger'
import { PageNumber } from '../../common/decorators/page-number.decorator'
import { PageLimit } from '../../common/decorators/page-limit.decorator'

@ApiSchema({ name: 'ListAuditLogsQuery' })
export class ListAuditLogsQueryDto {
  @PageNumber(1)
  page = 1

  @PageLimit(10)
  limit = 10
}
