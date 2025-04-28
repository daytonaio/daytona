/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, PrimaryGeneratedColumn, UpdateDateColumn } from 'typeorm'
import { NodeRegion } from '../enums/node-region.enum'
import { WorkspaceClass } from '../enums/workspace-class.enum'
import { NodeState } from '../enums/node-state.enum'

@Entity()
export class Node {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({ unique: true })
  domain: string

  @Column()
  apiUrl: string

  @Column()
  apiKey: string

  @Column()
  cpu: number

  @Column()
  memory: number

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

  @Column({
    default: 0,
  })
  used: number

  @Column()
  capacity: number

  @Column({
    type: 'enum',
    enum: NodeRegion,
  })
  region: NodeRegion

  @Column({
    type: 'enum',
    enum: NodeState,
    default: NodeState.INITIALIZING,
  })
  state: NodeState

  @Column({
    nullable: true,
  })
  lastChecked: Date

  @Column({
    default: false,
  })
  unschedulable: boolean

  @CreateDateColumn()
  createdAt: Date

  @UpdateDateColumn()
  updatedAt: Date
}
