/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, CreateDateColumn, Entity, PrimaryGeneratedColumn } from 'typeorm'

export enum DeviceAuthorizationStatus {
  PENDING = 'pending',
  APPROVED = 'approved',
  DENIED = 'denied',
  EXPIRED = 'expired',
}

@Entity()
export class DeviceAuthorization {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({ unique: true, length: 128 })
  deviceCode: string

  @Column({ unique: true, length: 16 })
  userCode: string

  @Column({ length: 128 })
  clientId: string

  @Column({ type: 'text', nullable: true })
  scope: string

  @Column({
    type: 'varchar',
    length: 32,
    default: DeviceAuthorizationStatus.PENDING,
  })
  status: DeviceAuthorizationStatus

  @Column({ type: 'uuid', nullable: true })
  userId: string

  @Column({ type: 'uuid', nullable: true })
  organizationId: string

  @CreateDateColumn()
  createdAt: Date

  @Column({ type: 'timestamp' })
  expiresAt: Date

  @Column({ type: 'timestamp', nullable: true })
  approvedAt: Date

  @Column({ type: 'timestamp', nullable: true })
  lastPollAt: Date
}
