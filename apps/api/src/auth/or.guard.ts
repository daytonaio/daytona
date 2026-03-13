/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, Logger, CanActivate, Type, mixin } from '@nestjs/common'
import { ModuleRef } from '@nestjs/core'

/**
 * Utility guard that allows access if at least one of the provided guards succeeds.
 *
 * Using guards that are not registered as providers in the module will result in a runtime error.
 *
 */
export function OrGuard(guards: Type<CanActivate>[]): Type<CanActivate> {
  @Injectable()
  class OrGuardMixin implements CanActivate {
    protected readonly logger = new Logger(`OrGuard`)

    constructor(private readonly moduleRef: ModuleRef) {}

    async canActivate(context: ExecutionContext): Promise<boolean> {
      for (const GuardClass of guards) {
        const guard = this.moduleRef.get(GuardClass)
        try {
          const result = await guard.canActivate(context)

          if (result) {
            this.logger.debug(`Guard ${GuardClass.name} succeeded`)
            return true
          }
        } catch (error) {
          this.logger.debug(`Guard ${GuardClass.name} failed: ${error.message}`)
        }
      }

      this.logger.debug('All guards in OrGuard failed')
      return false
    }
  }

  return mixin(OrGuardMixin)
}
