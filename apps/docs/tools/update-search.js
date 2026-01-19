import fs from 'fs'
import matter from 'gray-matter'
import path from 'path'
import { dirname } from 'path'
import { fileURLToPath } from 'url'
import https from 'https'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

const DOCS_PATH = path.join(__dirname, '../src/content/docs/en')
const SDK_FOLDER_NAMES = []
const SDK_FOLDERS = fs
  .readdirSync(DOCS_PATH)
  .filter(directoryName => {
    const fullPath = path.join(DOCS_PATH, directoryName)
    return fs.statSync(fullPath).isDirectory() && directoryName.endsWith('-sdk')
  })
  .map(directoryName => {
    SDK_FOLDER_NAMES.push(directoryName)
    return path.join(DOCS_PATH, directoryName)
  })
const CLI_FILE_PATH = path.join(DOCS_PATH, 'tools/cli.mdx')
const EXCLUDE_FILES = ['404.md', 'index.mdx', 'api.mdx']

function findWorkspaceRoot(startPath) {
  let current = path.resolve(startPath)
  const root = path.resolve(current, '/')

  while (current !== root) {
    const packageJson = path.join(current, 'package.json')
    const appsDocs = path.join(current, 'apps/docs')

    if (fs.existsSync(packageJson) && fs.existsSync(appsDocs)) {
      return current
    }

    const parent = path.resolve(current, '..')
    if (parent === current) break
    current = parent
  }

  return process.cwd()
}

const WORKSPACE_ROOT = findWorkspaceRoot(__dirname)
const MAIN_API_PATH = path.join(
  WORKSPACE_ROOT,
  'dist/apps/api/openapi.3.1.0.json'
)
const TOOLBOX_API_PATH = path.join(
  WORKSPACE_ROOT,
  'apps/daemon/pkg/toolbox/docs/swagger.json'
)

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
  // Return text
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
  return (
    /^[A-Z0-9\[]/.test(trimmed) &&
    !trimmed.includes('\n') &&
    !/[.!?]$/.test(trimmed)
  )
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
    /```(?:bash|shell|sh)\n([\s\S]*?)```/g,
  ]
  const codeSnippets = []

  codeRegexes.forEach(regex => {
    let match
    while ((match = regex.exec(content)) !== null) {
      const code = match[1].trim()
      if (code) codeSnippets.push(code)
    }
  })

  return codeSnippets.join('\n\n')
}

function extractHeadings(content, tag, slug) {
  // First, temporarily replace code blocks with placeholders to avoid matching # inside code
  const codeBlockRegex = /```[\s\S]*?```/g
  const codeBlocks = []
  let codeBlockIndex = 0

  const contentWithoutCode = content.replace(codeBlockRegex, match => {
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
      length: match[0].length,
    })
  }

  // Process each heading content
  for (let i = 0; i < headingMatches.length; i++) {
    const current = headingMatches[i]
    const next = headingMatches[i + 1]

    // Content below current heading and the next heading (or end)
    const startIndex = current.index + current.length
    const endIndex = next ? next.index : contentWithoutCode.length
    let currentTextBelow = contentWithoutCode.substring(startIndex, endIndex)

    // Restore code blocks
    currentTextBelow = currentTextBelow.replace(
      /___CODE_BLOCK_(\d+)___/g,
      (match, index) => {
        return codeBlocks[parseInt(index)]
      }
    )

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
  const description =
    data.description || extractHyperlinks(extractRealSentence(cleanContent))
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

function parseOpenAPISpec(specPath, apiName, baseUrl) {
  const records = []

  if (!fs.existsSync(specPath)) {
    console.warn(`OpenAPI spec not found: ${specPath}`)
    return records
  }

  try {
    const specContent = fs.readFileSync(specPath, 'utf8')
    const spec = JSON.parse(specContent)

    // Handle both OpenAPI 3.x and Swagger 2.0
    const isSwagger2 = spec.swagger === '2.0'
    const paths = spec.paths || {}
    const urlPrefix = baseUrl || '/docs/tools/api'

    Object.entries(paths).forEach(([path, pathItem]) => {
      if (!pathItem) return

      const operations = [
        'get',
        'post',
        'put',
        'delete',
        'patch',
        'head',
        'options',
      ]

      operations.forEach(method => {
        const operation = pathItem[method]
        if (!operation) return

        const operationId = operation.operationId || `${method}${path}`
        const summary = operation.summary || ''
        const description = operation.description || ''
        const tags = Array.isArray(operation.tags) ? operation.tags : []

        // Extract parameter descriptions
        const parameters = (operation.parameters || [])
          .map(param => {
            const paramDesc = param.description || ''
            const paramName = param.name || ''
            const paramIn = param.in || ''
            return `${paramIn} ${paramName}: ${paramDesc}`
          })
          .filter(Boolean)
          .join('; ')

        // Extract response descriptions
        const responseDescriptions = Object.entries(operation.responses || {})
          .map(([code, response]) => {
            const respDesc =
              response?.description ||
              (typeof response === 'string' ? response : '')
            return respDesc ? `${code}: ${respDesc}` : null
          })
          .filter(Boolean)
          .join('; ')

        // Build searchable content
        const searchableContent = [
          summary,
          description,
          ...tags,
          path,
          method.toUpperCase(),
          operationId,
          parameters,
          responseDescriptions,
        ]
          .filter(Boolean)
          .join(' ')

        // Create title (e.g., "GET /api/sandboxes - Create sandbox")
        const title = `${method.toUpperCase()} ${path}${summary ? ` - ${summary}` : ''}`

        // Build description from available fields
        const recordDescription =
          description ||
          summary ||
          `${method.toUpperCase()} endpoint for ${path}`

        // Create slug for linking
        const slug = `${path}${operationId ? `#${operationId}` : ''}`

        // Use first tag for URL, or fallback to path-based tag
        const primaryTag =
          tags.length > 0
            ? tags[0]
            : path.split('/').filter(Boolean)[0] || 'default'

        // For toolbox API, use 'daytona-toolbox' in the URL hash
        const hashApiName = apiName === 'toolbox' ? 'daytona-toolbox' : apiName

        // Format URL to match the hash format: #apiName/tag/tagName/METHOD/path
        const scalarHash = `${hashApiName}/tag/${primaryTag}/${method.toUpperCase()}${path}`

        records.push({
          title,
          description: recordDescription,
          tag: 'API',
          url: `${urlPrefix}#${scalarHash}`,
          slug: `api-${apiName}-${slug.replace(/\//g, '-').replace(/[^a-z0-9-]/gi, '')}`,
          apiName,
          method: method.toUpperCase(),
          path,
          operationId,
          tags: tags.join(', '),
          searchableContent,
        })
      })
    })
  } catch (error) {
    console.error(`Error parsing OpenAPI spec ${specPath}:`, error.message)
  }

  return records
}

function searchDocs() {
  const docsRecords = []
  const cliRecords = []
  const sdkRecords = []
  const apiRecords = []
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
            if (SDK_FOLDER_NAMES.includes(directoryName)) return
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
        parseMarkdownFile(fullPath, tag).forEach(data =>
          recordsData.push({
            ...data,
            objectID: objectID++,
          })
        )
      }
    })
  }

  traverseDirectory(DOCS_PATH, 'Documentation', docsRecords)

  SDK_FOLDERS.forEach(sdkFolderPath =>
    traverseDirectory(sdkFolderPath, 'SDK', sdkRecords)
  )

  parseMarkdownFile(CLI_FILE_PATH, 'CLI').forEach(data =>
    cliRecords.push({
      ...data,
      objectID: objectID++,
    })
  )

  const mainApiRecords = parseOpenAPISpec(
    MAIN_API_PATH,
    'daytona',
    '/docs/tools/api'
  )
  const toolboxApiRecords = parseOpenAPISpec(
    TOOLBOX_API_PATH,
    'toolbox',
    '/docs/tools/api'
  )

  const allApiRecords = mainApiRecords.concat(toolboxApiRecords)
  allApiRecords.forEach(data => {
    apiRecords.push({
      ...data,
      objectID: objectID++,
    })
  })

  return { docsRecords, cliRecords, sdkRecords, apiRecords }
}

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms))
}

function getEnv(name, defaultValue) {
  const v = process.env[name]
  return v === undefined || v === '' ? defaultValue : v
}

function loadDotEnvFileIfPresent(filePath) {
  if (!filePath) return
  if (!fs.existsSync(filePath)) return

  const content = fs.readFileSync(filePath, 'utf8')
  content
    .split('\n')
    .map(line => line.trim())
    .filter(line => line.length > 0 && !line.startsWith('#'))
    .forEach(line => {
      const eq = line.indexOf('=')
      if (eq <= 0) return
      const key = line.slice(0, eq).trim()
      if (!key) return
      if (process.env[key] !== undefined) return

      let value = line.slice(eq + 1).trim()
      if (
        (value.startsWith('"') && value.endsWith('"')) ||
        (value.startsWith("'") && value.endsWith("'"))
      ) {
        value = value.slice(1, -1)
      }
      process.env[key] = value
    })
}

function normalizeAlgoliaHost(hostOrUrl) {
  const value = String(hostOrUrl || '').trim()
  if (!value) return ''
  if (value.startsWith('http://') || value.startsWith('https://')) {
    try {
      return new URL(value).hostname
    } catch {
      return value.replace(/^https?:\/\//, '').split('/')[0]
    }
  }
  return value.split('/')[0]
}

function parseBoolEnv(value) {
  if (!value) return false
  return ['1', 'true', 'yes', 'y', 'on'].includes(String(value).toLowerCase())
}

function algoliaRequest({ host, appId, apiKey, method, requestPath, body }) {
  return new Promise((resolve, reject) => {
    const jsonBody = body === undefined ? undefined : JSON.stringify(body)
    const req = https.request(
      {
        hostname: host,
        method,
        path: requestPath,
        headers: {
          'x-algolia-application-id': appId,
          'x-algolia-api-key': apiKey,
          accept: 'application/json',
          ...(jsonBody
            ? {
                'content-type': 'application/json',
                'content-length': Buffer.byteLength(jsonBody),
              }
            : {}),
        },
      },
      res => {
        let data = ''
        res.setEncoding('utf8')
        res.on('data', chunk => (data += chunk))
        res.on('end', () => {
          const status = res.statusCode || 0
          const ok = status >= 200 && status < 300
          const parsed = data ? safeJsonParse(data) : null
          if (!ok) {
            const message =
              (parsed && parsed.message) || `Algolia request failed (${status})`
            const err = new Error(message)
            err.statusCode = status
            err.responseBody = parsed || data
            return reject(err)
          }
          resolve(parsed)
        })
      }
    )

    req.on('error', reject)
    if (jsonBody) req.write(jsonBody)
    req.end()
  })
}

function safeJsonParse(text) {
  try {
    return JSON.parse(text)
  } catch {
    return null
  }
}

async function algoliaRequestWithRetry(opts) {
  const maxAttempts = Number(getEnv('ALGOLIA_MAX_ATTEMPTS', '5'))
  const baseDelayMs = Number(getEnv('ALGOLIA_RETRY_BASE_DELAY_MS', '500'))

  let lastErr
  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      return await algoliaRequest(opts)
    } catch (err) {
      lastErr = err
      const status = err?.statusCode
      const retryable =
        status === 429 || (typeof status === 'number' && status >= 500)
      if (!retryable || attempt === maxAttempts) break
      const jitter = Math.floor(Math.random() * 250)
      const backoff = baseDelayMs * Math.pow(2, attempt - 1) + jitter
      await sleep(backoff)
    }
  }
  throw lastErr
}

function buildAlgoliaObjectID(fileName, record) {
  return `${fileName}:${record.slug}`
}

function chunkArray(arr, size) {
  const chunks = []
  for (let i = 0; i < arr.length; i += size) chunks.push(arr.slice(i, i + size))
  return chunks
}

async function uploadAlgoliaIndex({
  fileName,
  indexName,
  records,
  host,
  appId,
  apiKey,
}) {
  const clearFirst = parseBoolEnv(getEnv('ALGOLIA_CLEAR_INDEX', ''))
  const batchSize = Number(getEnv('ALGOLIA_BATCH_SIZE', '1000'))

  if (clearFirst) {
    await algoliaRequestWithRetry({
      host,
      appId,
      apiKey,
      method: 'POST',
      requestPath: `/1/indexes/${encodeURIComponent(indexName)}/clear`,
      body: {},
    })
  }

  const algoliaRecords = records.map(r => ({
    ...r,
    objectID: buildAlgoliaObjectID(fileName, r),
  }))

  const batches = chunkArray(algoliaRecords, batchSize)
  for (const batch of batches) {
    const actions = batch.map(obj => ({ action: 'updateObject', body: obj }))
    await algoliaRequestWithRetry({
      host,
      appId,
      apiKey,
      method: 'POST',
      requestPath: `/1/indexes/${encodeURIComponent(indexName)}/batch`,
      body: { requests: actions },
    })
  }
}

function main() {
  loadDotEnvFileIfPresent(path.join(__dirname, '../.env'))

  const { docsRecords, cliRecords, sdkRecords, apiRecords } = searchDocs()
  const fileRecords = [
    { fileName: 'docs', records: docsRecords },
    { fileName: 'cli', records: cliRecords },
    { fileName: 'sdk', records: sdkRecords },
    { fileName: 'api', records: apiRecords },
  ]

  const outputDir = path.join(__dirname, '../../../dist/apps/docs/client')
  if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true })
  }

  fileRecords.forEach(({ fileName, records }) =>
    fs.writeFileSync(
      path.join(outputDir, `${fileName}.json`),
      JSON.stringify(records, null, 2),
      'utf8'
    )
  )

  const algoliaAppId =
    getEnv('ALGOLIA_APP_ID', '') ||
    getEnv('PUBLIC_ALGOLIA_APP_ID', '') ||
    getEnv('PUBLIC_ALGOLIA_APPLICATION_ID', '')

  const algoliaApiKey =
    getEnv('ALGOLIA_API_KEY', '') ||
    getEnv('ALGOLIA_ADMIN_API_KEY', '') ||
    getEnv('PUBLIC_ALGOLIA_ADMIN_API_KEY', '') ||
    getEnv('PUBLIC_ALGOLIA_API_KEY', '')

  const algoliaEnabled = Boolean(algoliaAppId && algoliaApiKey)
  const algoliaHost = normalizeAlgoliaHost(
    getEnv('ALGOLIA_HOST', `${algoliaAppId}.algolia.net`)
  )
  const indexPrefix = getEnv('ALGOLIA_INDEX_PREFIX', '')

  if (algoliaEnabled) {
    if (!algoliaHost) {
      console.error('Algolia sync failed: missing ALGOLIA_HOST')
      process.exitCode = 1
      return
    }

    const resolveIndexName = fileName => {
      const explicit = getEnv(`ALGOLIA_INDEX_${fileName.toUpperCase()}`, '')
      if (explicit) return explicit
      const publicIndexName =
        getEnv(`PUBLIC_ALGOLIA_${fileName.toUpperCase()}_INDEX_NAME`, '') ||
        getEnv(`NEXT_PUBLIC_ALGOLIA_${fileName.toUpperCase()}_INDEX_NAME`, '')
      if (publicIndexName) return publicIndexName
      if (indexPrefix) return `${indexPrefix}_${fileName}`
      return ''
    }

    Promise.resolve()
      .then(async () => {
        for (const { fileName, records } of fileRecords) {
          const indexName = resolveIndexName(fileName)
          if (!indexName) continue
          await uploadAlgoliaIndex({
            fileName,
            indexName,
            records,
            host: algoliaHost,
            appId: algoliaAppId,
            apiKey: algoliaApiKey,
          })
        }
      })
      .then(() => console.log('search index updated (dist + algolia)'))
      .catch(err => {
        console.error('Algolia sync failed:', err?.message || err)
        process.exitCode = 1
      })
    return
  }

  console.log('search index updated')
}

if (import.meta.url === `file://${process.argv[1]}`) {
  main()
}
