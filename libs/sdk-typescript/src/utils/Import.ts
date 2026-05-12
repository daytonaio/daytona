/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 *
 * Note: the ESM build of this file is prepended with a `createRequire` shim by
 * `scripts/post-build.js` so that synchronous `require()` works in ESM Node.js
 * (and any runtime that supports `createRequire(import.meta.url)`). The CJS
 * build uses the native `require`. See that script for the exact shim source.
 */

import { DaytonaError } from '../errors/DaytonaError'
import { RUNTIME } from './Runtime'

const loaderMap = {
  'fast-glob': async () => require('fast-glob'),
  '@iarna/toml': async () => require('@iarna/toml'),
  stream: async () => require('stream'),
  tar: async () => require('tar'),
  'expand-tilde': async () => require('expand-tilde'),
  // Stored in a variable + /* @vite-ignore */ so Vite/Rollup don't statically
  // chunk this; otherwise apps bundling the SDK source (e.g. the dashboard)
  // would pull @aws-sdk transitively. Node consumers resolve at runtime.
  ObjectStorage: () => {
    const path = '../ObjectStorage.js'
    return import(/* @vite-ignore */ path)
  },
  fs: async () => require('fs'),
  'form-data': async () => require('form-data'),
  util: async () => require('util'),
}

const requireMap = {
  'fast-glob': () => require('fast-glob'),
  '@iarna/toml': () => require('@iarna/toml'),
  stream: () => require('stream'),
  tar: () => require('tar'),
  'expand-tilde': () => require('expand-tilde'),
  fs: () => require('fs'),
  'form-data': () => require('form-data'),
  buffer: () => require('buffer'),
  busboy: () => require('busboy'),
  '@opentelemetry/api': () => require('@opentelemetry/api'),
  '@opentelemetry/sdk-node': () => require('@opentelemetry/sdk-node'),
  '@opentelemetry/instrumentation-http': () => require('@opentelemetry/instrumentation-http'),
  '@opentelemetry/sdk-trace-base': () => require('@opentelemetry/sdk-trace-base'),
  '@opentelemetry/exporter-trace-otlp-http': () => require('@opentelemetry/exporter-trace-otlp-http'),
  '@opentelemetry/otlp-exporter-base': () => require('@opentelemetry/otlp-exporter-base'),
  '@opentelemetry/semantic-conventions': () => require('@opentelemetry/semantic-conventions'),
  '@opentelemetry/resources': () => require('@opentelemetry/resources'),
}

const validateMap: Record<string, (mod: any) => boolean> = {
  'fast-glob': (mod: any) => typeof mod === 'function' && typeof mod?.sync === 'function',
  '@iarna/toml': (mod: any) => typeof mod.parse === 'function' && typeof mod.stringify === 'function',
  stream: (mod: any) => typeof mod.Readable === 'function' && typeof mod.Writable === 'function',
  tar: (mod: any) => typeof mod.extract === 'function' && typeof mod.create === 'function',
  'expand-tilde': (mod: any) => typeof mod === 'function',
  fs: (mod: any) => typeof mod.createReadStream === 'function' && typeof mod.readFile === 'function',
  'form-data': (mod: any) => typeof mod === 'function',
  util: (mod: any) => typeof mod.promisify === 'function',
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
    mod = unwrapInterop(mod)
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err)
    throw new DaytonaError(`${errorPrefix || ''} Module "${name}" is not available in the "${RUNTIME}" runtime: ${msg}`)
  }

  if (validateMap[name] && !validateMap[name](mod)) {
    throw new DaytonaError(
      `${errorPrefix || ''} Module "${name}" didn't pass import validation in the "${RUNTIME}" runtime`,
    )
  }

  return mod
}

type RequireMap = typeof requireMap

function unwrapInterop(mod: any): any {
  if (!mod || typeof mod !== 'object' || mod.default === undefined) return mod
  const namedKeys = Object.keys(mod).filter((k) => k !== 'default')
  if (namedKeys.length === 0) return mod.default
  return mod
}

export function dynamicRequire<K extends keyof RequireMap>(name: K, errorPrefix?: string): ReturnType<RequireMap[K]> {
  const loader = requireMap[name]
  if (!loader) {
    throw new DaytonaError(`${errorPrefix || ''} Unknown module "${name}"`)
  }

  let mod: any
  try {
    mod = loader()
    mod = unwrapInterop(mod)
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err)
    throw new DaytonaError(`${errorPrefix || ''} Module "${name}" is not available in the "${RUNTIME}" runtime: ${msg}`)
  }

  if (validateMap[name] && !validateMap[name](mod)) {
    throw new DaytonaError(
      `${errorPrefix || ''} Module "${name}" didn't pass import validation in the "${RUNTIME}" runtime`,
    )
  }

  return mod
}

let _packageInfo: { name: string; version: string } | null = null

export function getPackageInfo(): { name: string; version: string } {
  if (_packageInfo) return _packageInfo
  try {
    const pkg = require('../../package.json')
    _packageInfo = { name: pkg.name, version: pkg.version }
  } catch {
    _packageInfo = { name: '@daytona/sdk', version: '0.0.0' }
  }
  return _packageInfo
}
