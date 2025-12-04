import { Daytona, CodeLanguage, Sandbox } from '@daytonaio/sdk'
import OpenAI from 'openai'
import * as fs from 'fs'

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
    await sb.fs.uploadFile('cafe_sales_data.csv', 'cafe_sales_data.csv')

    // Define the user prompt
    const userPrompt = `Give the three highest revenue products for the month of January and show them as a bar chart.`
    console.log("Prompt:", userPrompt)

    // Generate the system prompt with the first few rows of data for context
    const csvSample = fs.readFileSync('cafe_sales_data.csv', 'utf8').split('\n').slice(0, 3).join('\n')
    const systemPrompt = `
You are a helpful assistant that analyzes data.
Generate Python code when necessary. Pandas and numpy are installed.
You have access to cafe_sales_data.csv. The first few rows are:
${csvSample}
After seeing the results of the code, give your final response.`

    // Generate the Python code with the LLM
    console.log("Generating code...")
    const messages: OpenAI.Chat.Completions.ChatCompletionMessageParam[] = [
      { role: 'system', content: systemPrompt },
      { role: 'user', content: userPrompt },
    ]
    const llmOutput = await openai.chat.completions.create({
      model: 'gpt-5.1',
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
      model: 'gpt-4o',
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
