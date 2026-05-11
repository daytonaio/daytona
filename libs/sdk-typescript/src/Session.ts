/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import axios, { AxiosInstance } from 'axios'
import { DaytonaError } from './errors/DaytonaError'

/**
 * HTTP status of an error, whether it's a raw AxiosError or a DaytonaError
 * already normalized by the shared axios interceptors.
 */
function errorStatus(err: unknown): number | undefined {
  if (axios.isAxiosError(err)) {
    return err.response?.status
  }
  if (err instanceof DaytonaError) {
    return err.statusCode
  }
  return undefined
}

/**
 * Raw response body of an error, from a raw AxiosError or a normalized
 * DaytonaError (which preserves the body on `.data`).
 */
function errorBody(err: unknown): unknown {
  if (axios.isAxiosError(err)) {
    return err.response?.data
  }
  if (err instanceof DaytonaError) {
    return err.data
  }
  return undefined
}

/**
 * One frame on the streaming WS protocol. Mirrors the daemon-side frame format 1:1.
 */
export type SessionFrameType = 'stdout' | 'stderr' | 'error' | 'display' | 'control'

export interface SessionDisplay {
  formats: string[]
  /** Mime → payload (base64 for binary). */
  data: Record<string, string>
}

export interface SessionExecutionError {
  name: string
  value?: string
  traceback?: string
}

export interface SessionRunResult {
  stdout: string
  stderr: string
  /** The execution error, or `null` when the code ran successfully. */
  error: SessionExecutionError | null
  displays: SessionDisplay[]
  durationMs: number
}

/**
 * Signed access bundle for a session, returned by the API.
 *
 * `token` is exposed for revocation / observability; the SDK does not send it
 * as an `Authorization` header.
 */
export interface SessionAccess {
  httpUrl: string
  wsUrl: string
  token: string
  /** ISO timestamp; the access is refreshed shortly before this. */
  tokenExpiresAt: string
}

export interface SessionRef {
  id: string
  language: string
  cwd?: string
  createdAt: string
  lastUsedAt?: string
  expiresAt: string
  /** Present on createSession / refresh responses; omitted on listSessions. */
  access?: SessionAccess
}

/**
 * Reference to an existing session to run code against, optionally carrying a
 * primed access bundle to skip an extra access lookup.
 */
export interface SessionContext {
  id: string
  access?: SessionAccess
}

/** Options for {@link SessionService.createSession}. */
export interface CreateSessionOptions {
  template?: string
  language?: string
  cwd?: string
}

/** A runtime template available for sessions. */
export interface SessionTemplate {
  name: string
  description?: string
  languages: string[]
  packages?: string[]
}

/** A package available within a session template. */
export interface SessionPackage {
  name: string
  version: string
  hasNativeBindings?: boolean
}

export interface SessionRunOptions {
  language?: 'python' | 'typescript' | 'javascript'
  template?: string
  context?: SessionContext
  env?: Record<string, string>
  timeout?: number
}

export interface SessionRunStreamOptions extends SessionRunOptions {
  onStdout?: (chunk: string) => void
  onStderr?: (chunk: string) => void
  onError?: (err: SessionExecutionError) => void
  onDisplay?: (display: SessionDisplay) => void
  onControl?: (text: string) => void
  signal?: AbortSignal
}

/**
 * SessionInvalidatedError is the SDK projection of HTTP 410
 * `error.name=SessionInvalidated`. Surfaces when the underlying sandbox
 * has been rolled (death / snapshot drift / autostop).
 */
export class SessionInvalidatedError extends DaytonaError {
  sessionId: string
  invalidatedAt: string
  constructor(sessionId: string, invalidatedAt: string) {
    super(`Session ${sessionId} has been invalidated at ${invalidatedAt}`)
    this.name = 'SessionInvalidatedError'
    this.sessionId = sessionId
    this.invalidatedAt = invalidatedAt
  }
}

/**
 * SessionExpiredError is the SDK projection of HTTP 410
 * `error.name=SessionExpired`. `reason` distinguishes idle vs absolute TTL.
 */
export class SessionExpiredError extends DaytonaError {
  sessionId: string
  expiredAt: string
  reason: 'idle' | 'absolute'
  constructor(sessionId: string, expiredAt: string, reason: 'idle' | 'absolute') {
    super(`Session ${sessionId} expired (${reason}) at ${expiredAt}`)
    this.name = 'SessionExpiredError'
    this.sessionId = sessionId
    this.expiredAt = expiredAt
    this.reason = reason
  }
}

const REFRESH_SKEW_SECONDS = 60

interface Handlers {
  onStdout?: (chunk: string) => void
  onStderr?: (chunk: string) => void
  onError?: (err: SessionExecutionError) => void
  onDisplay?: (display: SessionDisplay) => void
  onControl?: (text: string) => void
  signal?: AbortSignal
}

interface DaemonFrame {
  type: SessionFrameType
  text?: string
  name?: string
  value?: string
  traceback?: string
  formats?: string[]
  data?: Record<string, string>
}

class LegacyFallback extends Error {}
class WsAuthError extends Error {}

/**
 * SessionService is the user-facing SDK surface for the sessions product.
 *
 * Use `createSession` to start a persistent session and `run` / `runStream`
 * to execute code against it. Calls to `run` / `runStream` can also be made
 * without an explicit session — passing only a `template` / `language` — to
 * execute one-shot code transparently. `listTemplates` and `listPackages`
 * describe the runtimes available for sessions.
 */
export class SessionService {
  // SDK-side caches keyed by context id and `${template}:${language}` for
  // one-shot transients. Survive across calls; eviction happens on auth
  // failure, expiry, or explicit deleteSession.
  private readonly ctxAccess = new Map<string, SessionAccess>()
  private readonly transientAccess = new Map<string, { sessionId: string; access: SessionAccess }>()
  // Latches once the API confirms /transients is missing (older server).
  private transientSupported: boolean | null = null

  constructor(private readonly http: AxiosInstance) {}

  async run(code: string, options: SessionRunOptions = {}): Promise<SessionRunResult> {
    return this.runInternal(code, options, undefined)
  }

  async runStream(code: string, options: SessionRunStreamOptions = {}): Promise<SessionRunResult> {
    const { onStdout, onStderr, onError, onDisplay, onControl, signal, ...runOpts } = options
    return this.runInternal(code, runOpts, { onStdout, onStderr, onError, onDisplay, onControl, signal })
  }

  async createSession(options: CreateSessionOptions): Promise<SessionRef> {
    try {
      const { data } = await this.http.post<SessionRef>('/sessions', options)
      if (data.access) {
        this.ctxAccess.set(data.id, data.access)
      }
      return data
    } catch (err) {
      throw this.translateError(err)
    }
  }

  async listSessions(template?: string): Promise<SessionRef[]> {
    try {
      const { data } = await this.http.get<SessionRef[]>('/sessions', {
        params: template ? { template } : undefined,
      })
      return data
    } catch (err) {
      throw this.translateError(err)
    }
  }

  async deleteSession(id: string): Promise<void> {
    try {
      await this.http.delete(`/sessions/${encodeURIComponent(id)}`)
      this.ctxAccess.delete(id)
    } catch (err) {
      throw this.translateError(err)
    }
  }

  async listTemplates(): Promise<SessionTemplate[]> {
    try {
      const { data } = await this.http.get('/sessions/templates')
      return data
    } catch (err) {
      throw this.translateError(err)
    }
  }

  async listPackages(templateName: string, language: string): Promise<SessionPackage[]> {
    try {
      const { data } = await this.http.get(`/sessions/templates/${encodeURIComponent(templateName)}/packages`, {
        params: { language },
      })
      return data
    } catch (err) {
      throw this.translateError(err)
    }
  }

  // -- direct-to-sandbox hot path ----------------------------------------

  private async runInternal(
    code: string,
    options: SessionRunOptions,
    handlers: Handlers | undefined,
  ): Promise<SessionRunResult> {
    if (!options.context && this.transientSupported === false) {
      return this.runLegacy(code, options, handlers)
    }

    let sessionId: string
    let access: SessionAccess
    let reset: boolean
    try {
      ;({ sessionId, access, reset } = await this.ensureAccess(options))
    } catch (err) {
      if (err instanceof LegacyFallback) {
        if (!options.context) {
          this.transientSupported = false
        }
        return this.runLegacy(code, options, handlers)
      }
      throw err
    }

    try {
      return await this.runWsDirect(sessionId, access, code, options, reset, handlers)
    } catch (err) {
      if (err instanceof WsAuthError) {
        this.evictAccess(options, sessionId)
        try {
          ;({ sessionId, access, reset } = await this.ensureAccess(options))
        } catch (refreshErr) {
          if (refreshErr instanceof LegacyFallback) {
            if (!options.context) {
              this.transientSupported = false
            }
            return this.runLegacy(code, options, handlers)
          }
          throw refreshErr
        }
        return this.runWsDirect(sessionId, access, code, options, reset, handlers)
      }
      throw err
    }
  }

  private async ensureAccess(
    options: SessionRunOptions,
  ): Promise<{ sessionId: string; access: SessionAccess; reset: boolean }> {
    if (options.context) {
      const sessionId = options.context.id
      // Inline access from a freshly-created context primes the cache —
      // avoids an immediate GET /access right after createSession().
      if (options.context.access && !this.ctxAccess.has(sessionId)) {
        this.ctxAccess.set(sessionId, options.context.access)
      }
      let access = this.ctxAccess.get(sessionId)
      if (!access || this.isExpired(access)) {
        access = await this.fetchSessionAccess(sessionId)
        this.ctxAccess.set(sessionId, access)
      }
      return { sessionId, access, reset: false }
    }

    const key = `${options.template ?? ''}:${options.language ?? ''}`
    const cached = this.transientAccess.get(key)
    if (cached && !this.isExpired(cached.access)) {
      return { sessionId: cached.sessionId, access: cached.access, reset: true }
    }
    const fresh = await this.fetchTransient(options.template, options.language)
    this.transientAccess.set(key, fresh)
    this.transientSupported = true
    return { sessionId: fresh.sessionId, access: fresh.access, reset: true }
  }

  private async runWsDirect(
    sessionId: string,
    access: SessionAccess,
    code: string,
    options: SessionRunOptions,
    reset: boolean,
    handlers: Handlers | undefined,
  ): Promise<SessionRunResult> {
    const WebSocketImpl = await this.resolveWebSocketImpl()

    return new Promise<SessionRunResult>((resolve, reject) => {
      const ws = new WebSocketImpl(access.wsUrl)
      const aggregated: SessionRunResult = {
        stdout: '',
        stderr: '',
        error: null,
        displays: [],
        durationMs: 0,
      }
      const startedAt = Date.now()
      let resolved = false
      let opened = false

      const finish = (resolvedOrErr: SessionRunResult | Error) => {
        if (resolved) return
        resolved = true
        void options
        handlers?.signal?.removeEventListener('abort', onAbort)
        try {
          ws.close()
        } catch {
          /* close errors are non-fatal here */
        }
        if (resolvedOrErr instanceof Error) {
          reject(resolvedOrErr)
        } else {
          resolvedOrErr.durationMs = Date.now() - startedAt
          resolve(resolvedOrErr)
        }
      }

      const onAbort = () => finish(new DaytonaError('runStream aborted by caller'))
      if (handlers?.signal?.aborted) {
        finish(new DaytonaError('runStream aborted by caller'))
        return
      }
      handlers?.signal?.addEventListener('abort', onAbort, { once: true })

      ws.onopen = () => {
        opened = true
        ws.send(
          JSON.stringify({
            code,
            envs: options.env ?? {},
            timeout: options.timeout ?? 0,
            reset,
          }),
        )
      }
      ws.onmessage = (ev: { data: string | ArrayBuffer }) => {
        try {
          const frame = JSON.parse(
            typeof ev.data === 'string' ? ev.data : Buffer.from(ev.data as ArrayBuffer).toString(),
          ) as DaemonFrame
          this.applyFrame(frame, aggregated, handlers)
        } catch {
          // malformed frame — ignore, daemon never emits non-JSON frames in normal operation
        }
      }
      // Node's `ws` exposes the HTTP status code on the unexpected-response
      // event before `error` fires. Hook it if available; the browser
      // WebSocket gives us no such hook.
      const wsAny = ws as unknown as {
        on?: (event: string, listener: (...args: any[]) => void) => void
      }
      if (typeof wsAny.on === 'function') {
        wsAny.on('unexpected-response', (_req: unknown, res: { statusCode?: number }) => {
          const status = res?.statusCode
          if (status === 401 || status === 403) {
            finish(new WsAuthError())
          } else {
            // 404 (context gone), 400 (proxy: "Is the Sandbox started?"),
            // 5xx — all map to invalidation. The caller's recovery action
            // is the same: drop the context and create a fresh one.
            finish(new SessionInvalidatedError(sessionId, new Date().toISOString()))
          }
        })
      }
      ws.onerror = () => {
        if (!opened) {
          // Pre-open error with no status — treat as sandbox-gone.
          finish(new SessionInvalidatedError(sessionId, new Date().toISOString()))
        } else {
          finish(new DaytonaError('session websocket error'))
        }
      }
      ws.onclose = () => finish(aggregated)
    })
  }

  private async fetchSessionAccess(sessionId: string): Promise<SessionAccess> {
    try {
      const { data } = await this.http.get<SessionAccess>(`/sessions/${encodeURIComponent(sessionId)}/access`)
      return data
    } catch (err) {
      const status = errorStatus(err)
      if (status === 404) {
        if (this.isRouteMissing404(errorBody(err))) {
          throw new LegacyFallback()
        }
        throw new SessionInvalidatedError(sessionId, new Date().toISOString())
      }
      if (status === 410) {
        throw this.translateError(err)
      }
      throw err
    }
  }

  private async fetchTransient(
    template: string | undefined,
    language: string | undefined,
  ): Promise<{ sessionId: string; access: SessionAccess }> {
    try {
      const body: Record<string, string> = {}
      if (template !== undefined) body.template = template
      if (language !== undefined) body.language = language
      const { data } = await this.http.post<SessionRef>('/sessions/transients', body)
      if (!data.access) {
        throw new LegacyFallback()
      }
      return { sessionId: data.id, access: data.access }
    } catch (err) {
      if (err instanceof LegacyFallback) {
        throw err
      }
      if (errorStatus(err) === 404 && this.isRouteMissing404(errorBody(err))) {
        throw new LegacyFallback()
      }
      throw err
    }
  }

  private async runLegacy(
    code: string,
    options: SessionRunOptions,
    handlers: Handlers | undefined,
  ): Promise<SessionRunResult> {
    if (handlers) {
      return this.runStreamLegacy(code, options, handlers)
    }
    return this.runCodeRunLegacy(code, options)
  }

  private async runCodeRunLegacy(code: string, options: SessionRunOptions): Promise<SessionRunResult> {
    try {
      const { data } = await this.http.post<SessionRunResult>('/sessions/code-run', { code, ...options })
      return {
        ...data,
        error: data.error ?? null,
        displays: data.displays ?? [],
      }
    } catch (err) {
      throw this.translateError(err)
    }
  }

  private async runStreamLegacy(
    code: string,
    options: SessionRunOptions,
    handlers: Handlers,
  ): Promise<SessionRunResult> {
    let data: { wsUrl: string; token: string; sessionId: string; expiresAt: string }
    try {
      ;({ data } = await this.http.post<{ wsUrl: string; token: string; sessionId: string; expiresAt: string }>(
        '/sessions/connect',
        options,
      ))
    } catch (err) {
      throw this.translateError(err)
    }
    const WebSocketImpl = await this.resolveWebSocketImpl()
    return new Promise<SessionRunResult>((resolve, reject) => {
      const ws = new WebSocketImpl(data.wsUrl)
      const aggregated: SessionRunResult = {
        stdout: '',
        stderr: '',
        error: null,
        displays: [],
        durationMs: 0,
      }
      const startedAt = Date.now()
      let resolved = false

      const finish = (resolvedOrErr: SessionRunResult | Error) => {
        if (resolved) return
        resolved = true
        handlers.signal?.removeEventListener('abort', onAbort)
        try {
          ws.close()
        } catch {
          /* close errors are non-fatal here */
        }
        if (resolvedOrErr instanceof Error) {
          reject(resolvedOrErr)
        } else {
          resolvedOrErr.durationMs = Date.now() - startedAt
          resolve(resolvedOrErr)
        }
      }

      const onAbort = () => finish(new DaytonaError('runStream aborted by caller'))
      if (handlers.signal?.aborted) {
        finish(new DaytonaError('runStream aborted by caller'))
        return
      }
      handlers.signal?.addEventListener('abort', onAbort, { once: true })

      ws.onopen = () => {
        ws.send(JSON.stringify({ code, envs: options.env, timeout: options.timeout }))
      }
      ws.onmessage = (ev: { data: string | ArrayBuffer }) => {
        try {
          const frame = JSON.parse(
            typeof ev.data === 'string' ? ev.data : Buffer.from(ev.data as ArrayBuffer).toString(),
          ) as DaemonFrame
          this.applyFrame(frame, aggregated, handlers)
        } catch {
          /* ignore malformed frames */
        }
      }
      ws.onerror = () => finish(new DaytonaError('session websocket error'))
      ws.onclose = () => finish(aggregated)
    })
  }

  // -- helpers -----------------------------------------------------------

  private isExpired(access: SessionAccess): boolean {
    const expiry = new Date(access.tokenExpiresAt).getTime()
    return Number.isFinite(expiry) ? Date.now() + REFRESH_SKEW_SECONDS * 1000 >= expiry : true
  }

  private evictAccess(options: SessionRunOptions, sessionId: string): void {
    if (options.context) {
      this.ctxAccess.delete(sessionId)
      return
    }
    this.transientAccess.delete(`${options.template ?? ''}:${options.language ?? ''}`)
  }

  private isRouteMissing404(body: unknown): boolean {
    // A non-JSON / empty body means we never reached the route layer — treat as route-missing.
    if (!body || typeof body !== 'object') return true
    const b = body as { message?: unknown; error?: unknown }
    // Nest's unknown-route 404 carries `message:"Cannot GET /..."`. A genuine
    // missing-session NotFoundException carries the SAME `error:"Not Found"` but
    // a descriptive message (e.g. "Session <id> not found."), so we must NOT
    // key off `error` — only the message (or its absence) is a reliable signal.
    if (typeof b.message === 'string') {
      return b.message.startsWith('Cannot ') || b.message.trim() === ''
    }
    // No message field at all → not a recognizable missing-session envelope; assume route-missing.
    return b.message === undefined || b.message === null
  }

  private applyFrame(frame: DaemonFrame, aggregated: SessionRunResult, handlers: Handlers | undefined): void {
    switch (frame.type) {
      case 'stdout':
        aggregated.stdout += frame.text ?? ''
        handlers?.onStdout?.(frame.text ?? '')
        break
      case 'stderr':
        aggregated.stderr += frame.text ?? ''
        handlers?.onStderr?.(frame.text ?? '')
        break
      case 'error': {
        const e: SessionExecutionError = { name: frame.name ?? 'Error', value: frame.value, traceback: frame.traceback }
        aggregated.error = e
        handlers?.onError?.(e)
        break
      }
      case 'display': {
        const d: SessionDisplay = { formats: frame.formats ?? [], data: frame.data ?? {} }
        aggregated.displays.push(d)
        handlers?.onDisplay?.(d)
        break
      }
      case 'control':
        handlers?.onControl?.(frame.text ?? '')
        break
    }
  }

  private async resolveWebSocketImpl(): Promise<typeof WebSocket> {
    if (typeof WebSocket !== 'undefined') {
      return WebSocket
    }
    const wsModule = (await import('ws').catch(() => null)) as { default?: typeof WebSocket } | null
    if (!wsModule?.default) {
      throw new DaytonaError("WebSocket support not available; install the 'ws' package or run in a browser/Node 22+")
    }
    return wsModule.default
  }

  private translateError(err: unknown): Error {
    if (errorStatus(err) === 410) {
      const body = errorBody(err) as {
        error?: {
          name?: string
          sessionId?: string
          invalidatedAt?: string
          expiredAt?: string
          reason?: 'idle' | 'absolute'
        }
      }
      const e = body?.error
      if (e?.name === 'SessionInvalidated' && e.sessionId && e.invalidatedAt) {
        return new SessionInvalidatedError(e.sessionId, e.invalidatedAt)
      }
      if (e?.name === 'SessionExpired' && e.sessionId && e.expiredAt && e.reason) {
        return new SessionExpiredError(e.sessionId, e.expiredAt, e.reason)
      }
    }
    return err as Error
  }
}
