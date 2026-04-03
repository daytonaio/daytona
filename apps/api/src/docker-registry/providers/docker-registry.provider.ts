/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { HttpService } from '@nestjs/axios'
import { AxiosRequestConfig, AxiosResponse } from 'axios'
import { firstValueFrom } from 'rxjs'
import {
  IDockerRegistryProvider,
  RegistryCredentialsValidationError,
  RegistryCredentialsValidationErrorCode,
} from './docker-registry.provider.interface'

const VALIDATION_TIMEOUT_MS = 5000

interface RegistryAuthChallenge {
  scheme: string
  params: Map<string, string[]>
}

@Injectable()
export class DockerRegistryProvider implements IDockerRegistryProvider {
  constructor(private readonly httpService: HttpService) {}

  async validateCredentials(baseUrl: string, auth: { username: string; password: string }): Promise<void> {
    const registryUrl = `${baseUrl.replace(/\/+$/, '')}/v2/`
    const challengeResponse = await this.request({
      method: 'get',
      url: registryUrl,
    })

    // A public or misconfigured registry can return 200 without challenging us.
    // In that case the submitted credentials were never proven, so we must reject.
    if (challengeResponse.status === 200) {
      throw new RegistryCredentialsValidationError(
        RegistryCredentialsValidationErrorCode.UNVERIFIED_CREDENTIALS,
        'Registry is reachable but submitted credentials could not be verified',
      )
    }

    // A secured registry should challenge unauthenticated /v2/ with 401.
    // Any other status means we can't perform a trustworthy auth check.
    if (challengeResponse.status !== 401) {
      throw new RegistryCredentialsValidationError(
        RegistryCredentialsValidationErrorCode.UNREACHABLE,
        `Registry returned unexpected status ${challengeResponse.status} during validation`,
      )
    }

    const challenge = this.parseAuthenticateChallenge(challengeResponse)

    if (challenge.scheme === 'basic') {
      await this.validateBasicChallenge(registryUrl, auth)
      return
    }

    if (challenge.scheme === 'bearer') {
      await this.validateBearerChallenge(registryUrl, auth, challenge)
      return
    }

    throw new RegistryCredentialsValidationError(
      RegistryCredentialsValidationErrorCode.UNSUPPORTED_CHALLENGE,
      `Unsupported registry authentication scheme: ${challenge.scheme}`,
    )
  }

  async createRobotAccount(
    url: string,
    auth: { username: string; password: string },
    robotConfig: {
      name: string
      description: string
      duration: number
      level: string
      permissions: Array<{
        kind: string
        namespace: string
        access: Array<{ resource: string; action: string }>
      }>
    },
  ): Promise<{ name: string; secret: string }> {
    const response = await firstValueFrom(this.httpService.post(url, robotConfig, { auth }))
    return {
      name: response.data.name,
      secret: response.data.secret,
    }
  }

  async deleteArtifact(
    baseUrl: string,
    auth: { username: string; password: string },
    params: { project: string; repository: string; tag: string },
  ): Promise<void> {
    const url = `${baseUrl}/api/v2.0/projects/${params.project}/repositories/${params.repository}/artifacts/${params.tag}`

    try {
      await firstValueFrom(this.httpService.delete(url, { auth }))
    } catch (error) {
      if (error.response?.status === 404) {
        return // Artifact not found, consider it a success
      }
      throw error
    }
  }

  /**
   * Retries the registry probe using HTTP Basic authentication.
   */
  private async validateBasicChallenge(
    registryUrl: string,
    auth: { username: string; password: string },
  ): Promise<void> {
    const response = await this.request({
      method: 'get',
      url: registryUrl,
      auth,
    })

    // Basic auth is valid once the authenticated retry is accepted.
    if (response.status === 200) {
      return
    }

    // The registry challenged us, we retried with credentials, and it still refused.
    if (response.status === 401 || response.status === 403) {
      throw new RegistryCredentialsValidationError(
        RegistryCredentialsValidationErrorCode.INVALID_CREDENTIALS,
        'Registry rejected the submitted credentials',
      )
    }

    throw new RegistryCredentialsValidationError(
      RegistryCredentialsValidationErrorCode.UNVERIFIED_CREDENTIALS,
      `Registry returned unexpected status ${response.status} after Basic authentication`,
    )
  }

  /**
   * Completes the token-service challenge, then retries the registry probe with a Bearer token.
   */
  private async validateBearerChallenge(
    registryUrl: string,
    auth: { username: string; password: string },
    challenge: RegistryAuthChallenge,
  ): Promise<void> {
    const realm = challenge.params.get('realm')?.[0]
    if (!realm) {
      throw new RegistryCredentialsValidationError(
        RegistryCredentialsValidationErrorCode.UNSUPPORTED_CHALLENGE,
        'Registry Bearer challenge did not include a realm',
      )
    }

    let tokenUrl: URL
    try {
      tokenUrl = new URL(realm)
    } catch {
      throw new RegistryCredentialsValidationError(
        RegistryCredentialsValidationErrorCode.UNSUPPORTED_CHALLENGE,
        'Registry Bearer challenge included an invalid realm URL',
      )
    }

    // Prevent SSRF: only follow HTTPS realm URLs (or HTTP for localhost).
    const scheme = tokenUrl.protocol
    const hostname = tokenUrl.hostname
    if (
      scheme !== 'https:' &&
      !(scheme === 'http:' && (hostname === 'localhost' || hostname === '127.0.0.1' || hostname === 'registry'))
    ) {
      throw new RegistryCredentialsValidationError(
        RegistryCredentialsValidationErrorCode.UNSUPPORTED_CHALLENGE,
        `Registry Bearer realm uses disallowed scheme: ${scheme}`,
      )
    }

    const service = challenge.params.get('service')?.[0]
    const scopes = challenge.params.get('scope') ?? []

    if (service) {
      tokenUrl.searchParams.set('service', service)
    }

    for (const scope of scopes) {
      tokenUrl.searchParams.append('scope', scope)
    }

    const tokenResponse = await this.request({
      method: 'get',
      url: tokenUrl.toString(),
      auth,
    })

    // The token service refused the submitted username/password.
    if (tokenResponse.status === 401 || tokenResponse.status === 403) {
      throw new RegistryCredentialsValidationError(
        RegistryCredentialsValidationErrorCode.INVALID_CREDENTIALS,
        'Token service rejected the submitted credentials',
      )
    }

    if (tokenResponse.status !== 200) {
      throw new RegistryCredentialsValidationError(
        RegistryCredentialsValidationErrorCode.UNVERIFIED_CREDENTIALS,
        `Token service returned unexpected status ${tokenResponse.status}`,
      )
    }

    const token = tokenResponse.data?.token ?? tokenResponse.data?.access_token
    if (!token || typeof token !== 'string') {
      throw new RegistryCredentialsValidationError(
        RegistryCredentialsValidationErrorCode.UNSUPPORTED_CHALLENGE,
        'Token service did not return a usable Bearer token',
      )
    }

    const response = await this.request({
      method: 'get',
      url: registryUrl,
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })

    // Bearer auth is valid once the registry accepts the token-backed retry.
    if (response.status === 200) {
      return
    }

    // The registry challenged us, we exchanged the credentials for a token, and the retry still failed.
    if (response.status === 401 || response.status === 403) {
      throw new RegistryCredentialsValidationError(
        RegistryCredentialsValidationErrorCode.INVALID_CREDENTIALS,
        'Registry rejected the token generated from the submitted credentials',
      )
    }

    throw new RegistryCredentialsValidationError(
      RegistryCredentialsValidationErrorCode.UNVERIFIED_CREDENTIALS,
      `Registry returned unexpected status ${response.status} after Bearer authentication`,
    )
  }

  /**
   * Parses the registry auth challenge and preserves repeated keys such as scope.
   */
  private parseAuthenticateChallenge(response: AxiosResponse): RegistryAuthChallenge {
    const rawHeader = response.headers['www-authenticate']
    const header = Array.isArray(rawHeader) ? rawHeader.find((value) => typeof value === 'string') : rawHeader

    if (typeof header !== 'string' || header.length === 0) {
      throw new RegistryCredentialsValidationError(
        RegistryCredentialsValidationErrorCode.UNSUPPORTED_CHALLENGE,
        'Registry did not provide a usable WWW-Authenticate challenge',
      )
    }

    const schemeMatch = header.match(/^\s*([A-Za-z]+)\s+/)
    if (!schemeMatch) {
      throw new RegistryCredentialsValidationError(
        RegistryCredentialsValidationErrorCode.UNSUPPORTED_CHALLENGE,
        'Registry authentication challenge is malformed',
      )
    }

    const params = new Map<string, string[]>()
    const paramPattern = /([A-Za-z][A-Za-z0-9_-]*)="([^"]*)"/g
    let match: RegExpExecArray | null

    while ((match = paramPattern.exec(header)) !== null) {
      const [, key, value] = match
      params.set(key.toLowerCase(), [...(params.get(key.toLowerCase()) ?? []), value])
    }

    return {
      scheme: schemeMatch[1].toLowerCase(),
      params,
    }
  }

  /**
   * Sends a registry-validation HTTP request and converts transport failures into typed errors.
   */
  private async request(config: AxiosRequestConfig): Promise<AxiosResponse> {
    try {
      return await firstValueFrom(
        this.httpService.request({
          ...config,
          timeout: VALIDATION_TIMEOUT_MS,
          validateStatus: () => true,
        }),
      )
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error)
      throw new RegistryCredentialsValidationError(
        RegistryCredentialsValidationErrorCode.UNREACHABLE,
        `Registry validation request failed: ${message}`,
      )
    }
  }
}
