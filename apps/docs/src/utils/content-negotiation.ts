export function acceptsMarkdown(acceptHeader: string): boolean {
  return acceptHeader.toLowerCase().includes('text/markdown')
}

export function rewrite404ForMarkdownAccept(
  response: Response,
  request: Request
): Response {
  if (response.status !== 404) return response
  if (!acceptsMarkdown(request.headers.get('accept') ?? '')) return response

  return new Response(response.body, {
    status: 200,
    statusText: 'OK',
    headers: response.headers,
  })
}
