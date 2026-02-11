/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, Index, PrimaryGeneratedColumn, UpdateDateColumn } from 'typeorm'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../constants/sandbox.constants'

@Entity()
@Index('warm_pool_find_idx', [
  'organizationId',
  'snapshot',
  'target',
  'class',
  'cpu',
  'mem',
  'disk',
  'gpu',
  'osUser',
  'env',
])
export class WarmPool {
  @PrimaryGeneratedColumn('uuid')
  id: string

  //  The organization this warm pool belongs to.
  //  If set to SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION, any organization can use this pool.
  //  If set to a specific organization ID, only that organization can use this pool.
  @Column({
    type: 'uuid',
    default: SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION,
  })
  organizationId: string

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

  @Column({
    type: 'enum',
    enum: SandboxClass,
    default: SandboxClass.SMALL,
  })
  class: SandboxClass

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
