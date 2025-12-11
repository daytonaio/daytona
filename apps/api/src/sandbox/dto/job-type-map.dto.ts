/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { JobType } from '../enums/job-type.enum'
import { ResourceType } from '../enums/resource-type.enum'
import { SandboxJobPayload, SnapshotJobPayload, BackupJobPayload } from './job.dto'

/**
 * Type-safe mapping between JobType and its corresponding ResourceType + Payload
 * This ensures compile-time safety when creating jobs
 */
export interface JobTypeMap {
  [JobType.CREATE_SANDBOX]: {
    resourceType: ResourceType.SANDBOX
    payload: SandboxJobPayload
  }
  [JobType.START_SANDBOX]: {
    resourceType: ResourceType.SANDBOX
    payload?: Pick<SandboxJobPayload, 'metadata'> // Optional metadata for start
  }
  [JobType.STOP_SANDBOX]: {
    resourceType: ResourceType.SANDBOX
    payload?: undefined // No payload needed for stop
  }
  [JobType.DESTROY_SANDBOX]: {
    resourceType: ResourceType.SANDBOX
    payload?: undefined // No payload needed for destroy
  }
  [JobType.CREATE_BACKUP]: {
    resourceType: ResourceType.SANDBOX
    payload: BackupJobPayload
  }
  [JobType.BUILD_SNAPSHOT]: {
    resourceType: ResourceType.SANDBOX
    payload: SnapshotJobPayload
  }
  [JobType.PULL_SNAPSHOT]: {
    resourceType: ResourceType.SNAPSHOT
    payload: SnapshotJobPayload
  }
  [JobType.REMOVE_SNAPSHOT]: {
    resourceType: ResourceType.SNAPSHOT
    payload?: undefined // No payload needed for remove
  }
}

/**
 * Helper type to extract the payload type for a given JobType
 */
export type PayloadForJobType<T extends JobType> = JobTypeMap[T]['payload']

/**
 * Helper type to extract the resource type for a given JobType
 */
export type ResourceTypeForJobType<T extends JobType> = JobTypeMap[T]['resourceType']
