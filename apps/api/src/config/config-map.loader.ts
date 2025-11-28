/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { existsSync, readFileSync } from 'node:fs'
import { parse as parseYaml } from 'yaml'
import { Logger } from '@nestjs/common'

const logger = new Logger('ConfigMapLoader')

export interface ConfigMap {
  [key: string]: any
}

/**
 * Loads configuration from a JSON or YAML file.
 * Supports nested configuration objects.
 *
 * @param filePath - Path to the config file (JSON or YAML)
 * @returns Parsed configuration object or empty object if file doesn't exist
 */
export function loadConfigMap(filePath?: string): ConfigMap {
  if (!filePath) {
    logger.debug('No config map file path provided')
    return {}
  }

  if (!existsSync(filePath)) {
    throw new Error(`Config map file not found: ${filePath}`)
  }

  try {
    const fileContent = readFileSync(filePath, 'utf8')
    const isYaml = filePath.endsWith('.yaml') || filePath.endsWith('.yml')

    const config = isYaml ? parseYaml(fileContent) : JSON.parse(fileContent)

    logger.log(`Successfully loaded config map from: ${filePath}`)
    return config || {}
  } catch (error) {
    logger.error(`Failed to parse config map file: ${filePath}`, error.stack)
    throw new Error(`Failed to parse config map file: ${filePath}. ${error.message}`)
  }
}

/**
 * Gets a nested value from an object using dot notation.
 * Example: getNestedValue({ database: { host: 'localhost' } }, 'database.host') => 'localhost'
 *
 * @param obj - Source object
 * @param path - Dot-separated path (e.g., 'database.host')
 * @returns Value at the path or undefined
 */
export function getNestedValue(obj: any, path: string): any {
  if (!obj || typeof obj !== 'object') {
    return undefined
  }

  const keys = path.split('.')
  let current = obj

  for (const key of keys) {
    if (current === null || current === undefined || typeof current !== 'object') {
      return undefined
    }
    current = current[key]
  }

  return current
}

/**
 * Gets a value from environment variable or falls back to config map.
 * Supports nested config map paths using dot notation.
 *
 * @param envVar - Environment variable name
 * @param configMapPath - Dot-separated path in config map (e.g., 'database.host')
 * @param configMap - Loaded config map object
 * @returns Value from env or config map, or undefined
 */
export function getConfigValue(envVar: string, configMapPath: string, configMap: ConfigMap): string | undefined {
  // Environment variables take precedence
  const envValue = process.env[envVar]
  if (envValue !== undefined && envValue !== '') {
    return envValue
  }

  // Fall back to config map
  const configValue = getNestedValue(configMap, configMapPath)
  return configValue === undefined || configValue === null ? undefined : String(configValue)
}
