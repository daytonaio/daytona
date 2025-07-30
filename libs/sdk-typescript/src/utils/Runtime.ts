/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

declare global {
  /**
   * In Deno this global exists and has a `version.deno` string;
   * in all other runtimes it will be `undefined`.
   */
  var Deno:
    | {
        version: { deno: string }
        env: {
          get(name: string): string | undefined
          toObject(): Record<string, string>
        }
      }
    | undefined

  /**
   * In Bun this global exists and has a `version.bun` string;
   * in all other runtimes it will be `undefined`.
   */
  var Bun:
    | {
        version: { bun: string }
        file: (path: string) => File
      }
    | undefined
}

export enum Runtime {
  NODE = 'node',
  DENO = 'deno',
  BUN = 'bun',
  BROWSER = 'browser',
  UNKNOWN = 'unknown',
}

export const RUNTIME =
  typeof Deno !== 'undefined'
    ? Runtime.DENO
    : typeof Bun !== 'undefined' && !!Bun.version
      ? Runtime.BUN
      : typeof process !== 'undefined' && !!process.versions?.node
        ? Runtime.NODE
        : typeof window !== 'undefined'
          ? Runtime.BROWSER
          : Runtime.UNKNOWN

export function getEnvVar(name: string): string | undefined {
  if (RUNTIME === Runtime.NODE || RUNTIME === Runtime.BUN) {
    return process.env[name]
  }
  if (RUNTIME === Runtime.DENO) {
    return Deno.env.get(name)
  }

  return undefined
}
