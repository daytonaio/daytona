/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { createOpencodeClient } from '@opencode-ai/sdk'
import type { AssistantMessage, Event, OpencodeClient, TextPartInput } from '@opencode-ai/sdk'

// Extract a human-readable message from an AssistantMessage error (e.g. ApiError.data.message).
function messageFromError(err: AssistantMessage['error']): string {
  return (err as { data?: { message?: string } })?.data?.message ?? String(err) ?? ''
}

// Yield from async iterable until `until` resolves.
async function* takeUntil<T>(iterable: AsyncIterable<T>, until: Promise<unknown>): AsyncGenerator<T> {
  const it = iterable[Symbol.asyncIterator]()
  while (true) {
    const next = await Promise.race([
      it.next(),
      until.then(() => ({ done: true as const, value: undefined })),
    ])
    if (next.done) break
    yield next.value
  }
}

// Log tool completions and session errors for the given session to the console.
function printEvent(sessionId: string, event: Event) {
  if (event.type === 'message.part.updated') {
    const { part } = event.properties
    if (part.sessionID !== sessionId || part.type !== 'tool' || !part.tool) return
    const st = part.state
    if (st.status === 'completed') {
      const label = st.title ?? st.output.trim().split('\n')[0] ?? part.tool
      if (part.tool === 'write') console.log('📝 Add', label)
      else if (part.tool === 'bash' || part.tool === 'run') console.log('🔨 ✓ Run:', label)
      else console.log('✓', label)
    } else if (st.status === 'error') {
      console.log('✗', part.tool, st.error ?? '')
    }
  } else if (event.type === 'session.error') {
    const { error } = event.properties
    console.error('Session error:', error && 'data' in error ? error.data.message ?? error : error ?? event.properties)
  }
}

export class Session {
  private readonly client
  readonly sessionId
  private readonly events

  private constructor(client: OpencodeClient, sessionId: string, events: Awaited<ReturnType<OpencodeClient['event']['subscribe']>>) {
    this.client = client
    this.sessionId = sessionId
    this.events = events
  }

  // Create a new OpenCode session and subscribe to its events.
  static async create(baseUrl: string): Promise<Session> {
    const client = createOpencodeClient({ baseUrl })
    const sessionRes = await client.session.create({ body: { title: 'Daytona query' } })
    const sessionId = sessionRes.data?.id
    if (!sessionId) throw new Error('Failed to create OpenCode session:' + sessionRes.error)
    const events = await client.event.subscribe()
    return new Session(client, sessionId, events)
  }

  // Send a prompt, stream tool events to the console, then print the final text response.
  async runQuery(query: string): Promise<void> {
    console.log('Thinking...')

    const promptPromise = this.client.session.prompt({
      path: { id: this.sessionId },
      body: { parts: [{ type: 'text', text: query } satisfies TextPartInput] },
    })

    // Consume event stream until the prompt request completes.
    for await (const event of takeUntil(this.events.stream, promptPromise)) {
      printEvent(this.sessionId, event)
    }

    // Await result, handle errors, and print text parts.
    const promptRes = await promptPromise
    if (promptRes.error) throw new Error(String(promptRes.error))
    const response = promptRes.data
    if (!response) return
    if (response.info?.error) throw new Error(messageFromError(response.info.error))
    for (const part of response.parts)
      if (part.type === 'text' && part.text) console.log(part.text)
  }
}