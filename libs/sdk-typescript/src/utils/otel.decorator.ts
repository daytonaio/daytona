/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { trace, context, metrics, SpanStatusCode, Histogram } from '@opentelemetry/api'

// Lazy initialization to ensure SDK is started before getting tracer/meter
const getTracer = () => trace.getTracer('')
const getMeter = () => metrics.getMeter('')

const executionHistograms = new Map<string, Histogram>()

/**
 * Configuration options for span instrumentation
 */
export interface SpanConfig {
  /**
   * Custom name for the span. If not provided, uses `ClassName.methodName` format
   */
  name?: string
  /**
   * Additional attributes to attach to the span
   */
  attributes?: Record<string, string>
}

/**
 * Configuration options for metric instrumentation
 */
export interface MetricConfig {
  /**
   * Custom name for the metric. If not provided, uses `ClassName.methodName` format
   */
  name?: string
  /**
   * Description for the metrics being collected
   */
  description?: string
  /**
   * Additional labels to attach to the metrics
   */
  labels?: Record<string, string>
}

/**
 * Configuration options for the combined instrumentation decorator
 */
export interface InstrumentationConfig {
  /**
   * Custom name for the span and metric. If not provided, uses `ClassName.methodName` format
   */
  name?: string
  /**
   * Description for the metrics being collected
   */
  description?: string
  /**
   * Additional labels/attributes to attach to spans and metrics
   */
  labels?: Record<string, string>
  /**
   * Enable trace collection (default: true)
   */
  enableTraces?: boolean
  /**
   * Enable metrics collection (default: true)
   */
  enableMetrics?: boolean
}

/**
 * Converts a string to snake_case for Prometheus-friendly metric names
 */
function toSnakeCase(str: string): string {
  return str
    .replace(/([A-Z])/g, '_$1')
    .toLowerCase()
    .replace(/^_/, '')
    .replace(/\./g, '_')
}

/**
 * Decorator for instrumenting methods with OpenTelemetry spans (traces only)
 *
 * @param config - Configuration object or string name for the span
 *
 */
export function WithSpan(config?: string | SpanConfig) {
  return (target: object, propertyKey: string | symbol, descriptor: PropertyDescriptor) => {
    const originalMethod = descriptor.value
    const methodName = String(propertyKey)

    descriptor.value = async function (...args: any[]) {
      const cfg: SpanConfig = typeof config === 'string' ? { name: config } : config || {}
      const { name, attributes = {} } = cfg

      const spanName = name || `${target.constructor.name}.${methodName}`

      const allAttributes = {
        component: target.constructor.name,
        method: methodName,
        ...attributes,
      }

      const span = getTracer().startSpan(
        spanName,
        {
          attributes: allAttributes,
        },
        context.active(),
      )

      return context.with(trace.setSpan(context.active(), span), async () => {
        try {
          const result = await originalMethod.apply(this, args)
          span.setStatus({ code: SpanStatusCode.OK })
          return result
        } catch (error) {
          span.setStatus({
            code: SpanStatusCode.ERROR,
            message: error instanceof Error ? error.message : String(error),
          })
          span.recordException(error instanceof Error ? error : new Error(String(error)))
          throw error
        } finally {
          span.end()
        }
      })
    }
  }
}

/**
 * Decorator for instrumenting methods with OpenTelemetry metrics (metrics only)
 *
 * Collects two metrics:
 * - Counter: `{name}_executions` - tracks number of executions with status (success/error)
 * - Histogram: `{name}_duration` - tracks execution duration in milliseconds
 *
 * @param config - Configuration object or string name for the metric
 *
 */
export function WithMetric(config?: string | MetricConfig) {
  return (target: object, propertyKey: string | symbol, descriptor: PropertyDescriptor) => {
    const originalMethod = descriptor.value
    const methodName = String(propertyKey)

    descriptor.value = async function (...args: any[]) {
      const cfg: MetricConfig = typeof config === 'string' ? { name: config } : config || {}
      const { name, description, labels = {} } = cfg

      const metricName = toSnakeCase(name || `${target.constructor.name}.${methodName}`)
      const allLabels = {
        component: target.constructor.name,
        method: methodName,
        ...labels,
      }

      // Get or create histogram for this method
      if (!executionHistograms.has(metricName)) {
        executionHistograms.set(
          metricName,
          getMeter().createHistogram(`${metricName}_duration`, {
            description: description || `Duration of executions for ${metricName}`,
            unit: 'ms',
          }),
        )
      }
      const histogram = executionHistograms.get(metricName)
      if (!histogram) {
        throw new Error(`Histogram not found for metric: ${metricName}`)
      }

      const startTime = Date.now()

      let status: 'success' | 'error' = 'success'
      try {
        const result = await originalMethod.apply(this, args)
        return result
      } catch (error) {
        status = 'error'
        throw error
      } finally {
        const duration = Date.now() - startTime
        histogram.record(duration, { ...allLabels, status })
      }
    }
  }
}

/**
 * Decorator for instrumenting methods with both OpenTelemetry traces and metrics
 *
 * This decorator composes @WithSpan and @WithMetric to provide both trace and metric collection.
 * You can selectively enable/disable traces or metrics using the config options.
 *
 * @param config - Configuration object or string name for the instrumentation
 */
export function WithInstrumentation(config?: string | InstrumentationConfig): MethodDecorator {
  const cfg: InstrumentationConfig = typeof config === 'string' ? { name: config } : config || {}
  const { enableTraces = true, enableMetrics = true, name, description, labels } = cfg

  const decorators: MethodDecorator[] = []

  if (enableTraces) {
    decorators.push(WithSpan({ name, attributes: labels }))
  }

  if (enableMetrics) {
    decorators.push(WithMetric({ name, description, labels }))
  }

  return (target: object, propertyKey: string | symbol, descriptor: PropertyDescriptor) => {
    decorators.forEach((decorator) => decorator(target, propertyKey, descriptor))
  }
}
