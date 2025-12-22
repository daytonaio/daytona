/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Check,
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
import { v4 } from 'uuid'

@Entity()
@Index(['runnerId', 'status'])
@Index(['status', 'createdAt'])
@Index(['resourceType', 'resourceId'])
@Index('IDX_UNIQUE_INCOMPLETE_JOB', ['resourceType', 'resourceId'], {
  unique: true,
  where: '"completedAt" IS NULL',
})
@Check(
  'VALIDATE_JOB_TYPE',
  `"type" IN (${Object.values(JobType)
    .map((v) => `'${v}'`)
    .join(', ')})`,
)
export class Job {
  @PrimaryGeneratedColumn('uuid')
  id: string

  @VersionColumn()
  version: number

  @Column({
    type: 'character varying',
  })
  type: JobType

  @Column({
    type: 'enum',
    enum: JobStatus,
    default: JobStatus.PENDING,
  })
  status: JobStatus

  @Column()
  runnerId: string

  @Column({
    type: 'enum',
    enum: ResourceType,
  })
  resourceType: ResourceType

  @Column()
  resourceId: string

  @Column({
    nullable: true,
  })
  payload: string | null

  @Column({
    nullable: true,
  })
  resultMetadata: string | null

  @Column({
    type: 'jsonb',
    nullable: true,
  })
  traceContext: Record<string, string> | null

  @Column({
    nullable: true,
    type: 'text',
  })
  errorMessage: string | null

  @Column({
    nullable: true,
    type: 'timestamp with time zone',
  })
  startedAt: Date | null

  @Column({
    nullable: true,
    type: 'timestamp with time zone',
  })
  completedAt: Date | null

  @CreateDateColumn({
    type: 'timestamp with time zone',
  })
  createdAt: Date

  @UpdateDateColumn({
    type: 'timestamp with time zone',
  })
  updatedAt: Date

  constructor(params: {
    id?: string
    type: JobType
    status?: JobStatus
    runnerId: string
    resourceType: ResourceType
    resourceId: string
    payload?: string | null
    traceContext?: Record<string, string> | null
    errorMessage?: string | null
    startedAt?: Date | null
    completedAt?: Date | null
  }) {
    this.id = params.id || v4()
    this.type = params.type
    this.status = params.status || JobStatus.PENDING
    this.runnerId = params.runnerId
    this.resourceType = params.resourceType
    this.resourceId = params.resourceId
    this.payload = params.payload || null
    this.traceContext = params.traceContext || null
    this.errorMessage = params.errorMessage || null
    this.startedAt = params.startedAt || null
    this.completedAt = params.completedAt || null
    this.createdAt = new Date()
    this.updatedAt = new Date()
  }

  getResultMetadata(): Record<string, string> | null {
    if (!this.resultMetadata) {
      return null
    }

    try {
      return JSON.parse(this.resultMetadata)
    } catch {
      return null
    }
  }
}
