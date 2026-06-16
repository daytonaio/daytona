'use client'

import { useState } from 'react'
import { Inspector, InspectorToggle } from './Inspector'

type Result = { file?: string; success?: boolean; error?: string }

type Props = {
  status: string
  pattern?: string
  newValue?: string
  results?: Result[]
}

export function ReplaceCard({ status, pattern, newValue, results }: Props) {
  const [open, setOpen] = useState(false)
  const pending = status !== 'complete'
  const list = results ?? []
  const successCount = list.filter((r) => r.success).length

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
          background: 'var(--muted)',
          borderBottom: list.length ? '1px solid var(--border)' : 'none',
        }}
      >
        <span style={{ color: pending ? '#2563eb' : '#16a34a', fontWeight: 600 }}>
          {pending ? '…' : '✓'}
        </span>
        <span style={{ color: '#64748b' }}>replaced</span>
        {pattern || newValue ? (
          <code
            style={{
              fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
              fontSize: 12,
              color: 'var(--fg)',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
              maxWidth: 280,
            }}
          >
            {pattern} → {newValue}
          </code>
        ) : null}
        {!pending && list.length ? (
          <span style={{ marginLeft: 'auto', color: '#64748b', fontSize: 11 }}>
            {successCount}/{list.length} ok
          </span>
        ) : null}
        <InspectorToggle open={open} onClick={() => setOpen((o) => !o)} />
      </div>
      {list.length ? (
        <div style={{ maxHeight: 240, overflow: 'auto', padding: '6px 0' }}>
          {list.map((r, i) => (
            <div
              key={i}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                padding: '3px 14px',
                fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
                fontSize: 12,
                color: r.success ? 'var(--fg)' : '#b91c1c',
              }}
            >
              <span style={{ color: r.success ? '#16a34a' : '#b91c1c' }}>
                {r.success ? '✓' : '✗'}
              </span>
              <span>{r.file ?? '?'}</span>
              {r.error ? (
                <span style={{ marginLeft: 'auto', color: '#b91c1c', fontSize: 11 }}>
                  {r.error}
                </span>
              ) : null}
            </div>
          ))}
        </div>
      ) : null}
      {open ? (
        <Inspector
          rows={[
            ['pattern', pattern],
            ['newValue', newValue],
            ['files', list.length ? String(list.length) : undefined],
            ['successful', list.length ? `${successCount}/${list.length}` : undefined],
          ]}
        />
      ) : null}
    </div>
  )
}
