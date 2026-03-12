/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SystemRole } from '../../user/enums/system-role.enum'
import { ProxyAuthContext } from './proxy-auth-context.interface'
import { RunnerAuthContext } from './runner-auth-context.interface'
import { SshGatewayAuthContext } from './ssh-gateway-auth-context.interface'
import { RegionProxyAuthContext } from './region-proxy-auth-context.interface'
import { RegionSSHGatewayAuthContext } from './region-ssh-gateway-auth-context.interface'
import { OtelCollectorAuthContext } from './otel-collector-auth-context.interface'
import { HealthCheckAuthContext } from './health-check-auth-context.interface'
import { UserAuthContext } from './user-auth-context.interface'
import { OrganizationAuthContext } from './organization-auth-context.interface'

export interface BaseAuthContext {
  role: ApiRole
}

export function isBaseAuthContext(user: BaseAuthContext): user is BaseAuthContext {
  return 'role' in user
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

export type AuthContextType =
  | UserAuthContext
  | OrganizationAuthContext
  | ProxyAuthContext
  | RunnerAuthContext
  | SshGatewayAuthContext
  | RegionProxyAuthContext
  | RegionSSHGatewayAuthContext
  | OtelCollectorAuthContext
  | HealthCheckAuthContext
