/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
import { RegionDto } from '@daytonaio/api-client'
import { Trash2 } from 'lucide-react'
import { Button } from './ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table'
import { formatDistanceToNow } from 'date-fns'

interface RegionTableProps {
  data: RegionDto[]
  loading: boolean
  loadingRegions: Record<string, boolean>
  onDelete: (region: RegionDto) => void
  writePermitted: boolean
}

export const RegionTable: React.FC<RegionTableProps> = ({
  data,
  loading,
  loadingRegions,
  onDelete,
  writePermitted,
}) => {
  if (loading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 5 }).map((_, i) => (
          <div key={i} className="h-12 bg-muted animate-pulse rounded" />
        ))}
      </div>
    )
  }

  if (data.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">No regions found.</p>
        <p className="text-sm text-muted-foreground mt-1">Create your first region to get started.</p>
      </div>
    )
  }

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Docker Registry ID</TableHead>
            <TableHead>Created</TableHead>
            <TableHead>Updated</TableHead>
            {writePermitted && <TableHead className="w-[100px]">Actions</TableHead>}
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((region) => (
            <TableRow key={region.name}>
              <TableCell className="font-medium">{region.name}</TableCell>
              <TableCell>
                {region.dockerRegistryId ? (
                  <code className="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm">
                    {region.dockerRegistryId}
                  </code>
                ) : (
                  <span className="text-muted-foreground">—</span>
                )}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {formatDistanceToNow(new Date(region.createdAt), { addSuffix: true })}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {formatDistanceToNow(new Date(region.updatedAt), { addSuffix: true })}
              </TableCell>
              {writePermitted && (
                <TableCell>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onDelete(region)}
                    disabled={loadingRegions[region.name]}
                    className="h-8 w-8 p-0"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </TableCell>
              )}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
