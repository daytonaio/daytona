import { defineMiddleware } from 'astro:middleware'

import {
  loadDocsMarkdownBody,
  parseDocsContentPath,
  preferredMarkdownPlainFormat,
  shouldTryMarkdownPath,
} from './utils/acceptMarkdownNegotiation'
import { redirects } from './utils/redirects'

export const onRequest = defineMiddleware(
  async ({ request, redirect }, next) => {
    const url = new URL(request.url)
    const path = url.pathname.replace(/\/$/, '')

    if (path === '/docs/sitemap.xml') {
      return next(new Request(new URL('/docs/sitemap-index.xml', url), request))
    }

    // Match /docs/old-slug or /docs/{locale}/old-slug
    const match = path.match(/^\/docs(?:\/([a-z]{2}))?\/(.+)$/)
    if (match) {
      const locale = match[1]
      const slug = match[2]
      const newSlug = redirects[slug]
      if (newSlug) {
        const target = locale
          ? `/docs/${locale}/${newSlug}`
          : `/docs/${newSlug}`
        return redirect(target, 301)
      }
    }

    const textFormat = preferredMarkdownPlainFormat(
      request.headers.get('accept')
    )
    if (
      textFormat &&
      (request.method === 'GET' || request.method === 'HEAD') &&
      shouldTryMarkdownPath(url.pathname)
    ) {
      const parsed = parseDocsContentPath(url.pathname)
      if (parsed) {
        const body = await loadDocsMarkdownBody(parsed)
        if (body !== null) {
          const contentType =
            textFormat === 'plain'
              ? 'text/plain; charset=utf-8'
              : 'text/markdown; charset=utf-8'
          const headers = {
            'Content-Type': contentType,
            'Cache-Control': 'public, max-age=300',
            Vary: 'Accept',
          } as const
          if (request.method === 'HEAD') {
            return new Response(null, { status: 200, headers })
          }
          return new Response(body, {
            status: 200,
            headers,
          })
        }
      }
    }

    return next()
  }
)
