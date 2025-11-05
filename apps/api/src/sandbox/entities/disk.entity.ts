/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, PrimaryGeneratedColumn, Unique, UpdateDateColumn } from 'typeorm'
import { DiskState } from '../enums/disk-state.enum'

@Entity()
@Unique(['organizationId', 'name'])
export class Disk {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({
    type: 'uuid',
  })
  organizationId: string

  @Column()
  name: string

  @Column({
    type: 'int',
  })
  size: number

  @Column({
    type: 'enum',
    enum: DiskState,
    default: DiskState.FRESH,
  })
  state: DiskState

  @Column({
    type: 'uuid',
    nullable: true,
  })
  baseDiskId?: string

  @Column({
    type: 'uuid',
    nullable: true,
  })
  runnerId?: string

  @Column({
    type: 'uuid',
    nullable: true,
  })
  sandboxId?: string

  @Column({ nullable: true })
  errorReason?: string

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date
}
