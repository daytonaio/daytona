/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

type AppMode = 'api' | 'worker' | 'all'

let appMode = process.env.APP_MODE as AppMode

// Default to all mode if no app mode is set
if (!appMode) {
  appMode = 'all'
}

// Validate app mode
if (!Object.values(['api', 'worker', 'all']).includes(appMode)) {
  throw new Error(`Invalid app mode: ${appMode}`)
}

/**
 * Returns true if the API should be started
 */
export function isApiEnabled(): boolean {
  return appMode === 'api' || appMode === 'all'
}

/**
 * Returns true if the worker should be started
 */
export function isWorkerEnabled(): boolean {
  return appMode === 'worker' || appMode === 'all'
}

/**
 * Returns the app mode
 */
export function getAppMode(): AppMode {
  return appMode
}
