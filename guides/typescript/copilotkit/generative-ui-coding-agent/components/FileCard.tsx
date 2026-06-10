'use client'

import { useState } from 'react'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism'

type Props = {
  status: string
  verb: 'wrote' | 'read'
  path?: string
  content?: string
  bytes?: number
}

export function FileCard({ status, verb, path, content, bytes }: Props) {
  const [open, setOpen] = useState(false)
  const pending = status !== 'complete'
  const label = pending
    ? verb === 'wrote'
      ? 'writing'
      : 'reading'
    : verb === 'wrote'
      ? 'wrote'
      : 'read'

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
      <button
        onClick={() => setOpen((o) => !o)}
        disabled={!content}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 10,
          width: '100%',
          padding: '8px 12px',
          background: 'transparent',
          border: 'none',
          cursor: content ? 'pointer' : 'default',
          textAlign: 'left',
          color: 'var(--fg)',
        }}
      >
        <span style={{ color: pending ? '#2563eb' : '#16a34a', fontWeight: 600 }}>
          {pending ? '…' : '✓'}
        </span>
        <span style={{ color: '#64748b' }}>{label}</span>
        <code
          style={{
            fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
            fontSize: 12,
          }}
        >
          {path ?? '…'}
        </code>
        {bytes !== undefined ? (
          <span style={{ marginLeft: 'auto', color: '#64748b', fontSize: 11 }}>
            {bytes} B {content ? (open ? '▾' : '▸') : ''}
          </span>
        ) : null}
      </button>
      {open && content ? (
        <div style={{ borderTop: '1px solid var(--border)', maxHeight: 360, overflow: 'auto' }}>
          <SyntaxHighlighter
            language={languageFor(path ?? '')}
            style={vscDarkPlus}
            customStyle={{ margin: 0, fontSize: 12 }}
          >
            {content}
          </SyntaxHighlighter>
        </div>
      ) : null}
    </div>
  )
}

function languageFor(path: string): string {
  if (path.endsWith('.tsx') || path.endsWith('.ts')) return 'typescript'
  if (path.endsWith('.jsx') || path.endsWith('.js') || path.endsWith('.mjs')) return 'javascript'
  if (path.endsWith('.json')) return 'json'
  if (path.endsWith('.css')) return 'css'
  if (path.endsWith('.html')) return 'html'
  if (path.endsWith('.md')) return 'markdown'
  if (path.endsWith('.py')) return 'python'
  if (path.endsWith('.sh')) return 'bash'
  return 'text'
}
