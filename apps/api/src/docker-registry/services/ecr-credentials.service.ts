/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { ECRClient, GetAuthorizationTokenCommand } from '@aws-sdk/client-ecr'
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
  private readonly apiIdentityRoleArn: string

  constructor(
    @InjectRedis() private readonly redis: Redis,
    configService: TypedConfigService,
  ) {
    this.apiIdentityRoleArn = configService.get('ecr.apiIdentityRoleArn')?.trim() ?? ''
  }

  isEcrUrl(url: string): boolean {
    return ECR_HOST_REGEX.test(stripScheme(url))
  }

  /**
   * Resolves a fresh `AWS:<token>` Docker auth pair for an ECR registry.
   * AssumeRoles the supplied ARN, except when it matches the operator-declared
   * `ECR_API_IDENTITY_ROLE_ARN` — then the API's own AWS identity is used.
   * Cached in Redis with TTL derived from AWS's `expiresAt`.
   */
  async resolveEcrCredentials(url: string, roleArn: string, externalId: string): Promise<EcrAuth> {
    const match = ECR_HOST_REGEX.exec(stripScheme(url))
    if (!match) {
      throw new Error(`Not an ECR URL: ${url}`)
    }
    const region = match[1]
    const normalizedArn = roleArn.trim()
    const useApiIdentity = this.apiIdentityRoleArn !== '' && normalizedArn === this.apiIdentityRoleArn

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

    const ecr = new ECRClient({
      region,
      credentials: useApiIdentity
        ? undefined
        : fromTemporaryCredentials({
            params: {
              RoleArn: normalizedArn,
              RoleSessionName: `daytona-${externalId}-pull`,
              ExternalId: externalId,
            },
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
      `Resolved ECR credentials for region=${region} role=${useApiIdentity ? '(API identity)' : normalizedArn}`,
    )
    return auth
  }
}

function stripScheme(url: string): string {
  return url.replace(/^https?:\/\//, '')
}
