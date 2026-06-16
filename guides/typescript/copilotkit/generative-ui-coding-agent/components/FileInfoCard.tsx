'use client'

type Info = {
  path?: string
  name?: string
  isDir?: boolean
  size?: number
  mode?: string
  permissions?: string
  owner?: string
  group?: string
  modifiedAt?: string
}

type Props = {
  status: string
  info?: Info
}

export function FileInfoCard({ status, info }: Props) {
  const pending = status !== 'complete'
  const rawRows: Array<[string, string | number | undefined]> = info
    ? [
        ['type', info.isDir ? 'directory' : 'file'],
        ['size', info.size],
        ['mode', info.mode],
        ['permissions', info.permissions],
        ['owner', info.owner && info.group ? `${info.owner}:${info.group}` : info.owner],
        ['modified', info.modifiedAt],
      ]
    : []
  const rows = rawRows.filter(
    (entry): entry is [string, string | number] =>
      entry[1] !== undefined && entry[1] !== '',
  )

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
          borderBottom: rows.length ? '1px solid var(--border)' : 'none',
        }}
      >
        <span style={{ color: pending ? '#2563eb' : '#16a34a', fontWeight: 600 }}>
          {pending ? '…' : '✓'}
        </span>
        <span style={{ color: '#64748b' }}>info</span>
        <code
          style={{
            fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
            fontSize: 12,
            color: 'var(--fg)',
          }}
        >
          {info?.path ?? '…'}
        </code>
      </div>
      {rows.length ? (
        <div style={{ padding: '6px 14px' }}>
          {rows.map(([k, v]) => (
            <div
              key={k}
              style={{
                display: 'grid',
                gridTemplateColumns: '100px 1fr',
                gap: 8,
                padding: '2px 0',
                fontFamily: 'ui-monospace, SFMono-Regular, Menlo, monospace',
                fontSize: 12,
              }}
            >
              <span style={{ color: '#94a3b8' }}>{k}</span>
              <span style={{ color: 'var(--fg)' }}>{String(v)}</span>
            </div>
          ))}
        </div>
      ) : null}
    </div>
  )
}
