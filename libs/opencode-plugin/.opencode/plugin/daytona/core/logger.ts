/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Logger class for handling plugin logging
 */

import { appendFileSync, mkdirSync, readFileSync, statSync, writeFileSync } from 'fs'
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
    // Trim log file if it exceeds 3MB (keep last 1MB)
    try {
      const stats = statSync(this.logFile)
      const maxSize = 3 * 1024 * 1024
      const keepSize = 1024 * 1024
      if (stats.size > maxSize) {
        const content = readFileSync(this.logFile, 'utf8')
        const trimmed = content.slice(-keepSize)
        // Drop partial first line so we don't start mid-log
        const firstNewline = trimmed.indexOf('\n')
        writeFileSync(this.logFile, firstNewline >= 0 ? trimmed.slice(firstNewline + 1) : trimmed)
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
