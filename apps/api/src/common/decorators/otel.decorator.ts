/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { trace, context } from '@opentelemetry/api'

const tracer = trace.getTracer('daytona-api')

export function OtelSpan(name?: string) {
  return (target: object, propertyKey: string | symbol, descriptor: PropertyDescriptor) => {
    const originalMethod = descriptor.value
    descriptor.value = async function (...args: any[]) {
      const spanName = name || `${target.constructor.name}.${originalMethod.name}`
      const span = tracer.startSpan(spanName, {
        attributes: {
          component: target.constructor.name,
          method: originalMethod.name,
        },
      })
      return context.with(trace.setSpan(context.active(), span), async () => {
        try {
          return await originalMethod.apply(this, args)
        } finally {
          span.end()
        }
      })
    }
  }
}
