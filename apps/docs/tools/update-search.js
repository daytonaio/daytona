import fs from 'fs'
import matter from 'gray-matter'
import path from 'path'
import { dirname } from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

const DOCS_PATH = path.join(__dirname, '../src/content/docs/en')
const SDK_FOLDER_NAMES = []
const SDK_FOLDERS = fs.readdirSync(DOCS_PATH)
  .filter(directoryName => {
    const fullPath = path.join(DOCS_PATH, directoryName)
    return fs.statSync(fullPath).isDirectory() && directoryName.endsWith('-sdk')
  }).map((directoryName) => {
    SDK_FOLDER_NAMES.push(directoryName)
    return path.join(DOCS_PATH, directoryName)
  })
const CLI_FILE_PATH = path.join(DOCS_PATH, 'tools/cli.mdx')
const EXCLUDE_FILES = ['404.md', 'index.mdx', 'api.mdx']

function processContent(content) {
  return content
    .split('\n')
    .filter(
      line =>
        !line.trim().startsWith('import') && !line.trim().startsWith('export')
    )
    .join('\n')
    .trim()
}

function extractSentences(text) {
  // return text
  const match = text.match(/[^.!?]*[.!?]/g)
  const sentences = match ? match.map(m => m.trim()) : text.trim().split('\n')
  return sentences.length > 0
    ? sentences.filter(s => s.endsWith('.')).join(' ')
    : ''
}

function extractHyperlinks(text) {
  return text.replace(/\[([^\]]+)\]\(([^\)]+)\)/g, '$1')
}

function isSentence(sentence) {
  return /^[A-Z0-9\[]/.test(sentence.trim())
}

function isSentenceWithoutPunctuation(sentence) {
  const trimmed = sentence.trim()
  return /^[A-Z0-9\[]/.test(trimmed) && !trimmed.includes('\n') && !/[.!?]$/.test(trimmed)
}

function extractRealSentence(text) {
  const sentences = text
    .split(/\n\n+/)
    .map(s => s.trim())
    .filter(s => s.length > 0)
  for (const sentence of sentences) {
    if (isSentence(sentence)) {
      const extracted = extractSentences(sentence)
      if (extracted) {
        return extracted
      }
      if (isSentenceWithoutPunctuation(sentence)) {
        return sentence
      }
    }
  }
  return ''
}

function extractCodeSnippets(content) {
  const codeRegexes = [
    /```(?:python|py)\n([\s\S]*?)```/g,
    /```(?:typescript|ts|tsx)\n([\s\S]*?)```/g,
    /```(?:bash|shell|sh)\n([\s\S]*?)```/g
  ]
  const codeSnippets = []
  
  codeRegexes.forEach(regex => {
    let match
    while ((match = regex.exec(content)) !== null) {
      const code = match[1].trim()
      if (code)
        codeSnippets.push(code)
    }
  })
  
  return codeSnippets.join('\n\n')
}

function extractHeadings(content, tag, slug) {
  // First, temporarily replace code blocks with placeholders to avoid matching # inside code
  const codeBlockRegex = /```[\s\S]*?```/g
  const codeBlocks = []
  let codeBlockIndex = 0
  
  const contentWithoutCode = content.replace(codeBlockRegex, (match) => {
    const placeholder = `___CODE_BLOCK_${codeBlockIndex++}___`
    codeBlocks.push(match)
    return placeholder
  })
  
  // Extract headings from content without code blocks
  const headingRegex = /^(#{1,6})\s+(.+)$/gm
  const headings = []
  const headingMatches = []
  let match
  
  // Collect all heading positions from content WITHOUT code blocks -> we use them later to restore code blocks
  while ((match = headingRegex.exec(contentWithoutCode)) !== null) {
    headingMatches.push({
      title: match[2].trim().replace(/\\_/g, '_'),
      index: match.index,
      length: match[0].length
    })
  }
  
  // Process each heading content
  for (let i = 0; i < headingMatches.length; i++) {
    const current = headingMatches[i]
    const next = headingMatches[i + 1]
    
    //Content below current heading and the next heading (or end)
    const startIndex = current.index + current.length
    const endIndex = next ? next.index : contentWithoutCode.length
    let currentTextBelow = contentWithoutCode.substring(startIndex, endIndex)
    
    // Restore code blocks
    currentTextBelow = currentTextBelow.replace(/___CODE_BLOCK_(\d+)___/g, (match, index) => {
      return codeBlocks[parseInt(index)]
    })
    
    currentTextBelow = currentTextBelow.trim()
    
    const heading = current.title
    const description = extractHyperlinks(extractRealSentence(currentTextBelow))
    const codeSnippets = extractCodeSnippets(currentTextBelow)
    const headingSlug = `${slug}#${heading
      .toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9-_]/g, '')}`
    
    headings.push({
      title: heading,
      description,
      codeSnippets,
      tag,
      url: `/docs${headingSlug}`,
      slug: headingSlug,
    })
  }

  return headings
}

function parseMarkdownFile(filePath, tag) {
  const fileContent = fs.readFileSync(filePath, 'utf8')
  const { content, data } = matter(fileContent)

  const cleanContent = processContent(content)

  const title = data.title || cleanContent.match(/^#\s+(.*)/)?.[1] || 'Untitled'
  const description = data.description || extractHyperlinks(extractRealSentence(cleanContent))
  const slug = filePath
    .replace(DOCS_PATH, '')
    .replace(/\\/g, '/')
    .replace(/\.mdx?$/, '')
  const headings = extractHeadings(cleanContent, tag, slug)

  const mainData = {
    title,
    description,
    tag,
    url: `/docs${slug}`,
    slug,
  }

  return [mainData, ...headings]
}

function searchDocs() {
  const docsRecords = []
  const cliRecords = []
  const sdkRecords = []
  let objectID = 1

  function traverseDirectory(directory, tag, recordsData) {
    const files = fs.readdirSync(directory)

    files.forEach(file => {
      const fullPath = path.join(directory, file)
      const stat = fs.statSync(fullPath)

      if (stat.isDirectory()) {
        const directoryName = path.basename(fullPath)
        switch (tag) {
          case 'Documentation':
            if (SDK_FOLDER_NAMES.includes(directoryName))
              return
            break
          case 'SDK':
            // For SDK we traverse all subfolders inside given initial SDK directory
            break
        }
        // Traverse directory if we passed the checks
        traverseDirectory(fullPath, tag, recordsData)
      } else if (
        stat.isFile() &&
        (file.endsWith('.md') || file.endsWith('.mdx')) &&
        ![...EXCLUDE_FILES, path.basename(CLI_FILE_PATH)].includes(file) //CLI file is handled separately -> exclude it from directory traversals
      ) {
        parseMarkdownFile(fullPath, tag).forEach(data => recordsData.push({
            ...data,
            objectID: objectID++,
          }))
      }
    })
  }

  traverseDirectory(DOCS_PATH, 'Documentation', docsRecords)

  SDK_FOLDERS.forEach((sdkFolderPath) => traverseDirectory(sdkFolderPath, 'SDK', sdkRecords))

  parseMarkdownFile(CLI_FILE_PATH, 'CLI').forEach(data => cliRecords.push({
      ...data,
      objectID: objectID++,
  }))

  return {docsRecords, cliRecords, sdkRecords}
}

function main() {
  const {docsRecords, cliRecords, sdkRecords} = searchDocs()
  const fileRecords = [
    {fileName: 'docs', records: docsRecords}, 
    {fileName: 'cli', records: cliRecords}, 
    {fileName: 'sdk', records: sdkRecords}
  ]
  
  fileRecords.forEach(({fileName, records}) => 
    fs.writeFileSync(
    path.join(__dirname, `../../../dist/apps/docs/dist/client/${fileName}.json`),
    JSON.stringify(records, null, 2),
    'utf8'
  ))
  console.log('search index updated')
}

if (import.meta.url === `file://${process.argv[1]}`) {
  main()
}
