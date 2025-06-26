import { Daytona } from '@daytonaio/sdk'

async function main() {
  const daytona = new Daytona()

  // Create a sandbox for testing search functionality
  const sandbox = await daytona.create()

  try {
    console.log(`Created sandbox with ID: ${sandbox.id}`)

    // Create some sample files to search through
    console.log('Setting up sample files for search demonstration...')

    // Create a project structure
    await sandbox.fs.createFolder('~/search-demo', '755')
    await sandbox.fs.createFolder('~/search-demo/src', '755')
    await sandbox.fs.createFolder('~/search-demo/tests', '755')
    await sandbox.fs.createFolder('~/search-demo/docs', '755')

    // Upload sample files with different content
    await sandbox.fs.uploadFiles([
      {
        source: Buffer.from(`// Main application file
import { Logger } from './utils/logger'
import { Config } from './config'

class Application {
  private logger: Logger
  private config: Config

  constructor() {
    this.logger = new Logger()
    this.config = new Config()
    // TODO: Add error handling
  }

  async start(): Promise<void> {
    this.logger.info('Starting application...')
    // TODO: Implement startup logic
  }

  async stop(): Promise<void> {
    this.logger.info('Stopping application...')
  }
}

export default Application
`),
        destination: '~/search-demo/src/app.ts',
      },
      {
        source: Buffer.from(`# Project Documentation

This is a sample project for demonstrating search functionality.

## Features

- Advanced search capabilities
- File type filtering
- Pattern matching
- Context-aware results

## TODO Items

- [ ] Add more examples
- [ ] Improve documentation
- [ ] Add unit tests

## Configuration

The application uses a JSON configuration file for settings.
`),
        destination: '~/search-demo/docs/README.md',
      },
      {
        source: Buffer.from(`{
  "name": "search-demo",
  "version": "1.0.0",
  "description": "Demo project for search functionality",
  "main": "src/app.ts",
  "scripts": {
    "start": "node dist/app.js",
    "build": "tsc",
    "test": "jest"
  },
  "dependencies": {
    "express": "^4.18.0",
    "lodash": "^4.17.21"
  },
  "devDependencies": {
    "typescript": "^4.9.0",
    "jest": "^29.0.0",
    "@types/node": "^18.0.0"
  }
}
`),
        destination: '~/search-demo/package.json',
      },
      {
        source: Buffer.from(`import { Application } from '../src/app'

describe('Application', () => {
  let app: Application

  beforeEach(() => {
    app = new Application()
  })

  test('should start successfully', async () => {
    // TODO: Add proper test implementation
    await expect(app.start()).resolves.not.toThrow()
  })

  test('should stop successfully', async () => {
    // TODO: Add proper test implementation  
    await expect(app.stop()).resolves.not.toThrow()
  })
})
`),
        destination: '~/search-demo/tests/app.test.ts',
      },
      {
        source: Buffer.from(`#!/bin/bash

# Build script for the project
echo "Building project..."

# Install dependencies
npm install

# Run TypeScript compiler
npm run build

# Run tests
npm test

echo "Build completed successfully!"
`),
        destination: '~/search-demo/build.sh',
      },
    ])

    console.log('Sample files created. Starting search demonstrations...\n')

    // === SEARCH DEMONSTRATIONS ===

    // 1. Basic text search
    console.log('=== 1. Basic Text Search ===')
    const basicSearch = await sandbox.fs.search({
      query: 'TODO',
      path: '~/search-demo',
    })
    console.log(`Found ${basicSearch.total_matches} TODO items in ${basicSearch.total_files} files:`)
    basicSearch.matches.forEach((match) => {
      console.log(`  ${match.file}:${match.line_number}: ${match.line.trim()}`)
    })

    // 2. Case-insensitive search
    console.log('\n=== 2. Case-Insensitive Search ===')
    const caseInsensitiveSearch = await sandbox.fs.search({
      query: 'application',
      path: '~/search-demo',
      case_sensitive: false,
    })
    console.log(`Found ${caseInsensitiveSearch.total_matches} matches for "application" (case-insensitive):`)
    caseInsensitiveSearch.matches.forEach((match) => {
      console.log(`  ${match.file}:${match.line_number}: ${match.match}`)
    })

    // 3. File type filtering
    console.log('\n=== 3. File Type Filtering ===')
    const typeScriptSearch = await sandbox.fs.search({
      query: 'class|interface|function',
      path: '~/search-demo',
      file_types: ['ts'],
      max_results: 10,
    })
    console.log(`Found ${typeScriptSearch.total_matches} TypeScript definitions:`)
    typeScriptSearch.matches.forEach((match) => {
      console.log(`  ${match.file}:${match.line_number}: ${match.line.trim()}`)
    })

    // 4. Search with context
    console.log('\n=== 4. Search with Context ===')
    const contextSearch = await sandbox.fs.search({
      query: 'constructor',
      path: '~/search-demo',
      context: 2,
      max_results: 3,
    })
    console.log('Constructor definitions with context:')
    contextSearch.matches.forEach((match) => {
      console.log(`\n  ${match.file}:${match.line_number}:`)
      if (match.context_before) {
        match.context_before.forEach((line, i) => {
          console.log(`    ${match.line_number - match.context_before!.length + i}: ${line}`)
        })
      }
      console.log(`  > ${match.line_number}: ${match.line}`)
      if (match.context_after) {
        match.context_after.forEach((line, i) => {
          console.log(`    ${match.line_number + i + 1}: ${line}`)
        })
      }
    })

    // 5. Include/Exclude patterns
    console.log('\n=== 5. Include/Exclude Patterns ===')
    const patternSearch = await sandbox.fs.search({
      query: 'test',
      path: '~/search-demo',
      include_globs: ['*.ts', '*.js'],
      exclude_globs: ['*.test.*'],
      case_sensitive: false,
    })
    console.log(`Found ${patternSearch.total_matches} "test" matches in source files (excluding test files):`)
    patternSearch.matches.forEach((match) => {
      console.log(`  ${match.file}:${match.line_number}: ${match.line.trim()}`)
    })

    // 6. Count-only search
    console.log('\n=== 6. Count-Only Search ===')
    const countSearch = await sandbox.fs.search({
      query: 'import|require',
      path: '~/search-demo',
      count_only: true,
    })
    console.log(`Total import/require statements: ${countSearch.total_matches} in ${countSearch.total_files} files`)

    // 7. Filenames-only search
    console.log('\n=== 7. Filenames-Only Search ===')
    const filenamesSearch = await sandbox.fs.search({
      query: 'app',
      path: '~/search-demo',
      filenames_only: true,
    })
    console.log(`Files containing "app": ${filenamesSearch.files?.join(', ') || 'none'}`)

    // 8. Advanced regex search
    console.log('\n=== 8. Advanced Regex Search ===')
    const regexSearch = await sandbox.fs.search({
      query: '"[^"]*":\\s*"[^"]*"', // JSON key-value pairs
      path: '~/search-demo',
      file_types: ['json'],
      max_results: 5,
    })
    console.log(`Found ${regexSearch.total_matches} JSON key-value pairs:`)
    regexSearch.matches.forEach((match) => {
      console.log(`  ${match.file}:${match.line_number}: ${match.match}`)
    })

    // 9. Multiline search
    console.log('\n=== 9. Multiline Search ===')
    const multilineSearch = await sandbox.fs.search({
      query: 'class.*{[\\s\\S]*?constructor',
      path: '~/search-demo',
      multiline: true,
      max_results: 3,
    })
    console.log(`Found ${multilineSearch.total_matches} class-constructor patterns:`)
    multilineSearch.matches.forEach((match) => {
      console.log(`  ${match.file}:${match.line_number}: ${match.match.substring(0, 50)}...`)
    })

    // 10. Performance comparison
    console.log('\n=== 10. Performance Comparison ===')
    const startTime = Date.now()
    const performanceSearch = await sandbox.fs.search({
      query: '.', // Match any character (lots of matches)
      path: '~/search-demo',
      max_results: 1000,
    })
    const endTime = Date.now()
    console.log(`Performance test: Found ${performanceSearch.total_matches} matches in ${endTime - startTime}ms`)

    console.log('\n✅ All search demonstrations completed successfully!')
  } catch (error) {
    console.error('❌ Error during search demonstration:', error)
  } finally {
    // Cleanup
    await daytona.delete(sandbox)
    console.log('Sandbox cleaned up')
  }
}

main().catch(console.error)
