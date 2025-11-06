import type { APIRoute } from 'astro'
import { getEntry } from 'astro:content'

import config from '../../gt.config.json'
import { processMarkdownContent, rewriteLinksToMd } from '../utils/md'

export const prerender = false

const markdownHeaders = {
  'Content-Type': 'text/markdown; charset=utf-8',
}

const DEFAULT_LOCALE = config.defaultLocale
const SUPPORTED_LOCALES = new Set([config.defaultLocale, ...config.locales])

const tryGetEntry = async (slug: string) => {
  try {
    return await getEntry('docs', slug)
  } catch {
    return null
  }
}

const normalizeSlug = (slug: string): string => {
  if (!slug || slug === '/') return 'index'
  const sanitized = slug
    .split('/')
    .map(segment => segment.trim())
    .filter(segment => segment && segment !== '.' && segment !== '..')
    .join('/')
    .replace(/\/+$/, '')

  return sanitized || 'index'
}

export const GET: APIRoute = async ({ params, request, redirect }) => {
  const slugParam = params.slug ?? ''
  const segments = slugParam ? slugParam.split('/') : []

  if (!segments.length) {
    return redirect(`/docs/${DEFAULT_LOCALE}/index.md`, 302)
  }

  const firstSegment = segments[0]
  let locale = DEFAULT_LOCALE
  let contentSegments = segments

  if (SUPPORTED_LOCALES.has(firstSegment)) {
    locale = firstSegment
    contentSegments = segments.slice(1)
  }

  if (locale !== DEFAULT_LOCALE) {
    const redirectSlug = contentSegments.join('/') || 'index'
    return redirect(`/docs/${DEFAULT_LOCALE}/${redirectSlug}.md`, 302)
  }

  const normalizedSlug = normalizeSlug(contentSegments.join('/') || 'index')

  const entry =
    (await tryGetEntry(`${locale}/${normalizedSlug}`)) ??
    (normalizedSlug !== 'index'
      ? await tryGetEntry(`${locale}/${normalizedSlug}/index`)
      : await tryGetEntry(locale))

  if (!entry || !entry.body) {
    return new Response('Not Found', { status: 404 })
  }

  const rawContent =
    typeof entry.body === 'string' ? entry.body : String(entry.body)
  const cleanedContent = processMarkdownContent(rawContent)
  const origin = new URL(request.url)
  const docsBaseUrl = `${origin.origin}/docs`
  const rewrittenContent = rewriteLinksToMd(cleanedContent, docsBaseUrl)

  return new Response(rewrittenContent, {
    status: 200,
    headers: markdownHeaders,
  })
}
