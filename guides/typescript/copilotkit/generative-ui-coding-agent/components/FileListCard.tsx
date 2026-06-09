'use client'

import { useState } from 'react'
import { Inspector, InspectorToggle } from './Inspector'

type Entry = { name: string; isDir: boolean; size?: number; permissions?: string }

type Props = {
  status: string
  verb: 'listed' | 'searched'
  path?: string
  pattern?: string
  entries?: Entry[]
  files?: string[]
}

export function FileListCard({ status, verb, path, pattern, entries, files }: Props) {
  const [open, setOpen] = useState(false)
  const pending = status !== 'complete'
  const list: { name: string; isDir?: boolean; size?: number }[] = entries
    ? entries.map((e) => ({ name: e.name, isDir: e.isDir, size: e.size }))
    : (files ?? []).map((p) => ({ name: p }))

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
        <span style={{ color: '#64748b' }}>{verb}</span>
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
          {pattern && path ? `${path} • ${pattern}` : (path ?? '…')}
        </code>
        {!pending && list.length ? (
          <span style={{ marginLeft: 'auto', color: '#64748b', fontSize: 11 }}>
            {list.length} {list.length === 1 ? 'entry' : 'entries'}
          </span>
        ) : null}
        <InspectorToggle open={open} onClick={() => setOpen((o) => !o)} />
      </div>
      {list.length ? (
        <div style={{ maxHeight: 240, overflow: 'auto', padding: '6px 0' }}>
          {list.map((item, i) => (
            <div
              key={i}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                padding: '3px 14px',
                fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
                fontSize: 12,
                color: 'var(--fg)',
              }}
            >
              <span style={{ color: item.isDir ? '#2563eb' : '#64748b' }}>
                {item.isDir ? '▸' : '·'}
              </span>
              <span>{item.name}</span>
              {typeof item.size === 'number' && !item.isDir ? (
                <span style={{ marginLeft: 'auto', color: '#94a3b8' }}>{item.size} B</span>
              ) : null}
            </div>
          ))}
        </div>
      ) : !pending ? (
        <div style={{ padding: '8px 14px', color: '#94a3b8', fontSize: 12, fontStyle: 'italic' }}>
          {verb === 'searched' ? '(no matches)' : '(empty)'}
        </div>
      ) : null}
      {open ? (
        <Inspector
          rows={[
            ['path', path],
            ['pattern', pattern],
            ['count', String(list.length)],
          ]}
        />
      ) : null}
    </div>
  )
}
