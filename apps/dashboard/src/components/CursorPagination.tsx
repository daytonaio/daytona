/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ChevronLeft, ChevronRight } from 'lucide-react'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select'
import { Button } from './ui/button'
import { PAGE_SIZE_OPTIONS } from '../constants/Pagination'

interface CursorPaginationProps {
  pageSize: number
  onPageSizeChange: (pageSize: number) => void
  hasNextPage: boolean
  hasPreviousPage: boolean
  onNextPage: () => void
  onPreviousPage: () => void
  className?: string
}

export function CursorPagination({
  pageSize,
  onPageSizeChange,
  hasNextPage,
  hasPreviousPage,
  onNextPage,
  onPreviousPage,
  className,
}: CursorPaginationProps) {
  return (
    <div className={`flex items-center justify-start gap-2 ${className}`}>
      <Select value={`${pageSize}`} onValueChange={(value) => onPageSizeChange(Number(value))}>
        <SelectTrigger className="h-8 w-[164px]">
          <SelectValue placeholder={pageSize + ' per page'} />
        </SelectTrigger>
        <SelectContent side="top">
          {PAGE_SIZE_OPTIONS.map((size) => (
            <SelectItem key={size} value={`${size}`}>
              {size} per page
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <div className="flex items-center space-x-2">
        <Button variant="outline" className="h-8 w-8 p-0" onClick={onPreviousPage} disabled={!hasPreviousPage}>
          <span className="sr-only">Go to previous page</span>
          <ChevronLeft />
        </Button>
        <Button variant="outline" className="h-8 w-8 p-0" onClick={onNextPage} disabled={!hasNextPage}>
          <span className="sr-only">Go to next page</span>
          <ChevronRight />
        </Button>
      </div>
    </div>
  )
}
