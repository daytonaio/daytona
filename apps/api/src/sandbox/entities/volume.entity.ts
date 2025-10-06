/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, PrimaryGeneratedColumn, Unique, UpdateDateColumn } from 'typeorm'
import { VolumeState } from '../enums/volume-state.enum'
import { v4 } from 'uuid'

@Entity()
@Unique(['organizationId', 'name'])
export class Volume {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({
    nullable: true,
    type: 'uuid',
  })
  organizationId: string | null

  @Column()
  name: string

  @Column({
    type: 'enum',
    enum: VolumeState,
    default: VolumeState.PENDING_CREATE,
  })
  state: VolumeState

  @Column({ nullable: true, type: String })
  errorReason: string | null

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

  constructor(params: { organizationId: string | null; name?: string }) {
    this.id = v4()
    this.organizationId = params.organizationId
    this.name = params.name || this.id
    this.state = VolumeState.PENDING_CREATE
    this.errorReason = null
    this.createdAt = new Date()
    this.updatedAt = new Date()
  }

  public getBucketName(): string {
    return `daytona-volume-${this.id}`
  }
}
