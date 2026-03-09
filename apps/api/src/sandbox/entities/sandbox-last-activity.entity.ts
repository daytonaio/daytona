/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, Entity, JoinColumn, OneToOne, PrimaryColumn } from 'typeorm'
import { Sandbox } from './sandbox.entity'

@Entity('sandbox_last_activity')
export class SandboxLastActivity {
  @PrimaryColumn()
  sandboxId: string

  @Column({ nullable: true, type: 'timestamp with time zone' })
  lastActivityAt?: Date

  @OneToOne(() => Sandbox, { onDelete: 'CASCADE' })
  @JoinColumn({ name: 'sandboxId' })
  sandbox?: Sandbox
}
