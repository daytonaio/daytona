/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CanActivate, ExecutionContext, Logger, NotFoundException } from '@nestjs/common'
import { EntityNotFoundError } from 'typeorm'

/**
 *  Base class for guards that validate that the current request is authenticated with an auth context type that has access to the requested resource.
 */
export abstract class ResourceAccessGuard implements CanActivate {
  abstract canActivate(context: ExecutionContext): Promise<boolean>

  /**
   * Handles resource access errors by logging the error and throwing a NotFoundException.
   * @param error - The error to handle.
   * @param logger - The logger to use.
   * @param notFoundMessage - The message to use in the NotFoundException.
   */
  protected handleResourceAccessError(error: unknown, logger: Logger, notFoundMessage: string): never {
    if (!(error instanceof NotFoundException) && !(error instanceof EntityNotFoundError)) {
      logger.error(error)
    }
    throw new NotFoundException(notFoundMessage)
  }
}
