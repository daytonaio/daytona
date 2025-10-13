/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module, Global } from '@nestjs/common'
import { ClickHouseService } from './clickhouse.service'

@Global()
@Module({
  providers: [ClickHouseService],
  exports: [ClickHouseService],
})
export class ClickHouseModule {}
