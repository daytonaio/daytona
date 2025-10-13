/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { createCipheriv, createDecipheriv, randomBytes, scrypt } from 'crypto'
import { promisify } from 'node:util'
import { TypedConfigService } from '../config/typed-config.service'

export interface EncryptedData {
  data: string
  iv: string
}

@Injectable()
export class EncryptionService {
  private readonly algorithm = 'aes-256-ctr'
  private readonly encoding = 'base64'
  private readonly secret: string
  private readonly salt: string
  private readonly logger = new Logger(EncryptionService.name)

  constructor(configService: TypedConfigService) {
    this.logger.debug('Initializing encryption service')
    this.secret = configService.getOrThrow('encryption.key')
    this.salt = configService.getOrThrow('encryption.salt')
  }

  public async encrypt(input: string): Promise<string> {
    const key = (await promisify(scrypt)(this.secret, this.salt, 32)) as Buffer
    const iv = randomBytes(16)
    const cipher = createCipheriv(this.algorithm, key, iv)

    return this.serialize({
      data: Buffer.concat([cipher.update(input), cipher.final()]).toString(this.encoding),
      iv: iv.toString(this.encoding),
    })
  }

  /**
   * Decrypts the input string. If backwardsCompatible is true, it will return the input string
   * as is if decryption fails (for handling unencrypted data).
   */
  public async decrypt(input: string, backwardsCompatible = false): Promise<string> {
    if (backwardsCompatible) {
      try {
        return await this._decrypt(input)
      } catch {
        return input
      }
    }

    return this._decrypt(input)
  }

  private async _decrypt(input: string): Promise<string> {
    const encryptedData = this.deserialize(input)
    const key = (await promisify(scrypt)(this.secret, this.salt, 32)) as Buffer
    const encrypted = Buffer.from(encryptedData.data, this.encoding)
    const iv = Buffer.from(encryptedData.iv, this.encoding)
    const decipher = createDecipheriv(this.algorithm, key, iv)

    const decrypted = Buffer.concat([decipher.update(encrypted), decipher.final()])
    return decrypted.toString()
  }

  private serialize(data: EncryptedData): string {
    return JSON.stringify(data)
  }

  private deserialize(data: string): EncryptedData {
    return JSON.parse(data)
  }
}
