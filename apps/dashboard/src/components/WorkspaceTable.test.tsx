/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { render, screen } from '@testing-library/react'
import { WorkspaceTable } from './WorkspaceTable'
import { DEFAULT_PAGE_SIZE } from '@/lib/table-utils'

// Mock the dependencies
jest.mock('@/hooks/useSelectedOrganization', () => ({
  useSelectedOrganization: () => ({
    authenticatedUserHasPermission: () => true,
  }),
}))

// Mock the Pagination component to verify it receives the correct table prop
jest.mock('./Pagination', () => ({
  Pagination: ({ table }) => {
    // Check if the table has the correct default page size
    const pageSize = table.getState().pagination.pageSize
    return (
      <div data-testid="pagination" data-page-size={pageSize}>
        Pagination Component
      </div>
    )
  },
}))

describe('WorkspaceTable Component', () => {
  const mockProps = {
    data: [],
    loadingWorkspaces: {},
    loading: false,
    handleStart: jest.fn(),
    handleStop: jest.fn(),
    handleDelete: jest.fn(),
    handleBulkDelete: jest.fn(),
    handleArchive: jest.fn(),
  }

  it('initializes with the correct default page size', () => {
    render(<WorkspaceTable {...mockProps} />)

    // Check that the Pagination component received a table with the correct page size
    const paginationElement = screen.getByTestId('pagination')
    expect(paginationElement.getAttribute('data-page-size')).toBe(DEFAULT_PAGE_SIZE.toString())
  })
})
