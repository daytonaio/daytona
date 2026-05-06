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

  // Mount strategy currently used for this volume's data. Defaulted at
  // create time from the organization's `defaultVolumeBackend` and can
  // later be switched in place via `PUT /volumes/{id}/backend` (gated on
  // the `volume_backend_picker` feature flag). The backing S3 bucket is
  // identical across both backends — only the mounting layer changes.
  //
  //  - 's3fuse'       host-side mount on the runner backed by an S3 bucket
  //                   provisioned by the API in the configured object store.
  //  - 'experimental' in-container mount on the runner backed by an Archil
  //                   disk that mounts the *same* S3 bucket. The disk's
  //                   mount token is stored encrypted in
  //                   `archilMountTokenEnc` and forwarded to the daemon via
  //                   env vars at sandbox start time.
  //
  // Note: the `archilDiskId` / `archilRegion` / `archilMountTokenEnc`
  // columns are populated only while the volume is on `experimental` and
  // are cleared when migrating back to `s3fuse`.
  @Column({
    default: 's3fuse',
  })
  backend: string

  // Archil disk ID (e.g. "dsk-0123456789abcdef") for backend = 'experimental'.
  @Column({ nullable: true })
  archilDiskId?: string

  // Archil region the disk lives in (e.g. "aws-us-east-1") for backend =
  // 'experimental'. Used to route control-plane requests and to set the
  // `--region` flag passed to `archil mount` inside the sandbox.
  @Column({ nullable: true })
  archilRegion?: string

  // Per-disk Archil mount token, encrypted with `EncryptionService`. This is
  // the only credential needed to mount the disk and grants full read/write
  // access to it; never log or return it through the API. Decrypt at sandbox
  // start time, forward via env var to the runner, scrub after use.
  @Column({ nullable: true })
  archilMountTokenEnc?: string

  public getBucketName(): string {
    return `daytona-volume-${this.id}`
  }

  public getArchilDiskName(): string {
    // Stable, idempotent disk name derived from the Daytona volume ID.
    // Archil disk names allow [a-zA-Z0-9_-]+ up to 100 chars; the volume's
    // UUID fits comfortably and gives us 1:1 mapping for safe retries.
    return `daytona-vol-${this.id}`
  }
}
