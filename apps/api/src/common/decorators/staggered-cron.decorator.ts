/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Cron, CronOptions } from '@nestjs/schedule'

function simpleHash(str: string): number {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i)
    hash = (hash << 5) - hash + char
    hash |= 0
  }
  return Math.abs(hash)
}

// Fixed-time expressions (e.g. EVERY_DAY_AT_2AM) are not staggered
function hasStepPattern(cronExpression: string): boolean {
  return cronExpression.split(/\s+/).some((field) => field.includes('/'))
}

/**
 * Applies a deterministic offset to a cron expression based on a job name,
 * spreading jobs that share the same interval across different seconds
 * to avoid thundering-herd spikes.
 *
 * Fixed-time expressions (e.g. EVERY_DAY_AT_2AM) are left unchanged —
 * only recurring interval patterns with step syntax are staggered.
 */
function staggerCronExpression(cronExpression: string, jobName: string): string {
  if (!hasStepPattern(cronExpression)) {
    return cronExpression
  }

  const hash = simpleHash(jobName)
  const fields = cronExpression.trim().split(/\s+/)

  if (fields.length === 6) {
    const [seconds, ...rest] = fields

    const secStepMatch = seconds.match(/^\*\/(\d+)$/)
    if (secStepMatch) {
      const interval = parseInt(secStepMatch[1])
      if (interval > 1) {
        const offset = hash % interval
        return [`${offset}/${interval}`, ...rest].join(' ')
      }
      return cronExpression
    }

    if (seconds === '0') {
      const offset = hash % 60
      return [String(offset), ...rest].join(' ')
    }

    return cronExpression
  }

  if (fields.length === 5) {
    const offset = hash % 60
    return [String(offset), ...fields].join(' ')
  }

  return cronExpression
}

/**
 * Drop-in replacement for @Cron that staggers execution times.
 *
 * Uses a hash of the job name (or method name as fallback) to compute
 * a deterministic second-offset, so jobs with the same interval don't
 * all fire at the same wall-clock instant.
 *
 * The offset is stable across restarts — the same job name always
 * produces the same offset.
 */
export function StaggeredCron(cronTime: string | Date, options?: CronOptions): MethodDecorator {
  return (target: object, propertyKey: string | symbol, descriptor: PropertyDescriptor) => {
    const name = options?.name || String(propertyKey)

    let staggeredCronTime: string | Date = cronTime
    if (typeof staggeredCronTime === 'string') {
      staggeredCronTime = staggerCronExpression(staggeredCronTime, name)
    }

    const cronDecorator = Cron(staggeredCronTime, options)
    return cronDecorator(target, propertyKey, descriptor)
  }
}
