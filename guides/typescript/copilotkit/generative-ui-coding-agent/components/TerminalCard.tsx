'use client'

import { useState } from 'react'
import { Inspector, InspectorToggle } from './Inspector'

type Props = {
  status: string
  command?: string
  stdout?: string
  exitCode?: number
  background?: boolean
  sessionId?: string
  cmdId?: string
}

export function TerminalCard({
  status,
  command,
  stdout,
  exitCode,
  background,
  sessionId,
  cmdId,
}: Props) {
  const [open, setOpen] = useState(false)
  const pending = status !== 'complete'
  const failed = !pending && typeof exitCode === 'number' && exitCode !== 0

  return (
    <div
      style={{
        margin: '6px 0',
        border: '1px solid #1f2937',
        borderRadius: 10,
        overflow: 'hidden',
        background: '#0b1020',
        color: '#e2e8f0',
        fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
        fontSize: 12,
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          padding: '8px 12px',
          background: '#111827',
          borderBottom: '1px solid #1f2937',
        }}
      >
        <span style={{ color: pending ? '#60a5fa' : failed ? '#f87171' : '#4ade80' }}>
          {pending ? '●' : failed ? '✗' : '✓'}
        </span>
        <span style={{ color: '#94a3b8' }}>
          {pending && (background ? 'starting…' : 'running…')}
          {!pending && background && 'background'}
          {!pending && !background && (failed ? `exit ${exitCode}` : 'done')}
        </span>
        <span style={{ marginLeft: 'auto' }}>
          <InspectorToggle open={open} onClick={() => setOpen((o) => !o)} />
        </span>
      </div>
      <div style={{ padding: '10px 12px' }}>
        <div style={{ color: '#60a5fa', wordBreak: 'break-all' }}>
          <span style={{ color: '#475569', marginRight: 6 }}>$</span>
          {command ?? '…'}
        </div>
        {stdout && !background ? (
          <pre
            style={{
              margin: '8px 0 0',
              padding: 0,
              background: 'transparent',
              color: '#cbd5e1',
              whiteSpace: 'pre-wrap',
              wordBreak: 'break-word',
              maxHeight: 240,
              overflow: 'auto',
            }}
          >
            {stdout.length > 4000 ? `…${stdout.slice(-4000)}` : stdout}
          </pre>
        ) : null}
      </div>
      {open ? (
        <Inspector
          rows={[
            ['exitCode', typeof exitCode === 'number' ? String(exitCode) : undefined],
            ['background', background ? 'true' : 'false'],
            ['sessionId', sessionId],
            ['cmdId', cmdId],
            ['stdout bytes', stdout ? String(stdout.length) : '0'],
          ]}
        />
      ) : null}
    </div>
  )
}
