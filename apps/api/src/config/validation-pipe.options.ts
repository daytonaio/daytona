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
 * still send undeclared/legacy fields (e.g. the CLI sends `envVars` on sandbox
 * creation), and silently stripping them is backward-compatible whereas
 * rejecting them with a 400 is not.
 *
 * Kept in a shared constant so the regression spec validates the exact
 * production configuration.
 */
export const validationPipeOptions: ValidationPipeOptions = {
  transform: true,
  whitelist: true,
}
