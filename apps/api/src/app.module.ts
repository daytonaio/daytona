/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module, NestModule, MiddlewareConsumer, RequestMethod, ExecutionContext } from '@nestjs/common'
import { VersionHeaderMiddleware } from './common/middleware/version-header.middleware'
import { AppService } from './app.service'
import { UserModule } from './user/user.module'
import { TypeOrmModule } from '@nestjs/typeorm'
import { SandboxModule } from './sandbox/sandbox.module'
import { AuthModule } from './auth/auth.module'
import { ServeStaticModule } from '@nestjs/serve-static'
import { join } from 'path'
import { ApiKeyModule } from './api-key/api-key.module'
import { seconds, ThrottlerModule } from '@nestjs/throttler'
import { RateLimitGuard } from './common/guards/rate-limit.guard'
import { DockerRegistryModule } from './docker-registry/docker-registry.module'
import { RedisModule, getRedisConnectionToken } from '@nestjs-modules/ioredis'
import { ScheduleModule } from '@nestjs/schedule'
import { EventEmitterModule } from '@nestjs/event-emitter'
import { UsageModule } from './usage/usage.module'
import { AnalyticsModule } from './analytics/analytics.module'
import { OrganizationModule } from './organization/organization.module'
import { EmailModule } from './email/email.module'
import { TypedConfigService } from './config/typed-config.service'
import { TypedConfigModule } from './config/typed-config.module'
import { NotificationModule } from './notification/notification.module'
import { WebhookModule } from './webhook/webhook.module'
import { ObjectStorageModule } from './object-storage/object-storage.module'
import { CustomNamingStrategy } from './common/utils/naming-strategy.util'
import { MaintenanceMiddleware } from './common/middleware/maintenance.middleware'
import { AuditModule } from './audit/audit.module'
import { HealthModule } from './health/health.module'
import { OpenFeatureModule } from '@openfeature/nestjs-sdk'
import { OpenFeaturePostHogProvider } from './common/providers/openfeature-posthog.provider'
import { Redis } from 'ioredis'
import { ThrottlerStorageRedisService } from '@nest-lab/throttler-storage-redis'
import { APP_GUARD } from '@nestjs/core'

@Module({
  imports: [
    TypedConfigModule.forRoot({
      isGlobal: true,
    }),
    TypeOrmModule.forRootAsync({
      inject: [TypedConfigService],
      useFactory: (configService: TypedConfigService) => {
        return {
          type: 'postgres',
          host: configService.getOrThrow('database.host'),
          port: configService.getOrThrow('database.port'),
          username: configService.getOrThrow('database.username'),
          password: configService.getOrThrow('database.password'),
          database: configService.getOrThrow('database.database'),
          autoLoadEntities: true,
          migrations: [join(__dirname, 'migrations/**/*{.ts,.js}')],
          migrationsRun: configService.get('runMigrations') || !configService.getOrThrow('production'),
          namingStrategy: new CustomNamingStrategy(),
          manualInitialization: configService.get('skipConnections'),
        }
      },
    }),
    ServeStaticModule.forRoot({
      rootPath: join(__dirname, '..'),
      exclude: ['/api/*'],
      renderPath: '/runner-amd64',
      serveStaticOptions: {
        cacheControl: false,
      },
    }),
    ServeStaticModule.forRoot({
      rootPath: join(__dirname, '..', 'dashboard'),
      exclude: ['/api/*'],
      renderPath: '/',
      serveStaticOptions: {
        cacheControl: false,
      },
    }),
    RedisModule.forRootAsync({
      inject: [TypedConfigService],
      useFactory: (configService: TypedConfigService) => {
        return {
          type: 'single',
          options: {
            host: configService.getOrThrow('redis.host'),
            port: configService.getOrThrow('redis.port'),
            tls: configService.get('redis.tls'),
            lazyConnect: configService.get('skipConnections'),
          },
        }
      },
    }),
    RedisModule.forRootAsync(
      {
        inject: [TypedConfigService],
        useFactory: (configService: TypedConfigService) => {
          return {
            type: 'single',
            options: {
              host: configService.getOrThrow('redis.host'),
              port: configService.getOrThrow('redis.port'),
              tls: configService.get('redis.tls'),
              lazyConnect: configService.get('skipConnections'),
              db: 1,
            },
          }
        },
      },
      'throttler',
    ),
    ThrottlerModule.forRootAsync({
      useFactory: async (redis: Redis) => {
        return {
          throttlers: [
            {
              name: 'anonymous',
              ttl: seconds(100),
              limit: 5,
            },
            {
              name: 'authenticated',
              ttl: seconds(30),
              limit: 20000,
            },
          ],
          storage: new ThrottlerStorageRedisService(redis),
        }
      },
      inject: [getRedisConnectionToken('throttler')],
    }),
    EventEmitterModule.forRoot({
      maxListeners: 100,
    }),
    ApiKeyModule,
    AuthModule,
    UserModule,
    SandboxModule,
    DockerRegistryModule,
    ScheduleModule.forRoot(),
    UsageModule,
    AnalyticsModule,
    OrganizationModule,
    EmailModule.forRootAsync({
      inject: [TypedConfigService],
      useFactory: (configService: TypedConfigService) => {
        return {
          host: configService.get('smtp.host'),
          port: configService.get('smtp.port'),
          user: configService.get('smtp.user'),
          password: configService.get('smtp.password'),
          secure: configService.get('smtp.secure'),
          from: configService.get('smtp.from'),
          dashboardUrl: configService.getOrThrow('dashboardUrl'),
        }
      },
    }),
    NotificationModule,
    WebhookModule,
    ObjectStorageModule,
    AuditModule,
    HealthModule,
    OpenFeatureModule.forRoot({
      contextFactory: (request: ExecutionContext) => {
        const req = request.switchToHttp().getRequest()

        return {
          targetingKey: req.user?.userId,
          organizationId: req.user?.organizationId,
        }
      },
      defaultProvider: new OpenFeaturePostHogProvider({
        clientOptions: {
          host: process.env.POSTHOG_HOST,
        },
        apiKey: process.env.POSTHOG_API_KEY,
      }),
    }),
  ],
  controllers: [],
  providers: [
    AppService,
    {
      provide: APP_GUARD,
      useClass: RateLimitGuard,
    },
  ],
})
export class AppModule implements NestModule {
  configure(consumer: MiddlewareConsumer) {
    consumer.apply(VersionHeaderMiddleware).forRoutes({ path: '*', method: RequestMethod.ALL })
    consumer.apply(MaintenanceMiddleware).forRoutes({ path: '*', method: RequestMethod.ALL })
  }
}
