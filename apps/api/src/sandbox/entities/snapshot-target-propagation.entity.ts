/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Column,
  CreateDateColumn,
  Entity,
  JoinColumn,
  ManyToOne,
  PrimaryGeneratedColumn,
  UpdateDateColumn,
} from 'typeorm'
import { Min } from 'class-validator'

@Entity()
export class SnapshotTargetPropagation {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column()
  target: string

  @Column({ type: 'uuid' })
  snapshotId: string

  // TODO: consider adding snapshot ref to simplify queries
  // @Column({ nullable: true })
  // ref?: string

  @Column({ type: 'int' })
  @Min(0, { message: 'Desired concurrent sandboxes cannot be negative' })
  desiredConcurrentSandboxes: number

  @Column({ type: 'int', default: 0 })
  @Min(0, { message: 'User minimum cannot be negative' })
  userOverride: number

  @ManyToOne('Snapshot', 'targetPropagations')
  @JoinColumn({ name: 'snapshotId' })
  snapshot: any

  @CreateDateColumn()
  createdAt: Date

  @UpdateDateColumn()
  updatedAt: Date
}
