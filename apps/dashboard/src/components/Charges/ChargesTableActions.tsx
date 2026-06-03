/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ExternalLinkIcon, MoreHorizontalIcon } from 'lucide-react'
import { Button } from '../ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '../ui/dropdown-menu'
import { ChargesTableActionsProps } from './types'

export function ChargesTableActions({ charge }: ChargesTableActionsProps) {
  if (!charge.receiptUrl) {
    return null
  }

  return (
    <div className="flex items-center justify-center">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon-sm" aria-label="Open menu">
            <MoreHorizontalIcon className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem className="cursor-pointer" asChild>
            <a href={charge.receiptUrl} target="_blank" rel="noopener noreferrer">
              <ExternalLinkIcon className="mr-2 h-4 w-4" />
              View receipt
            </a>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
