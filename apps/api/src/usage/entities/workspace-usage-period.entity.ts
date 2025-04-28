/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, Entity, PrimaryGeneratedColumn } from 'typeorm'

@Entity('workspace_usage_periods')
export class WorkspaceUsagePeriod {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column()
  workspaceId: string

  @Column()
  // Redundant property to optimize billing queries
  organizationId: string

  @Column({ type: 'timestamp' })
  startAt: Date

  @Column({ type: 'timestamp', nullable: true })
  endAt: Date | null

  @Column({ type: 'float' })
  cpu: number

  @Column({ type: 'float' })
  gpu: number

  @Column({ type: 'float' })
  mem: number

  @Column({ type: 'float' })
  disk: number

  @Column()
  region: string
}
