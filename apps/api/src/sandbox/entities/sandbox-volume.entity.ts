/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, Index, PrimaryGeneratedColumn, Unique, UpdateDateColumn } from 'typeorm'

// One row per (sandbox, volume) layered mount: per-mount params (path,
// subpath, read-only) plus the encrypted mount token.
//
// s3fuse mounts are NOT stored here — they live on the `sandbox.volumes`
// JSONB column. A sandbox is all-s3fuse or all-layered, never mixed.
@Entity({ name: 'sandbox_volume' })
@Unique(['sandboxId', 'volumeId', 'mountPath'])
@Index(['sandboxId'])
@Index(['volumeId'])
export class SandboxVolumeMount {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({ type: 'character varying' })
  sandboxId: string

  @Column({ type: 'uuid' })
  volumeId: string

  @Column()
  mountPath: string

  @Column({ nullable: true })
  subpath?: string

  @Column({ default: false })
  readOnly: boolean

  // Encrypted per-(sandbox, volume) mount token. Minted on first start,
  // revoked on destroy. Decrypt only when forwarding to the runner.
  @Column({ type: 'text', nullable: true })
  mountKeyEnc?: string | null

  // Encrypted control-plane identifier returned when the token was minted;
  // required to revoke it.
  @Column({ type: 'text', nullable: true })
  mountIdentifierEnc?: string | null

  @CreateDateColumn({ type: 'timestamp with time zone' })
  createdAt: Date

  @UpdateDateColumn({ type: 'timestamp with time zone' })
  updatedAt: Date
}
