import { defineMiddleware } from 'astro:middleware'

import { redirects } from './utils/redirects'

export const onRequest = defineMiddleware(({ request, redirect }, next) => {
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
      const target = locale ? `/docs/${locale}/${newSlug}` : `/docs/${newSlug}`
      return redirect(target, 301)
    }
  }

  return next()
})
