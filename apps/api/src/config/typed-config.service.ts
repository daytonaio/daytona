/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ConfigService } from '@nestjs/config'
import { Injectable } from '@nestjs/common'
import { configuration } from './configuration'
import { KafkaConfig, Mechanism, SASLOptions } from 'kafkajs'
import { AwsSigv4Signer, AwsSigv4SignerResponse } from '@opensearch-project/opensearch/aws'
import { defaultProvider } from '@aws-sdk/credential-provider-node'
import { fromTemporaryCredentials } from '@aws-sdk/credential-providers'
import { ClientOptions } from '@opensearch-project/opensearch'
import { RedisOptions } from 'ioredis'

type Configuration = typeof configuration

// Helper type to get nested property paths
type Paths<T> = T extends object
  ? {
      [K in keyof T]: K extends string ? (T[K] extends object ? `${K}` | `${K}.${Paths<T[K]>}` : `${K}`) : never
    }[keyof T]
  : never

// Helper type to get the type of a property at a given path
type PathValue<T, P extends string> = P extends `${infer K}.${infer Rest}`
  ? K extends keyof T
    ? T[K] extends object
      ? PathValue<T[K], Rest>
      : never
    : never
  : P extends keyof T
    ? T[P]
    : never

@Injectable()
export class TypedConfigService {
  constructor(private configService: ConfigService) {}

  /**
   * Get a configuration value with type safety
   * @param key The configuration key (can be nested using dot notation)
   * @returns The configuration value with proper typing
   */
  get<K extends Paths<Configuration>>(key: K): PathValue<Configuration, K> {
    return this.configService.get(key)
  }

  /**
   * Get a configuration value with type safety, throwing an error if undefined
   * @param key The configuration key (can be nested using dot notation)
   * @returns The configuration value with proper typing
   * @throws Error if the configuration value is undefined
   */
  getOrThrow<K extends Paths<Configuration>>(key: K): NonNullable<PathValue<Configuration, K>> {
    const value = this.get(key)
    if (value === undefined) {
      throw new Error(`Configuration key "${key}" is undefined`)
    }
    return value as NonNullable<PathValue<Configuration, K>>
  }

  /**
   * Get the Kafka configuration
   * @returns The Kafka configuration
   */
  getKafkaClientConfig(): KafkaConfig {
    const mechanism = this.get('kafka.sasl.mechanism') || 'plain'
    const username = this.get('kafka.sasl.username')
    const password = this.get('kafka.sasl.password')

    if (mechanism !== 'plain' && mechanism !== 'scram-sha-256' && mechanism !== 'scram-sha-512') {
      throw new Error(`Invalid Kafka SASL mechanism: ${mechanism}`)
    }
    const sasl: SASLOptions | Mechanism | undefined =
      username && password
        ? ({
            mechanism,
            username,
            password,
          } as SASLOptions)
        : undefined

    return {
      brokers: this.get('kafka.brokers')
        .split(',')
        .map((broker) => broker.trim()),
      ssl: this.get('kafka.tls.enabled')
        ? {
            rejectUnauthorized: this.get('kafka.tls.rejectUnauthorized'),
          }
        : undefined,
      sasl,
    }
  }

  /**
   * Get the OpenSearch configuration
   * @returns The OpenSearch configuration
   */
  getOpenSearchConfig(): ClientOptions {
    const nodes = this.get('opensearch.nodes')
      .split(',')
      .map((node) => node.trim())
    const username = this.get('opensearch.username')
    const password = this.get('opensearch.password')

    // Basic auth
    if (username && password) {
      return {
        nodes,
        auth: {
          username,
          password,
        },
        ssl: {
          rejectUnauthorized: this.get('opensearch.tls.rejectUnauthorized'),
        },
      }
    }

    // AWS Sigv4 auth
    try {
      let signer: AwsSigv4SignerResponse
      if (this.get('opensearch.aws.roleArn')) {
        signer = AwsSigv4Signer({
          getCredentials: fromTemporaryCredentials({
            params: {
              RoleArn: this.get('opensearch.aws.roleArn'),
              RoleSessionName: 'daytona-opensearch',
            },
          }),
          service: 'es',
          region: this.get('opensearch.aws.region'),
        })
      } else {
        signer = AwsSigv4Signer({
          getCredentials() {
            const credentialsProvider = defaultProvider()
            return credentialsProvider()
          },
          service: 'es',
          region: this.get('opensearch.aws.region'),
        })
      }
      return {
        nodes,
        ssl: {
          rejectUnauthorized: this.get('opensearch.tls.rejectUnauthorized'),
        },
        ...signer,
      }
      // Try without auth if AWS credentials are not available
    } catch {
      return {
        nodes,
        ssl: {
          rejectUnauthorized: this.get('opensearch.tls.rejectUnauthorized'),
        },
      }
    }
  }

  /**
   * Get the Redis configuration
   * @param overrides Optional overrides for the Redis configuration
   * @returns The Redis configuration
   */
  getRedisConfig(overrides?: Partial<RedisOptions>): RedisOptions {
    return {
      host: this.getOrThrow('redis.host'),
      port: this.getOrThrow('redis.port'),
      username: this.get('redis.username'),
      password: this.get('redis.password'),
      tls: this.get('redis.tls'),
      lazyConnect: this.get('skipConnections'),
      ...overrides,
    }
  }

  /**
   * Get the ClickHouse configuration
   * @returns The ClickHouse configuration
   */
  getClickHouseConfig() {
    const host = this.get('clickhouse.host')
    if (!host) {
      return null
    }
    return {
      url: `${this.get('clickhouse.protocol')}://${host}:${this.get('clickhouse.port')}`,
      username: this.get('clickhouse.username'),
      password: this.get('clickhouse.password'),
      database: this.get('clickhouse.database'),
    }
  }
}
