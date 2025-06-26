import { Daytona } from '@daytonaio/sdk'
import * as fs from 'fs'
import * as path from 'path'

async function main() {
  const daytona = new Daytona()

  //  first, create a sandbox
  const sandbox = await daytona.create()

  try {
    console.log(`Created sandbox with ID: ${sandbox.id}`)

    //  list files in the sandbox
    const files = await sandbox.fs.listFiles('~')
    console.log('Initial files:', files)

    //  create a new directory in the sandbox
    const newDir = '~/project-files'
    await sandbox.fs.createFolder(newDir, '755')

    // Create a local file for demonstration
    const localFilePath = 'local-example.txt'
    fs.writeFileSync(localFilePath, 'This is a local file created for use case purposes')

    // Create a configuration file with JSON data
    const configData = JSON.stringify(
      {
        name: 'project-config',
        version: '1.0.0',
        settings: {
          debug: true,
          maxConnections: 10,
        },
      },
      null,
      2,
    )

    // Upload multiple files at once - both from local path and from buffers
    await sandbox.fs.uploadFiles([
      {
        source: localFilePath,
        destination: path.join(newDir, 'example.txt'),
      },
      {
        source: Buffer.from(configData),
        destination: path.join(newDir, 'config.json'),
      },
      {
        source: Buffer.from('#!/bin/bash\necho "Hello from script!"\nexit 0'),
        destination: path.join(newDir, 'script.sh'),
      },
    ])

    // Execute commands on the sandbox to verify files and make them executable
    console.log('Verifying uploaded files:')
    const lsResult = await sandbox.process.executeCommand(`ls -la ${newDir}`)
    console.log(lsResult.result)

    // Make the script executable
    await sandbox.process.executeCommand(`chmod +x ${path.join(newDir, 'script.sh')}`)

    // Run the script
    console.log('Running script:')
    const scriptResult = await sandbox.process.executeCommand(`${path.join(newDir, 'script.sh')}`)
    console.log(scriptResult.result)

    //  search for files by name pattern (temporarily using find_files until new endpoint is deployed)
    try {
      // This will work once the new search endpoint is available
      const fileMatches = await (sandbox.fs as any).search({
        query: '*.json',
        path: newDir,
        filenames_only: true,
      })
      console.log('JSON files found:', fileMatches.files)
    } catch (e) {
      console.log('Search endpoint not yet available, using fallback...')
      const files = await sandbox.fs.listFiles(newDir)
      const jsonFiles = files.filter((f) => f.name.endsWith('.json'))
      console.log(
        'JSON files found:',
        jsonFiles.map((f) => f.name),
      )
    }

    // === NEW ENHANCED SEARCH FUNCTIONALITY ===
    console.log('\n=== Enhanced Content Search Examples ===')
    console.log('Note: These examples will work once the new search endpoint is deployed')

    // 1. Basic content search
    console.log('\n1. Basic content search for "debug":')
    let basicSearch: any
    try {
      basicSearch = await (sandbox.fs as any).search({
        query: 'debug',
        path: newDir,
        case_sensitive: false,
      })
      console.log(`Found ${basicSearch.total_matches} matches in ${basicSearch.total_files} files`)
      basicSearch.matches.forEach((match: any) => {
        console.log(`  ${match.file}:${match.line_number}: ${match.line.trim()}`)
      })
    } catch (e) {
      console.log('Search not available yet:', (e as Error).message)
      basicSearch = { total_matches: 0, total_files: 0 }
    }

    // Remaining search examples (will work once endpoint is deployed)
    let shellSearch: any, contextSearch: any, countSearch: any, filenameSearch: any, advancedSearch: any
    try {
      // 2. Search with file type filtering
      console.log('\n2. Search for "echo" in shell scripts only:')
      shellSearch = await (sandbox.fs as any).search({
        query: 'echo',
        path: newDir,
        file_types: ['sh'],
        max_results: 5,
      })
      console.log(`Found ${shellSearch.total_matches} echo statements in shell files`)
      shellSearch.matches.forEach((match: any) => {
        console.log(`  ${match.file}:${match.line_number}: ${match.match}`)
      })

      // 3. Search with context lines
      console.log('\n3. Search for "version" with context:')
      contextSearch = await (sandbox.fs as any).search({
        query: 'version',
        path: newDir,
        context: 1,
        max_results: 3,
      })
      contextSearch.matches.forEach((match: any) => {
        console.log(`\n  Match in ${match.file}:${match.line_number}:`)
        if (match.context_before) {
          match.context_before.forEach((line: string, i: number) => {
            console.log(`    ${match.line_number - match.context_before!.length + i}: ${line}`)
          })
        }
        console.log(`  > ${match.line_number}: ${match.line}`)
        if (match.context_after) {
          match.context_after.forEach((line: string, i: number) => {
            console.log(`    ${match.line_number + i + 1}: ${line}`)
          })
        }
      })

      // 4. Count-only search for performance
      console.log('\n4. Count-only search for all words:')
      countSearch = await (sandbox.fs as any).search({
        query: '\\w+', // Regex for words
        path: newDir,
        count_only: true,
      })
      console.log(`Total word matches: ${countSearch.total_matches} in ${countSearch.total_files} files`)

      // 5. Filenames-only search
      console.log('\n5. Files containing "project":')
      filenameSearch = await (sandbox.fs as any).search({
        query: 'project',
        path: newDir,
        filenames_only: true,
      })
      console.log(`Files with "project": ${filenameSearch.files?.join(', ') || 'none'}`)

      // 6. Advanced search with include/exclude patterns
      console.log('\n6. Search in text files only, excluding scripts:')
      advancedSearch = await (sandbox.fs as any).search({
        query: 'file',
        path: newDir,
        include_globs: ['*.txt', '*.json'],
        exclude_globs: ['*.sh'],
        case_sensitive: false,
        max_results: 10,
      })
      console.log(`Found ${advancedSearch.total_matches} matches in text files`)
      advancedSearch.matches.forEach((match: any) => {
        console.log(`  ${match.file}: ${match.match}`)
      })
    } catch (e) {
      console.log('Advanced search examples not available yet:', (e as Error).message)
      // Create mock objects for the report
      shellSearch = { total_matches: 0 }
      contextSearch = { total_matches: 0 }
      countSearch = { total_matches: 0 }
      filenameSearch = { total_files: 0 }
      advancedSearch = { total_matches: 0 }
    }

    //  replace content in config file
    await sandbox.fs.replaceInFiles([path.join(newDir, 'config.json')], '"debug": true', '"debug": false')

    //  download the modified config file
    console.log('Downloading updated config file:')
    const configContent = await sandbox.fs.downloadFile(path.join(newDir, 'config.json'))
    console.log(configContent.toString())

    // Create a report of all operations including search results
    const reportData = `
Project Files Report:
---------------------
Time: ${new Date().toISOString()}
Files: JSON files found (search demo)
Config: ${configContent.includes('"debug": false') ? 'Production mode' : 'Debug mode'}
Script: ${scriptResult.exitCode === 0 ? 'Executed successfully' : 'Failed'}

Enhanced Search Results:
- Debug references: ${basicSearch.total_matches} matches
- Echo statements: ${shellSearch.total_matches} matches
- Version references: ${contextSearch.total_matches} matches
- Total words: ${countSearch.total_matches} matches
- Files with "project": ${filenameSearch.total_files} files
- Text file matches: ${advancedSearch.total_matches} matches
    `.trim()

    // Save the report
    await sandbox.fs.uploadFile(Buffer.from(reportData), path.join(newDir, 'report.txt'))

    // Clean up local file
    fs.unlinkSync(localFilePath)
  } catch (error) {
    console.error('Error:', error)
  } finally {
    //  cleanup
    await daytona.delete(sandbox)
  }
}

main()
