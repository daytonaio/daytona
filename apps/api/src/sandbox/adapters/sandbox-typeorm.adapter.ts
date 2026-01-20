/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BadRequestException, Logger } from '@nestjs/common'
import { Repository, Brackets } from 'typeorm'
import { Sandbox } from '../entities/sandbox.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { SandboxSearchSortDirection, SandboxSearchSortField } from '../dto/search-sandboxes-query.dto'
import {
  SandboxSearchAdapter,
  SandboxSearchFilters,
  SandboxSearchPagination,
  SandboxSearchResult,
  SandboxSearchSort,
} from '../interfaces/sandbox-search.interface'
import { SandboxDto } from '../dto/sandbox.dto'

export class SandboxTypeormSearchAdapter implements SandboxSearchAdapter {
  private readonly logger = new Logger(SandboxTypeormSearchAdapter.name)

  constructor(private readonly sandboxRepository: Repository<Sandbox>) {}

  async search(params: {
    filters: SandboxSearchFilters
    pagination: SandboxSearchPagination
    sort: SandboxSearchSort
  }): Promise<SandboxSearchResult> {
    const { filters, pagination, sort } = params

    const sortField = this.getSortFieldMapping(sort.field)
    const sortDirection = this.getSortDirectionMapping(sort.direction)

    const qb = this.sandboxRepository.createQueryBuilder('sandbox')

    // Base filters
    qb.andWhere('sandbox.organizationId = :organizationId', { organizationId: filters.organizationId })

    if (filters.idPrefix) {
      qb.andWhere('LOWER(sandbox.id) LIKE LOWER(:idPrefix)', { idPrefix: `${filters.idPrefix}%` })
    }
    if (filters.namePrefix) {
      qb.andWhere('LOWER(sandbox.name) LIKE LOWER(:namePrefix)', { namePrefix: `${filters.namePrefix}%` })
    }
    if (filters.labels) {
      qb.andWhere('sandbox.labels @> :labels', { labels: filters.labels })
    }
    if (filters.snapshots?.length) {
      qb.andWhere('sandbox.snapshot IN (:...snapshots)', { snapshots: filters.snapshots })
    }
    if (filters.regionIds?.length) {
      qb.andWhere('sandbox.region IN (:...regionIds)', { regionIds: filters.regionIds })
    }
    if (filters.isPublic !== undefined) {
      qb.andWhere('sandbox.public = :isPublic', { isPublic: filters.isPublic })
    }
    if (filters.isRecoverable !== undefined) {
      qb.andWhere('sandbox.recoverable = :isRecoverable', { isRecoverable: filters.isRecoverable })
    }

    // Range filters
    if (filters.minCpu !== undefined) {
      qb.andWhere('sandbox.cpu >= :minCpu', { minCpu: filters.minCpu })
    }
    if (filters.maxCpu !== undefined) {
      qb.andWhere('sandbox.cpu <= :maxCpu', { maxCpu: filters.maxCpu })
    }
    if (filters.minMemoryGiB !== undefined) {
      qb.andWhere('sandbox.mem >= :minMemoryGiB', { minMemoryGiB: filters.minMemoryGiB })
    }
    if (filters.maxMemoryGiB !== undefined) {
      qb.andWhere('sandbox.mem <= :maxMemoryGiB', { maxMemoryGiB: filters.maxMemoryGiB })
    }
    if (filters.minDiskGiB !== undefined) {
      qb.andWhere('sandbox.disk >= :minDiskGiB', { minDiskGiB: filters.minDiskGiB })
    }
    if (filters.maxDiskGiB !== undefined) {
      qb.andWhere('sandbox.disk <= :maxDiskGiB', { maxDiskGiB: filters.maxDiskGiB })
    }
    if (filters.createdAtAfter) {
      qb.andWhere('sandbox.createdAt >= :createdAtAfter', { createdAtAfter: filters.createdAtAfter })
    }
    if (filters.createdAtBefore) {
      qb.andWhere('sandbox.createdAt <= :createdAtBefore', { createdAtBefore: filters.createdAtBefore })
    }
    if (filters.lastEventAfter) {
      qb.andWhere('sandbox.lastActivityAt >= :lastEventAfter', { lastEventAfter: filters.lastEventAfter })
    }
    if (filters.lastEventBefore) {
      qb.andWhere('sandbox.lastActivityAt <= :lastEventBefore', { lastEventBefore: filters.lastEventBefore })
    }

    // State filtering with error state handling
    const errorStates = [SandboxState.ERROR, SandboxState.BUILD_FAILED]
    const statesToInclude = (filters.states || Object.values(SandboxState)).filter(
      (state) => state !== SandboxState.DESTROYED,
    )
    const nonErrorStatesToInclude = statesToInclude.filter((state) => !errorStates.includes(state))
    const errorStatesToInclude = statesToInclude.filter((state) => errorStates.includes(state))

    if (nonErrorStatesToInclude.length > 0 || errorStatesToInclude.length > 0) {
      qb.andWhere(
        new Brackets((stateQb) => {
          if (nonErrorStatesToInclude.length > 0) {
            stateQb.orWhere('sandbox.state IN (:...nonErrorStates)', { nonErrorStates: nonErrorStatesToInclude })
          }
          if (errorStatesToInclude.length > 0) {
            if (filters.includeErroredDeleted) {
              stateQb.orWhere('sandbox.state IN (:...errorStates)', { errorStates: errorStatesToInclude })
            } else {
              stateQb.orWhere(
                new Brackets((errorQb) => {
                  errorQb
                    .where('sandbox.state IN (:...errorStates)', { errorStates: errorStatesToInclude })
                    .andWhere('sandbox.desiredState != :destroyedState', {
                      destroyedState: SandboxDesiredState.DESTROYED,
                    })
                }),
              )
            }
          }
        }),
      )
    }

    // Cursor-based pagination
    if (pagination.cursor) {
      const { sortValue, id } = this.decodeCursor(pagination.cursor)
      const op = sortDirection === 'ASC' ? '>' : '<'

      qb.andWhere(
        new Brackets((cursorQb) => {
          cursorQb.where(`sandbox.${sortField} ${op} :cursorSortValue`, { cursorSortValue: sortValue }).orWhere(
            new Brackets((tieBreaker) => {
              tieBreaker
                .where(`sandbox.${sortField} = :cursorSortValue`, { cursorSortValue: sortValue })
                .andWhere(`sandbox.id ${op} :cursorId`, { cursorId: id })
            }),
          )
        }),
      )
    }

    // Sorting with id as tie-breaker
    qb.orderBy(`sandbox.${sortField}`, sortDirection as 'ASC' | 'DESC', 'NULLS LAST')
    qb.addOrderBy('sandbox.id', sortDirection as 'ASC' | 'DESC')

    // Fetch one extra to determine if there are more items
    qb.take(pagination.limit + 1)

    const items = await qb.getMany()

    const hasMore = items.length > pagination.limit
    const returnItems = hasMore ? items.slice(0, pagination.limit) : items

    let nextCursor: string | null = null
    if (hasMore && returnItems.length > 0) {
      const lastItem = returnItems[returnItems.length - 1]
      nextCursor = this.encodeCursor(lastItem, sort.field)
    }

    return {
      items: returnItems.map((sandbox) => SandboxDto.fromSandbox(sandbox)),
      nextCursor,
    }
  }

  private getSortFieldMapping(sortField: SandboxSearchSortField): string {
    const fieldMapping: Record<SandboxSearchSortField, string> = {
      [SandboxSearchSortField.NAME]: 'name',
      [SandboxSearchSortField.CPU]: 'cpu',
      [SandboxSearchSortField.MEMORY]: 'mem',
      [SandboxSearchSortField.DISK]: 'disk',
      [SandboxSearchSortField.LAST_ACTIVITY_AT]: 'lastActivityAt',
      [SandboxSearchSortField.CREATED_AT]: 'createdAt',
    }
    return fieldMapping[sortField]
  }

  private getSortDirectionMapping(sortDirection: SandboxSearchSortDirection): string {
    const directionMapping: Record<SandboxSearchSortDirection, string> = {
      [SandboxSearchSortDirection.ASC]: 'ASC',
      [SandboxSearchSortDirection.DESC]: 'DESC',
    }
    return directionMapping[sortDirection]
  }

  private encodeCursor(sandbox: Sandbox, sortField: SandboxSearchSortField): string {
    const field = this.getSortFieldMapping(sortField)
    const sortValue = sandbox[field as keyof Sandbox]
    const cursorData = {
      sortValue: sortValue instanceof Date ? sortValue.toISOString() : sortValue,
      id: sandbox.id,
    }
    return Buffer.from(JSON.stringify(cursorData)).toString('base64')
  }

  private decodeCursor(cursor: string): { sortValue: any; id: string } {
    try {
      const decoded = JSON.parse(Buffer.from(cursor, 'base64').toString())
      return decoded
    } catch {
      throw new BadRequestException(`Invalid cursor provided: ${cursor}`)
    }
  }
}
