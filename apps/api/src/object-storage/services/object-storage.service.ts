/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BadRequestException, Injectable, Logger } from '@nestjs/common'
import { TypedConfigService } from '../../config/typed-config.service'
import { StorageAccessDto } from '../../sandbox/dto/storage-access-dto'
import axios from 'axios'
import * as aws4 from 'aws4'
import * as xml2js from 'xml2js'
import { STSClient, AssumeRoleCommand } from '@aws-sdk/client-sts'

interface S3Config {
  endpoint: string
  stsEndpoint: string
  accessKey: string
  secretKey: string
  bucket: string
  region: string
  accountId?: string
  roleName?: string
  organizationId: string
  policy: any
}

@Injectable()
export class ObjectStorageService {
  private readonly logger = new Logger(ObjectStorageService.name)

  constructor(private readonly configService: TypedConfigService) {}

  async getPushAccess(organizationId: string): Promise<StorageAccessDto> {
    try {
      const bucket = this.configService.getOrThrow('s3.defaultBucket')
      const s3Config: S3Config = {
        endpoint: this.configService.getOrThrow('s3.endpoint'),
        stsEndpoint: this.configService.getOrThrow('s3.stsEndpoint'),
        accessKey: this.configService.getOrThrow('s3.accessKey'),
        secretKey: this.configService.getOrThrow('s3.secretKey'),
        bucket,
        region: this.configService.getOrThrow('s3.region'),
        accountId: this.configService.getOrThrow('s3.accountId'),
        roleName: this.configService.getOrThrow('s3.roleName'),
        organizationId,
        policy: {
          Version: '2012-10-17',
          Statement: [
            {
              Effect: 'Allow',
              Action: ['s3:PutObject', 's3:GetObject'],
              Resource: [`arn:aws:s3:::${bucket}/${organizationId}/*`],
            },
            // ListBucket only shows object keys and some metadata, not the actual objects
            {
              Effect: 'Allow',
              Action: ['s3:ListBucket'],
              Resource: [`arn:aws:s3:::${bucket}`],
            },
          ],
        },
      }

      const isMinioServer = s3Config.endpoint.includes('minio')

      if (isMinioServer) {
        return this.getMinioCredentials(s3Config)
      } else {
        return this.getAwsCredentials(s3Config)
      }
    } catch (error) {
      this.logger.error('Storage push access error:', error.response?.data || error.message)
      throw new BadRequestException(`Failed to get temporary credentials: ${error.message}`)
    }
  }

  private async getMinioCredentials(config: S3Config): Promise<StorageAccessDto> {
    const body = new URLSearchParams({
      Action: 'AssumeRole',
      Version: '2011-06-15',
      DurationSeconds: '3600', // 1 hour (in seconds)
      Policy: JSON.stringify(config.policy),
    })

    const requestOptions = {
      host: new URL(config.endpoint).hostname,
      path: '/minio/v1/assume-role',
      service: 'sts',
      method: 'POST',
      body: body.toString(),
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
    }

    aws4.sign(requestOptions, {
      accessKeyId: config.accessKey,
      secretAccessKey: config.secretKey,
    })

    const response = await axios.post(config.stsEndpoint, body.toString(), {
      headers: requestOptions.headers,
    })

    const parser = new xml2js.Parser({ explicitArray: false })
    const parsedData = await parser.parseStringPromise(response.data)

    if (!parsedData.AssumeRoleResponse.AssumeRoleResult.Credentials) {
      throw new BadRequestException('MinIO STS response did not return expected credentials')
    }

    const creds = parsedData.AssumeRoleResponse.AssumeRoleResult.Credentials

    return {
      accessKey: creds.AccessKeyId,
      secret: creds.SecretAccessKey,
      sessionToken: creds.SessionToken,
      storageUrl: config.endpoint,
      organizationId: config.organizationId,
      bucket: config.bucket,
    }
  }

  private async getAwsCredentials(config: S3Config): Promise<StorageAccessDto> {
    try {
      const stsClient = new STSClient({
        region: config.region,
        endpoint: config.stsEndpoint,
        credentials: {
          accessKeyId: config.accessKey,
          secretAccessKey: config.secretKey,
        },
        maxAttempts: 3,
      })

      const command = new AssumeRoleCommand({
        RoleArn: `arn:aws:iam::${config.accountId}:role/${config.roleName}`,
        RoleSessionName: `daytona-${config.organizationId}-${Date.now()}`,
        DurationSeconds: 3600, // One hour
        Policy: JSON.stringify(config.policy),
      })

      try {
        const response = await stsClient.send(command)

        if (!response.Credentials) {
          throw new BadRequestException('AWS STS response did not return expected credentials')
        }

        return {
          accessKey: response.Credentials.AccessKeyId,
          secret: response.Credentials.SecretAccessKey,
          sessionToken: response.Credentials.SessionToken,
          storageUrl: config.endpoint,
          organizationId: config.organizationId,
          bucket: config.bucket,
        }
      } catch (error: any) {
        throw new BadRequestException(`Failed to assume role: ${error.message || 'Unknown AWS error'}`)
      }
    } catch (error: any) {
      this.logger.error(`AWS STS client setup error: ${error.message}`, error.stack)
      throw new BadRequestException(`Failed to setup AWS client: ${error.message}`)
    }
  }
}
