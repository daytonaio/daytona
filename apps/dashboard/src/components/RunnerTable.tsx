/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
import { Runner, RunnerState, SandboxClass } from '@daytonaio/api-client'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Server, Cpu, MemoryStick, HardDrive, Zap } from 'lucide-react'

interface RunnerTableProps {
  data: Runner[]
  loading: boolean
  writePermitted: boolean
}

export const RunnerTable: React.FC<RunnerTableProps> = ({ data, loading, writePermitted }) => {
  const getStateBadgeVariant = (state: RunnerState) => {
    switch (state) {
      case RunnerState.READY:
        return 'default'
      case RunnerState.INITIALIZING:
        return 'secondary'
      case RunnerState.DISABLED:
        return 'outline'
      case RunnerState.DECOMMISSIONED:
        return 'outline'
      case RunnerState.UNRESPONSIVE:
        return 'destructive'
      default:
        return 'secondary'
    }
  }

  const getClassBadgeVariant = (sandboxClass: SandboxClass) => {
    switch (sandboxClass) {
      case SandboxClass.SMALL:
        return 'default'
      case SandboxClass.MEDIUM:
        return 'secondary'
      case SandboxClass.LARGE:
        return 'outline'
      default:
        return 'secondary'
    }
  }

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
        <Server className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
        <p className="text-muted-foreground">No runners found in this region.</p>
        <p className="text-sm text-muted-foreground mt-1">
          Runners will appear here once they are registered in the selected region.
        </p>
      </div>
    )
  }

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Domain</TableHead>
            <TableHead>State</TableHead>
            <TableHead>Class</TableHead>
            <TableHead>Resources</TableHead>
            <TableHead>Usage</TableHead>
            <TableHead>Availability Score</TableHead>
            <TableHead>Last Checked</TableHead>
            {writePermitted && <TableHead className="w-[100px]">Actions</TableHead>}
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((runner) => (
            <TableRow key={runner.id}>
              <TableCell className="font-mono text-sm">{runner.domain}</TableCell>
              <TableCell>
                <Badge variant={getStateBadgeVariant(runner.state)}>{runner.state}</Badge>
              </TableCell>
              <TableCell>
                <Badge variant={getClassBadgeVariant(runner.class)}>{runner.class}</Badge>
              </TableCell>
              <TableCell>
                <div className="flex flex-col gap-1 text-sm">
                  <div className="flex items-center gap-2">
                    <Cpu className="w-4 h-4 text-muted-foreground" />
                    <span>{runner.cpu} CPU</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <MemoryStick className="w-4 h-4 text-muted-foreground" />
                    <span>{runner.memory} GiB RAM</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <HardDrive className="w-4 h-4 text-muted-foreground" />
                    <span>{runner.disk} GiB Disk</span>
                  </div>
                  {runner.gpu > 0 && (
                    <div className="flex items-center gap-2">
                      <Zap className="w-4 h-4 text-muted-foreground" />
                      <span>
                        {runner.gpu} {runner.gpuType}
                      </span>
                    </div>
                  )}
                </div>
              </TableCell>
              <TableCell>
                <div className="flex flex-col gap-1 text-sm">
                  <div className="flex items-center gap-2">
                    <span>CPU: {runner.currentCpuUsagePercentage?.toFixed(1) || 0}%</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span>RAM: {runner.currentMemoryUsagePercentage?.toFixed(1) || 0}%</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span>Disk: {runner.currentDiskUsagePercentage?.toFixed(1) || 0}%</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span>Snapshots: {runner.currentSnapshotCount || 0}</span>
                  </div>
                </div>
              </TableCell>
              <TableCell>
                <div className="text-sm">
                  <span className="font-medium">{runner.availabilityScore || 0}</span>
                  <span className="text-muted-foreground">/100</span>
                </div>
              </TableCell>
              <TableCell className="text-sm text-muted-foreground">
                {runner.lastChecked ? new Date(runner.lastChecked).toLocaleDateString() : 'Never'}
              </TableCell>
              {writePermitted && (
                <TableCell>
                  <div className="flex items-center gap-2">{/* Add action buttons here when needed */}</div>
                </TableCell>
              )}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
