/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { NodeSDK } from '@opentelemetry/sdk-node'
import { HttpInstrumentation } from '@opentelemetry/instrumentation-http'
import { NestInstrumentation } from '@opentelemetry/instrumentation-nestjs-core'
import { BatchSpanProcessor } from '@opentelemetry/sdk-trace-base'
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http'
import { resourceFromAttributes } from '@opentelemetry/resources'
import { ATTR_SERVICE_NAME } from '@opentelemetry/semantic-conventions'
import { ATTR_DEPLOYMENT_ENVIRONMENT_NAME } from '@opentelemetry/semantic-conventions/incubating'
import { IORedisInstrumentation } from '@opentelemetry/instrumentation-ioredis'
import { PgInstrumentation } from '@opentelemetry/instrumentation-pg'

if (process.env.OTEL_ENABLED === 'true') {
  const traceExporter = new OTLPTraceExporter({
    url: process.env.OTEL_COLLECTOR_URL,
    headers:
      process.env.OTEL_AUTH_USERNAME && process.env.OTEL_AUTH_PASSWORD
        ? {
            Authorization: `Basic ${Buffer.from(`${process.env.OTEL_AUTH_USERNAME}:${process.env.OTEL_AUTH_PASSWORD}`).toString('base64')}`,
          }
        : undefined,
  })

  const sdk = new NodeSDK({
    resource: resourceFromAttributes({
      [ATTR_SERVICE_NAME]: 'api',
      [ATTR_DEPLOYMENT_ENVIRONMENT_NAME]: process.env.ENVIRONMENT,
    }),
    traceExporter,
    instrumentations: [
      new HttpInstrumentation({ requireParentforOutgoingSpans: true, requireParentforIncomingSpans: true }),
      new NestInstrumentation(),
      new IORedisInstrumentation({ requireParentSpan: true }),
      new PgInstrumentation({ requireParentSpan: true }),
    ],
    spanProcessors: [new BatchSpanProcessor(traceExporter)],
  })

  sdk.start()
}
