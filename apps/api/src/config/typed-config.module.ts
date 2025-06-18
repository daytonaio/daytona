/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Global, Module, DynamicModule } from '@nestjs/common'
import { ConfigModule as NestConfigModule, ConfigModuleOptions } from '@nestjs/config'
import { TypedConfigService } from './typed-config.service'
import { configuration } from './configuration'
import { ConfigController } from './config.controller'

@Global()
@Module({
  imports: [
    NestConfigModule.forRoot({
      isGlobal: true,
      load: [() => configuration],
    }),
  ],
  controllers: [ConfigController],
  providers: [TypedConfigService],
  exports: [TypedConfigService],
})
export class TypedConfigModule {
  static forRoot(options: Partial<ConfigModuleOptions> = {}): DynamicModule {
    return {
      module: TypedConfigModule,
      imports: [
        NestConfigModule.forRoot({
          ...options,
        }),
      ],
      providers: [TypedConfigService],
      exports: [TypedConfigService],
    }
  }
}
