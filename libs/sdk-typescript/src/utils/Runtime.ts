/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

export enum Runtime {
  NODE = 'node',
  DENO = 'deno',
  BUN = 'bun',
  BROWSER = 'browser',
  SERVERLESS = 'serverless',
  UNKNOWN = 'unknown',
}

const denoGlobal = (
  globalThis as {
    Deno?: {
      version?: { deno?: string }
      env?: { get(name: string): string | undefined; toObject(): Record<string, string> }
    }
  }
).Deno
const bunGlobal = (globalThis as { Bun?: { version?: { bun?: string }; file?: (path: string) => File } }).Bun

export const RUNTIME =
  typeof denoGlobal !== 'undefined' && !!denoGlobal?.version?.deno
    ? Runtime.DENO
    : typeof bunGlobal !== 'undefined' && !!bunGlobal?.version?.bun
      ? Runtime.BUN
      : isServerlessRuntime()
        ? Runtime.SERVERLESS
        : typeof window !== 'undefined'
          ? Runtime.BROWSER
          : typeof process !== 'undefined' && !!process.versions?.node
            ? Runtime.NODE
            : Runtime.UNKNOWN

export function getEnvVar(name: string): string | undefined {
  if (RUNTIME === Runtime.NODE || RUNTIME === Runtime.BUN) {
    return process.env[name]
  }
  if (RUNTIME === Runtime.DENO) {
    return denoGlobal?.env?.get(name)
  }

  return undefined
}

export function isServerlessRuntime(): boolean {
  // Safely grab env vars, even if `process` is undeclared
  const env = typeof process !== 'undefined' ? process.env : {}

  // Worker-specific globals
  const globalObj = globalThis as any

  return Boolean(
    // Cloudflare Workers (V8 isolate API)
    typeof globalObj.WebSocketPair === 'function' ||
      // Cloudflare Pages
      env.CF_PAGES === '1' ||
      // AWS Lambda (incl. SAM local)
      env.AWS_EXECUTION_ENV?.startsWith('AWS_Lambda') ||
      env.LAMBDA_TASK_ROOT !== undefined ||
      env.AWS_SAM_LOCAL === 'true' ||
      // Azure Functions
      env.FUNCTIONS_WORKER_RUNTIME !== undefined ||
      // Google Cloud Functions / Cloud Run
      (env.FUNCTION_TARGET !== undefined && env.FUNCTION_SIGNATURE_TYPE !== undefined) ||
      // Vercel
      env.VERCEL === '1' ||
      // Netlify Functions
      env.SITE_NAME !== undefined,
  )
}
