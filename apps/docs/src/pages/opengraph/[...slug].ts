import type { APIRoute } from 'astro'
import { getEntry } from 'astro:content'

import config from '../../../gt.config.json'
import { generateDocsOgImagePng } from '../../lib/generate-docs-og-image'

export const prerender = false

const localeSet = new Set(config.locales)

export const GET: APIRoute = async ({ params }) => {
  const slug = params.slug
  if (!slug || typeof slug !== 'string') {
    return new Response('Bad Request', { status: 400 })
  }

  const firstSegment = slug.split('/')[0] ?? ''
  const hasLocalePrefix = localeSet.has(firstSegment)

  let entry = await getEntry('docs', slug)
  if (!entry && !hasLocalePrefix) {
    entry = await getEntry('docs', `${config.defaultLocale}/${slug}`)
  }
  if (!entry) {
    return new Response('Not Found', { status: 404 })
  }

  const title =
    typeof entry.data.title === 'string' && entry.data.title.trim()
      ? entry.data.title.trim()
      : slug
  const png = await generateDocsOgImagePng({ title })

  return new Response(new Uint8Array(png), {
    headers: {
      'Content-Type': 'image/png',
      'Cache-Control': 'public, max-age=86400, s-maxage=86400',
    },
  })
}
