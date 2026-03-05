/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { PassportModule } from '@nestjs/passport'
import { JwtModule } from '@nestjs/jwt'
import { AdminAuthStrategy } from './admin-auth.strategy'
import { ApiKeyStrategy } from './api-key.strategy'
import { UserModule } from '../user/user.module'
import { ApiKeyModule } from '../api-key/api-key.module'
import { SandboxModule } from '../sandbox/sandbox.module'
import { TypedConfigService } from '../config/typed-config.service'
import { UserService } from '../user/user.service'
import { TypedConfigModule } from '../config/typed-config.module'
import { FailedAuthTrackerService } from './failed-auth-tracker.service'
import { RegionModule } from '../region/region.module'
import { AdminController } from './admin.controller'

@Module({
  imports: [
    PassportModule.register({
      defaultStrategy: ['admin-jwt', 'api-key'],
      property: 'user',
      session: false,
    }),
    JwtModule.registerAsync({
      imports: [TypedConfigModule],
      useFactory: (configService: TypedConfigService) => ({
        secret: configService.getOrThrow('jwtSecret'),
        signOptions: {
          expiresIn: '64h',
          algorithm: 'HS256',
        },
      }),
      inject: [TypedConfigService],
    }),
    TypedConfigModule,
    UserModule,
    ApiKeyModule,
    SandboxModule,
    RegionModule,
  ],
  providers: [
    ApiKeyStrategy,
    {
      provide: AdminAuthStrategy,
      useFactory: (userService: UserService, configService: TypedConfigService) => {
        return new AdminAuthStrategy(userService, configService)
      },
      inject: [UserService, TypedConfigService],
    },
    FailedAuthTrackerService,
  ],
  controllers: [AdminController],
  exports: [PassportModule, AdminAuthStrategy, ApiKeyStrategy, FailedAuthTrackerService, JwtModule],
})
export class AuthModule { }
