/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiKey } from '../../api-key/api-key.entity'
import { OrganizationUser } from '../../organization/entities/organization-user.entity'
import { Organization } from '../../organization/entities/organization.entity'
import { SystemRole } from '../../user/enums/system-role.enum'
import { ProxyContext } from './proxy-context.interface'
import { RunnerContext } from './runner-context.interface'
import { SshGatewayContext } from './ssh-gateway-context.interface'
import { RegionProxyContext } from './region-proxy.interface'
import { RegionSSHGatewayContext } from './region-ssh-gateway.interface'
import { OtelCollectorContext } from './otel-collector-context.interface'
import { HealthCheckContext } from './health-check-context.interface'

export interface BaseAuthContext {
  role: ApiRole
}

export type ApiRole =
  | SystemRole
  | 'proxy'
  | 'runner'
  | 'ssh-gateway'
  | 'region-proxy'
  | 'region-ssh-gateway'
  | 'otel-collector'
  | 'health-check'

export interface AuthContext extends BaseAuthContext {
  userId: string
  email: string
  apiKey?: ApiKey
  organizationId?: string
  runnerId?: string
}

export function isAuthContext(user: BaseAuthContext): user is AuthContext {
  return 'userId' in user
}

export interface OrganizationAuthContext extends AuthContext {
  organizationId: string
  organization: Organization
  organizationUser?: OrganizationUser
}

export function isOrganizationAuthContext(user: BaseAuthContext): user is OrganizationAuthContext {
  return 'organizationId' in user
}

export type AuthContextType =
  | AuthContext
  | OrganizationAuthContext
  | ProxyContext
  | RunnerContext
  | SshGatewayContext
  | RegionProxyContext
  | RegionSSHGatewayContext
  | OtelCollectorContext
  | HealthCheckContext
