/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Entity, PrimaryGeneratedColumn, Column, CreateDateColumn, Index } from 'typeorm'

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
}
