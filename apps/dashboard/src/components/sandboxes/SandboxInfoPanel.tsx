/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React, { useMemo } from 'react'
import { CopyButton } from '@/components/CopyButton'
import { ResourceChip } from '@/components/ResourceChip'
import { TimestampTooltip } from '@/components/TimestampTooltip'
import { Skeleton } from '@/components/ui/skeleton'
import { formatDuration, getRelativeTimeString } from '@/lib/utils'
import { Sandbox } from '@daytonaio/api-client'
import { Tag } from 'lucide-react'

export function InfoSection({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="px-5 py-4 border-b border-border last:border-b-0">
      <p className="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground mb-2">{title}</p>
      {children}
    </div>
  )
}

export function InfoRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-start justify-between gap-3 py-1.5">
      <span className="text-[13px] text-muted-foreground shrink-0 pt-px leading-tight">{label}</span>
      <div className="min-w-0 text-[13px] font-medium text-right">{children}</div>
    </div>
  )
}

interface SandboxInfoPanelProps {
  sandbox: Sandbox
  getRegionName: (id: string) => string | undefined
}

export function SandboxInfoPanel({ sandbox, getRegionName }: SandboxInfoPanelProps) {
  const labelEntries = useMemo(() => {
    return sandbox.labels ? Object.entries(sandbox.labels) : []
  }, [sandbox.labels])

  return (
    <div className="flex flex-col">
      <InfoSection title="General">
        <InfoRow label="Region">
          <div className="flex items-center gap-1">
            <span className="truncate">{getRegionName(sandbox.target) ?? sandbox.target}</span>
            <CopyButton value={sandbox.target} tooltipText="Copy" size="icon-xs" />
          </div>
        </InfoRow>
        <InfoRow label="Snapshot">
          {sandbox.snapshot ? (
            <div className="flex items-center gap-1 min-w-0">
              <span className="truncate font-mono text-[13px]">{sandbox.snapshot}</span>
              <CopyButton value={sandbox.snapshot} tooltipText="Copy" size="icon-xs" />
            </div>
          ) : (
            <span className="text-muted-foreground font-normal">—</span>
          )}
        </InfoRow>
      </InfoSection>

      <InfoSection title="Resources">
        <div className="flex flex-wrap gap-1.5 py-1">
          <ResourceChip resource="cpu" value={sandbox.cpu} />
          <ResourceChip resource="memory" value={sandbox.memory} />
          <ResourceChip resource="disk" value={sandbox.disk} />
        </div>
      </InfoSection>

      <InfoSection title="Lifecycle">
        <InfoRow label="Auto-stop">
          {sandbox.autoStopInterval ? (
            formatDuration(sandbox.autoStopInterval)
          ) : (
            <span className="text-muted-foreground font-normal">Disabled</span>
          )}
        </InfoRow>
        <InfoRow label="Auto-archive">
          {sandbox.autoArchiveInterval ? (
            formatDuration(sandbox.autoArchiveInterval)
          ) : (
            <span className="text-muted-foreground font-normal">Disabled</span>
          )}
        </InfoRow>
        <InfoRow label="Auto-delete">
          {sandbox.autoDeleteInterval !== undefined && sandbox.autoDeleteInterval >= 0 ? (
            sandbox.autoDeleteInterval === 0 ? (
              'On stop'
            ) : (
              formatDuration(sandbox.autoDeleteInterval)
            )
          ) : (
            <span className="text-muted-foreground font-normal">Disabled</span>
          )}
        </InfoRow>
      </InfoSection>

      <InfoSection title="Labels">
        {labelEntries.length > 0 ? (
          <div className="max-h-[250px] overflow-y-auto scrollbar-sm">
            <div className="flex flex-wrap gap-1.5 py-1">
              {labelEntries.map(([key, value]) => (
                <code
                  key={key}
                  className="flex items-center gap-1 bg-muted border border-border rounded px-2 py-0.5 text-xs font-mono"
                >
                  <span className="text-muted-foreground">{key}:</span>
                  <span>{value}</span>
                </code>
              ))}
            </div>
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center gap-1.5 py-5 text-muted-foreground">
            <Tag className="size-4" />
            <span className="text-[13px]">No labels</span>
          </div>
        )}
      </InfoSection>

      <InfoSection title="Timestamps">
        <InfoRow label="Created">
          <TimestampTooltip timestamp={sandbox.createdAt}>
            <span>{getRelativeTimeString(sandbox.createdAt).relativeTimeString}</span>
          </TimestampTooltip>
        </InfoRow>
        <InfoRow label="Last event">
          <TimestampTooltip timestamp={sandbox.updatedAt}>
            <span>{getRelativeTimeString(sandbox.updatedAt).relativeTimeString}</span>
          </TimestampTooltip>
        </InfoRow>
      </InfoSection>
    </div>
  )
}

export function InfoPanelSkeleton() {
  return (
    <div className="flex flex-col">
      <div className="px-5 py-4 border-b border-border">
        <Skeleton className="h-2.5 w-16 mb-3" />
        <div className="space-y-3">
          <div className="flex justify-between">
            <Skeleton className="h-4 w-12" />
            <Skeleton className="h-4 w-20" />
          </div>
          <div className="flex justify-between">
            <Skeleton className="h-4 w-16" />
            <Skeleton className="h-4 w-32" />
          </div>
        </div>
      </div>
      <div className="px-5 py-4 border-b border-border">
        <Skeleton className="h-2.5 w-20 mb-3" />
        <div className="flex gap-1.5">
          <Skeleton className="h-6 w-16 rounded-full" />
          <Skeleton className="h-6 w-16 rounded-full" />
          <Skeleton className="h-6 w-16 rounded-full" />
        </div>
      </div>
      <div className="px-5 py-4 border-b border-border">
        <Skeleton className="h-2.5 w-18 mb-3" />
        <div className="space-y-3">
          <div className="flex justify-between">
            <Skeleton className="h-4 w-20" />
            <Skeleton className="h-4 w-16" />
          </div>
          <div className="flex justify-between">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-4 w-16" />
          </div>
          <div className="flex justify-between">
            <Skeleton className="h-4 w-22" />
            <Skeleton className="h-4 w-16" />
          </div>
        </div>
      </div>
      <div className="px-5 py-4 border-b border-border">
        <Skeleton className="h-2.5 w-14 mb-3" />
        <Skeleton className="h-4 w-full" />
      </div>
      <div className="px-5 py-4">
        <Skeleton className="h-2.5 w-24 mb-3" />
        <div className="space-y-3">
          <div className="flex justify-between">
            <Skeleton className="h-4 w-16" />
            <Skeleton className="h-4 w-24" />
          </div>
          <div className="flex justify-between">
            <Skeleton className="h-4 w-20" />
            <Skeleton className="h-4 w-24" />
          </div>
        </div>
      </div>
    </div>
  )
}
