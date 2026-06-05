/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ValidationPipeOptions } from '@nestjs/common'

/**
 * Options for the global request ValidationPipe.
 *
 * `whitelist` + `forbidNonWhitelisted` reduce request bodies to the properties
 * declared (and validated) on their DTO and reject any others, preventing
 * mass-assignment of server-authoritative fields (e.g. organizationId).
 *
 * Kept in a shared constant so the regression spec validates the exact
 * production configuration.
 */
export const validationPipeOptions: ValidationPipeOptions = {
  transform: true,
  whitelist: true,
  forbidNonWhitelisted: true,
}
