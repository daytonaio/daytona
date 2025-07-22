/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export enum AuditTarget {
  API_KEY = 'api_key',
  ORGANIZATION = 'organization',
  ORGANIZATION_INVITATION = 'organization_invitation',
  ORGANIZATION_ROLE = 'organization_role',
  ORGANIZATION_USER = 'organization_user',
  DOCKER_REGISTRY = 'docker_registry',
  RUNNER = 'runner',
  SANDBOX = 'sandbox',
  SNAPSHOT = 'snapshot',
  USER = 'user',
  VOLUME = 'volume',
}
