/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, Logger, CanActivate, Type, mixin } from '@nestjs/common'
import { ModuleRef } from '@nestjs/core'

/**
 * Creates an OrGuard that allows access if at least one of the provided guards allows access.
 * It tries each guard in sequence and returns true on the first successful guard.
 * If all guards fail, it returns false.
 *
 * Usage:
 * ```typescript
 * @UseGuards(OrGuard([GuardA, GuardB]))
 * ```
 */
export function OrGuard(guards: Type<CanActivate>[]): Type<CanActivate> {
  @Injectable()
  class OrGuardMixin implements CanActivate {
    protected readonly logger = new Logger(`OrGuard`)

    constructor(private readonly moduleRef: ModuleRef) {}

    async canActivate(context: ExecutionContext): Promise<boolean> {
      for (const GuardClass of guards) {
        try {
          const guard = this.moduleRef.get(GuardClass, { strict: false })
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
