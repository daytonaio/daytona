/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export class OrganizationUserRemovedEvent {
  constructor(
    public readonly organizationId: string,
    public readonly userId: string,
  ) {}
}
