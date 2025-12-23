/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { otelSdk } from './tracing'
import { readdirSync, readFileSync, statSync, writeFileSync } from 'node:fs'
import { NestFactory } from '@nestjs/core'
import { NestExpressApplication } from '@nestjs/platform-express'
import { AppModule } from './app.module'
import { SwaggerModule } from '@nestjs/swagger'
import { INestApplication, Logger, ValidationPipe } from '@nestjs/common'
import { AllExceptionsFilter } from './filters/all-exceptions.filter'
import { MetricsInterceptor } from './interceptors/metrics.interceptor'
import { HttpsOptions } from '@nestjs/common/interfaces/external/https-options.interface'
import { TypedConfigService } from './config/typed-config.service'
import { FailedAuthTrackerService } from './auth/failed-auth-tracker.service'
import { DataSource, MigrationExecutor } from 'typeorm'
import { getOpenApiConfig } from './openapi.config'
import { AuditInterceptor } from './audit/interceptors/audit.interceptor'
import { join } from 'node:path'
import { ApiKeyService } from './api-key/api-key.service'
import { DAYTONA_ADMIN_USER_ID } from './app.service'
import { OrganizationService } from './organization/services/organization.service'
import { MicroserviceOptions, Transport } from '@nestjs/microservices'
import { Partitioners } from 'kafkajs'
import { isApiEnabled, isWorkerEnabled } from './common/utils/app-mode'
import cluster from 'node:cluster'
import { Logger as PinoLogger, LoggerErrorInterceptor } from 'nestjs-pino'

// https options
const httpsEnabled = process.env.CERT_PATH && process.env.CERT_KEY_PATH
const httpsOptions: HttpsOptions = {
  cert: process.env.CERT_PATH ? readFileSync(process.env.CERT_PATH) : undefined,
  key: process.env.CERT_KEY_PATH ? readFileSync(process.env.CERT_KEY_PATH) : undefined,
}

async function bootstrap() {
  if (process.env.OTEL_ENABLED === 'true') {
    await otelSdk.start()
  }
  const app = await NestFactory.create<NestExpressApplication>(AppModule, {
    bufferLogs: true,
    httpsOptions: httpsEnabled ? httpsOptions : undefined,
  })
  app.useLogger(app.get(PinoLogger))
  app.flushLogs()
  app.enableCors({
    origin: true,
    methods: 'GET,HEAD,PUT,PATCH,POST,DELETE,OPTIONS',
    credentials: true,
  })

  const configService = app.get(TypedConfigService)
  const failedAuthTracker = app.get(FailedAuthTrackerService)
  app.set('trust proxy', true)
  app.useGlobalFilters(new AllExceptionsFilter(failedAuthTracker))
  app.useGlobalInterceptors(new LoggerErrorInterceptor())
  app.useGlobalInterceptors(new MetricsInterceptor(configService))
  app.useGlobalInterceptors(app.get(AuditInterceptor))
  app.useGlobalPipes(
    new ValidationPipe({
      transform: true,
    }),
  )

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

  // Replace dashboard api url before serving
  if (configService.get('production')) {
    const dashboardDir = join(__dirname, '..', 'dashboard')
    const replaceInDirectory = (dir: string) => {
      for (const file of readdirSync(dir)) {
        const filePath = join(dir, file)
        if (statSync(filePath).isDirectory()) {
          if (file === 'assets') {
            replaceInDirectory(filePath)
          }
          continue
        }
        Logger.log(`Replacing %DAYTONA_BASE_API_URL% in ${filePath}`)
        const fileContent = readFileSync(filePath, 'utf8')
        const newFileContent = fileContent.replaceAll(
          '%DAYTONA_BASE_API_URL%',
          configService.get('dashboardBaseApiUrl'),
        )
        writeFileSync(filePath, newFileContent)
      }
    }
    replaceInDirectory(dashboardDir)
  }

  // Starts listening for shutdown hooks
  app.enableShutdownHooks()

  const host = '0.0.0.0'
  const port = configService.get('port')

  if (isApiEnabled()) {
    await app.listen(port, host)
    Logger.log(`🚀 Daytona API is running on: http://${host}:${port}/${globalPrefix}`)
  } else {
    await app.init()
    app.flushLogs()
  }

  if (isWorkerEnabled() && configService.get('kafka.enabled')) {
    app.connectMicroservice<MicroserviceOptions>({
      transport: Transport.KAFKA,
      options: {
        client: configService.getKafkaClientConfig(),
        producer: {
          allowAutoTopicCreation: true,
          createPartitioner: Partitioners.DefaultPartitioner,
          idempotent: true,
        },
        consumer: {
          allowAutoTopicCreation: true,
          groupId: 'daytona',
        },
        run: {
          autoCommit: false,
        },
        subscribe: {
          fromBeginning: true,
        },
      },
    })
    await app.startAllMicroservices()
  }

  // If app running in cluster mode, send ready signal
  if (cluster.isWorker) {
    process.send('ready')
  }
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
