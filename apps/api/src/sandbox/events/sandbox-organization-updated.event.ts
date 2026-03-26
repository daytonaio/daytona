/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Sandbox } from '../entities/sandbox.entity'

export class SandboxOrganizationUpdatedEvent {
  constructor(
    public readonly sandbox: Sandbox,
    public readonly oldOrganizationId: string,
    public readonly newOrganizationId: string,
  ) {}
}
