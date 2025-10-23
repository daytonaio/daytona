/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { NodeSDK } from '@opentelemetry/sdk-node'
import { HttpInstrumentation } from '@opentelemetry/instrumentation-http'
import { ExpressInstrumentation } from '@opentelemetry/instrumentation-express'
import { NestInstrumentation } from '@opentelemetry/instrumentation-nestjs-core'
import { BatchSpanProcessor } from '@opentelemetry/sdk-trace-base'
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http'
import { CompressionAlgorithm, OTLPExporterNodeConfigBase } from '@opentelemetry/otlp-exporter-base'
import { resourceFromAttributes } from '@opentelemetry/resources'
import { ATTR_SERVICE_NAME } from '@opentelemetry/semantic-conventions'
import {
  ATTR_DEPLOYMENT_ENVIRONMENT_NAME,
  ATTR_SERVICE_INSTANCE_ID,
} from '@opentelemetry/semantic-conventions/incubating'
import { IORedisInstrumentation } from '@opentelemetry/instrumentation-ioredis'
import { PgInstrumentation } from '@opentelemetry/instrumentation-pg'
import { KafkaJsInstrumentation } from '@opentelemetry/instrumentation-kafkajs'
import { getAppMode } from './common/utils/app-mode'
import { diag, DiagConsoleLogger, DiagLogLevel } from '@opentelemetry/api'
import { hostname } from 'os'
import { OTLPMetricExporter } from '@opentelemetry/exporter-metrics-otlp-http'
import { PeriodicExportingMetricReader } from '@opentelemetry/sdk-metrics'
import { PinoInstrumentation } from '@opentelemetry/instrumentation-pino'
import { RuntimeNodeInstrumentation } from '@opentelemetry/instrumentation-runtime-node'
import { BatchLogRecordProcessor } from '@opentelemetry/sdk-logs'
import { OTLPLogExporter } from '@opentelemetry/exporter-logs-otlp-http'

// Enable OpenTelemetry diagnostics
diag.setLogger(new DiagConsoleLogger(), DiagLogLevel.WARN)

const appMode = getAppMode()
const serviceNameSuffix = appMode === 'api' ? 'api' : appMode === 'worker' ? 'worker' : 'api'

const otlpExporterConfig: OTLPExporterNodeConfigBase = {
  compression: CompressionAlgorithm.GZIP,
  keepAlive: true,
}

const otelSdk = new NodeSDK({
  resource: resourceFromAttributes({
    [ATTR_SERVICE_NAME]: `daytona-${serviceNameSuffix}`,
    [ATTR_DEPLOYMENT_ENVIRONMENT_NAME]: process.env.ENVIRONMENT,
    [ATTR_SERVICE_INSTANCE_ID]: process.env.NODE_APP_INSTANCE
      ? `${hostname()}-${process.env.NODE_APP_INSTANCE}`
      : hostname(),
  }),
  instrumentations: [
    new PinoInstrumentation(),
    new HttpInstrumentation({ requireParentforOutgoingSpans: true }),
    new ExpressInstrumentation(),
    new NestInstrumentation(),
    new IORedisInstrumentation({ requireParentSpan: true }),
    new PgInstrumentation({ requireParentSpan: true }),
    new KafkaJsInstrumentation(),
    new RuntimeNodeInstrumentation(),
  ],
  logRecordProcessors: [new BatchLogRecordProcessor(new OTLPLogExporter(otlpExporterConfig))],
  spanProcessors: [new BatchSpanProcessor(new OTLPTraceExporter(otlpExporterConfig))],
  metricReaders: [
    new PeriodicExportingMetricReader({
      exporter: new OTLPMetricExporter(otlpExporterConfig),
      exportIntervalMillis: 30 * 1000,
    }),
  ],
})

export { otelSdk }

process.on('SIGTERM', async () => {
  console.log('SIGTERM received, shutting down OpenTelemetry SDK')
  await otelSdk.shutdown()
  console.log('OpenTelemetry SDK shut down')
})
