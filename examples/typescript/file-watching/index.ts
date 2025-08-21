import { Daytona, Sandbox, Image, FilesystemEventType } from '@daytonaio/sdk'
import { execSync } from 'child_process'

// Set up environment variables before running this example:
// export DAYTONA_API_KEY="your-api-key"
// export DAYTONA_API_URL="https://your-api-url"
// export DAYTONA_TARGET="your-target"
// export DAYTONA_ORGANIZATION_ID="your-organization-id"

async function basicFileWatching(sandbox: Sandbox) {
  console.log('Starting basic file watching...')

  // Watch a directory for all changes
  const handle = await sandbox.fs.watchDir('/workspace', (event) => {
    console.log(`${event.type}: ${event.name}`)
  })

  // Create some files to trigger events
  await sandbox.fs.uploadFile(Buffer.from('Hello World!'), '/workspace/test.txt')
  await sandbox.fs.uploadFile(Buffer.from('Another file'), '/workspace/another.txt')
  await sandbox.fs.deleteFile('/workspace/test.txt')

  // Wait a bit for events to be processed
  await new Promise((resolve) => setTimeout(resolve, 2000))

  // Stop watching
  await handle.close()
  console.log('Basic file watching completed')
}

async function recursiveFileWatching(sandbox: Sandbox) {
  console.log('Starting recursive file watching...')

  // Watch recursively with filtering
  const handle = await sandbox.fs.watchDir(
    '/workspace',
    (event) => {
      // Only log TypeScript file changes
      if (event.type === FilesystemEventType.WRITE && event.name.endsWith('.ts')) {
        console.log(`TypeScript file changed: ${event.name}`)
      }
      // Log all directory creation events
      if (event.type === FilesystemEventType.CREATE && event.isDir) {
        console.log(`Directory created: ${event.name}`)
      }
    },
    { recursive: true },
  )

  // Create nested directory structure
  await sandbox.fs.createFolder('/workspace/src', '755')
  await sandbox.fs.uploadFile(Buffer.from('console.log("Hello from app!")'), '/workspace/src/app.ts')
  await sandbox.fs.createFolder('/workspace/src/components', '755')
  await sandbox.fs.uploadFile(Buffer.from('export class Button {}'), '/workspace/src/components/button.ts')

  // Wait for events to be processed
  await new Promise((resolve) => setTimeout(resolve, 2000))

  // Stop watching
  await handle.close()
  console.log('Recursive file watching completed')
}

async function fileWatchingWithErrorHandling(sandbox: Sandbox) {
  console.log('Starting file watching with error handling...')

  try {
    const handle = await sandbox.fs.watchDir('/workspace', (event) => {
      console.log(`Event: ${event.type} - ${event.name} (${event.isDir ? 'dir' : 'file'})`)
    })

    // Create files to trigger events
    await sandbox.fs.uploadFile(Buffer.from('Testing error handling'), '/workspace/error-test.txt')
    await sandbox.fs.setFilePermissions('/workspace/error-test.txt', { mode: '644' })
    await sandbox.fs.deleteFile('/workspace/error-test.txt')

    // Wait for events
    await new Promise((resolve) => setTimeout(resolve, 2000))

    await handle.close()
    console.log('File watching with error handling completed')
  } catch (error) {
    console.error('File watching error:', error)
  }
}

async function fileWatchingWithAsyncCallback(sandbox: Sandbox) {
  console.log('Starting file watching with async callback...')

  let eventCount = 0
  const handle = await sandbox.fs.watchDir('/workspace', async (event) => {
    eventCount++
    console.log(`Event ${eventCount}: ${event.type} - ${event.name}`)

    // Simulate async processing
    await new Promise((resolve) => setTimeout(resolve, 100))

    // Log event details
    console.log(`  Timestamp: ${event.timestamp}`)
    console.log(`  Is Directory: ${event.isDir}`)
  })

  // Create multiple files quickly
  for (let i = 0; i < 3; i++) {
    await sandbox.fs.uploadFile(Buffer.from(`Content ${i}`), `/workspace/async-test-${i}.txt`)
  }

  // Wait for all events to be processed
  await new Promise((resolve) => setTimeout(resolve, 3000))

  await handle.close()
  console.log(`Async file watching completed. Processed ${eventCount} events.`)
}

async function main() {
  const daytona = new Daytona()

  // Create an image with the workspace directory
  const image = Image.base('ubuntu:22.04').runCommands(
    'apt-get update && apt-get install -y --no-install-recommends nodejs npm coreutils',
    'curl -fsSL https://deb.nodesource.com/setup_20.x | bash -',
    'apt-get install -y nodejs',
    'npm install -g ts-node typescript',
    'mkdir -p /workspace', // Create the workspace directory
  )

  // Create a new sandbox
  const sandbox = await daytona.create({
    image,
    language: 'typescript',
    resources: {
      cpu: 1,
      memory: 1,
      disk: 3,
    },
  })

  // Local Hack for DNS resolution
  // execSync('hack/file-watching/dns_fix.sh', { stdio: 'inherit' })

  try {
    await basicFileWatching(sandbox)
    await recursiveFileWatching(sandbox)
    await fileWatchingWithErrorHandling(sandbox)
    await fileWatchingWithAsyncCallback(sandbox)
  } catch (error) {
    console.error('Error with file watching:', error)
  } finally {
    // Cleanup
    await daytona.delete(sandbox)
    console.log('File watching demo completed. Sandbox cleaned up.')
  }
}

main()
