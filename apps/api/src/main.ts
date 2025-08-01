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
import { ConsoleLogger, INestApplication, Logger, LogLevel, ValidationPipe } from '@nestjs/common'
import { HttpAdapterHost } from '@nestjs/core'
import { AllExceptionsFilter } from './filters/all-exceptions.filter'
import { NotFoundExceptionFilter } from './common/middleware/frontend.middleware'
import { MetricsInterceptor } from './interceptors/metrics.interceptor'
import { HttpsOptions } from '@nestjs/common/interfaces/external/https-options.interface'
import { TypedConfigService } from './config/typed-config.service'
import { DataSource, MigrationExecutor } from 'typeorm'
import { RunnerService } from './sandbox/services/runner.service'
import { getOpenApiConfig } from './openapi.config'
import { SchedulerRegistry } from '@nestjs/schedule'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { AuditInterceptor } from './audit/interceptors/audit.interceptor'
import { ApiKeyService } from './api-key/api-key.service'
import { DAYTONA_ADMIN_USER_ID } from './app.service'
import { OrganizationService } from './organization/services/organization.service'

// https options
const httpsEnabled = process.env.CERT_PATH && process.env.CERT_KEY_PATH
const httpsOptions: HttpsOptions = {
  cert: process.env.CERT_PATH ? readFileSync(process.env.CERT_PATH) : undefined,
  key: process.env.CERT_KEY_PATH ? readFileSync(process.env.CERT_KEY_PATH) : undefined,
}

// Default log level
const logLevels: LogLevel[] = ['log', 'error', 'warn']
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
  app.useGlobalInterceptors(new MetricsInterceptor(configService))
  app.useGlobalInterceptors(app.get(AuditInterceptor))
  app.useGlobalPipes(new ValidationPipe())

  const eventEmitter = app.get(EventEmitter2)
  eventEmitter.setMaxListeners(100)

  // Runtime flags for migrations for run and revert migrations
  if (process.argv.length > 2) {
    if (process.argv[2].startsWith('--migration-')) {
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
    } else if (process.argv[2] === '--create-admin-api-key') {
      if (process.argv.length < 4) {
        Logger.error('Invalid flag. API key name is required.')
        process.exit(1)
      }
      await createAdminApiKey(app, process.argv[3])
    } else {
      Logger.error('Invalid flag')
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
  if (configService.get('defaultRunner.domain')) {
    const runnerService = app.get(RunnerService)
    const runners = await runnerService.findAll()
    if (!runners.find((runner) => runner.domain === configService.getOrThrow('defaultRunner.domain'))) {
      Logger.log(`Creating default runner: ${configService.getOrThrow('defaultRunner.domain')}`)
      await runnerService.create({
        apiUrl: configService.getOrThrow('defaultRunner.apiUrl'),
        proxyUrl: configService.getOrThrow('defaultRunner.proxyUrl'),
        apiKey: configService.getOrThrow('defaultRunner.apiKey'),
        cpu: configService.getOrThrow('defaultRunner.cpu'),
        memoryGiB: configService.getOrThrow('defaultRunner.memory'),
        diskGiB: configService.getOrThrow('defaultRunner.disk'),
        gpu: configService.getOrThrow('defaultRunner.gpu'),
        gpuType: configService.getOrThrow('defaultRunner.gpuType'),
        capacity: configService.getOrThrow('defaultRunner.capacity'),
        region: configService.getOrThrow('defaultRunner.region'),
        class: configService.getOrThrow('defaultRunner.class'),
        domain: configService.getOrThrow('defaultRunner.domain'),
        version: configService.get('defaultRunner.version') || '0'
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

async function createAdminApiKey(app: INestApplication, apiKeyName: string) {
  const apiKeyService = app.get(ApiKeyService)
  const organizationService = app.get(OrganizationService)

  const personalOrg = await organizationService.findPersonal(DAYTONA_ADMIN_USER_ID)
  const { value } = await apiKeyService.createApiKey(personalOrg.id, DAYTONA_ADMIN_USER_ID, apiKeyName, [])
  Logger.log(
    `
=========================================
=========================================
Admin API key created: ${value}
=========================================
=========================================`,
  )
}

bootstrap()
