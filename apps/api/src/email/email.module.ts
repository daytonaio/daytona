/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DynamicModule, Module } from '@nestjs/common'
import { EmailService } from './services/email.service'
import { EMAIL_MODULE_OPTIONS } from './constants'

export interface EmailModuleOptions {
  host: string
  port: number
  user?: string
  password?: string
  secure?: boolean
  from: string
  dashboardUrl: string
}

@Module({})
export class EmailModule {
  static forRoot(options: EmailModuleOptions): DynamicModule {
    return {
      module: EmailModule,
      providers: [
        {
          provide: EMAIL_MODULE_OPTIONS,
          useValue: options,
        },
        EmailService,
      ],
      exports: [EmailService],
    }
  }

  static forRootAsync(options: {
    useFactory: (...args: any[]) => Promise<EmailModuleOptions> | EmailModuleOptions
    inject?: any[]
  }): DynamicModule {
    return {
      module: EmailModule,
      providers: [
        {
          provide: EMAIL_MODULE_OPTIONS,
          useFactory: options.useFactory,
          inject: options.inject || [],
        },
        EmailService,
      ],
      exports: [EmailService],
    }
  }
}
