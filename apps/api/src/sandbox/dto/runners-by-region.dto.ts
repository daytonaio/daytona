/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'RunnersByRegion' })
export class RunnersByRegionDto {
  [region: string]: string[]
}
