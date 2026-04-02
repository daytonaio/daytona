import fsSync from 'node:fs'
import fs from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import config from '../../gt.config.json'
import { processMarkdownContent, rewriteLinksToMd } from './md.js'

const defaultLocale = config.defaultLocale as string
const locales = new Set<string>([defaultLocale, ...config.locales])

let docsProjectRootCache: string | null = null

/**
 * Resolves the docs app root (folder containing src/content/docs). Nx or other runners may use
 * monorepo cwd; production uses dist/apps/docs without src (prebuilt client/*.md only).
 */
function getDocsProjectRoot(): string {
  if (docsProjectRootCache) return docsProjectRootCache
  const candidates = [
    process.cwd(),
    path.join(process.cwd(), 'apps', 'docs'),
    path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..'),
  ]
  for (const root of candidates) {
    if (fsSync.existsSync(path.join(root, 'src', 'content', 'docs'))) {
      docsProjectRootCache = root
      return root
    }
  }
  docsProjectRootCache = process.cwd()
  return docsProjectRootCache
}

function docsBaseUrl(): string {
  const base = (import.meta.env.PUBLIC_WEB_URL || 'https://daytona.io').replace(
    /\/$/,
    ''
  )
  return `${base}/docs`
}

type AcceptEntry = { type: string; q: number; index: number }

function parseAcceptHeader(accept: string): AcceptEntry[] {
  return accept.split(',').map((part, index) => {
    const bits = part
      .trim()
      .split(';')
      .map(s => s.trim())
    const type = (bits[0] || '').toLowerCase()
    let q = 1
    for (const p of bits.slice(1)) {
      if (p.startsWith('q=')) {
        const n = Number.parseFloat(p.slice(2))
        if (!Number.isNaN(n)) q = n
      }
    }
    return { type, q, index }
  })
}

/**
 * Returns whether the client prefers a raw text markdown or plain response over HTML,
 * using q-values and header order (RFC 7231 style).
 */
export function preferredMarkdownPlainFormat(
  acceptHeader: string | null
): 'markdown' | 'plain' | null {
  if (!acceptHeader?.trim()) return null

  const entries = parseAcceptHeader(acceptHeader)
    .filter(e => e.q > 0)
    .sort((a, b) => b.q - a.q || a.index - b.index)

  for (const { type } of entries) {
    if (type === 'text/markdown' || type === 'text/x-markdown')
      return 'markdown'
    if (type === 'text/plain') return 'plain'
    if (type === 'text/html' || type === 'application/xhtml+xml') return null
  }

  return null
}

export function shouldTryMarkdownPath(pathname: string): boolean {
  if (!pathname.startsWith('/docs')) return false
  if (pathname.startsWith('/docs/_astro')) return false

  const trimmed = pathname.replace(/\/$/, '') || pathname
  const last = trimmed.split('/').pop() || ''
  if (last.includes('.') && !last.endsWith('/')) {
    return false
  }

  return true
}

export type ParsedDocsPath = {
  locale: string
  relKey: string
}

/**
 * Maps /docs[/locale]/[...rest] to locale and a slash-separated content key (no leading slash).
 * Paths without a locale segment are treated as default locale (matches static English export).
 */
function isSafeDocRelKey(relKey: string): boolean {
  if (!relKey) return true
  return !relKey.split('/').some(s => s === '..' || s === '.')
}

export function parseDocsContentPath(pathname: string): ParsedDocsPath | null {
  const raw = pathname.replace(/\/$/, '') || pathname
  if (!raw.startsWith('/docs')) return null

  let rest = raw.slice('/docs'.length).replace(/^\/+/, '')
  if (!rest) {
    return { locale: defaultLocale, relKey: '' }
  }

  const segments = rest.split('/').filter(Boolean)
  const first = segments[0]
  if (!first) {
    return { locale: defaultLocale, relKey: '' }
  }

  if (locales.has(first)) {
    const relKey = segments.slice(1).join('/')
    if (!isSafeDocRelKey(relKey)) return null
    return { locale: first, relKey }
  }

  const relKey = segments.join('/')
  if (!isSafeDocRelKey(relKey)) return null
  return { locale: defaultLocale, relKey }
}

function clientMarkdownPath(locale: string, relKey: string): string {
  const fileName = relKey ? `${relKey}.md` : 'docs.md'
  return path.join(getDocsProjectRoot(), 'client', locale, fileName)
}

function sourceDocCandidates(locale: string, relKey: string): string[] {
  const base = path.join(getDocsProjectRoot(), 'src/content/docs', locale)
  if (!relKey) {
    return [path.join(base, 'index.mdx'), path.join(base, 'index.md')]
  }
  const nested = relKey.split('/').join(path.sep)
  return [
    path.join(base, `${nested}.mdx`),
    path.join(base, `${nested}.md`),
    path.join(base, nested, 'index.mdx'),
    path.join(base, nested, 'index.md'),
  ]
}

async function readFirstExisting(paths: string[]): Promise<string | null> {
  for (const p of paths) {
    try {
      const buf = await fs.readFile(p, 'utf8')
      return buf
    } catch {
      /* try next */
    }
  }
  return null
}

export async function loadDocsMarkdownBody(
  parsed: ParsedDocsPath
): Promise<string | null> {
  const { locale, relKey } = parsed

  const prebuilt = await readFirstExisting([clientMarkdownPath(locale, relKey)])
  if (prebuilt !== null) {
    return prebuilt
  }

  const raw = await readFirstExisting(sourceDocCandidates(locale, relKey))
  if (raw === null) {
    return null
  }

  const cleaned = processMarkdownContent(raw)
  return rewriteLinksToMd(cleaned, docsBaseUrl())
}
