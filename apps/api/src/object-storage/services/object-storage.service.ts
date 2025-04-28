/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BadRequestException, Injectable, Logger } from '@nestjs/common'
import { TypedConfigService } from '../../config/typed-config.service'
import { StorageAccessDto } from '../../workspace/dto/storage-access-dto'
import axios from 'axios'
import * as aws4 from 'aws4'
import * as xml2js from 'xml2js'

@Injectable()
export class ObjectStorageService {
  private readonly logger = new Logger(ObjectStorageService.name)

  constructor(private readonly configService: TypedConfigService) {}

  async getPushAccess(organizationId: string): Promise<StorageAccessDto> {
    try {
      const s3Endpoint = this.configService.getOrThrow('s3.endpoint')
      const s3AccessKey = this.configService.getOrThrow('s3.accessKey')
      const s3SecretKey = this.configService.getOrThrow('s3.secretKey')
      const s3DefaultBucket = this.configService.getOrThrow('s3.defaultBucket') || 'daytona'

      const policy = {
        Version: '2012-10-17',
        Statement: [
          {
            Effect: 'Allow',
            Action: ['s3:PutObject', 's3:GetObject'],
            Resource: [`arn:aws:s3:::${s3DefaultBucket}/${organizationId}/*`],
          },
          // ListBucket only shows object keys and some metadata, not the actual objects
          {
            Effect: 'Allow',
            Action: ['s3:ListBucket'],
            Resource: [`arn:aws:s3:::${s3DefaultBucket}`],
          },
        ],
      }

      const stsEndpoint = new URL('/minio/v1/assume-role', s3Endpoint).toString()

      const body = new URLSearchParams({
        Action: 'AssumeRole',
        Version: '2011-06-15',
        DurationSeconds: '86400', // 24 hours (in seconds)
        Policy: JSON.stringify(policy),
      })

      const requestOptions = {
        host: new URL(s3Endpoint).hostname,
        path: '/minio/v1/assume-role',
        service: 'sts',
        method: 'POST',
        body: body.toString(),
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
      }

      // Sign the request with AWS SigV4
      aws4.sign(requestOptions, {
        accessKeyId: s3AccessKey,
        secretAccessKey: s3SecretKey,
      })

      // Send the request to MinIO
      const response = await axios.post(stsEndpoint, body.toString(), {
        headers: requestOptions.headers,
      })

      const parser = new xml2js.Parser({ explicitArray: false })
      const parsedData = await parser.parseStringPromise(response.data)

      // Check if Credentials exist in the parsed response
      if (!parsedData.AssumeRoleResponse.AssumeRoleResult.Credentials) {
        throw new BadRequestException('MinIO STS response did not return expected credentials')
      }

      const creds = parsedData.AssumeRoleResponse.AssumeRoleResult.Credentials

      return {
        accessKey: creds.AccessKeyId,
        secret: creds.SecretAccessKey,
        sessionToken: creds.SessionToken,
        storageUrl: s3Endpoint,
        registryId: organizationId,
        organizationId,
      }
    } catch (error) {
      this.logger.error('Storage push access error:', error.response?.data || error.message)
      throw new BadRequestException(`Failed to get temporary credentials: ${error.message}`)
    }
  }
}
