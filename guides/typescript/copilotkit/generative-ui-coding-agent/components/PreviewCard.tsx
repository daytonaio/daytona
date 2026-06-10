'use client'

import { useState } from 'react'
import { Inspector, InspectorToggle } from './Inspector'

type Props = {
  status: string
  url?: string
  port?: number
}

export function PreviewCard({ status, url, port }: Props) {
  const ready = status === 'complete' && !!url
  const [reloadKey, setReloadKey] = useState(0)
  const [open, setOpen] = useState(false)

  return (
    <div
      style={{
        margin: '8px 0',
        border: '1px solid var(--border)',
        borderRadius: 12,
        overflow: 'hidden',
        background: '#fff',
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 10,
          padding: '10px 14px',
          borderBottom: '1px solid var(--border)',
          background: 'var(--muted)',
          fontSize: 13,
        }}
      >
        <span style={{ color: ready ? '#16a34a' : '#2563eb', fontWeight: 600 }}>
          {ready ? '●' : '…'}
        </span>
        <span style={{ color: 'var(--fg)' }}>
          {ready ? 'Live preview' : `Opening port ${port ?? '…'}…`}
        </span>
        {ready ? (
          <div style={{ marginLeft: 'auto', display: 'flex', gap: 12, alignItems: 'center' }}>
            <button
              onClick={() => setReloadKey((k) => k + 1)}
              style={{
                background: 'transparent',
                border: 'none',
                color: 'var(--accent)',
                fontSize: 12,
                cursor: 'pointer',
                padding: 0,
              }}
            >
              reload ↻
            </button>
            <a
              href={url}
              target="_blank"
              rel="noreferrer noopener"
              style={{
                fontSize: 12,
                color: 'var(--accent)',
                textDecoration: 'none',
              }}
            >
              open in new tab ↗
            </a>
            <InspectorToggle open={open} onClick={() => setOpen((o) => !o)} />
          </div>
        ) : (
          <span style={{ marginLeft: 'auto' }}>
            <InspectorToggle open={open} onClick={() => setOpen((o) => !o)} />
          </span>
        )}
      </div>
      <div style={{ padding: 12 }}>
        {ready ? (
          <iframe
            key={reloadKey}
            src={url}
            title="Sandbox preview"
            sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
            style={{
              width: '100%',
              height: 520,
              border: '1px solid var(--border)',
              borderRadius: 8,
              background: '#fff',
            }}
          />
        ) : (
          <div
            style={{
              height: 520,
              borderRadius: 8,
              border: '1px dashed var(--border)',
              background:
                'linear-gradient(90deg, #f1f5f9 0%, #e2e8f0 50%, #f1f5f9 100%) 0 0 / 200% 100%',
              animation: 'cpkit-shimmer 1.4s linear infinite',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#64748b',
              fontSize: 13,
            }}
          >
            Waiting for the dev server…
            <style>{`@keyframes cpkit-shimmer { 0%{background-position:100% 0} 100%{background-position:-100% 0} }`}</style>
          </div>
        )}
      </div>
      {open ? (
        <Inspector
          rows={[
            ['url', url],
            ['port', typeof port === 'number' ? String(port) : undefined],
          ]}
        />
      ) : null}
    </div>
  )
}
