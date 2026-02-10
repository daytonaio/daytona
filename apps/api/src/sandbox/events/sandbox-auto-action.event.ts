/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { EntityManager } from 'typeorm'
import { Sandbox } from '../entities/sandbox.entity'

export class SandboxAutoActionEvent {
  constructor(
    public readonly sandbox: Sandbox,
    public readonly entityManager: EntityManager,
  ) {}
}
