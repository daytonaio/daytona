/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateDateColumn, Column, Entity, JoinColumn, ManyToOne, PrimaryGeneratedColumn, Unique } from 'typeorm'
import { Sandbox } from './sandbox.entity'

@Entity()
@Unique(['childId'])
export class SandboxFork {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column()
  parentId: string

  @Column()
  childId: string

  @ManyToOne(() => Sandbox, { onDelete: 'CASCADE' })
  @JoinColumn({ name: 'parentId' })
  parent: Sandbox

  @ManyToOne(() => Sandbox, { onDelete: 'CASCADE' })
  @JoinColumn({ name: 'childId' })
  child: Sandbox

  @CreateDateColumn({ type: 'timestamp with time zone' })
  createdAt: Date
}
