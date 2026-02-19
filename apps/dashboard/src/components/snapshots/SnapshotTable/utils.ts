/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DEFAULT_SNAPSHOT_SORTING, SnapshotSorting } from '@/hooks/queries/useSnapshotsQuery'
import { GetAllSnapshotsOrderEnum, GetAllSnapshotsSortEnum } from '@daytonaio/api-client'
import { SortingState } from '@tanstack/react-table'

export const convertApiSortingToTableSorting = (sorting: SnapshotSorting): SortingState => {
  let id: string
  switch (sorting.field) {
    case GetAllSnapshotsSortEnum.NAME:
      id = 'name'
      break
    case GetAllSnapshotsSortEnum.STATE:
      id = 'state'
      break
    case GetAllSnapshotsSortEnum.CREATED_AT:
      id = 'createdAt'
      break
    case GetAllSnapshotsSortEnum.LAST_USED_AT:
    default:
      id = 'lastUsedAt'
      break
  }

  return [{ id, desc: sorting.direction === GetAllSnapshotsOrderEnum.DESC }]
}

export const convertTableSortingToApiSorting = (sorting: SortingState): SnapshotSorting => {
  if (!sorting.length) {
    return DEFAULT_SNAPSHOT_SORTING
  }

  const sort = sorting[0]
  let field: GetAllSnapshotsSortEnum

  switch (sort.id) {
    case 'name':
      field = GetAllSnapshotsSortEnum.NAME
      break
    case 'state':
      field = GetAllSnapshotsSortEnum.STATE
      break
    case 'createdAt':
      field = GetAllSnapshotsSortEnum.CREATED_AT
      break
    case 'lastUsedAt':
    default:
      field = GetAllSnapshotsSortEnum.LAST_USED_AT
      break
  }

  return {
    field,
    direction: sort.desc ? GetAllSnapshotsOrderEnum.DESC : GetAllSnapshotsOrderEnum.ASC,
  }
}
