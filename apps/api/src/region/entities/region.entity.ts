/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, PrimaryColumn, Unique, UpdateDateColumn } from 'typeorm'
import { nanoid } from 'nanoid'

@Entity()
@Unique(['organizationId', 'name'])
export class Region {
  @PrimaryColumn()
  id: string

  @Column()
  name: string

  @Column({
    type: 'uuid',
  })
  organizationId: string

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date

  constructor(name: string, organizationId: string) {
    this.id = nanoid(12)
    this.name = name
    this.organizationId = organizationId
  }
}
