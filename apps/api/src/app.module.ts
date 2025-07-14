/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module, NestModule, MiddlewareConsumer, RequestMethod } from '@nestjs/common'
import { VersionHeaderMiddleware } from './common/middleware/version-header.middleware'
import { AppService } from './app.service'
import { UserModule } from './user/user.module'
import { TypeOrmModule } from '@nestjs/typeorm'
import { SandboxModule } from './sandbox/sandbox.module'
import { AuthModule } from './auth/auth.module'
import { ServeStaticModule } from '@nestjs/serve-static'
import { join } from 'path'
import { ApiKeyModule } from './api-key/api-key.module'
import { ThrottlerModule } from '@nestjs/throttler'
import { DockerRegistryModule } from './docker-registry/docker-registry.module'
import { RedisModule } from '@nestjs-modules/ioredis'
import { ScheduleModule } from '@nestjs/schedule'
import { EventEmitterModule } from '@nestjs/event-emitter'
import { UsageModule } from './usage/usage.module'
import { AnalyticsModule } from './analytics/analytics.module'
import { OrganizationModule } from './organization/organization.module'
import { EmailModule } from './email/email.module'
import { TypedConfigService } from './config/typed-config.service'
import { TypedConfigModule } from './config/typed-config.module'
import { NotificationModule } from './notification/notification.module'
import { ObjectStorageModule } from './object-storage/object-storage.module'
import { CustomNamingStrategy } from './common/utils/naming-strategy.util'
import { MaintenanceMiddleware } from './common/middleware/maintenance.middleware'

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
          migrationsRun: !configService.getOrThrow('production'),
          namingStrategy: new CustomNamingStrategy(),
          manualInitialization: configService.get('skipConnections'),
        }
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
    ThrottlerModule.forRoot([
      {
        ttl: 1000,
        limit: 10,
      },
    ]),
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
    EventEmitterModule.forRoot(),
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
    ObjectStorageModule,
  ],
  controllers: [],
  providers: [AppService],
})
export class AppModule implements NestModule {
  configure(consumer: MiddlewareConsumer) {
    consumer.apply(VersionHeaderMiddleware).forRoutes({ path: '*', method: RequestMethod.ALL })
    consumer.apply(MaintenanceMiddleware).forRoutes({ path: '*', method: RequestMethod.ALL })
  }
}
