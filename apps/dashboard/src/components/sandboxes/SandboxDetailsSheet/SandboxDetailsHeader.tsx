/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CopyButton } from '@/components/CopyButton'
import type { Sandbox } from '@daytona/api-client'
import type React from 'react'
import { InfoRow, InfoSection } from '../SandboxInfoPanel'
import { SandboxState as SandboxStateComponent } from '../SandboxState'

export function SandboxDetailsHeader({ sandbox, actions }: { sandbox: Sandbox; actions?: React.ReactNode }) {
  const hasCustomName = !!sandbox.name && sandbox.name !== sandbox.id

  return (
    <InfoSection className="shrink-0 last:border-b">
      <InfoRow label={hasCustomName ? 'Name' : 'Name / UUID'} className="-mr-2">
        <div className="flex items-center gap-1 min-w-0">
          <span className="truncate">{hasCustomName ? sandbox.name : sandbox.id}</span>
          <CopyButton
            value={hasCustomName ? sandbox.name : sandbox.id}
            tooltipText={hasCustomName ? 'Copy name' : 'Copy name / UUID'}
            size="icon-xs"
          />
        </div>
      </InfoRow>
      {hasCustomName && (
        <InfoRow label="UUID" className="-mr-2">
          <div className="flex items-center gap-1 min-w-0">
            <span className="truncate">{sandbox.id}</span>
            <CopyButton value={sandbox.id} tooltipText="Copy UUID" size="icon-xs" />
          </div>
        </InfoRow>
      )}
      <div className="flex items-center justify-between gap-3 pt-3">
        <SandboxStateComponent
          state={sandbox.state}
          errorReason={sandbox.errorReason}
          recoverable={sandbox.recoverable}
          animate
        />
        {actions ? <div className="flex justify-end">{actions}</div> : null}
      </div>
    </InfoSection>
  )
}
