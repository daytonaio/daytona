/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import './tracing'
import { readFileSync } from 'node:fs'
import { NestFactory } from '@nestjs/core'
import { NestExpressApplication } from '@nestjs/platform-express'
import { AppModule } from './app.module'
import { SwaggerModule } from '@nestjs/swagger'
import { ConsoleLogger, Logger, LogLevel, ValidationPipe } from '@nestjs/common'
import { HttpAdapterHost } from '@nestjs/core'
import { AllExceptionsFilter } from './filters/all-exceptions.filter'
import { NotFoundExceptionFilter } from './common/middleware/frontend.middleware'
import { MetricsInterceptor } from './interceptors/metrics.interceptor'
import { HttpsOptions } from '@nestjs/common/interfaces/external/https-options.interface'
import { TypedConfigService } from './config/typed-config.service'
import { DataSource, MigrationExecutor } from 'typeorm'
import { RunnerService } from './sandbox/services/runner.service'
import { RunnerRegion } from './sandbox/enums/runner-region.enum'
import { SandboxClass } from './sandbox/enums/sandbox-class.enum'
import { getOpenApiConfig } from './openapi.config'
import { SchedulerRegistry } from '@nestjs/schedule'
import { EventEmitter2 } from '@nestjs/event-emitter'

// https options
const httpsEnabled = process.env.CERT_PATH && process.env.CERT_KEY_PATH
const httpsOptions: HttpsOptions = {
  cert: process.env.CERT_PATH ? readFileSync(process.env.CERT_PATH) : undefined,
  key: process.env.CERT_KEY_PATH ? readFileSync(process.env.CERT_KEY_PATH) : undefined,
}

// Default log level
const logLevels: LogLevel[] = ['log', 'error']
if (process.env.LOG_LEVEL) {
  logLevels.push(process.env.LOG_LEVEL as LogLevel)
}

async function bootstrap() {
  const app = await NestFactory.create<NestExpressApplication>(AppModule, {
    logger: new ConsoleLogger({
      prefix: 'API',
      logLevels,
    }),
    httpsOptions: httpsEnabled ? httpsOptions : undefined,
  })
  app.enableCors({
    origin: true,
    methods: 'GET,HEAD,PUT,PATCH,POST,DELETE,OPTIONS',
    credentials: true,
  })

  const configService = app.get(TypedConfigService)
  const httpAdapter = app.get(HttpAdapterHost)
  app.useGlobalFilters(new AllExceptionsFilter(httpAdapter))
  app.useGlobalFilters(new NotFoundExceptionFilter())
  app.useGlobalInterceptors(new MetricsInterceptor())
  app.useGlobalPipes(new ValidationPipe())

  const eventEmitter = app.get(EventEmitter2)
  eventEmitter.setMaxListeners(100)

  // Runtime flags for migrations for run and revert migrations
  if (process.argv.length > 2 && process.argv[2].startsWith('--migration-')) {
    const dataSource = app.get(DataSource)
    dataSource.setOptions({ logging: true })
    const migrationExecutor = new MigrationExecutor(dataSource)

    switch (process.argv[2]) {
      case '--migration-run':
        await migrationExecutor.executePendingMigrations()
        break
      case '--migration-revert':
        await migrationExecutor.undoLastMigration()
        break
      default:
        Logger.error('Invalid migration flag')
        process.exit(1)
    }
    process.exit(0)
  }

  const globalPrefix = 'api'
  app.setGlobalPrefix(globalPrefix)

  const documentFactory = () => SwaggerModule.createDocument(app, getOpenApiConfig(configService.get('oidc.issuer')))
  SwaggerModule.setup('api', app, documentFactory, {
    swaggerOptions: {
      initOAuth: {
        clientId: configService.get('oidc.clientId'),
        appName: 'Daytona AI',
        scopes: ['openid', 'profile', 'email'],
        additionalQueryStringParams: {
          audience: configService.get('oidc.audience'),
        },
      },
    },
  })

  // Auto create runners only in local development environment
  if (!configService.get('production')) {
    const runnerService = app.get(RunnerService)
    const runners = await runnerService.findAll()
    if (!runners.find((runner) => runner.domain === 'localtest.me:3003')) {
      await runnerService.create({
        apiUrl: 'http://localhost:3003',
        apiKey: 'secret_api_token',
        cpu: 4,
        memory: 8192,
        disk: 50,
        gpu: 0,
        gpuType: 'none',
        capacity: 100,
        region: RunnerRegion.US,
        class: SandboxClass.SMALL,
        domain: 'localtest.me:3003',
      })
    }
  }

  // Stop all cron jobs if maintenance mode is enabled
  if (configService.get('maintananceMode')) {
    await app.init()
    const schedulerRegistry = app.get(SchedulerRegistry)
    for (const cronName of schedulerRegistry.getCronJobs().keys()) {
      schedulerRegistry.deleteCronJob(cronName)
    }
  }

  const host = '0.0.0.0'
  const port = configService.get('port')
  await app.listen(port, host)
  Logger.log(`🚀 Daytona API is running on: http://${host}:${port}/${globalPrefix}`)
}

bootstrap()
