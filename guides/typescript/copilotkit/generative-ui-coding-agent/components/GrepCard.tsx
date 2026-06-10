'use client'

import { useState } from 'react'
import { Inspector, InspectorToggle } from './Inspector'

type Match = { file: string; line: number; content: string }

type Props = {
  status: string
  path?: string
  pattern?: string
  matches?: Match[]
}

export function GrepCard({ status, path, pattern, matches }: Props) {
  const [open, setOpen] = useState(false)
  const pending = status !== 'complete'
  const list = matches ?? []

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
        <span style={{ color: '#64748b' }}>found</span>
        <code
          style={{
            fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
            fontSize: 12,
            color: 'var(--fg)',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
            maxWidth: 320,
          }}
        >
          {pattern && path ? `${pattern} in ${path}` : (path ?? '…')}
        </code>
        {!pending ? (
          <span style={{ marginLeft: 'auto', color: '#64748b', fontSize: 11 }}>
            {list.length} {list.length === 1 ? 'match' : 'matches'}
          </span>
        ) : null}
        <InspectorToggle open={open} onClick={() => setOpen((o) => !o)} />
      </div>
      {list.length ? (
        <div style={{ maxHeight: 280, overflow: 'auto' }}>
          {list.map((m, i) => (
            <div
              key={i}
              style={{
                display: 'grid',
                gridTemplateColumns: 'auto 1fr',
                gap: 10,
                padding: '4px 14px',
                fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
                fontSize: 12,
                borderTop: i > 0 ? '1px solid #f1f5f9' : 'none',
              }}
            >
              <span style={{ color: '#94a3b8' }}>
                {m.file}:{m.line}
              </span>
              <span style={{ color: 'var(--fg)', whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                {m.content}
              </span>
            </div>
          ))}
        </div>
      ) : !pending ? (
        <div style={{ padding: '8px 14px', color: '#94a3b8', fontSize: 12, fontStyle: 'italic' }}>
          (no matches)
        </div>
      ) : null}
      {open ? (
        <Inspector
          rows={[
            ['path', path],
            ['pattern', pattern],
            ['matches', String(list.length)],
          ]}
        />
      ) : null}
    </div>
  )
}
