/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, PrimaryGeneratedColumn, UpdateDateColumn } from 'typeorm'
import { NodeRegion } from '../enums/node-region.enum'
import { WorkspaceClass } from '../enums/workspace-class.enum'

@Entity()
export class WarmPool {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column()
  pool: number

  @Column()
  image: string

  @Column({
    type: 'enum',
    enum: NodeRegion,
    default: NodeRegion.EU,
  })
  target: NodeRegion

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
    enum: WorkspaceClass,
    default: WorkspaceClass.SMALL,
  })
  class: WorkspaceClass

  @Column()
  osUser: string

  @Column({ nullable: true })
  errorReason?: string

  @Column({
    type: 'simple-json',
    default: {},
  })
  env: { [key: string]: string }

  @CreateDateColumn()
  createdAt: Date

  @UpdateDateColumn()
  updatedAt: Date
}
