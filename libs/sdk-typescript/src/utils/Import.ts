/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { DaytonaError } from '../errors/DaytonaError'
import { RUNTIME } from './Runtime'

const loaderMap = {
  'fast-glob': () => import('fast-glob'),
  '@iarna/toml': () => import('@iarna/toml'),
  stream: () => import('stream'),
  tar: () => import('tar'),
  'expand-tilde': () => import('expand-tilde'),
  ObjectStorage: () => import('../ObjectStorage.js'),
  fs: (): Promise<typeof import('fs')> => import('fs'),
  'form-data': () => import('form-data'),
}

const requireMap = {
  'fast-glob': () => require('fast-glob'),
  '@iarna/toml': () => require('@iarna/toml'),
  stream: () => require('stream'),
  tar: () => require('tar'),
  'expand-tilde': () => require('expand-tilde'),
  fs: () => require('fs'),
  'form-data': () => require('form-data'),
}

const validateMap: Record<string, (mod: any) => boolean> = {
  fs: (mod: any) => typeof mod.createReadStream === 'function',
  'form-data': (mod: any) => typeof mod.default === 'function',
}

type ModuleMap = typeof loaderMap

export async function dynamicImport<K extends keyof ModuleMap>(
  name: K,
  errorPrefix?: string,
): Promise<Awaited<ReturnType<ModuleMap[K]>>> {
  const loader = loaderMap[name]
  if (!loader) {
    throw new DaytonaError(`${errorPrefix || ''} Unknown module "${name}"`)
  }

  let mod: any
  try {
    mod = (await loader()) as any
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err)
    throw new DaytonaError(`${errorPrefix || ''} Module "${name}" is not available in the "${RUNTIME}" runtime: ${msg}`)
  }

  if (validateMap[name] && !validateMap[name](mod)) {
    throw new DaytonaError(
      `${errorPrefix || ''} Module "${name}" didn't pass import validation in the "${RUNTIME}" runtime`,
    )
  }

  return mod?.default ?? mod
}

type RequireMap = typeof requireMap

export function dynamicRequire<K extends keyof RequireMap>(name: K, errorPrefix?: string): ReturnType<RequireMap[K]> {
  const loader = requireMap[name]
  if (!loader) {
    throw new DaytonaError(`${errorPrefix || ''} Unknown module "${name}"`)
  }

  let mod: any
  try {
    mod = loader()
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err)
    throw new DaytonaError(`${errorPrefix || ''} Module "${name}" is not available in the "${RUNTIME}" runtime: ${msg}`)
  }

  if (validateMap[name] && !validateMap[name](mod)) {
    throw new DaytonaError(
      `${errorPrefix || ''} Module "${name}" didn't pass import validation in the "${RUNTIME}" runtime`,
    )
  }

  return mod?.default ?? mod
}
