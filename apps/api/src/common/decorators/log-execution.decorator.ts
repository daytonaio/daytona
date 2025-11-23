/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Logger } from '@nestjs/common'

// Parse threshold once at module load time
let LOG_THRESHOLD = parseInt(process.env.LOG_EXECUTION_THRESHOLD_MILLISECONDS, 10)
if (isNaN(LOG_THRESHOLD) || LOG_THRESHOLD <= 0) {
  LOG_THRESHOLD = 1000 // Default to 1000ms if not set or invalid
}

export function LogExecution(name?: string) {
  return function (target: any, propertyKey: string, descriptor: PropertyDescriptor) {
    const shouldLogExecutions = process.env.LOG_EXECUTIONS === 'true'
    if (!shouldLogExecutions) {
      return descriptor
    }

    // Wrap the original method with logging
    const originalMethod = descriptor.value
    const logger = new Logger(`Function:${target.constructor.name}`)

    descriptor.value = async function (...args: any[]) {
      const startTime = Date.now()
      const functionName = name || propertyKey

      try {
        const result = await originalMethod.apply(this, args)
        const duration = Date.now() - startTime

        if (duration > LOG_THRESHOLD) {
          logger.warn(`Function ${functionName} took a long time: ${duration}ms`)
        }

        return result
      } catch (error) {
        const duration = Date.now() - startTime
        logger.error(`Failed function: ${functionName} (took ${duration}ms)`, error.stack)
        throw error
      }
    }

    return descriptor
  }
}
