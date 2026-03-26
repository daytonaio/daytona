/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { RegistryType } from './../../docker-registry/enums/registry-type.enum'
import { DockerRegistry } from '../entities/docker-registry.entity'

@ApiSchema({ name: 'DockerRegistry' })
export class DockerRegistryDto {
  @ApiProperty({
    description: 'Registry ID',
    example: '123e4567-e89b-12d3-a456-426614174000',
  })
  id: string

  @ApiProperty({
    description: 'Registry name',
    example: 'My Docker Hub',
  })
  name: string

  @ApiProperty({
    description: 'Registry URL',
    example: 'https://registry.hub.docker.com',
  })
  url: string

  @ApiProperty({
    description: 'Registry username',
    example: 'username',
  })
  username: string

  @ApiProperty({
    description: 'Registry project',
    example: 'my-project',
  })
  project: string

  @ApiProperty({
    description: 'Registry type',
    enum: RegistryType,
    example: RegistryType.INTERNAL,
  })
  registryType: RegistryType

  @ApiProperty({
    description: 'Creation timestamp',
    example: '2024-01-31T12:00:00Z',
  })
  createdAt: Date

  @ApiProperty({
    description: 'Last update timestamp',
    example: '2024-01-31T12:00:00Z',
  })
  updatedAt: Date

  static fromDockerRegistry(dockerRegistry: DockerRegistry): DockerRegistryDto {
    const dto: DockerRegistryDto = {
      id: dockerRegistry.id,
      name: dockerRegistry.name,
      url: dockerRegistry.url,
      username: dockerRegistry.username,
      project: dockerRegistry.project,
      registryType: dockerRegistry.registryType,
      createdAt: dockerRegistry.createdAt,
      updatedAt: dockerRegistry.updatedAt,
    }

    return dto
  }
}
