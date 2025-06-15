import * as _fs from 'fs'
import { dirname } from 'path'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)
const fs = _fs.promises

const SDK_DOCS_PATH = path.join(__dirname, '../src/content/docs')
const PYTHON_SDK_PATH = path.join(SDK_DOCS_PATH, 'python-sdk')
const TYPESCRIPT_SDK_PATH = path.join(SDK_DOCS_PATH, 'typescript-sdk')

const GITHUB_API_BASE = 'https://api.github.com/repos/daytonaio/sdk/contents/docs'
const GITHUB_RAW_BASE = 'https://raw.githubusercontent.com/daytonaio/sdk/main/docs'

async function fetchFile(url) {
  const response = await fetch(url)
  if (!response.ok) {
    throw new Error(`Failed to fetch ${url}: ${response.statusText}`)
  }
  return await response.text()
}

async function fetchDirectoryContents(path) {
  const response = await fetch(`${GITHUB_API_BASE}/${path}`)
  if (!response.ok) {
    throw new Error(`Failed to fetch directory contents: ${response.statusText}`)
  }
  return await response.json()
}

async function cleanDirectory(dir, keepIndex = true) {
  const files = await fs.readdir(dir)
  for (const file of files) {
    if (file === 'index.mdx' && keepIndex) continue
    await fs.unlink(path.join(dir, file))
  }
}

async function downloadAndSaveFile(url, targetPath) {
  const content = await fetchFile(url)
  await fs.writeFile(targetPath, content)
  console.log(`Downloaded and saved: ${targetPath}`)
}

async function getMdxFilesFromDirectory(dirPath) {
  const contents = await fetchDirectoryContents(dirPath)
  return contents
    .filter(item => item.type === 'file' && item.name.endsWith('.mdx'))
    .map(item => item.name)
}

async function updateSDKDocs() {
  console.log('Starting SDK documentation update...')

  // Clean directories while preserving index.mdx files
  console.log('Cleaning existing documentation...')
  await cleanDirectory(PYTHON_SDK_PATH)
  await cleanDirectory(TYPESCRIPT_SDK_PATH)

  // Get list of files from GitHub
  console.log('Fetching file lists from GitHub...')
  const pythonFiles = await getMdxFilesFromDirectory('python-sdk')
  const typescriptFiles = await getMdxFilesFromDirectory('typescript-sdk')

  // Download Python SDK docs
  console.log('Downloading Python SDK documentation...')
  for (const file of pythonFiles) {
    const url = `${GITHUB_RAW_BASE}/python-sdk/${file}`
    const targetPath = path.join(PYTHON_SDK_PATH, file)
    await downloadAndSaveFile(url, targetPath)
  }

  // Download TypeScript SDK docs
  console.log('Downloading TypeScript SDK documentation...')
  for (const file of typescriptFiles) {
    const url = `${GITHUB_RAW_BASE}/typescript-sdk/${file}`
    const targetPath = path.join(TYPESCRIPT_SDK_PATH, file)
    await downloadAndSaveFile(url, targetPath)
  }

  console.log('SDK documentation update completed successfully!')
}

// Execute the update
updateSDKDocs().catch(error => {
  console.error('Error updating SDK documentation:', error)
  process.exit(1)
})
