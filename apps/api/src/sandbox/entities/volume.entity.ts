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

  // Mount strategy backing this volume's data. Locked at create time from
  // the organization's `defaultVolumeBackend` and never switched in place
  // (the two backends use incompatible storage layouts).
  //
  //  - 's3fuse'  Host-side mount on the runner backed by a dedicated S3
  //              bucket per volume (`daytona-volume-<id>`).
  //  - 'layered' In-container mount backed by a single per-organization
  //              S3 bucket where this volume is a `<volumeId>/` prefix.
  //              The control plane creates one disk per volume scoped to
  //              that prefix, and the per-sandbox mount tokens are stored
  //              on the `sandbox_volume` table — never on this row.
  @Column({
    default: 's3fuse',
  })
  backend: string

  // Layered disk ID (e.g. "dsk-0123456789abcdef") for backend = 'layered'.
  @Column({ nullable: true })
  layeredDiskId?: string

  // Region the layered disk lives in (e.g. "aws-us-east-1") for backend =
  // 'layered'. Used to route control-plane requests for adding/removing
  // per-sandbox mount tokens.
  @Column({ nullable: true })
  layeredRegion?: string

  // Bucket name used when backend = 's3fuse'. Each s3fuse volume gets its
  // own bucket.
  public getBucketName(): string {
    return `daytona-volume-${this.id}`
  }

  // Bucket prefix used inside the per-organization layered bucket when
  // backend = 'layered'. Trailing slash so S3 list/delete operations match
  // exactly the volume's namespace and not sibling volumes.
  public getLayeredBucketPrefix(): string {
    return `${this.id}/`
  }

  // Stable, idempotent disk name derived from the Daytona volume ID.
  // Disk names allow [a-zA-Z0-9_-]+ up to 100 chars; the volume's UUID
  // fits comfortably and gives us 1:1 mapping for safe retries.
  public getLayeredDiskName(): string {
    return `daytona-vol-${this.id}`
  }
}
