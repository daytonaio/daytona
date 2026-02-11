import fs from 'fs'
import path from 'path'
import { dirname } from 'path'
import { fileURLToPath } from 'url'
import https from 'https'

import {
  buildSearchFileRecords,
  writeSearchFilesToDist,
} from './update-search-local.js'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

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

function safeJsonParse(text) {
  try {
    return JSON.parse(text)
  } catch {
    return null
  }
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

export function main() {
  loadDotEnvFileIfPresent(path.join(__dirname, '../.env'))

  const fileRecords = buildSearchFileRecords()
  writeSearchFilesToDist(fileRecords)

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

