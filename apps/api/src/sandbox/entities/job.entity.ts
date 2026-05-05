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
import { parseProto } from '../proto/proto-registry'
import { JobStatus } from '../enums/job-status.enum'
import { JobType } from '../enums/job-type.enum'
import { ResourceType } from '../enums/resource-type.enum'
import { v4 } from 'uuid'

@Entity()
@Index(['runnerId', 'status'])
@Index(['status', 'createdAt'])
@Index(['resourceType', 'resourceId'])
@Index('IDX_UNIQUE_INCOMPLETE_JOB', ['resourceType', 'resourceId', 'runnerId'], {
  unique: true,
  where: `"completedAt" IS NULL AND "type" != '${JobType.CREATE_BACKUP}'`,
})
@Index('IDX_UNIQUE_INCOMPLETE_BACKUP_JOB', ['resourceType', 'resourceId', 'runnerId'], {
  unique: true,
  where: `"completedAt" IS NULL AND "type" = '${JobType.CREATE_BACKUP}'`,
})
// FIXME: Add this once https://github.com/typeorm/typeorm/issues/11714 is resolved
// @Check(
//   'VALIDATE_JOB_TYPE',
//   `"type" IN (${Object.values(JobType)
//     .map((v) => `'${v}'`)
//     .join(', ')})`,
// )
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
  })
  payloadType: string | null

  @Column({
    nullable: true,
  })
  resultType: string | null

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
    payloadType?: string | null
    resultType?: string | null
    traceContext?: Record<string, string> | null
    errorMessage?: string | null
    startedAt?: Date | null
    completedAt?: Date | null
  }) {
    this.id = params.id || v4()
    this.version = 1
    this.type = params.type
    this.status = params.status || JobStatus.PENDING
    this.runnerId = params.runnerId
    this.resourceType = params.resourceType
    this.resourceId = params.resourceId
    this.payload = params.payload || null
    this.payloadType = params.payloadType || null
    this.resultType = params.resultType || null
    this.traceContext = params.traceContext || null
    this.errorMessage = params.errorMessage || null
    this.startedAt = params.startedAt || null
    this.completedAt = params.completedAt || null
    this.createdAt = new Date()
    this.updatedAt = new Date()
  }

  getPayload<T = Record<string, any>>(): T | null {
    if (!this.payload) {
      return null
    }

    try {
      const parsed = JSON.parse(this.payload)
      if (this.payloadType) {
        return (parseProto(this.payloadType, parsed) ?? parsed) as T
      }
      return parsed
    } catch {
      return null
    }
  }

  getResultMetadata(): Record<string, any> | null {
    if (!this.resultMetadata) {
      return null
    }

    try {
      const parsed = JSON.parse(this.resultMetadata)
      if (this.resultType) {
        return parseProto(this.resultType, parsed) ?? parsed
      }
      return parsed
    } catch {
      return null
    }
  }
}
