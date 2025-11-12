import fs from 'fs'
import matter from 'gray-matter'
import path, { dirname } from 'path'
import { fileURLToPath } from 'url'

import { processMarkdownContent, rewriteLinksToMd } from '../src/utils/md.js'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

const PUBLIC_WEB_URL = (
  process.env.PUBLIC_WEB_URL || 'https://daytona.io'
).replace(/\/$/, '')
const DOCS_BASE_URL = `${PUBLIC_WEB_URL}/docs`

const BUILD_OUTPUT_DIR = path.join(
  __dirname,
  '../../../dist/apps/docs/dist/client'
)
const STATIC_LOCALE = 'en'
const STATIC_DOCS_OUTPUT_DIR = path.join(BUILD_OUTPUT_DIR, STATIC_LOCALE)
const STATIC_DOCS_OUTPUT_DIR_ROOT = BUILD_OUTPUT_DIR

const packageJson = JSON.parse(
  fs.readFileSync(path.join(__dirname, '../../../package.json'), 'utf8')
)
const version = packageJson.version

const getCurrentDate = () => {
  const date = new Date()
  return date.toISOString().split('T')[0] // YYYY-MM-DD format
}

const getVersionHeader = () => {
  return [
    `# Daytona Documentation v${version}`,
    `# Generated on: ${getCurrentDate()}`,
    '',
  ].join('\n')
}

// Only include English docs
const DOCS_PATH = path.join(__dirname, '../src/content/docs/en')
const EXCLUDE_FILES = new Set(['404.md', 'api.mdx'])

const ensureLeadingSlash = value =>
  value.startsWith('/') ? value : `/${value}`

const normalizeSlugPath = slug => {
  if (!slug) return '/'

  let value = ensureLeadingSlash(slug.replace(/\\/g, '/'))
  value = value.replace(/\/+/g, '/')

  if (value.endsWith('/index')) {
    value = value.slice(0, -'/index'.length) || '/'
  }

  if (value !== '/' && value.endsWith('/')) {
    value = value.slice(0, -1)
  }

  return value || '/'
}

const getOutputRelativeSlug = normalizedSlug =>
  !normalizedSlug || normalizedSlug === '/'
    ? 'docs'
    : normalizedSlug.replace(/^\//, '')

const getSlugFromFilePath = filePath =>
  filePath
    .replace(DOCS_PATH, '')
    .replace(/\\/g, '/')
    .replace(/\.mdx?$/, '')

const extractSubHeadings = (content, slug) => {
  const headingRegex = /^(#{2,3})\s+(.*)/gm
  const headings = []
  let match

  while ((match = headingRegex.exec(content)) !== null) {
    const headingSlug = `${slug}#${match[2]
      .toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9-]/g, '')}`
    headings.push({ title: match[2].trim(), url: `/docs/en${headingSlug}` })
  }

  return headings
}

const parseMarkdownFile = (filePath, fileContent) => {
  const { content, data } = matter(fileContent)
  const cleanContent = processMarkdownContent(content)
  const rewrittenContent = rewriteLinksToMd(cleanContent, DOCS_BASE_URL)
  const slug = getSlugFromFilePath(filePath)
  const normalizedSlug = normalizeSlugPath(slug)
  const title =
    data.title || cleanContent.match(/^#\s+(.*)/)?.[1]?.trim() || 'Untitled'

  return {
    entries: [
      {
        title,
        url: normalizedSlug === '/' ? '/docs/en' : `/docs/en${normalizedSlug}`,
      },
      ...extractSubHeadings(cleanContent, normalizedSlug),
    ],
    doc: {
      slug: normalizedSlug,
      content: rewrittenContent,
    },
    cleanContent,
  }
}

const searchDocs = () => {
  const results = []
  const docs = []
  const fullContentArray = []

  const traverseDirectory = directory => {
    fs.readdirSync(directory).forEach(file => {
      const fullPath = path.join(directory, file)
      const stat = fs.statSync(fullPath)

      if (stat.isDirectory()) {
        traverseDirectory(fullPath)
      } else if (
        stat.isFile() &&
        (file.endsWith('.md') || file.endsWith('.mdx')) &&
        !EXCLUDE_FILES.has(file)
      ) {
        const fileContent = fs.readFileSync(fullPath, 'utf8')
        const { entries, doc, cleanContent } = parseMarkdownFile(
          fullPath,
          fileContent
        )

        fullContentArray.push(cleanContent)
        results.push(...entries)
        docs.push(doc)
      }
    })
  }

  traverseDirectory(DOCS_PATH)
  return {
    results,
    docs,
    fullContent: fullContentArray.join('\n\n'),
  }
}

const generateLlmsTxtFile = docsData => {
  const llmsContent = [
    getVersionHeader(),
    '# Daytona',
    '',
    '> Secure and Elastic Infrastructure for Running Your Al-Generated Code.',
    '',
    '## Docs',
    '',
    ...docsData.map(doc => `- [${doc.title}](${PUBLIC_WEB_URL}${doc.url})`),
  ]
  const output = rewriteLinksToMd(llmsContent.join('\n'), DOCS_BASE_URL)
  fs.writeFileSync(
    path.join(__dirname, '../../../dist/apps/docs/dist/client/llms.txt'),
    output,
    'utf8'
  )
  console.log('llms.txt index updated')
}

const generateLlmsFullTxtFile = fullContent => {
  const content = [getVersionHeader(), fullContent].join('\n\n')
  const output = rewriteLinksToMd(content, DOCS_BASE_URL)
  fs.writeFileSync(
    path.join(__dirname, '../../../dist/apps/docs/dist/client/llms-full.txt'),
    output,
    'utf8'
  )
  console.log('llms-full.txt index updated')
}

const generateStaticMarkdownFiles = docs => {
  const writeMarkdownFile = (outputPath, content) => {
    fs.mkdirSync(path.dirname(outputPath), { recursive: true })
    fs.writeFileSync(outputPath, content, 'utf8')
    console.log(
      `Static markdown page generated for "${path.relative(
        STATIC_DOCS_OUTPUT_DIR_ROOT,
        outputPath
      )}"`
    )
  }

  docs.forEach(({ slug, content }) => {
    const relativeSlug = getOutputRelativeSlug(slug)
    const localeOutputPath = path.join(
      STATIC_DOCS_OUTPUT_DIR,
      `${relativeSlug}.md`
    )
    const rootOutputPath = path.join(
      STATIC_DOCS_OUTPUT_DIR_ROOT,
      `${relativeSlug}.md`
    )

    writeMarkdownFile(localeOutputPath, content)

    if (rootOutputPath !== localeOutputPath) {
      writeMarkdownFile(rootOutputPath, content)
    }

    if (slug === '/' || slug === '') {
      const rootLocaleAlias = path.join(
        STATIC_DOCS_OUTPUT_DIR_ROOT,
        `${STATIC_LOCALE}.md`
      )
      writeMarkdownFile(rootLocaleAlias, content)
    }
  })
}

const main = () => {
  const { results, fullContent, docs } = searchDocs()
  generateLlmsTxtFile(results)
  generateLlmsFullTxtFile(fullContent)
  generateStaticMarkdownFiles(docs)
}

main()
