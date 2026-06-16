'use client'

import { useState } from 'react'
import { Inspector, InspectorToggle, renderMap } from './Inspector'

type Props = {
  status: string
  sandboxId?: string
  envVars?: Record<string, string>
  labels?: Record<string, string>
  autoStopInterval?: number
}

export function SandboxCard({
  status,
  sandboxId,
  envVars,
  labels,
  autoStopInterval,
}: Props) {
  const [open, setOpen] = useState(false)
  const ready = status === 'complete' && !!sandboxId

  return (
    <div
      style={{
        margin: '6px 0',
        border: '1px solid var(--border)',
        borderRadius: 8,
        overflow: 'hidden',
        background: '#fff',
        fontSize: 13,
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 10,
          padding: '8px 12px',
        }}
      >
        <span style={{ color: ready ? '#16a34a' : '#2563eb', fontWeight: 600 }}>
          {ready ? '✓' : '…'}
        </span>
        <span style={{ color: 'var(--fg)' }}>
          {ready ? 'Sandbox ready' : 'Creating sandbox…'}
        </span>
        {sandboxId ? (
          <code
            style={{
              marginLeft: 'auto',
              fontSize: 11,
              color: '#64748b',
              fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
            }}
          >
            {sandboxId}
          </code>
        ) : null}
        <InspectorToggle open={open} onClick={() => setOpen((o) => !o)} />
      </div>
      {open ? (
        <Inspector
          rows={[
            ['sandboxId', sandboxId],
            ['envVars', renderMap(envVars) ?? '(none)'],
            ['labels', renderMap(labels) ?? '(none)'],
            [
              'autoStopInterval',
              typeof autoStopInterval === 'number' ? `${autoStopInterval} min` : '15 min (default)',
            ],
          ]}
        />
      ) : null}
    </div>
  )
}
