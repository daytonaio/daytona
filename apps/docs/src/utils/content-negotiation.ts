export function acceptsMarkdown(acceptHeader: string): boolean {
  for (const entry of acceptHeader.split(',')) {
    const [mediaType, ...params] = entry.split(';')
    if (mediaType.trim().toLowerCase() !== 'text/markdown') continue

    const qParam = params
      .map(p => p.trim().toLowerCase())
      .find(p => p.startsWith('q='))
    if (!qParam) return true

    const q = Number.parseFloat(qParam.slice(2))
    return Number.isNaN(q) || q > 0
  }
  return false
}

function withVaryAccept(headers: Headers): Headers {
  const result = new Headers(headers)
  const vary = result.get('vary')
  if (!vary) {
    result.set('vary', 'Accept')
    return result
  }

  const fields = vary.split(',').map(f => f.trim().toLowerCase())
  if (!fields.includes('*') && !fields.includes('accept')) {
    result.set('vary', `${vary}, Accept`)
  }
  return result
}

export function rewrite404ForMarkdownAccept(
  response: Response,
  request: Request
): Response {
  if (response.status !== 404) return response

  const wantsMarkdown = acceptsMarkdown(request.headers.get('accept') ?? '')
  return new Response(response.body, {
    status: wantsMarkdown ? 200 : 404,
    statusText: wantsMarkdown ? 'OK' : response.statusText,
    headers: withVaryAccept(response.headers),
  })
}
