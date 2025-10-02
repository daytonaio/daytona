/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { IsEnum, IsNumber, IsString } from 'class-validator'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { ApiProperty, ApiSchema } from '@nestjs/swagger'

@ApiSchema({ name: 'CreateRunner' })
export class CreateRunnerDto {
  @ApiProperty()
  @IsString()
  domain: string

  @IsString()
  @ApiProperty()
  apiUrl: string

  @IsString()
  @ApiProperty()
  proxyUrl: string

  @IsString()
  @ApiProperty()
  apiKey: string

  @IsNumber()
  @ApiProperty()
  cpu: number

  @IsNumber()
  @ApiProperty()
  memoryGiB: number

  @IsNumber()
  @ApiProperty()
  diskGiB: number

  @IsNumber()
  @ApiProperty()
  gpu: number

  @IsString()
  @ApiProperty()
  gpuType: string

  @IsEnum(SandboxClass)
  @ApiProperty({
    enum: SandboxClass,
    example: Object.values(SandboxClass)[0],
  })
  class: SandboxClass

  @IsString()
  @ApiProperty({
    example: 'us',
  })
  region: string

  @IsString()
  @ApiProperty()
  version: string

  constructor(createParams: {
    domain: string
    apiUrl: string
    proxyUrl: string
    apiKey: string
    cpu: number
    memoryGiB: number
    diskGiB: number
    gpu: number
    gpuType: string
    class: SandboxClass
    region: string
    version: string
  }) {
    this.domain = createParams.domain
    this.apiUrl = createParams.apiUrl
    this.proxyUrl = createParams.proxyUrl
    this.apiKey = createParams.apiKey
    this.cpu = createParams.cpu
    this.memoryGiB = createParams.memoryGiB
    this.diskGiB = createParams.diskGiB
    this.gpu = createParams.gpu
    this.gpuType = createParams.gpuType
    this.class = createParams.class
    this.region = createParams.region
    this.version = createParams.version
  }
}
