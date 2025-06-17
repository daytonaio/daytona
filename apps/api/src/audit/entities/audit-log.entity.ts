/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Entity, PrimaryGeneratedColumn, Column, CreateDateColumn, Index } from 'typeorm'

@Entity()
@Index(['createdAt'])
@Index(['userId', 'createdAt'])
@Index(['organizationId', 'createdAt'])
@Index(['organizationId', 'userId', 'createdAt'])
@Index(['targetId', 'createdAt'])
export class AuditLog {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @Column()
  userId: string

  @Column()
  userEmail: string

  @Column({ nullable: true })
  organizationId?: string

  @Column()
  action: string

  @Column({ nullable: true })
  targetType?: string

  @Column({ nullable: true })
  targetId?: string

  @Column()
  outcome: string

  @Column({ nullable: true })
  errorMessage?: string

  @Column({ nullable: true })
  ipAddress?: string

  @Column({ type: 'text', nullable: true })
  userAgent?: string

  @Column({ nullable: true })
  source?: string

  @CreateDateColumn()
  createdAt: Date
}
