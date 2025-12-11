/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Column,
  CreateDateColumn,
  Entity,
  Index,
  PrimaryGeneratedColumn,
  UpdateDateColumn,
  VersionColumn,
} from 'typeorm'
import { JobStatus } from '../enums/job-status.enum'
import { JobType } from '../enums/job-type.enum'
import { ResourceType } from '../enums/resource-type.enum'

@Entity()
@Index(['runnerId', 'status'])
@Index(['status', 'createdAt'])
@Index(['resourceType', 'resourceId'])
export class Job {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @VersionColumn()
  version: number

  @Column({
    type: 'enum',
    enum: JobType,
  })
  type: JobType

  @Column({
    type: 'enum',
    enum: JobStatus,
    default: JobStatus.PENDING,
  })
  @Index()
  status: JobStatus

  @Column()
  @Index()
  runnerId: string

  @Column({
    type: 'enum',
    enum: ResourceType,
    nullable: true,
  })
  resourceType: ResourceType

  @Column({
    nullable: true,
  })
  resourceId: string

  @Column({
    type: 'jsonb',
    nullable: true,
  })
  payload: Record<string, any>

  @Column({
    type: 'jsonb',
    nullable: true,
  })
  traceContext: Record<string, string>

  @Column({
    nullable: true,
    type: 'text',
  })
  errorMessage: string

  @Column({
    nullable: true,
    type: 'timestamp with time zone',
  })
  startedAt: Date

  @Column({
    nullable: true,
    type: 'timestamp with time zone',
  })
  completedAt: Date

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date
}
