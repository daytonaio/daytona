/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { v4 } from 'uuid'
import { RegistryType } from './../../docker-registry/enums/registry-type.enum'
import { Column, CreateDateColumn, Entity, PrimaryGeneratedColumn, UpdateDateColumn } from 'typeorm'

@Entity()
export class DockerRegistry {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column()
  name: string

  @Column()
  url: string

  @Column()
  username: string

  @Column()
  password: string

  @Column({ default: false })
  isDefault: boolean

  @Column({ default: false })
  isFallback: boolean

  @Column({ default: '' })
  project: string

  @Column({ nullable: true, type: 'uuid' })
  organizationId: string | null

  @Column({ nullable: true, type: String })
  region: string | null

  @Column({
    type: 'enum',
    enum: RegistryType,
    default: RegistryType.INTERNAL,
  })
  registryType: RegistryType

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date

  constructor(createParams: {
    name: string
    url: string
    username: string
    password: string
    isDefault?: boolean
    isFallback?: boolean
    project?: string
    organizationId: string | null
    region?: string | null
    registryType?: RegistryType
  }) {
    this.id = v4()
    this.name = createParams.name
    this.url = createParams.url
    this.username = createParams.username
    this.password = createParams.password
    this.isDefault = createParams.isDefault ?? false
    this.isFallback = createParams.isFallback ?? false
    this.project = createParams.project ?? ''
    this.organizationId = createParams.organizationId
    this.region = createParams.region ?? null
    this.registryType = createParams.registryType ?? RegistryType.INTERNAL
    this.createdAt = new Date()
    this.updatedAt = new Date()
  }
}
