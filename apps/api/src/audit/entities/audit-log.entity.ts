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
  organizationId: string | null

  @Column()
  action: string

  @Column({ nullable: true })
  targetType: string | null

  @Column({ nullable: true })
  targetId: string | null

  @Column({ nullable: true })
  statusCode: number | null

  @Column({ nullable: true })
  errorMessage: string | null

  @Column({ nullable: true })
  ipAddress: string | null

  @Column({ type: 'text', nullable: true })
  userAgent: string | null

  @Column({ nullable: true })
  source: string | null

  @Column({ type: 'jsonb', nullable: true })
  metadata: AuditLogMetadata | null

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  constructor(createParams: {
    actorId: string
    actorEmail: string
    organizationId?: string | null
    action: string
    targetType?: string | null
    targetId?: string | null
    statusCode?: number | null
    errorMessage?: string | null
    ipAddress?: string | null
    userAgent?: string | null
    source?: string | null
    metadata?: AuditLogMetadata | null
  }) {
    this.id = v4()
    this.actorId = createParams.actorId
    this.actorEmail = createParams.actorEmail
    this.organizationId = createParams.organizationId ?? null
    this.action = createParams.action
    this.targetType = createParams.targetType ?? null
    this.targetId = createParams.targetId ?? null
    this.statusCode = createParams.statusCode ?? null
    this.errorMessage = createParams.errorMessage ?? null
    this.ipAddress = createParams.ipAddress ?? null
    this.userAgent = createParams.userAgent ?? null
    this.source = createParams.source ?? null
    this.metadata = createParams.metadata ?? null

    this.createdAt = new Date()
  }
}
