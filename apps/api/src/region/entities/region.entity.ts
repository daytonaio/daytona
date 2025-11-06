/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, PrimaryColumn, Unique, UpdateDateColumn } from 'typeorm'
import { nanoid } from 'nanoid'

export const REGION_NAME_REGEX = /^[a-zA-Z0-9_.-]+$/

@Entity()
@Unique(['organizationId', 'name'])
export class Region {
  @PrimaryColumn()
  id: string

  @Column()
  name: string

  @Column({
    type: 'uuid',
    nullable: true,
  })
  organizationId?: string

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date

  constructor(name: string, id?: string, organizationId?: string) {
    this.name = name
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
