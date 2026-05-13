/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { ECRClient, GetAuthorizationTokenCommand } from '@aws-sdk/client-ecr'
import { STSClient, GetCallerIdentityCommand } from '@aws-sdk/client-sts'
import { fromTemporaryCredentials } from '@aws-sdk/credential-providers'
import { TypedConfigService } from '../../config/typed-config.service'

interface EcrAuth {
  username: string
  password: string
}

const ECR_HOST_REGEX = /^\d+\.dkr\.ecr\.([a-z0-9-]+)\.amazonaws\.com$/
// Refresh ahead of AWS-side expiry (12h) to absorb clock skew.
const REFRESH_BUFFER_SEC = 30 * 60

@Injectable()
export class EcrCredentialsService {
  private readonly logger = new Logger(EcrCredentialsService.name)
  private readonly brokerRoleArn: string
  private apiArn = ''

  constructor(
    @InjectRedis() private readonly redis: Redis,
    configService: TypedConfigService,
  ) {
    this.brokerRoleArn = configService.get('ecr.brokerRoleArn')?.trim() ?? ''
  }

  isEcrUrl(url: string): boolean {
    return ECR_HOST_REGEX.test(stripScheme(url))
  }

  /**
   * Resolves a fresh `AWS:<token>` Docker auth pair for an ECR registry.
   * AssumeRoles the supplied ARN, unless it matches the API's own identity
   * (auto-detected via STS) — then the API's creds are used directly.
   * Optionally hops through a broker role first when set. Cached in Redis
   * with TTL derived from AWS's `expiresAt`.
   */
  async resolveEcrCredentials(url: string, roleArn: string, externalId: string): Promise<EcrAuth> {
    const match = ECR_HOST_REGEX.exec(stripScheme(url))
    if (!match) {
      throw new Error(`Not an ECR URL: ${url}`)
    }
    const region = match[1]
    const normalizedArn = roleArn.trim()

    const apiArn = await this.getApiArn()
    const useApiIdentity = apiArn !== '' && apiArn === normalizedArn

    const cacheKey = useApiIdentity
      ? `ecr:token:default:${region}`
      : `ecr:token:${externalId}:${normalizedArn}:${region}`
    const cached = await this.redis.get(cacheKey)
    if (cached) {
      try {
        return JSON.parse(cached) as EcrAuth
      } catch {
        // corrupted cache entry — refetch
      }
    }

    // When set, broker creds are used as the source identity for the customer AssumeRole below.
    const baseCredentials = this.brokerRoleArn
      ? fromTemporaryCredentials({
          params: { RoleArn: this.brokerRoleArn, RoleSessionName: `daytona-${externalId}-broker` },
        })
      : undefined

    const ecr = new ECRClient({
      region,
      credentials: useApiIdentity
        ? baseCredentials
        : fromTemporaryCredentials({
            params: {
              RoleArn: normalizedArn,
              RoleSessionName: `daytona-${externalId}-pull`,
              ExternalId: externalId,
            },
            masterCredentials: baseCredentials,
          }),
    })

    const resp = await ecr.send(new GetAuthorizationTokenCommand({}))
    const data = resp.authorizationData?.[0]
    if (!data?.authorizationToken) {
      throw new Error('ECR returned no authorization data')
    }

    const decoded = Buffer.from(data.authorizationToken, 'base64').toString('utf-8')
    const sep = decoded.indexOf(':')
    if (sep < 0) {
      throw new Error('Unexpected ECR token format')
    }
    const auth = {
      username: decoded.slice(0, sep),
      password: decoded.slice(sep + 1),
    }

    if (data.expiresAt) {
      const ttl = Math.floor((data.expiresAt.getTime() - Date.now()) / 1000) - REFRESH_BUFFER_SEC
      if (ttl > 0) {
        await this.redis.setex(cacheKey, ttl, JSON.stringify(auth))
      }
    }

    this.logger.log(
      `Resolved ECR credentials for region=${region} role=${useApiIdentity ? '(API identity)' : normalizedArn}${this.brokerRoleArn ? ' via broker' : ''}`,
    )
    return auth
  }

  // STS returns `arn:aws:sts::N:assumed-role/X/sess`; convert to the IAM role form `arn:aws:iam::N:role/X` so it can be compared against the user-supplied role ARN.
  private async getApiArn(): Promise<string> {
    if (this.apiArn) return this.apiArn
    try {
      const { Arn } = await new STSClient({}).send(new GetCallerIdentityCommand({}))
      this.apiArn = Arn?.replace(/^arn:aws:sts::(\d+):assumed-role\/([^/]+)\/.+$/, 'arn:aws:iam::$1:role/$2') ?? ''
      this.logger.log(`Resolved API identity: ${this.apiArn || `(empty STS response, raw=${Arn ?? 'undefined'})`}`)
    } catch (err) {
      this.logger.warn(`STS GetCallerIdentity failed: ${(err as Error).message}`)
    }
    return this.apiArn
  }
}

function stripScheme(url: string): string {
  return url.replace(/^https?:\/\//, '')
}
