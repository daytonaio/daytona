/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { randomUUID } from 'crypto'
import WebSocket from 'ws'
import { TypedConfigService } from '../../config/typed-config.service'
import { SandboxService } from '../../sandbox/services/sandbox.service'
import { RunnerService } from '../../sandbox/services/runner.service'
import { Organization } from '../../organization/entities/organization.entity'
import { SessionInstance } from '../entities/session-instance.entity'
import { SessionRepository } from './session-repository.service'
import { SessionPoolService } from './session-pool.service'
import { SessionLoadService } from './session-load.service'
import { SessionTemplateService } from './session-template.service'
import { DaemonAccess, buildDaemonAccess } from '../common/daemon-access'
import {
  DisplayDataDto,
  ExecutionErrorDto,
  SessionCodeRunRequestDto,
  SessionCodeRunResponseDto,
} from '../dto/code-run.dto'
import { SessionConnectRequestDto, SessionConnectResponseDto } from '../dto/connect.dto'
import { CreateSessionDto } from '../dto/create-session.dto'
import { CreateSessionTransientDto } from '../dto/create-session-transient.dto'
import { SessionAccessDto, SessionDto } from '../dto/session.dto'
import { SessionPackageDto } from '../dto/session-package.dto'
import { SessionInvalidatedError } from '../errors/session-errors'

/**
 * Wire shape of a single frame the in-sandbox session-daemon emits on the
 * /sessions/:id/execute WebSocket. Mirrors `OutputMessage` in
 * apps/session-daemon/internal/interpreter/types.go — keep them in sync.
 */
interface DaemonFrame {
  type: 'stdout' | 'stderr' | 'error' | 'display' | 'control'
  text?: string
  name?: string
  value?: string
  traceback?: string
  formats?: string[]
  data?: Record<string, string>
}

/**
 * SessionService is a thin facade that routes user-facing requests to:
 *  - SessionTemplateService for template resolution / listing.
 *  - SessionPoolService for warm-sandbox lookup-or-create.
 *  - SessionRepository for context identity and the cache layer.
 *  - The in-sandbox session-daemon for actual code execution.
 *
 * If a request carries `context.id`, `template`/`language` are ignored — derived from the
 * context row.
 *
 * One-shot codeRun (no context): mints an in-memory uuid, calls daemon-side POST /sessions +
 * exec + DELETE /sessions; **no Session row is persisted**. Connect (streaming) without
 * a context **does** persist a row so the SDK has a stable handle for post-WS-close cleanup.
 */
@Injectable()
export class SessionService {
  private readonly logger = new Logger(SessionService.name)

  // In-process memo for "we already POSTed /sessions to the daemon for this
  // (instance, language) transient" — saves a daemon round-trip on every
  // subsequent one-shot codeRun. Cleared if a 4xx ever comes back, so a rolled
  // sandbox just re-creates lazily on the next call.
  private readonly transientInitialized = new Set<string>()

  constructor(
    private readonly templates: SessionTemplateService,
    private readonly pool: SessionPoolService,
    private readonly sessions: SessionRepository,
    private readonly sandboxService: SandboxService,
    private readonly runnerService: RunnerService,
    private readonly config: TypedConfigService,
    private readonly load: SessionLoadService,
  ) {}

  async codeRun(
    orgId: string,
    organization: Organization,
    req: SessionCodeRunRequestDto,
  ): Promise<SessionCodeRunResponseDto> {
    const start = Date.now()

    if (req.context) {
      const { context, instance } = await this.sessions.resolve(orgId, req.context.id)
      const access = await this.getDaemonAccess(instance)
      await this.load.incrInflight(instance.id)
      try {
        const out = await this.runOnDaemon(access, context.id, req.code, req.env, req.timeout, false)
        return { ...out, durationMs: Date.now() - start }
      } catch (err) {
        // A daemon handshake 404 means the daemon no longer knows this context (sandbox/daemon
        // rolled out from under the still-live DB row). Translate to 410 SessionInvalidatedError —
        // same signal the resolve()/transient paths emit — instead of a generic Error → HTTP 500.
        if (this.isDaemonSessionNotFound(err)) {
          throw new SessionInvalidatedError(context.id, instance.updatedAt ?? new Date())
        }
        throw err
      } finally {
        await this.load.decrInflight(instance.id)
      }
    }

    const tpl = await this.templates.resolve(orgId, req.template ?? this.config.getOrThrow('session.defaultTemplate'))
    // acquire() atomically claims one in-flight slot on the chosen instance; we own releasing it.
    const instance = await this.pool.acquire(orgId, organization, tpl)
    const access = await this.getDaemonAccess(instance)
    const language = req.language ?? this.firstLanguage(tpl.languages)
    const sandboxId = this.requireSandboxId(instance)

    // One-shot path: reuse a deterministic "transient" context instead of creating +
    // tearing down a UUID context per call. Subsequent calls pass `reset: true` on the
    // WS exec frame so the worker rebuilds its globals — same isolation semantics,
    // ~14× faster for Python (no per-call CPython spawn) and equivalent for TypeScript.
    //
    // Scale-out: a single transient context serializes concurrent one-shot ops on one
    // daemon worker. We instead check out a free *slot* per (instance, language) so up to
    // `targetConcurrencyPerSandbox` ops run on distinct daemon contexts in parallel; under
    // contention we fall back to slot 0 (which just serializes, as before).
    const target = this.config.get('session.scale.targetConcurrencyPerSandbox') ?? 4
    const slot = await this.load.checkoutSlot(instance.id, language, target)
    // A reused slot recycles a warm context (reset=true wipes its globals). When the slot pool is
    // exhausted (slot < 0) we use a unique ephemeral context instead of colliding on a shared one
    // — two ops on one daemon context would evict each other's WS client and return empty output.
    const ephemeral = slot < 0
    const transientId = ephemeral
      ? `transient-${instance.id}-${language}-op-${randomUUID()}`
      : `transient-${instance.id}-${language}-${slot}`
    // Memo is keyed by sandbox identity, not just instance id: the pool reuses the same
    // SessionInstance row (same instance.id) when it rolls to a fresh sandbox, so an
    // instance-keyed memo would wrongly report the daemon-side session as live and skip the
    // re-create — leaving the WS exec to 404 on a sandbox that never saw a POST /sessions.
    const memoKey = `${sandboxId}:${transientId}`
    try {
      await this.ensureTransientContext(access, transientId, language, memoKey)
      try {
        const out = await this.runOnDaemon(access, transientId, req.code, req.env, req.timeout, !ephemeral)
        return { ...out, durationMs: Date.now() - start }
      } catch (err) {
        // Self-heal: if the daemon no longer has the transient (sandbox/daemon rolled out
        // from under a stale memo, or the daemon restarted and lost its in-memory sessions),
        // the WS handshake 404s. Drop the memo, recreate the daemon-side context once, and
        // retry. Matches the documented intent that a rolled sandbox re-creates lazily.
        if (!this.isDaemonSessionNotFound(err)) throw err
        this.transientInitialized.delete(memoKey)
        await this.ensureTransientContext(access, transientId, language, memoKey)
        const out = await this.runOnDaemon(access, transientId, req.code, req.env, req.timeout, !ephemeral)
        return { ...out, durationMs: Date.now() - start }
      }
    } finally {
      if (ephemeral) {
        // Ephemeral contexts aren't pooled — tear down the daemon-side session and drop the memo.
        this.transientInitialized.delete(memoKey)
        await this.daemonDeleteSession(access, transientId).catch(() => undefined)
      } else {
        await this.load.releaseSlot(instance.id, language, slot)
      }
      await this.load.decrInflight(instance.id)
    }
  }

  async connect(
    orgId: string,
    organization: Organization,
    req: SessionConnectRequestDto,
  ): Promise<SessionConnectResponseDto> {
    let sessionId: string
    let sandboxId: string
    if (req.context) {
      const { context, instance } = await this.sessions.resolve(orgId, req.context.id)
      sessionId = context.id
      sandboxId = this.requireSandboxId(instance)
    } else {
      const tpl = await this.templates.resolve(orgId, req.template ?? this.config.getOrThrow('session.defaultTemplate'))
      const instance = await this.pool.acquire(orgId, organization, tpl)
      try {
        const language = req.language ?? this.firstLanguage(tpl.languages)
        const ctx = await this.sessions.create(orgId, instance, { language })
        sessionId = ctx.id
        sandboxId = this.requireSandboxId(instance)

        // Pre-create the daemon-side context so the SDK's first WS frame doesn't 404.
        // Best-effort: a daemon-side failure here will surface on the SDK's first send.
        try {
          const internalAccess = await this.getDaemonAccess(instance)
          await this.daemonCreateSession(internalAccess, sessionId, language)
        } catch (err) {
          this.logger.warn(`daemon-side context create failed during connect: ${(err as Error).message}`)
        }
      } finally {
        // Release the optimistic acquire claim; the persistent stream's ongoing load is
        // tracked thereafter via the daemon's polled busy-context count.
        await this.load.decrInflight(instance.id)
      }
    }

    // The SDK still talks to the proxy (not directly to the runner) because the
    // SDK lives outside the cluster — that's the public WS path. Only the
    // API-internal `daemonCreateSession` / `runOnDaemon` calls bypass the
    // proxy (see knob #3).
    const access = await this.buildSandboxAccess(sandboxId, orgId, sessionId)
    return { wsUrl: access.wsUrl, token: access.token, sessionId, expiresAt: access.tokenExpiresAt }
  }

  async createSession(orgId: string, organization: Organization, dto: CreateSessionDto): Promise<SessionDto> {
    const tpl = await this.templates.resolve(orgId, dto.template ?? this.config.getOrThrow('session.defaultTemplate'))
    const instance = await this.pool.acquire(orgId, organization, tpl)
    try {
      const language = dto.language ?? this.firstLanguage(tpl.languages)
      const ctx = await this.sessions.create(orgId, instance, { language, cwd: dto.cwd })

      // Best-effort: pre-create the daemon-side context. If this fails, the row is still valid
      // and the next exec will surface ContextInvalidatedError (translated from the daemon's
      // 404), which is the same clean failure as a rolled context. Keeping this best-effort
      // avoids a complex two-phase commit on a path that's not the steady-state.
      try {
        const internalAccess = await this.getDaemonAccess(instance)
        await this.daemonCreateSession(internalAccess, ctx.id, language, dto.cwd)
      } catch (err) {
        this.logger.warn(`daemon-side context create failed (will retry on next exec): ${err.message}`)
      }

      // Mint the SDK's direct-to-sandbox handle inline — saves the SDK an
      // immediate GET /access round-trip on the very first run().
      const sandboxId = this.requireSandboxId(instance)
      const dtoOut = this.sessions.toDto(ctx)
      try {
        dtoOut.access = await this.buildSandboxAccess(sandboxId, orgId, ctx.id)
      } catch (err) {
        this.logger.warn(`buildSandboxAccess failed (SDK will refresh on first use): ${(err as Error).message}`)
      }
      return dtoOut
    } finally {
      await this.load.decrInflight(instance.id)
    }
  }

  async listSessions(orgId: string, templateName?: string): Promise<SessionDto[]> {
    let templateId: string | undefined
    if (templateName) {
      const tpl = await this.templates.resolve(orgId, templateName)
      templateId = tpl.id
    }
    return this.sessions.list(orgId, templateId)
  }

  async deleteSession(orgId: string, sessionId: string): Promise<void> {
    // Best-effort daemon-side delete first, then DB.
    try {
      const { context, instance } = await this.sessions.resolve(orgId, sessionId)
      if (instance.sandboxId) {
        const access = await this.getDaemonAccess(instance)
        await this.daemonDeleteSession(access, context.id).catch(() => undefined)
      }
    } catch {
      // Resolve may throw 410/404 — proceed with DB delete regardless.
    }
    await this.sessions.delete(orgId, sessionId)
  }

  async listTemplates(orgId: string) {
    return this.templates.list(orgId)
  }

  async listPackages(
    orgId: string,
    organization: Organization,
    templateName: string,
    language: string,
  ): Promise<SessionPackageDto[]> {
    const tpl = await this.templates.resolve(orgId, templateName)
    const instance = await this.pool.acquire(orgId, organization, tpl)
    try {
      const access = await this.getDaemonAccess(instance)

      const resp = await fetch(`${access.url}/packages?language=${encodeURIComponent(language)}`, {
        headers: { Authorization: `Bearer ${access.runnerApiKey}` },
      })
      if (!resp.ok) {
        this.logger.warn(`listPackages from daemon ${resp.status}`)
        return []
      }
      return (await resp.json()) as SessionPackageDto[]
    } finally {
      await this.load.decrInflight(instance.id)
    }
  }

  // -- daemon helpers (the real WS client lives in the SDK; the API talks HTTP only) -----

  /**
   * Build the SDK-facing direct-to-sandbox access bundle (signed proxy URL +
   * token). This is the public path the SDK uses to talk to the daemon
   * straight from outside the cluster — same chain `sandbox.process.code_run`
   * uses against the classic daytona-daemon. Different from `getDaemonAccess`
   * below which is an API-internal shortcut that bypasses the proxy.
   */
  async buildSandboxAccess(sandboxId: string, orgId: string, sessionId: string): Promise<SessionAccessDto> {
    const port = this.config.get('session.daemonPort') ?? 2281
    const ttl = this.config.get('session.connectTokenTtlSeconds') ?? 300
    const signed = await this.sandboxService.getSignedPortPreviewUrl(sandboxId, orgId, port, ttl)
    const httpUrl = signed.url.replace(/\/$/, '')
    return {
      httpUrl,
      wsUrl: `${httpUrl.replace(/^http/, 'ws')}/sessions/${encodeURIComponent(sessionId)}/execute`,
      token: signed.token,
      tokenExpiresAt: new Date(Date.now() + ttl * 1000).toISOString(),
    }
  }

  /**
   * Refresh-or-mint the SDK's direct-to-sandbox handle for an existing context.
   * Acts as a keep-alive: bumps `lastUsedAt` so the idle GC respects the SDK's
   * continued use of the context even though no exec is hitting the API.
   * Surfaces the same {Not}{Invalidated}{Expired} errors `resolve()` does.
   */
  async getSessionAccess(orgId: string, sessionId: string): Promise<SessionAccessDto> {
    const { context, instance } = await this.sessions.resolve(orgId, sessionId)
    const sandboxId = this.requireSandboxId(instance)
    return this.buildSandboxAccess(sandboxId, orgId, context.id)
  }

  /**
   * SDK one-shot entrypoint: returns an `SessionDto` for the deterministic
   * transient context (`transient-${instance.id}-${language}`) bound to a warm
   * pool sandbox + an `access` handle the SDK can WS-stream against directly.
   * Reuses `ensureTransientContext` — same daemon-side primitive `codeRun` uses,
   * just with the access bundle attached for the SDK's WS path.
   */
  async createTransientSession(
    orgId: string,
    organization: Organization,
    dto: CreateSessionTransientDto,
  ): Promise<SessionDto> {
    const tpl = await this.templates.resolve(orgId, dto.template ?? this.config.getOrThrow('session.defaultTemplate'))
    const instance = await this.pool.acquire(orgId, organization, tpl)
    try {
      const language = dto.language ?? this.firstLanguage(tpl.languages)
      const sandboxId = this.requireSandboxId(instance)
      // SDK transient is a stable, reusable handle: deterministic per (instance, language) so
      // repeat calls return the same id (the SDK streams against it directly). Intra-sandbox
      // parallelism via the slot pool applies to the one-shot codeRun path, not this one.
      const transientId = `transient-${instance.id}-${language}`

      const internalAccess = await this.getDaemonAccess(instance)
      await this.ensureTransientContext(internalAccess, transientId, language, `${sandboxId}:${transientId}`)

      const access = await this.buildSandboxAccess(sandboxId, orgId, transientId)
      const now = new Date()
      return {
        id: transientId,
        language,
        cwd: undefined,
        createdAt: now.toISOString(),
        lastUsedAt: now.toISOString(),
        // Transients have no DB row, so we surface the token expiry as the
        // context's `expiresAt` — refresh /access to renew (lazy in the SDK).
        expiresAt: access.tokenExpiresAt,
        access,
      }
    } finally {
      await this.load.decrInflight(instance.id)
    }
  }

  /**
   * Resolve the runner once and build the direct `runner.apiUrl` URL that
   * proxies into the in-sandbox session-daemon. Skips the public proxy on
   * port 4000 (knob #3) — saves one local TCP hop, one auth round-trip, and
   * one URL signing pass per API-initiated daemon call.
   */
  private async getDaemonAccess(instance: SessionInstance): Promise<DaemonAccess> {
    const sandboxId = this.requireSandboxId(instance)
    const port = this.config.get('session.daemonPort') ?? 2281
    const runner = await this.runnerService.findBySandboxId(sandboxId)
    return buildDaemonAccess(runner, sandboxId, port)
  }

  /**
   * Idempotent POST /sessions with an in-process memo so the second one-shot
   * call per (instance, language) doesn't pay the round-trip. The daemon
   * already returns 409 on duplicates; the memo just avoids the HTTP at all.
   */
  private async ensureTransientContext(
    access: DaemonAccess,
    id: string,
    language: string,
    memoKey: string,
  ): Promise<void> {
    if (this.transientInitialized.has(memoKey)) return
    try {
      await this.daemonCreateSession(access, id, language)
      this.transientInitialized.add(memoKey)
    } catch (err) {
      // Drop the optimistic memo so the next call retries the create.
      this.transientInitialized.delete(memoKey)
      throw err
    }
  }

  /**
   * True when an error from `runOnDaemon` is the daemon reporting the session
   * id is unknown (WS handshake 404). Distinct from a runner-proxy 400 ("sandbox
   * not started / no container IP"), which is a not-ready condition we must NOT
   * paper over with a re-create.
   */
  private isDaemonSessionNotFound(err: unknown): boolean {
    const status = (err as { status?: number })?.status
    if (status === 404) return true
    const msg = (err as Error)?.message ?? ''
    return /\b404\b/.test(msg)
  }

  private async daemonCreateSession(access: DaemonAccess, id: string, language: string, cwd?: string): Promise<void> {
    const body = JSON.stringify({ id, language, cwd })
    const resp = await fetch(`${access.url}/sessions`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${access.runnerApiKey}`, 'Content-Type': 'application/json' },
      body,
    })
    if (!resp.ok && resp.status !== 409) {
      // Surface the daemon's response body so callers (and exception logs)
      // see the actual reason (`unsupported language`, `id is required`, the
      // raw bind error, etc.) instead of a bare status code.
      const detail = await this.readBodyForError(resp)
      throw new Error(`daemon POST /sessions failed: ${resp.status} ${detail} (id=${id}, language=${language})`)
    }
  }

  private async daemonDeleteSession(access: DaemonAccess, id: string): Promise<void> {
    const resp = await fetch(`${access.url}/sessions/${encodeURIComponent(id)}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${access.runnerApiKey}` },
    })
    // 404 is fine (already gone); 5xx surfaces as a logged warning at call sites.
    if (!resp.ok && resp.status !== 404) {
      const detail = await this.readBodyForError(resp)
      throw new Error(`daemon DELETE /sessions/${id} returned ${resp.status} ${detail}`)
    }
    this.transientInitialized.delete(id)
  }

  /**
   * Best-effort read of a non-OK daemon response body. Truncated so a huge
   * accidental HTML page doesn't blow up the log, and tolerant of bodies that
   * have already been consumed or are unreadable.
   */
  private async readBodyForError(resp: Response): Promise<string> {
    try {
      const text = await resp.text()
      const trimmed = text.trim()
      if (!trimmed) return ''
      return trimmed.length > 500 ? trimmed.slice(0, 500) + '…' : trimmed
    } catch {
      return ''
    }
  }

  private async runOnDaemon(
    access: DaemonAccess,
    sessionId: string,
    code: string,
    env?: Record<string, string>,
    timeout?: number,
    reset = false,
  ): Promise<Omit<SessionCodeRunResponseDto, 'durationMs'>> {
    // Convert the runner http(s) URL into the matching ws(s) URL and aggregate the
    // streamed frames into a synchronous response. We reuse the same daemon WS endpoint the
    // SDK streams against — there's no separate "run-once" REST surface to maintain.
    const wsUrl = access.url.replace(/^http/, 'ws') + `/sessions/${encodeURIComponent(sessionId)}/execute`
    const ws = new WebSocket(wsUrl, {
      headers: { Authorization: `Bearer ${access.runnerApiKey}` },
      handshakeTimeout: this.config.get('session.healthcheckTimeoutMs') ?? 60000,
    })

    // Aggregate output is buffered in-memory in the shared multi-tenant API process,
    // so an unbounded run would be a heap-pressure/DoS vector. Cap each stream (and the
    // displays array) at 5 MiB; once a stream hits the cap we stop appending and add a
    // single truncation marker. Small-output runs are unaffected.
    const MAX_AGGREGATE_OUTPUT_BYTES = 5 * 1024 * 1024
    const TRUNCATION_MARKER = '\n…[output truncated]…'
    let stdout = ''
    let stderr = ''
    let stdoutBytes = 0
    let stderrBytes = 0
    let stdoutTruncated = false
    let stderrTruncated = false
    let execError: ExecutionErrorDto | undefined
    const displays: DisplayDataDto[] = []
    let displaysBytes = 0
    let displaysTruncated = false

    return new Promise((resolve, reject) => {
      const hardTimeoutMs =
        timeout && timeout > 0 ? (timeout + 5) * 1000 : (this.config.get('session.execTimeoutSeconds') ?? 600) * 1000
      const timer = setTimeout(() => {
        try {
          ws.close(4008, 'timeout')
        } catch {
          /* swallow */
        }
        reject(new Error(`session code-run exceeded ${hardTimeoutMs}ms`))
      }, hardTimeoutMs)

      // Surface the daemon's HTTP handshake status (e.g. 404 = unknown session)
      // as a typed error so callers can distinguish "recreate the context and
      // retry" from a transport failure. Fires before 'error'; settling here
      // makes the later 'error'/'close' rejects no-ops.
      ws.once('unexpected-response', (_req, res) => {
        clearTimeout(timer)
        const e = new Error(`session daemon WS handshake failed: ${res.statusCode}`) as Error & { status?: number }
        e.status = res.statusCode
        try {
          ws.terminate()
        } catch {
          /* swallow */
        }
        reject(e)
      })

      ws.once('open', () => {
        const firstFrame = JSON.stringify({ code, envs: env ?? {}, timeout: timeout ?? 0, reset })
        ws.send(firstFrame)
      })

      ws.on('message', (data) => {
        let frame: DaemonFrame
        try {
          frame = JSON.parse(data.toString()) as DaemonFrame
        } catch (e) {
          this.logger.warn(`session daemon emitted non-JSON frame: ${(e as Error).message}`)
          return
        }
        switch (frame.type) {
          case 'stdout':
            if (!stdoutTruncated) {
              const chunk = frame.text ?? ''
              stdoutBytes += Buffer.byteLength(chunk, 'utf8')
              if (stdoutBytes <= MAX_AGGREGATE_OUTPUT_BYTES) {
                stdout += chunk
              } else {
                stdout += TRUNCATION_MARKER
                stdoutTruncated = true
              }
            }
            break
          case 'stderr':
            if (!stderrTruncated) {
              const chunk = frame.text ?? ''
              stderrBytes += Buffer.byteLength(chunk, 'utf8')
              if (stderrBytes <= MAX_AGGREGATE_OUTPUT_BYTES) {
                stderr += chunk
              } else {
                stderr += TRUNCATION_MARKER
                stderrTruncated = true
              }
            }
            break
          case 'error':
            execError = {
              name: frame.name ?? 'Error',
              value: frame.value,
              traceback: frame.traceback,
            }
            break
          case 'display':
            if (frame.formats && frame.data && !displaysTruncated) {
              displaysBytes += Object.values(frame.data).reduce((acc, d) => acc + Buffer.byteLength(d ?? '', 'utf8'), 0)
              if (displaysBytes <= MAX_AGGREGATE_OUTPUT_BYTES) {
                displays.push({ formats: frame.formats, data: frame.data })
              } else {
                displays.push({ formats: ['text/plain'], data: { 'text/plain': TRUNCATION_MARKER } })
                displaysTruncated = true
              }
            }
            break
          case 'control':
            // The daemon emits a 'completed' control frame and then closes the socket;
            // we don't have to act on it here.
            break
          default:
            this.logger.debug(`session daemon: unknown frame type ${frame.type}`)
        }
      })

      ws.once('close', (code: number, reason: Buffer) => {
        clearTimeout(timer)
        // The daemon closes 1000 (normal) on success. A non-normal code means the daemon
        // tore the stream down on an internal failure (gorilla CloseInternalServerErr=1011,
        // with an error reason) or the transport dropped abnormally (1006). Resolving those
        // as success would hand the caller partial stdout/stderr as if the run completed —
        // reject so it surfaces as an error instead.
        if (code !== 1000) {
          const detail = reason?.toString().trim()
          reject(new Error(`session daemon WS closed abnormally: ${code}${detail ? ` ${detail}` : ''}`))
          return
        }
        resolve({
          stdout,
          stderr,
          error: execError,
          displays: displays.length > 0 ? displays : undefined,
        })
      })

      ws.once('error', (err) => {
        clearTimeout(timer)
        reject(err)
      })
    })
  }

  private firstLanguage(langs: string[]): string {
    return langs[0] ?? 'python'
  }

  /**
   * READY instances always carry a sandboxId (the pool sets it before flipping state).
   * The throw guards against a stale cache hit returning an instance whose sandboxId
   * wasn't part of the cached blob — a defensive net rather than an expected case.
   */
  private requireSandboxId(instance: { sandboxId?: string }): string {
    if (!instance.sandboxId) {
      throw new Error('session instance has no sandboxId yet (likely PROVISIONING or stale cache)')
    }
    return instance.sandboxId
  }
}
