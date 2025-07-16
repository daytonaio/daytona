/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Logger, LogLevel } from '@nestjs/common'
import { NodeSDK } from '@opentelemetry/sdk-node'
import { HttpInstrumentation } from '@opentelemetry/instrumentation-http'
import { ExpressInstrumentation } from '@opentelemetry/instrumentation-express'
import { NestInstrumentation } from '@opentelemetry/instrumentation-nestjs-core'
import { BatchSpanProcessor } from '@opentelemetry/sdk-trace-base'
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http'
import { resourceFromAttributes } from '@opentelemetry/resources'
import { ATTR_SERVICE_NAME } from '@opentelemetry/semantic-conventions'
import { ATTR_DEPLOYMENT_ENVIRONMENT_NAME } from '@opentelemetry/semantic-conventions/incubating'
import { IORedisInstrumentation } from '@opentelemetry/instrumentation-ioredis'
import { PgInstrumentation } from '@opentelemetry/instrumentation-pg'
import { diag, DiagLogger, DiagLogLevel } from '@opentelemetry/api'
import { CompressionAlgorithm } from '@opentelemetry/otlp-exporter-base'

// Custom OpenTelemetry logger that uses NestJS Logger
class NestJSDiagLogger extends Logger implements DiagLogger {
  info(message: string, ...args: unknown[]): void {
    super.log(message, ...args)
  }
}

// Default log level
const logLevels: LogLevel[] = ['log', 'error']
if (process.env.LOG_LEVEL) {
  logLevels.push(process.env.LOG_LEVEL as LogLevel)
}

const logger = new NestJSDiagLogger('OpenTelemetry')
logger.localInstance.setLogLevels(logLevels)

if (process.env.OTEL_ENABLED === 'true') {
  // Enable OpenTelemetry diagnostics
  diag.setLogger(logger, DiagLogLevel.ALL)
  logger.debug(`OpenTelemetry diagnostics enabled with log level: ${DiagLogLevel.INFO}`)

  const traceExporter = new OTLPTraceExporter({
    url: process.env.OTEL_COLLECTOR_URL,
    headers:
      process.env.OTEL_AUTH_USERNAME && process.env.OTEL_AUTH_PASSWORD
        ? {
            Authorization: `Basic ${Buffer.from(`${process.env.OTEL_AUTH_USERNAME}:${process.env.OTEL_AUTH_PASSWORD}`).toString('base64')}`,
          }
        : undefined,
    keepAlive: true,
    compression: CompressionAlgorithm.GZIP,
  })

  const sdk = new NodeSDK({
    resource: resourceFromAttributes({
      [ATTR_SERVICE_NAME]: 'api',
      [ATTR_DEPLOYMENT_ENVIRONMENT_NAME]: process.env.ENVIRONMENT,
    }),
    traceExporter,
    instrumentations: [
      new HttpInstrumentation({ requireParentforOutgoingSpans: true }),
      new ExpressInstrumentation(),
      new NestInstrumentation(),
      new IORedisInstrumentation({ requireParentSpan: true }),
      new PgInstrumentation({ requireParentSpan: true }),
    ],
    spanProcessors: [
      new BatchSpanProcessor(traceExporter, {
        maxQueueSize: 10000,
        maxExportBatchSize: 2048,
        scheduledDelayMillis: 2000,
        exportTimeoutMillis: 30000,
      }),
    ],
  })

  try {
    sdk.start()
    logger.log('OpenTelemetry SDK initialized')
  } catch (error) {
    logger.error('Failed to start OpenTelemetry SDK:', error)
  }
}
