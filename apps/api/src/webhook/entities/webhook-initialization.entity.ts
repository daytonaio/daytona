/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Entity, PrimaryColumn, CreateDateColumn, UpdateDateColumn, Column } from 'typeorm'

@Entity()
export class WebhookInitialization {
  @PrimaryColumn()
  organizationId: string

  @Column({
    nullable: true,
  })
  svixApplicationId?: string

  @Column({
    type: 'text',
    nullable: true,
  })
  lastError?: string

  @Column({
    type: 'int',
    default: 0,
  })
  retryCount: number

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date
}
