/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ConfigService } from '@nestjs/config'
import { Injectable } from '@nestjs/common'
import { configuration } from './configuration'

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
}
