/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
import { Badge } from './ui/badge'
import { Button } from './ui/button'
import { Eye } from 'lucide-react'
import {
  summarizePermissions,
  getCategoryLevelLabel,
  getCategoryLevelVariant,
  type PermissionSummary,
} from '@/utils/permissionSummary'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from './ui/dialog'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from './ui/tooltip'

interface PermissionBadgesProps {
  permissions: string[]
  showDetails?: boolean
}

export function PermissionBadges({ permissions, showDetails = false }: PermissionBadgesProps) {
  const summary = summarizePermissions(permissions)

  if (permissions.length === 0) {
    return <span className="text-muted-foreground text-sm">No permissions</span>
  }

  // For simple cases (full access or read-only), show a single badge
  if (summary.type === 'full' || summary.type === 'readonly') {
    return (
      <div className="flex items-center gap-2">
        <Badge variant={summary.variant}>{summary.label}</Badge>
        {showDetails && <PermissionDetailsDialog permissions={permissions} summary={summary} />}
      </div>
    )
  }

  // For custom permissions, show category badges
  return (
    <div className="flex items-center gap-1 flex-wrap">
      {summary.categories.slice(0, 3).map((category) => (
        <TooltipProvider key={category.name}>
          <Tooltip>
            <TooltipTrigger>
              <Badge variant={getCategoryLevelVariant(category.level)} className="text-xs">
                {category.name} {getCategoryLevelLabel(category.level)}
              </Badge>
            </TooltipTrigger>
            <TooltipContent>
              <div className="text-sm">
                <div className="font-medium">{category.name}</div>
                <div className="text-xs text-muted-foreground">{category.permissions.join(', ')}</div>
              </div>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      ))}

      {summary.categories.length > 3 && (
        <Badge variant="secondary" className="text-xs">
          +{summary.categories.length - 3} more
        </Badge>
      )}

      {showDetails && <PermissionDetailsDialog permissions={permissions} summary={summary} />}
    </div>
  )
}

interface PermissionDetailsDialogProps {
  permissions: string[]
  summary: PermissionSummary
}

function PermissionDetailsDialog({ permissions, summary }: PermissionDetailsDialogProps) {
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="ghost" size="icon" className="h-6 w-6" onMouseDown={(e) => e.preventDefault()}>
          <Eye className="h-3 w-3" />
          <span className="sr-only">View permission details</span>
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-md [&_button:focus]:!outline-none [&_button:focus]:!ring-0">
        <DialogHeader>
          <DialogTitle>Permission Details</DialogTitle>
          <DialogDescription>Detailed breakdown of permissions for this API key.</DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div>
            <h4 className="font-medium mb-2">Summary</h4>
            <Badge variant={summary.variant}>{summary.label}</Badge>
          </div>

          {summary.categories.length > 0 && (
            <div>
              <h4 className="font-medium mb-2">Categories</h4>
              <div className="space-y-2">
                {summary.categories.map((category) => (
                  <div key={category.name} className="border rounded-md p-3">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="font-medium text-sm">{category.name}</span>
                      <Badge variant={getCategoryLevelVariant(category.level)} className="text-xs">
                        {getCategoryLevelLabel(category.level)}
                      </Badge>
                    </div>
                    <div className="text-xs text-muted-foreground">{category.permissions.join(', ')}</div>
                  </div>
                ))}
              </div>
            </div>
          )}

          <div>
            <h4 className="font-medium mb-2">Raw Permissions</h4>
            <div className="bg-muted rounded-md p-3 text-xs font-mono">{permissions.join(', ')}</div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
