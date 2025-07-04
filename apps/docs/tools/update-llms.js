import fs from 'fs'
import matter from 'gray-matter'
import path, { dirname } from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

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

const DOCS_PATH = path.join(__dirname, '../src/content/docs')
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

const processContent = content =>
  content
    .split('\n')
    .filter(line => {
      const trimmed = line.trim()
      return !(
        /^(import\s+.*from\s+['"](@components\/|@assets\/).*['"];?|export\s+(default|const|let|function|class)\b)/.test(
          trimmed
        ) ||
        trimmed.startsWith('<Tab') ||
        trimmed.startsWith('</Tab') ||
        trimmed.startsWith('---')
      )
    })
    .join('\n')
    .trim()

const extractSubHeadings = (content, slug) => {
  const headingRegex = /^(#{2,3})\s+(.*)/gm
  const headings = []
  let match

  while ((match = headingRegex.exec(content)) !== null) {
    const headingSlug = `${slug}#${match[2]
      .toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9-]/g, '')}`
    headings.push({ title: match[2].trim(), url: `/docs${headingSlug}` })
  }

  return headings
}

const parseMarkdownFile = filePath => {
  const { content, data } = matter(fs.readFileSync(filePath, 'utf8'))
  const cleanContent = processContent(content)
  const title = data.title || cleanContent.match(/^#\s+(.*)/)?.[1] || 'Untitled'
  const slug = filePath
    .replace(DOCS_PATH, '')
    .replace(/\\/g, '/')
    .replace(/\.mdx?$/, '')

  return [
    { title, url: `/docs${slug}` },
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
        const cleanContent = processContent(fileContent)
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
    ...docsData.map(doc => `- [${doc.title}](https://daytona.io${doc.url})`),
  ]
  fs.writeFileSync(
    path.join(__dirname, '../public/llms.txt'),
    llmsContent.join('\n'),
    'utf8'
  )
  console.log('llms.txt index updated')
}

const generateLlmsFullTxtFile = fullContent => {
  const content = [getVersionHeader(), fullContent].join('\n\n')

  fs.writeFileSync(
    path.join(__dirname, '../public/llms-full.txt'),
    content,
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
