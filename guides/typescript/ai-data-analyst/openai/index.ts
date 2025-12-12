/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Daytona, CodeLanguage, Sandbox } from '@daytonaio/sdk'
import OpenAI from 'openai'
import * as fs from 'fs'

const CODING_MODEL = "gpt-5.1"
const SUMMARY_MODEL = "gpt-4o"

// Helper function to extract Python code from a given string
function extractPython(text: string): string {
  const m = text.match(/```python([\s\S]*?)```/)
  return m ? m[1].trim() : ''
}

// Make sure you have the DAYTONA_API_KEY and OPENAI_API_KEY environment variables set
const dt = new Daytona()
const openai = new OpenAI()

async function run() {
  let sb: Sandbox | null = null

  try {
    sb = await dt.create({ language: CodeLanguage.PYTHON })
  
    // Upload the CSV file to the sandbox
    const csvPath = 'cafe_sales_data.csv'
    const sandboxCsvPath = csvPath
    await sb.fs.uploadFile(csvPath, sandboxCsvPath)

    // Define the user prompt
    const userPrompt = `Give the three highest revenue products for the month of January and show them as a bar chart.`
    console.log("Prompt:", userPrompt)

    // Generate the system prompt with the first few rows of data for context
    const csvSample = fs.readFileSync(csvPath, 'utf8').split('\n').slice(0, 3).join('\n')
    const systemPrompt = `
You are a helpful assistant that analyzes data.
To run Python code in a sandbox, output a single block of code.
The sandbox:
 - has pandas and numpy installed.
 - contains ${sandboxCsvPath}.
Plot any charts that you create.
The first few rows of ${sandboxCsvPath} are:
${csvSample}
After seeing the results of the code, answer the user's query.`

    // Generate the Python code with the LLM
    console.log("Generating code...")
    const messages: OpenAI.Chat.Completions.ChatCompletionMessageParam[] = [
      { role: 'system', content: systemPrompt },
      { role: 'user', content: userPrompt },
    ]
    const llmOutput = await openai.chat.completions.create({
      model: CODING_MODEL,
      messages: messages,
    })
    messages.push(llmOutput.choices[0].message)

    // Extract and execute Python code from the LLM's response
    console.log("Running code...")
    const code = extractPython(llmOutput.choices[0].message.content || '')
    const exec = await sb.process.codeRun(code)
    messages.push({ role: 'user', content: `Code execution result:\n${exec.result}.` })

    if (exec.artifacts?.charts) {
      exec.artifacts.charts.forEach((chart: { png?: string }, index: number) => {
        if (chart.png) {
          const filename = `chart-${index}.png`
          fs.writeFileSync(filename, chart.png, { encoding: 'base64' })
          console.log(`✓ Chart saved to ${filename}`)
        }
      })
    }

    // Generate the final response with the LLM
    const summaryOutput = await openai.chat.completions.create({
      model: SUMMARY_MODEL,
      messages: messages,
    })
    console.log('Response:', summaryOutput.choices[0].message.content)
  } catch (error) {
    console.error('Error executing example:', error)
  } finally {
    if (sb) {
      await sb.delete()
    }
  }
}

run()
