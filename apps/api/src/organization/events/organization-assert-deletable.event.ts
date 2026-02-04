/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export class OrganizationAssertDeletableEvent {
  constructor(public readonly organizationId: string) {}
}
