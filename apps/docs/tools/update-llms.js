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
const SUBFOLDERS = new Set([
  'about',
  'configuration',
  'installation',
  'misc',
  'usage',
  'tools',
  'sdk',
])
const EXCLUDE_FILES = new Set(['404.md', 'api.mdx'])

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

const parseMarkdownFile = filePath => {
  const { content, data } = matter(fs.readFileSync(filePath, 'utf8'))
  const cleanContent = processMarkdownContent(content)
  const title = data.title || cleanContent.match(/^#\s+(.*)/)?.[1] || 'Untitled'
  const slug = filePath
    .replace(DOCS_PATH, '')
    .replace(/\\/g, '/')
    .replace(/\.mdx?$/, '')

  return [
    { title, url: `/docs/en${slug}` },
    ...extractSubHeadings(cleanContent, slug),
  ]
}

const searchDocs = () => {
  const results = []
  const fullContentArray = []

  const traverseDirectory = directory => {
    fs.readdirSync(directory).forEach(file => {
      const fullPath = path.join(directory, file)
      const stat = fs.statSync(fullPath)

      if (stat.isDirectory() && SUBFOLDERS.has(path.basename(fullPath))) {
        traverseDirectory(fullPath)
      } else if (
        stat.isFile() &&
        (file.endsWith('.md') || file.endsWith('.mdx')) &&
        !EXCLUDE_FILES.has(file)
      ) {
        const fileContent = fs.readFileSync(fullPath, 'utf8')
        const cleanContent = processMarkdownContent(fileContent)
        fullContentArray.push(cleanContent)
        results.push(...parseMarkdownFile(fullPath))
      }
    })
  }

  traverseDirectory(DOCS_PATH)
  return { results, fullContent: fullContentArray.join('\n\n') }
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

const main = () => {
  const { results, fullContent } = searchDocs()
  generateLlmsTxtFile(results)
  generateLlmsFullTxtFile(fullContent)
}

main()
