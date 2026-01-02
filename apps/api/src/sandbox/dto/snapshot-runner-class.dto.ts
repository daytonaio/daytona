/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'
import { RunnerClass } from '../enums/runner-class'

export class SnapshotRunnerClassDto {
  @ApiProperty({
    enum: RunnerClass,
    enumName: 'RunnerClass',
    description: 'The runner class of the snapshot',
  })
  runnerClass: RunnerClass
}
