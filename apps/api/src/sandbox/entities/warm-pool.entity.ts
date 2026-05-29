/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, Index, PrimaryGeneratedColumn, UpdateDateColumn } from 'typeorm'

@Entity()
@Index('warm_pool_find_idx', ['snapshot', 'target', 'cpu', 'mem', 'disk', 'gpu', 'osUser', 'env'])
export class WarmPool {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column()
  pool: number

  @Column()
  snapshot: string

  @Column()
  target: string

  @Column()
  cpu: number

  @Column()
  mem: number

  @Column()
  disk: number

  @Column()
  gpu: number

  @Column()
  gpuType: string

  @Column()
  osUser: string

  @Column({ nullable: true })
  errorReason?: string

  @Column({
    type: 'simple-json',
    default: {},
  })
  env: { [key: string]: string }

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date
}
