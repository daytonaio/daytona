/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { createOpencodeClient } from '@opencode-ai/sdk'

type Part = {
  type: string
  sessionID?: string
  tool?: string
  state?: { status: string; input?: Record<string, unknown>; output?: string; error?: string; title?: string }
  text?: string
}

function getSessionId(res: unknown): string | undefined {
  if (res == null || typeof res !== 'object') return undefined
  const o = res as Record<string, unknown>
  const data = o.data as { id?: string } | undefined
  return data?.id ?? (o.id as string | undefined)
}

function printEvent(sessionId: string, ev: { type: string; properties?: Record<string, unknown> }): void {
  const props = ev.properties ?? {}
  const eventSessionId = (props.sessionID as string) ?? (props.part as Part | undefined)?.sessionID
  if (eventSessionId !== sessionId) return
  if (ev.type === 'message.part.updated') {
    const part = props.part as Part | undefined
    if (!part) return
    if (part.type === 'tool' && part.tool) {
      const st = part.state
      const status = st?.status ?? '?'
      const title = st && 'title' in st ? (st as { title?: string }).title : undefined
      const firstLine = st?.output?.trim().split('\n')[0]
      const label = title ?? firstLine ?? part.tool
      if (status === 'completed') {
        if (part.tool === 'write') console.log('📝 Add', label)
        else if (part.tool === 'bash' || part.tool === 'run') console.log('🔨 ✓ Run:', label)
        else console.log('✓', label)
      } else if (status === 'error') {
        console.log('✗', label, st?.error ?? '')
      }
    }
  } else if (ev.type === 'session.error') {
    const err = ev.properties?.error as { message?: string } | undefined
    console.error('Session error:', err?.message ?? ev.properties?.error)
  }
}

async function streamEventsUntilDone(
  stream: AsyncIterable<{ type: string; properties?: Record<string, unknown> }>,
  sessionId: string,
  done: Promise<unknown>,
): Promise<void> {
  const it = stream[Symbol.asyncIterator]()
  while (true) {
    const next = await Promise.race([
      it.next(),
      done.then(() => ({ done: true as const, value: undefined })),
    ])
    if (next.done) break
    const ev = next.value as { type: string; properties?: Record<string, unknown> } | undefined
    if (ev) printEvent(sessionId, ev)
  }
}

function parsePromptResponse(promptRes: unknown): Array<{ type: string; text?: string }> {
  const res = ((promptRes as Record<string, unknown>)?.data ?? promptRes) as Record<string, unknown>
  const info = res?.info as { error?: { data?: { message?: string } } } | undefined
  if (info?.error) throw new Error(String(info.error?.data?.message ?? info.error))
  return (res?.parts ?? []) as Array<{ type: string; text?: string }>
}

function printResponseParts(parts: Array<{ type: string; text?: string }>): void {
  const textParts = parts.filter((p) => p.type === 'text' && p.text)
  for (const part of textParts) {
    if (part.text) console.log(part.text)
  }
}

export class Session {
  private readonly client: ReturnType<typeof createOpencodeClient>
  readonly sessionId: string
  private readonly events: Awaited<ReturnType<ReturnType<typeof createOpencodeClient>['event']['subscribe']>>

  private constructor(
    client: ReturnType<typeof createOpencodeClient>,
    sessionId: string,
    events: Awaited<ReturnType<ReturnType<typeof createOpencodeClient>['event']['subscribe']>>,
  ) {
    this.client = client
    this.sessionId = sessionId
    this.events = events
  }

  static async create(baseUrl: string): Promise<Session> {
    const client = createOpencodeClient({ baseUrl })
    const sessionRes = await client.session.create({ body: { title: 'Daytona query' } }) as unknown
    const sessionId = getSessionId(sessionRes)
    if (!sessionId) throw new Error('Failed to create OpenCode session')
    const events = await client.event.subscribe()
    return new Session(client, sessionId, events)
  }

  async runQuery(query: string): Promise<void> {
    console.log('Thinking...')
    const promptPromise = this.client.session.prompt({
      path: { id: this.sessionId },
      body: { parts: [{ type: 'text', text: query }] },
    }) as Promise<unknown>

    await streamEventsUntilDone(
      this.events.stream as AsyncIterable<{ type: string; properties?: Record<string, unknown> }>,
      this.sessionId,
      promptPromise,
    )

    const promptRes = await promptPromise
    const parts = parsePromptResponse(promptRes)
    printResponseParts(parts)
  }
}
