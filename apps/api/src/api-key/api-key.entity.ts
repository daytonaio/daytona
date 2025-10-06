/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, Entity, PrimaryColumn } from 'typeorm'
import { OrganizationResourcePermission } from '../organization/enums/organization-resource-permission.enum'

@Entity()
export class ApiKey {
  @PrimaryColumn({
    type: 'uuid',
  })
  organizationId: string

  @PrimaryColumn()
  userId: string

  @PrimaryColumn()
  name: string

  @Column({ unique: true, default: '' })
  keyHash: string

  @Column({
    default: '',
  })
  keyPrefix: string

  @Column({
    default: '',
  })
  keySuffix: string

  @Column({
    type: 'enum',
    enum: OrganizationResourcePermission,
    array: true,
  })
  permissions: OrganizationResourcePermission[]

  @Column()
  createdAt: Date

  @Column({ nullable: true, type: 'timestamp with time zone' })
  lastUsedAt: Date | null

  @Column({ nullable: true, type: 'timestamp with time zone' })
  expiresAt: Date | null

  constructor(createParams: {
    organizationId: string
    userId: string
    name: string
    keyHash: string
    keyPrefix: string
    keySuffix: string
    permissions: OrganizationResourcePermission[]
    expiresAt?: Date | null
  }) {
    this.organizationId = createParams.organizationId
    this.userId = createParams.userId
    this.name = createParams.name
    this.keyHash = createParams.keyHash
    this.keyPrefix = createParams.keyPrefix
    this.keySuffix = createParams.keySuffix
    this.permissions = createParams.permissions
    this.createdAt = new Date()
    this.expiresAt = createParams.expiresAt || null
    this.lastUsedAt = null
  }
}
