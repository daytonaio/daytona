/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, PrimaryGeneratedColumn, Unique, UpdateDateColumn } from 'typeorm'
import { VolumeState } from '../enums/volume-state.enum'

@Entity()
@Unique(['organizationId', 'name'])
export class Volume {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({
    nullable: true,
    type: 'uuid',
  })
  organizationId?: string

  @Column()
  name: string

  @Column({
    type: 'enum',
    enum: VolumeState,
    default: VolumeState.PENDING_CREATE,
  })
  state: VolumeState

  @Column({ nullable: true })
  errorReason?: string

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date

  @Column({ nullable: true })
  lastUsedAt?: Date

  // Storage backend, locked at create time (the two layouts are
  // incompatible, so it never switches in place).
  //  - 's3fuse'  host-side runner mount, one S3 bucket per volume.
  //  - 'layered' in-container mount; a `<volumeId>/` prefix of the per-org
  //              bucket, with mount tokens on the `sandbox_volume` table.
  @Column({
    default: 's3fuse',
  })
  backend: string

  @Column({ type: 'float', nullable: true })
  currentStorageMb?: number

  @Column({ type: 'timestamp with time zone', nullable: true })
  storageCheckedAt?: Date

  // Layered disk ID (e.g. "dsk-0123456789abcdef") for backend = 'layered'.
  @Column({ nullable: true })
  layeredDiskId?: string

  // Region the layered disk lives in (e.g. "aws-us-east-1"); routes
  // control-plane mount-token requests.
  @Column({ nullable: true })
  layeredRegion?: string

  // Daytona Region.id the volume is pinned to (layered only). NULL for
  // s3fuse and for legacy layered volumes that pre-date region pinning.
  @Column({ nullable: true })
  regionId?: string | null

  // s3fuse bucket name; one bucket per volume.
  public getBucketName(): string {
    return `daytona-volume-${this.id}`
  }

  // Prefix inside the per-org layered bucket. Trailing slash so list/delete
  // match only this volume's namespace, not siblings.
  public getLayeredBucketPrefix(): string {
    return `${this.id}/`
  }

  // Stable disk name derived from the volume ID for idempotent retries.
  public getLayeredDiskName(): string {
    return `daytona-vol-${this.id}`
  }
}
