/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, ServiceUnavailableException } from '@nestjs/common'
import { TypedConfigService } from '../../../config/typed-config.service'
import {
  LayeredVolumeProvider,
  CreateDiskOptions,
  CreateDiskResult,
  MintMountKeyOptions,
  MintMountKeyResult,
} from './layered-volume.provider'

// Fallback per-region control-plane URLs; overridden by `LAYERED_CONTROL_URL_<REGION>`.
const DEFAULT_CONTROL_URLS: Record<string, string> = {
  'aws-us-east-1': 'https://control.green.us-east-1.aws.prod.example.com',
  'aws-eu-west-1': 'https://control.green.eu-west-1.aws.prod.example.com',
  'aws-us-west-2': 'https://control.green.us-west-2.aws.prod.example.com',
  'gcp-us-central1': 'https://control.blue.us-central1.gcp.prod.example.com',
}

export type {
  DiskMount as LayeredDiskMount,
  CreateDiskOptions as CreateLayeredDiskOptions,
} from './layered-volume.provider'
export type { CreateDiskResult as LayeredDisk } from './layered-volume.provider'

interface ApiResponseEnvelope<T> {
  success: boolean
  error?: string
  data?: T
}

interface CreateDiskResponseData {
  diskId: string
  authorizedUsers?: Array<{
    type?: string
    token?: string
    identifier?: string
    nickname?: string
  }>
}

interface AddDiskUserResponseData {
  type?: string
  token?: string
  identifier?: string
  nickname?: string
}

export type { MintMountKeyResult, MintMountKeyOptions } from './layered-volume.provider'

const MAX_ATTEMPTS = 4
const PER_ATTEMPT_TIMEOUT_MS = 30_000
const BACKOFF_BASE_MS = 500
const BACKOFF_CAP_MS = 5_000
const RETRY_AFTER_CAP_MS = 30_000
const RETRYABLE_STATUSES: ReadonlySet<number> = new Set([408, 429, 500, 502, 503, 504])

@Injectable()
export class LayeredVolumeClient implements LayeredVolumeProvider {
  private readonly logger = new Logger(LayeredVolumeClient.name)
  private readonly apiKey?: string
  private readonly defaultRegion: string

  constructor(private readonly configService: TypedConfigService) {
    this.apiKey = this.configService.get('layered.apiKey')
    this.defaultRegion = this.configService.get('layered.defaultRegion') || 'aws-us-east-1'
  }

  isConfigured(): boolean {
    return Boolean(this.apiKey) || this.hasAnyRegionApiKey()
  }

  getDefaultRegion(): string {
    return this.defaultRegion
  }

  async createDisk(opts: CreateDiskOptions): Promise<CreateDiskResult> {
    this.assertConfigured()
    const region = opts.region || this.defaultRegion
    const baseUrl = this.resolveControlUrl(region)
    const apiKey = this.resolveApiKey(region)

    const res = await this.request<CreateDiskResponseData>(`${baseUrl}/api/disks`, {
      method: 'POST',
      apiKey,
      body: JSON.stringify({
        name: opts.name,
        // One S3 mount per disk: a 1:1 view of a Daytona-owned bucket.
        mounts: [opts.mount],
      }),
    })

    if (!res.diskId) {
      throw new Error('createDisk response missing diskId')
    }

    // Token users only; AWS STS users are out of scope.
    const tokenUser = res.authorizedUsers?.find((u) => u.type === 'token' && u.token)
    if (!tokenUser?.token) {
      throw new Error(
        `createDisk response did not include a generated token (diskId=${res.diskId}). ` +
          `The disk exists but is not mountable; delete it manually or open a support ticket.`,
      )
    }

    return {
      diskId: res.diskId,
      region,
      mountToken: tokenUser.token,
    }
  }

  // 404 treated as success so retries after a partial delete are safe.
  async deleteDisk(diskId: string, region: string): Promise<void> {
    this.assertConfigured()
    const baseUrl = this.resolveControlUrl(region)
    const apiKey = this.resolveApiKey(region)

    await this.request<unknown>(`${baseUrl}/api/disks/${encodeURIComponent(diskId)}`, {
      method: 'DELETE',
      apiKey,
      treat404AsOk: true,
    })
  }

  async mintMountKey(opts: MintMountKeyOptions): Promise<MintMountKeyResult> {
    this.assertConfigured()
    const baseUrl = this.resolveControlUrl(opts.region)
    const apiKey = this.resolveApiKey(opts.region)

    const res = await this.request<AddDiskUserResponseData>(
      `${baseUrl}/api/disks/${encodeURIComponent(opts.diskId)}/users`,
      {
        method: 'POST',
        apiKey,
        body: JSON.stringify({
          type: 'token',
          nickname: opts.nickname,
        }),
      },
    )

    if (!res?.token) {
      throw new Error(`mintMountKey response missing token for disk ${opts.diskId}`)
    }
    if (!res?.identifier) {
      throw new Error(`mintMountKey response missing identifier for disk ${opts.diskId}`)
    }
    return { token: res.token, identifier: res.identifier }
  }

  // 404 treated as success so double-revokes after partial destroy are safe.
  async revokeMountKey(diskId: string, region: string, identifier: string): Promise<void> {
    this.assertConfigured()
    const baseUrl = this.resolveControlUrl(region)
    const apiKey = this.resolveApiKey(region)

    const params = new URLSearchParams({ identifier })
    await this.request<unknown>(`${baseUrl}/api/disks/${encodeURIComponent(diskId)}/users/token?${params.toString()}`, {
      method: 'DELETE',
      apiKey,
      treat404AsOk: true,
    })
  }

  private assertConfigured(): void {
    if (!this.isConfigured()) {
      throw new ServiceUnavailableException(
        'Layered volume control plane is not configured. Set LAYERED_API_KEY (or a per-region ' +
          'LAYERED_API_KEY_<REGION>) to enable the layered volume backend.',
      )
    }
  }

  private hasAnyRegionApiKey(): boolean {
    return Object.keys(process.env).some((k) => k.startsWith('LAYERED_API_KEY_') && Boolean(process.env[k]))
  }

  // Archil keys are region-scoped: `LAYERED_API_KEY_<REGION>` wins, else global `LAYERED_API_KEY`.
  private resolveApiKey(region: string): string {
    const overrideKey = `LAYERED_API_KEY_${region.toUpperCase().replace(/-/g, '_')}`
    const override = process.env[overrideKey]
    if (override) {
      return override
    }
    if (this.apiKey) {
      return this.apiKey
    }
    throw new ServiceUnavailableException(
      `No layered API key configured for region "${region}". Archil API keys are region-scoped; ` +
        `set ${overrideKey} (or a global LAYERED_API_KEY fallback).`,
    )
  }

  private resolveControlUrl(region: string): string {
    const overrideKey = `LAYERED_CONTROL_URL_${region.toUpperCase().replace(/-/g, '_')}`
    const override = process.env[overrideKey]
    if (override) {
      return override.replace(/\/$/, '')
    }
    const fallback = DEFAULT_CONTROL_URLS[region]
    if (!fallback) {
      throw new Error(
        `Unknown layered region "${region}". Set ${overrideKey} to its control-plane URL or pick a known region key.`,
      )
    }
    return fallback
  }

  // Retries POSTs too, so createDisk/mintMountKey can leak a duplicate disk or
  // token on a missed response. Accepted tradeoff: orphans are cheap, failures aren't.
  private async request<T>(
    url: string,
    init: { method: 'GET' | 'POST' | 'DELETE'; apiKey: string; body?: string; treat404AsOk?: boolean },
  ): Promise<T> {
    let lastError: Error | undefined

    for (let attempt = 1; attempt <= MAX_ATTEMPTS; attempt++) {
      const controller = new AbortController()
      const timeoutHandle = setTimeout(() => controller.abort(), PER_ATTEMPT_TIMEOUT_MS)

      let response: Response
      try {
        response = await fetch(url, {
          method: init.method,
          headers: {
            Authorization: `key-${init.apiKey}`,
            'Content-Type': 'application/json',
          },
          body: init.body,
          signal: controller.signal,
        })
      } catch (err) {
        const message = controller.signal.aborted
          ? `request timed out after ${PER_ATTEMPT_TIMEOUT_MS}ms`
          : errorMessage(err)
        lastError = new Error(message)

        if (attempt < MAX_ATTEMPTS) {
          const delayMs = backoffDelayMs(attempt)
          this.logger.warn(
            `Layered ${init.method} ${url} network error on attempt ${attempt}/${MAX_ATTEMPTS} (${message}); retrying in ${delayMs}ms`,
          )
          await sleep(delayMs)
          continue
        }
        throw new Error(
          `Layered control plane unreachable: ${init.method} ${url} failed after ${MAX_ATTEMPTS} attempts (${message})`,
        )
      } finally {
        clearTimeout(timeoutHandle)
      }

      if (init.treat404AsOk && response.status === 404) {
        return undefined as T
      }

      if (RETRYABLE_STATUSES.has(response.status) && attempt < MAX_ATTEMPTS) {
        // Drain the body so the socket can be reused.
        const raw = await response.text().catch(() => '')
        const delayMs = retryAfterMs(response) ?? backoffDelayMs(attempt)
        this.logger.warn(
          `Layered ${init.method} ${url} returned ${response.status} on attempt ${attempt}/${MAX_ATTEMPTS}; retrying in ${delayMs}ms${
            raw ? `: ${raw.slice(0, 200)}` : ''
          }`,
        )
        lastError = new Error(`${response.status} ${response.statusText || 'transient error'}`)
        await sleep(delayMs)
        continue
      }

      let envelope: ApiResponseEnvelope<T>
      const raw = await response.text()
      try {
        envelope = raw ? (JSON.parse(raw) as ApiResponseEnvelope<T>) : { success: response.ok }
      } catch {
        throw new Error(
          `Layered ${init.method} ${url} returned non-JSON response (status ${response.status}): ${raw.slice(0, 200)}`,
        )
      }

      if (!response.ok || envelope.success === false) {
        const message = envelope.error || `${init.method} ${url} failed with status ${response.status}`
        throw new Error(`Layered control plane error: ${message}`)
      }

      return envelope.data as T
    }

    throw lastError ?? new Error(`Layered ${init.method} ${url} failed after ${MAX_ATTEMPTS} attempts`)
  }
}

// Exponential backoff with full jitter to avoid thundering-herd retries.
function backoffDelayMs(attempt: number): number {
  const upper = Math.min(BACKOFF_BASE_MS * 2 ** (attempt - 1), BACKOFF_CAP_MS)
  return Math.floor(Math.random() * upper)
}

// Retry-After per RFC 9110 §10.2.3 (delta-seconds or HTTP-date), capped.
function retryAfterMs(response: Response): number | undefined {
  const header = response.headers.get('Retry-After')
  if (!header) return undefined

  const seconds = Number(header)
  if (Number.isFinite(seconds) && seconds >= 0) {
    return Math.min(seconds * 1000, RETRY_AFTER_CAP_MS)
  }

  const dateMs = Date.parse(header)
  if (!Number.isNaN(dateMs)) {
    return Math.max(0, Math.min(dateMs - Date.now(), RETRY_AFTER_CAP_MS))
  }

  return undefined
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

function errorMessage(err: unknown): string {
  if (err instanceof Error) return err.message
  return String(err)
}
