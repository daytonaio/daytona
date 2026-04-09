/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Entity, PrimaryGeneratedColumn, Column, CreateDateColumn, Index } from 'typeorm'
import { v4 } from 'uuid'

export type AuditLogMetadata = Record<string, any>

@Entity()
@Index(['createdAt'])
@Index(['organizationId', 'createdAt'])
export class AuditLog {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column()
  actorId: string

  @Column({
    default: '',
  })
  actorEmail: string

  @Column({ nullable: true })
  actorApiKeyPrefix?: string

  @Column({ nullable: true })
  actorApiKeySuffix?: string

  @Column({ nullable: true })
  organizationId?: string

  @Column()
  action: string

  @Column({ nullable: true })
  targetType?: string

  @Column({ nullable: true })
  targetId?: string

  @Column({ nullable: true })
  statusCode?: number

  @Column({ nullable: true })
  errorMessage?: string

  @Column({ nullable: true })
  ipAddress?: string

  @Column({ type: 'text', nullable: true })
  userAgent?: string

  @Column({ nullable: true })
  source?: string

  @Column({ type: 'jsonb', nullable: true })
  metadata?: AuditLogMetadata

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  constructor(params: {
    id?: string
    actorId: string
    actorEmail: string
    actorApiKeyPrefix?: string
    actorApiKeySuffix?: string
    organizationId?: string
    action: string
    targetType?: string
    targetId?: string
    statusCode?: number
    errorMessage?: string
    ipAddress?: string
    userAgent?: string
    source?: string
    metadata?: AuditLogMetadata
    createdAt?: Date
  }) {
    this.id = params.id || v4()
    this.actorId = params.actorId
    this.actorEmail = params.actorEmail
    this.actorApiKeyPrefix = params.actorApiKeyPrefix
    this.actorApiKeySuffix = params.actorApiKeySuffix
    this.organizationId = params.organizationId
    this.action = params.action
    this.targetType = params.targetType
    this.targetId = params.targetId
    this.statusCode = params.statusCode
    this.errorMessage = params.errorMessage
    this.ipAddress = params.ipAddress
    this.userAgent = params.userAgent
    this.source = params.source
    this.metadata = params.metadata
    this.createdAt = params.createdAt || new Date()
  }
}
