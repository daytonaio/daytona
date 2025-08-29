/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Column,
  CreateDateColumn,
  Entity,
  Generated,
  JoinColumn,
  ManyToOne,
  PrimaryColumn,
  UpdateDateColumn,
} from 'typeorm'
import { Sandbox } from './sandbox.entity'

@Entity()
export class SshAccess {
  @PrimaryColumn()
  @Generated('uuid')
  id: string

  @Column({
    type: 'uuid',
  })
  sandboxId: string

  @Column({
    type: 'text',
  })
  token: string

  @Column({
    type: 'timestamp',
  })
  expiresAt: Date

  @CreateDateColumn()
  createdAt: Date

  @UpdateDateColumn()
  updatedAt: Date

  @ManyToOne(() => Sandbox, { onDelete: 'CASCADE' })
  @JoinColumn({ name: 'sandboxId' })
  sandbox: Sandbox
}
