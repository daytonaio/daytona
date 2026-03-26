/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Daytona, Sandbox } from '@daytonaio/sdk'
import * as dotenv from 'dotenv'
import * as readline from 'readline'
import { Session } from './session.js'
import { Server } from './server.js'

dotenv.config()

// Create sandbox, start OpenCode server, and run an interactive query loop.
async function main(): Promise<void> {
  const apiKey = process.env.DAYTONA_API_KEY
  if (!apiKey) {
    console.error('Error: DAYTONA_API_KEY environment variable is not set')
    process.exit(1)
  }

  const daytona = new Daytona({ apiKey })
  let sandbox: Sandbox | undefined
  
  // Delete sandbox and exit on Ctrl+C or error.
  const cleanup = async () => {
    try {
      console.log('\nCleaning up...')
      if (sandbox) await sandbox.delete()
    } catch (e) {
      console.error('Error deleting sandbox:', e)
    } finally {
      process.exit(0)
    }
  }

  try {
    console.log('Creating sandbox...')
    sandbox = await daytona.create({ public: true })
    process.once('SIGINT', cleanup)

    console.log('Installing OpenCode in sandbox...')
    await sandbox.process.executeCommand('npm i -g opencode-ai@1.1.1')

    // Start OpenCode server and wait until it is listening.
    const { baseUrl, ready } = await Server.start(sandbox)
    await ready
    console.log('Preview:', baseUrl)
    console.log('Press Ctrl+C at any time to exit.')

    // Create OpenCode session and run interactive prompt loop.
    const session = await Session.create(baseUrl)
    const rl = readline.createInterface({ input: process.stdin, output: process.stdout })
    rl.once('SIGINT', cleanup)

    while (true) {
      const query = await new Promise<string>((resolve) => rl.question('User: ', resolve))
      if (!query.trim()) continue
      await session.runQuery(query)
      console.log('')
    }
  } catch (error) {
    console.error('An error occurred:', error)
    if (sandbox) await sandbox.delete()
    process.exit(1)
  }
}

main().catch(console.error)
