/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxState } from '../enums/sandbox-state.enum'
import { SearchSandboxesResultDto } from '../dto/search-sandboxes-result.dto'
import { SandboxSearchSortField, SandboxSearchSortDirection } from '../dto/search-sandboxes-query.dto'

export interface SandboxSearchFilters {
  /**
   * Filter by organization ID
   */
  organizationId: string
  /**
   * Filter by ID prefix (case-insensitive)
   */
  idPrefix?: string
  /**
   * Filter by name prefix (case-insensitive)
   */
  namePrefix?: string
  /**
   * Filter by labels
   */
  labels?: { [key: string]: string }
  /**
   * Include results with errored state and deleted desired state
   */
  includeErroredDeleted?: boolean
  /**
   * Filter by states
   */
  states?: SandboxState[]
  /**
   * Filter by snapshots
   */
  snapshots?: string[]
  /**
   * Filter by region IDs
   */
  regionIds?: string[]
  /**
   * Filter by minimum CPU
   */
  minCpu?: number
  /**
   * Filter by maximum CPU
   */
  maxCpu?: number
  /**
   * Filter by minimum memory in GiB
   */
  minMemoryGiB?: number
  /**
   * Filter by maximum memory in GiB
   */
  maxMemoryGiB?: number
  /**
   * Filter by minimum disk space in GiB
   */
  minDiskGiB?: number
  /**
   * Filter by maximum disk space in GiB
   */
  maxDiskGiB?: number
  /**
   * Filter by public status
   */
  isPublic?: boolean
  /**
   * Filter by recoverable status
   */
  isRecoverable?: boolean
  /**
   * Filter by created after this timestamp
   */
  createdAtAfter?: Date
  /**
   * Filter by created before this timestamp
   */
  createdAtBefore?: Date
  /**
   * Filter by last event after this timestamp
   */
  lastEventAfter?: Date
  /**
   * Filter by last event before this timestamp
   */
  lastEventBefore?: Date
}

export interface SandboxSearchPagination {
  /**
   * Number of results per page
   */
  limit: number
  /**
   * Pagination cursor from a previous response
   */
  cursor?: string
}

export interface SandboxSearchSort {
  /**
   * Field to sort by
   */
  field?: SandboxSearchSortField
  /**
   * Direction to sort by
   */
  direction?: SandboxSearchSortDirection
}

export type SandboxSearchResult = SearchSandboxesResultDto

/**
 * Interface for sandbox search operations
 * Provides search functionality for sandboxes with filtering and cursor-based pagination
 */
export interface SandboxSearchAdapter {
  /**
   * Search sandboxes for an organization
   * @param params - Search parameters
   * @param params.filters - Filters to apply
   * @param params.pagination - Pagination parameters
   * @param params.sort - Sort parameters
   * @returns Paginated search results with cursor for next page
   */
  search(params: {
    filters: SandboxSearchFilters
    pagination: SandboxSearchPagination
    sort: SandboxSearchSort
  }): Promise<SandboxSearchResult>
}
