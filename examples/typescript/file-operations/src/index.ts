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

    //  search for files in the project
    const matches = await sandbox.fs.searchFiles(newDir, '*.json')
    console.log('JSON files found:', matches)

    //  replace content in config file
    await sandbox.fs.replaceInFiles([path.join(newDir, 'config.json')], '"debug": true', '"debug": false')

    //  download the modified config file
    console.log('Downloading updated config file:')
    const configContent = await sandbox.fs.downloadFile(path.join(newDir, 'config.json'))
    console.log(configContent.toString())

    // Create a report of all operations
    const reportData = `
Project Files Report:
---------------------
Time: ${new Date().toISOString()}
Files: ${matches.files.length} JSON files found
Config: ${configContent.includes('"debug": false') ? 'Production mode' : 'Debug mode'}
Script: ${scriptResult.exitCode === 0 ? 'Executed successfully' : 'Failed'}
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
