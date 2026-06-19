/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Daytona, Sandbox } from '@daytona/sdk'
import * as dotenv from 'dotenv'
import * as readline from 'readline'
import { GeminiSession } from './session.js'

dotenv.config()

async function main() {
  const apiKey = process.env.DAYTONA_API_KEY
  if (!apiKey) {
    console.error('Error: DAYTONA_API_KEY environment variable is not set')
    console.error('Create a .env file with your Daytona API key (see .env.example)')
    process.exit(1)
  }

  if (!process.env.SANDBOX_GEMINI_API_KEY) {
    console.error('Error: SANDBOX_GEMINI_API_KEY environment variable is not set')
    console.error('Get a free Gemini API key from https://aistudio.google.com/apikey')
    process.exit(1)
  }

  const daytona = new Daytona({ apiKey })

  let sandbox: Sandbox | undefined
  let session: GeminiSession | undefined

  const cleanup = async (exitCode = 0) => {
    try {
      console.log('\nCleaning up...')
      if (session) await session.cleanup()
      if (sandbox) await sandbox.delete()
    } catch (e) {
      console.error('Error during cleanup:', e)
    } finally {
      process.exit(exitCode)
    }
  }

  try {
    // Inject the Gemini API key at create time so the CLI runs headless with no
    // browser OAuth. The host-side SANDBOX_GEMINI_API_KEY maps to the bare
    // GEMINI_API_KEY the CLI expects inside the sandbox.
    // GEMINI_CLI_TRUST_WORKSPACE bypasses the CLI's workspace-trust prompt,
    // which otherwise blocks headless runs in a fresh sandbox directory.
    console.log('Creating sandbox...')
    sandbox = await daytona.create({
      envVars: {
        GEMINI_API_KEY: process.env.SANDBOX_GEMINI_API_KEY,
        GEMINI_CLI_TRUST_WORKSPACE: 'true',
      },
    })

    process.once('SIGINT', () => cleanup())

    console.log('Installing Gemini CLI...')
    const install = await sandbox.process.executeCommand('npm install -g @google/gemini-cli')
    if (install.exitCode !== 0) {
      throw new Error('Error installing Gemini CLI: ' + install.result)
    }

    console.log('Starting Gemini CLI...\n')
    session = new GeminiSession(sandbox)
    await session.initialize()

    const rl = readline.createInterface({ input: process.stdin, output: process.stdout })
    rl.once('SIGINT', () => cleanup())

    console.log('Agent ready. Press Ctrl+C at any time to exit.\n')

    while (true) {
      const prompt = await new Promise<string>((resolve) => rl.question('User: ', resolve))
      if (prompt.trim()) {
        await session.processPrompt(prompt)
      }
    }
  } catch (error) {
    console.error(error)
    await cleanup(1)
  }
}

main().catch(console.error)
