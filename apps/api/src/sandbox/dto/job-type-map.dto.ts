/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { JobType } from '../enums/job-type.enum'
import { ResourceType } from '../enums/resource-type.enum'

/**
 * Type-safe mapping between JobType and its corresponding ResourceType + Payload
 * This ensures compile-time safety when creating jobs
 */
export interface JobTypeMap {
  [JobType.CREATE_SANDBOX]: {
    resourceType: ResourceType.SANDBOX
  }
  [JobType.START_SANDBOX]: {
    resourceType: ResourceType.SANDBOX
  }
  [JobType.STOP_SANDBOX]: {
    resourceType: ResourceType.SANDBOX
  }
  [JobType.DESTROY_SANDBOX]: {
    resourceType: ResourceType.SANDBOX
  }
  [JobType.CREATE_BACKUP]: {
    resourceType: ResourceType.SANDBOX
  }
  [JobType.BUILD_SNAPSHOT]: {
    resourceType: ResourceType.SANDBOX
  }
  [JobType.PULL_SNAPSHOT]: {
    resourceType: ResourceType.SNAPSHOT
  }
  [JobType.REMOVE_SNAPSHOT]: {
    resourceType: ResourceType.SNAPSHOT
  }
  [JobType.UPDATE_SANDBOX_NETWORK_SETTINGS]: {
    resourceType: ResourceType.SANDBOX
  }
}

/**
 * Helper type to extract the resource type for a given JobType
 */
export type ResourceTypeForJobType<T extends JobType> = JobTypeMap[T]['resourceType']
