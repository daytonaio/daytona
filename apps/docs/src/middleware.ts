import { defineMiddleware } from 'astro:middleware'

import { redirects } from './utils/redirects'

function filterProxyHeaders(headers: Headers): Headers {
  const filteredHeaders = new Headers()
  for (const [key, value] of headers.entries()) {
    if (
      !['content-encoding', 'content-length', 'transfer-encoding'].includes(
        key.toLowerCase()
      )
    ) {
      filteredHeaders.set(key, value)
    }
  }
  return filteredHeaders
}

export const onRequest = defineMiddleware(
  async ({ request, redirect }, next) => {
    const url = new URL(request.url)
    const path = url.pathname.replace(/\/$/, '')

    const proxyRequest = async (targetUrl: URL): Promise<Response> => {
      const response = await fetch(targetUrl.toString(), {
        method: request.method,
        body:
          request.method === 'GET' || request.method === 'HEAD'
            ? undefined
            : request.body,
      })

      return new Response(response.body, {
        status: response.status,
        statusText: response.statusText,
        headers: filterProxyHeaders(response.headers),
      })
    }

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

    if (path === '/docs') {
      const targetUrl = new URL('/docs/en', url)
      targetUrl.search = url.search
      return await proxyRequest(targetUrl)
    }

    const docsPath = path.match(/^\/docs\/(.+)$/)
    if (docsPath) {
      const slug = docsPath[1]
      const firstSegment = slug.split('/')[0]
      const isLocalePrefixed = /^[a-z]{2}$/.test(firstSegment)
      const looksLikeStaticAsset = /\.[^/]+$/.test(slug)

      if (!isLocalePrefixed && !looksLikeStaticAsset) {
        const targetUrl = new URL(`/docs/en/${slug}`, url)
        targetUrl.search = url.search
        return await proxyRequest(targetUrl)
      }
    }

    return next()
  }
)
