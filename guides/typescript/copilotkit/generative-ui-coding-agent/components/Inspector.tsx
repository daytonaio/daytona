'use client'

import type { ReactNode } from 'react'

export type InspectorRow = [label: string, value: ReactNode | undefined]

export function Inspector({ rows }: { rows: InspectorRow[] }) {
  const visible = rows.filter(([, v]) => v !== undefined && v !== null && v !== '')
  if (visible.length === 0) return null
  return (
    <div
      style={{
        padding: '8px 14px',
        background: '#f8fafc',
        borderTop: '1px solid var(--border)',
        fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
        fontSize: 11,
      }}
    >
      {visible.map(([k, v], i) => (
        <div
          key={i}
          style={{
            display: 'grid',
            gridTemplateColumns: '110px 1fr',
            gap: 10,
            padding: '3px 0',
          }}
        >
          <span style={{ color: '#94a3b8' }}>{k}</span>
          <span style={{ color: 'var(--fg)', wordBreak: 'break-all', whiteSpace: 'pre-wrap' }}>
            {v}
          </span>
        </div>
      ))}
    </div>
  )
}

export function InspectorToggle({
  open,
  onClick,
}: {
  open: boolean
  onClick: () => void
}) {
  return (
    <button
      onClick={onClick}
      aria-label={open ? 'hide details' : 'show details'}
      style={{
        background: 'transparent',
        border: 'none',
        cursor: 'pointer',
        color: '#64748b',
        fontSize: 11,
        padding: '0 4px',
        fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
      }}
    >
      {open ? '▾' : '▸'}
    </button>
  )
}

export function renderMap(map: Record<string, string> | undefined): string | undefined {
  if (!map || Object.keys(map).length === 0) return undefined
  return Object.entries(map)
    .map(([k, v]) => `${k}=${v}`)
    .join('\n')
}
