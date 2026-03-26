/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CpuIcon, HardDriveIcon, MemoryStickIcon } from 'lucide-react'

interface Props {
  resource: 'cpu' | 'memory' | 'disk'
  value: number
  unit?: string
  icon?: React.ReactNode
}

const resourceUnits = {
  cpu: 'vCPU',
  memory: 'GiB',
  disk: 'GiB',
}

const resourceIcons = {
  cpu: CpuIcon,
  memory: MemoryStickIcon,
  disk: HardDriveIcon,
}

export function ResourceChip({ resource, value, unit, icon }: Props) {
  const resourceUnit = unit ?? resourceUnits[resource]
  const ResourceIcon = resourceIcons[resource]

  return (
    <div className="flex items-center gap-1 bg-muted/80 border border-border rounded-full px-2 py-[2px] text-sm whitespace-nowrap">
      {icon === null ? null : (icon ?? <ResourceIcon className="w-4 h-4 flex-shrink-0" strokeWidth={1.5} />)} {value}{' '}
      {resourceUnit}
    </div>
  )
}
