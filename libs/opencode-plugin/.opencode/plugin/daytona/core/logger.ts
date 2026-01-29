/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Logger class for handling plugin logging
 */

import { appendFileSync, mkdirSync, statSync, truncateSync } from 'fs'
import { dirname } from 'path'
import type { LogLevel } from './types'
import { LOG_LEVEL_INFO, LOG_LEVEL_ERROR, LOG_LEVEL_WARN } from './types'

let logFilePath: string | undefined

export function setLogFilePath(path: string) {
  logFilePath = path
}

class Logger {
  private get logFile() {
    if (!logFilePath) throw new Error('Logger file path not set. Call setLogFilePath(path) before use.')
    return logFilePath
  }

  log(message: string, level: LogLevel = LOG_LEVEL_INFO): void {
    // Ensure log directory exists
    try {
      mkdirSync(dirname(this.logFile), { recursive: true })
    } catch (err) {
      // Directory may already exist, ignore
    }
    // Truncate log file if it exceeds 5MB
    try {
      const stats = statSync(this.logFile)
      const maxSize = 5 * 1024 * 1024 // 5MB
      if (stats.size > maxSize) {
        truncateSync(this.logFile, 0)
      }
    } catch (err) {
      // File may not exist yet, ignore
    }
    const timestamp = new Date().toISOString()
    const logEntry = `[${timestamp}] [${level}] ${message}\n`
    appendFileSync(this.logFile, logEntry)
  }

  info(message: string): void {
    this.log(message, LOG_LEVEL_INFO)
  }

  error(message: string): void {
    this.log(message, LOG_LEVEL_ERROR)
  }

  warn(message: string): void {
    this.log(message, LOG_LEVEL_WARN)
  }
}

export const logger = new Logger()
