/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, Entity, JoinColumn, OneToOne, PrimaryColumn } from 'typeorm'
import { nanoid } from 'nanoid'
import { Sandbox } from './sandbox.entity'

@Entity('sandbox_metadata')
export class SandboxMetadata {
  @PrimaryColumn()
  sandboxId: string

  // General-purpose HMAC key for signed sandbox URLs (currently pre-signed file
  // URLs). Stable across start/stop so signatures survive restarts; rotated only
  // via the dedicated rotate endpoint, which invalidates every prior signature.
  // Each signing purpose MUST use a distinct domain label in its canonical string
  // (e.g. "v1:files:...") so signatures can never be replayed across purposes.
  @Column({ type: 'character varying' })
  signingKey: string = nanoid(32)

  @OneToOne(() => Sandbox, { onDelete: 'CASCADE' })
  @JoinColumn({ name: 'sandboxId' })
  sandbox?: Sandbox
}
