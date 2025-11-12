export const processMarkdownContent = content =>
  content
    .replace(/^---[\s\S]*?---\s*/, '')
    .split('\n')
    .filter(line => {
      const trimmed = line.trim()
      return !(
        /^(import\s+.*from\s+['"](@components\/|@assets\/|@astrojs\/starlight\/components).*['"];?|export\s+(default|const|let|function|class)\b)/.test(
          trimmed
        ) ||
        trimmed.startsWith('<Tab') ||
        trimmed.startsWith('</Tab')
      )
    })
    .join('\n')
    .trim()

const ensureTrailingSlash = (value = '') =>
  value.endsWith('/') ? value : `${value}/`

const isHttpUrl = url => url.startsWith('http://') || url.startsWith('https://')

export const toMarkdownUrl = (href, docsBaseUrl) => {
  if (!href) return href

  const base = ensureTrailingSlash(docsBaseUrl || 'https://daytona.io/docs')
  let baseUrl
  let targetUrl

  try {
    baseUrl = new URL(base)
    targetUrl = new URL(href, baseUrl)
  } catch {
    return href
  }

  if (!isHttpUrl(targetUrl.href)) return href
  if (!targetUrl.href.startsWith(baseUrl.href)) return href
  if (/\.mdx?$/i.test(targetUrl.pathname)) return targetUrl.href

  let docPath = targetUrl.pathname.slice(baseUrl.pathname.length)
  docPath = docPath.replace(/\/+/g, '/').replace(/\/+$/, '')

  if (!docPath) docPath = 'index'
  if (docPath.endsWith('/index')) {
    const trimmed = docPath.slice(0, -'/index'.length)
    docPath = trimmed || 'index'
  }

  const mdPath = docPath === 'index' ? 'index.md' : `${docPath}.md`
  const anchor = targetUrl.hash || ''

  return `${baseUrl.origin}${baseUrl.pathname}${mdPath}${anchor}`
}

export const rewriteLinksToMd = (markdown, docsBaseUrl) => {
  if (!markdown) return markdown

  return markdown.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (match, text, href) => {
    const trimmedHref = href.trim()

    if (
      !trimmedHref ||
      trimmedHref.startsWith('#') ||
      trimmedHref.startsWith('mailto:') ||
      trimmedHref.startsWith('tel:')
    ) {
      return match
    }

    const rewritten = toMarkdownUrl(trimmedHref, docsBaseUrl)
    if (!rewritten || rewritten === trimmedHref) {
      return match
    }

    return `[${text}](${rewritten})`
  })
}

export default rewriteLinksToMd
