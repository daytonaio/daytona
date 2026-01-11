/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Column, Entity, PrimaryGeneratedColumn, CreateDateColumn, Index } from 'typeorm'

export enum DeviceAuthStatus {
  PENDING = 'pending',
  APPROVED = 'approved',
  DENIED = 'denied',
  EXPIRED = 'expired',
}

@Entity('device_authorization_request')
export class DeviceAuthorizationRequest {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column({ unique: true })
  @Index()
  deviceCode: string

  @Column({ unique: true })
  @Index()
  userCode: string

  @Column()
  clientId: string

  @Column({ nullable: true })
  scope: string

  @Column({
    type: 'enum',
    enum: DeviceAuthStatus,
    default: DeviceAuthStatus.PENDING,
  })
  status: DeviceAuthStatus

  @Column({ nullable: true })
  userId: string

  @Column({ nullable: true })
  organizationId: string

  @Column({ nullable: true })
  accessToken: string

  @CreateDateColumn()
  createdAt: Date

  @Column()
  expiresAt: Date

  @Column({ nullable: true })
  approvedAt: Date

  @Column({ nullable: true })
  lastPolledAt: Date
}
