/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, Index, PrimaryGeneratedColumn, Unique } from 'typeorm'

/**
 * Tracks per-(volume, subpath) Archil disks. Each disk is created with a
 * `bucketPrefix` matching the subpath so its mount token inherently grants
 * access ONLY to that prefix in S3 — preventing cross-subpath data leakage
 * when multiple sandboxes share the same volume with different subpaths.
 *
 * A NULL subpath represents a "root" disk that covers the entire bucket
 * (no prefix restriction). This is used when a sandbox mounts a volume
 * without specifying a subpath.
 */
@Entity('volume_disk')
@Unique(['volumeId', 'subpath'])
@Index('idx_volume_disk_volume_id', ['volumeId'])
export class VolumeDisk {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({ type: 'uuid' })
  volumeId: string

  // NULL means "root" / entire bucket (no bucketPrefix).
  // Non-null values correspond to the S3 prefix the Archil disk is scoped to.
  @Column({ type: 'varchar', nullable: true })
  subpath: string | null

  @Column()
  archilDiskId: string

  @Column()
  archilRegion: string

  // Encrypted with EncryptionService. Decrypt at sandbox start time only.
  @Column()
  archilMountTokenEnc: string

  @CreateDateColumn({ type: 'timestamp with time zone' })
  createdAt: Date
}
