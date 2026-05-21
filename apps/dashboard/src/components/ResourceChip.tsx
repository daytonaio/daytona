/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { CpuIcon, HardDriveIcon, MemoryStickIcon, SparklesIcon } from 'lucide-react'

type Resource = 'cpu' | 'memory' | 'disk' | 'gpu'

interface Props {
  resource: Resource
  value: number
  unit?: string
  icon?: React.ReactNode
}

const resourceUnits: Record<Resource, string> = {
  cpu: 'vCPU',
  memory: 'GiB',
  disk: 'GiB',
  gpu: 'GPU',
}

const resourceIcons: Record<Resource, React.ComponentType<{ className?: string; strokeWidth?: number }>> = {
  cpu: CpuIcon,
  memory: MemoryStickIcon,
  disk: HardDriveIcon,
  gpu: SparklesIcon,
}

const resourceChipClassName: Record<Resource, string> = {
  cpu: 'bg-muted/80 border border-border',
  memory: 'bg-muted/80 border border-border',
  disk: 'bg-muted/80 border border-border',
  gpu: 'bg-purple-100 text-purple-600 dark:bg-purple-950 dark:text-purple-200',
}

export function ResourceChip({ resource, value, unit, icon }: Props) {
  const resourceUnit = unit ?? resourceUnits[resource]
  const ResourceIcon = resourceIcons[resource]

  return (
    <div
      className={cn(
        'flex items-center gap-1 rounded-full px-2 py-[2px] text-sm whitespace-nowrap',
        resourceChipClassName[resource],
      )}
    >
      {icon === null ? null : (icon ?? <ResourceIcon className="w-4 h-4 flex-shrink-0" strokeWidth={1.5} />)} {value}{' '}
      {resourceUnit}
    </div>
  )
}
