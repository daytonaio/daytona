/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CreateDateColumn, Entity, JoinColumn, ManyToOne, PrimaryColumn, UpdateDateColumn } from 'typeorm'
import { Snapshot } from './snapshot.entity'
import { Region } from '../../region/entities/region.entity'

@Entity()
export class SnapshotRegion {
  @PrimaryColumn('uuid')
  snapshotId: string

  @PrimaryColumn()
  regionId: string

  @ManyToOne(() => Snapshot, (snapshot) => snapshot.snapshotRegions, {
    onDelete: 'CASCADE',
  })
  @JoinColumn({ name: 'snapshotId' })
  snapshot: Snapshot

  @ManyToOne(() => Region, {
    onDelete: 'CASCADE',
  })
  @JoinColumn({ name: 'regionId' })
  region: Region

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date
}
