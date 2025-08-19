/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Request } from 'express'
import { AuthContext, OrganizationAuthContext } from '../interfaces/auth-context.interface'
import { DockerRegistry } from '../../docker-registry/entities/docker-registry.entity'
import { Sandbox } from '../../sandbox/entities/sandbox.entity'
import { Snapshot } from '../../sandbox/entities/snapshot.entity'

// Request with optional user - used in interceptors
export type RequestWithUser = Request & {
  user?: AuthContext
}

// Request with required user - used in basic guards
export type RequestWithAuthContext = Request & {
  user: AuthContext
}

// Request with organization user and resource entities - used in organization guards
export type RequestWithOrganizationContext = Request & {
  user: OrganizationAuthContext
  dockerRegistry?: DockerRegistry
  sandbox?: Sandbox
  snapshot?: Snapshot
}
