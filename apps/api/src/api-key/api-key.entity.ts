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

  @Column({ nullable: true })
  lastUsedAt?: Date

  @Column({ nullable: true })
  expiresAt?: Date
}
