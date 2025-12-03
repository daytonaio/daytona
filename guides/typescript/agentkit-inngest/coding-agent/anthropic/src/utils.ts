import { Daytona, Sandbox } from '@daytonaio/sdk'
import { AgentResult, NetworkRun } from '@inngest/agent-kit'
import type { TextMessage } from '@inngest/agent-kit'

export function extractTextMessageContent(result: AgentResult | undefined): string {
  const textMessage = result?.output.find((msg) => msg.type === 'text') as TextMessage
  if (!textMessage || !textMessage.content) return ''
  if (typeof textMessage.content === 'string') return textMessage.content
  if (Array.isArray(textMessage.content)) {
    return textMessage.content.map((c) => c.text).join('')
  }
  return ''
}

export async function createSandbox(network?: NetworkRun<Record<string, any>>) {
  const daytona = new Daytona()
  let sandbox: Sandbox
  try {
    sandbox = await daytona.create()
  } catch (error) {
    throw new Error(`Failed to create Daytona sandbox: ${error}`)
  }
  if (network) network.state.data.sandbox = sandbox
  return sandbox
}

export async function getSandbox(network?: NetworkRun<Record<string, any>>) {
  let sandbox = network?.state.data.sandbox as Sandbox
  if (!sandbox) sandbox = await createSandbox(network)
  return sandbox
}

export const logDebug = (message: string) => {
  const enableDebugLogs = false
  if (enableDebugLogs) console.log(message)
}
