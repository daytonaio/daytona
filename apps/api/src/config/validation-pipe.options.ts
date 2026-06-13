/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ValidationPipeOptions } from '@nestjs/common'

/**
 * Options for the global request ValidationPipe.
 *
 * `whitelist` strips any request-body property that is not declared (with a
 * class-validator decorator) on its DTO, preventing mass-assignment of
 * server-authoritative fields (e.g. organizationId).
 *
 * `forbidNonWhitelisted` is intentionally NOT enabled: some first-party clients
 * still send undeclared fields the API never reads (e.g. a stray `envVars` on
 * sandbox creation — the service only consumes `env`). Stripping such fields is
 * behaviorally identical to the previous pipe, whereas rejecting them with a 400
 * would be a new, breaking regression.
 *
 * Kept in a shared constant so the regression spec validates the exact
 * production configuration.
 */
export const validationPipeOptions: ValidationPipeOptions = {
  transform: true,
  whitelist: true,
}
