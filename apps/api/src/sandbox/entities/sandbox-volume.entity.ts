/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, Index, PrimaryGeneratedColumn, Unique, UpdateDateColumn } from 'typeorm'

// SandboxVolumeMount tracks an attachment between a sandbox and a layered
// volume. Each row represents one mount inside one sandbox, holds the
// per-mount runtime parameters (path, subpath, read-only), and stores the
// encrypted per-attachment mount token minted from the volume's layered
// disk.
//
// Legacy s3fuse mounts are NOT stored here — they continue to live on the
// `sandbox.volumes` JSONB column. A sandbox is therefore either fully
// s3fuse (JSONB rows) or fully layered (sandbox_volume rows), never a mix.
@Entity({ name: 'sandbox_volume' })
@Unique(['sandboxId', 'volumeId', 'mountPath'])
@Index(['sandboxId'])
@Index(['volumeId'])
export class SandboxVolumeMount {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({ type: 'uuid' })
  sandboxId: string

  @Column({ type: 'uuid' })
  volumeId: string

  @Column()
  mountPath: string

  @Column({ nullable: true })
  subpath?: string

  @Column({ default: false })
  readOnly: boolean

  // Encrypted per-(sandbox, volume) mount token. Minted on first sandbox
  // start that needs this mount, revoked on sandbox destroy. Stored in the
  // shape produced by `EncryptionService.encrypt`; decrypt only at the
  // point of forwarding to the runner.
  @Column({ type: 'text', nullable: true })
  mountKeyEnc?: string | null

  @CreateDateColumn({ type: 'timestamp with time zone' })
  createdAt: Date

  @UpdateDateColumn({ type: 'timestamp with time zone' })
  updatedAt: Date
}
