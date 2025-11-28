/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, Index, PrimaryColumn, UpdateDateColumn } from 'typeorm'
import { nanoid } from 'nanoid'

@Entity()
@Index('region_organizationId_name_unique', ['organizationId', 'name'], {
  unique: true,
  where: '"organizationId" IS NOT NULL',
})
@Index('region_organizationId_null_name_unique', ['name'], {
  unique: true,
  where: '"organizationId" IS NULL',
})
export class Region {
  @PrimaryColumn()
  id: string

  @Column()
  name: string

  @Column({
    type: 'uuid',
    nullable: true,
  })
  organizationId: string | null

  @Column({ default: false })
  hidden: boolean

  @Column({
    type: 'boolean',
    default: true,
  })
  enforceQuotas: boolean

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date

  constructor(name: string, enforceQuotas: boolean, id?: string, organizationId?: string) {
    this.name = name
    this.enforceQuotas = enforceQuotas

    if (id) {
      this.id = id
    } else {
      this.id = nanoid(12)
    }

    if (organizationId) {
      this.organizationId = organizationId
    }
  }
}
