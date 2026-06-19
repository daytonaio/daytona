/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import 'reflect-metadata'
import { validate } from 'class-validator'
import { plainToInstance } from 'class-transformer'
import { UpdateSandboxDegradedReasonDto } from './update-sandbox-degraded-reason.dto'

describe('UpdateSandboxDegradedReasonDto', () => {
  it('accepts a non-empty reason', async () => {
    const dto = plainToInstance(UpdateSandboxDegradedReasonDto, {
      degradedReason: 'fd-exhaustion: too many open files',
    })
    const errors = await validate(dto)
    expect(errors).toHaveLength(0)
  })

  it('accepts null to clear the reason', async () => {
    const dto = plainToInstance(UpdateSandboxDegradedReasonDto, { degradedReason: null })
    const errors = await validate(dto)
    expect(errors).toHaveLength(0)
  })

  it('accepts an omitted reason', async () => {
    const dto = plainToInstance(UpdateSandboxDegradedReasonDto, {})
    const errors = await validate(dto)
    expect(errors).toHaveLength(0)
  })

  it('rejects an empty string', async () => {
    const dto = plainToInstance(UpdateSandboxDegradedReasonDto, { degradedReason: '' })
    const errors = await validate(dto)
    expect(errors).toHaveLength(1)
    expect(errors[0].constraints).toHaveProperty('isNotEmpty')
  })

  it('rejects a non-string value', async () => {
    const dto = plainToInstance(UpdateSandboxDegradedReasonDto, { degradedReason: 42 })
    const errors = await validate(dto)
    expect(errors).toHaveLength(1)
    expect(errors[0].constraints).toHaveProperty('isString')
  })
})
