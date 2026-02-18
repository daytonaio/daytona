/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty } from '@nestjs/swagger'
import { IsOptional, IsString, IsArray, IsEnum } from 'class-validator'
import { SandboxState } from '../enums/sandbox-state.enum'
import { ToArray } from '../../common/decorators/to-array.decorator'
import { PageLimit } from '../../common/decorators/page-limit.decorator'

const VALID_QUERY_STATES = Object.values(SandboxState).filter((state) => state !== SandboxState.DESTROYED)

export class ListSandboxesQueryDto {
  @ApiProperty({
    name: 'cursor',
    description: 'Pagination cursor from a previous response',
    required: false,
    type: String,
  })
  @IsOptional()
  @IsString()
  cursor?: string

  @PageLimit(100)
  limit = 100

  @ApiProperty({
    name: 'states',
    description: 'List of states to filter by.',
    required: false,
    enum: VALID_QUERY_STATES,
    isArray: true,
  })
  @IsOptional()
  @ToArray()
  @IsArray()
  @IsEnum(VALID_QUERY_STATES, {
    each: true,
    message: `each value must be one of the following values: ${VALID_QUERY_STATES.join(', ')}`,
  })
  states?: SandboxState[]
}
